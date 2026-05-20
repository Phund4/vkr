package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"

	"router/internal/adapters/capture"
	"router/internal/adapters/image"
	"router/internal/adapters/metrics"
	"router/internal/core/domain"
)

// enqueueFrame кладёт кадр в канал; при переполнении отбрасывает самый старый кадр.
func enqueueFrame(ch chan []byte, frame []byte) {
	select {
	case ch <- frame:
	default:
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- frame:
		default:
		}
	}
}

// frameChanCapacity вычисляет буфер канала кадров от числа воркеров обработки.
func frameChanCapacity(processWorkers int) int {
	c := processWorkers * frameChanCapacityMultiplier
	if c < minFrameChanCapacity {
		return minFrameChanCapacity
	}
	return c
}

// RunCamera бесконечный цикл: RTSP → Kafka (video meta + кадр для pusher + ML in-топики).
func RunCamera(
	ctx context.Context,
	cam domain.Camera,
	framePub FramePublisher,
	videoPub VideoMetaPublisher,
	mlPub MLFramePublisher,
	keyPrefix string,
	ffmpegPath string,
	targetFPS float64,
	processWorkers int,
) {
	var frameNo atomic.Uint64
	var lastUpstreamLog, lastMLLog time.Time
	backoff := reconnectBackoffDuration()

	s := &cameraSession{
		ctx:             ctx,
		cam:             cam,
		framePub:        framePub,
		videoPub:        videoPub,
		mlPub:           mlPub,
		prefix:          strings.Trim(keyPrefix, "/"),
		ffmpegPath:      ffmpegPath,
		targetFPS:       targetFPS,
		processWorkers:  processWorkers,
		logger:          zlog.With().Str("segment", cam.SegmentID).Str("camera", cam.CameraID).Logger(),
		frameNo:         &frameNo,
		lastUpstreamLog: &lastUpstreamLog,
		lastMLLog:       &lastMLLog,
	}

	for ctx.Err() == nil {
		s.runOnce(backoff)
		if ctx.Err() != nil {
			return
		}
	}
}

// cameraSession состояние одной камеры внутри RunCamera.
type cameraSession struct {
	ctx             context.Context
	cam             domain.Camera
	framePub        FramePublisher
	videoPub        VideoMetaPublisher
	mlPub           MLFramePublisher
	prefix          string
	ffmpegPath      string
	targetFPS       float64
	processWorkers  int
	logger          zerolog.Logger
	frameNo         *atomic.Uint64
	lastUpstreamLog *time.Time
	lastMLLog       *time.Time
}

// runOnce один проход: подключение к RTSP, чтение кадров до EOF/отмены, обработка воркерами.
func (s *cameraSession) runOnce(backoff time.Duration) {
	subCtx, cancel := context.WithCancel(s.ctx)
	pipe, err := capture.FFmpegPipe(subCtx, s.ffmpegPath, s.cam.RTSPURL, s.targetFPS)
	if err != nil {
		metrics.OperationErrors.WithLabelValues(MetricStageFfmpegStart).Inc()
		logSourceIssueThrottled(s.lastUpstreamLog, s.logger, "ffmpeg start (source not ready or invalid URL)", "err", err)
		cancel()
		sleepBackoff(s.ctx, backoff)
		return
	}

	sc := capture.NewScanner(pipe)
	frameCh := make(chan []byte, frameChanCapacity(s.processWorkers))

	var readerWG sync.WaitGroup
	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		defer close(frameCh)
		s.readFramesIntoChannel(subCtx, sc, frameCh)
	}()

	var workersWG sync.WaitGroup
	for range s.processWorkers {
		workersWG.Add(1)
		go func() {
			defer workersWG.Done()
			for frame := range frameCh {
				s.handleFrame(frame)
			}
		}()
	}

	readerWG.Wait()
	workersWG.Wait()
	_ = pipe.Close()
	cancel()

	if s.ctx.Err() != nil {
		return
	}
	logSourceIssueThrottled(s.lastUpstreamLog, s.logger, "source stopped delivering frames, reconnecting", "rtsp_url", s.cam.RTSPURL)
	sleepBackoff(s.ctx, backoff)
}

// readFramesIntoChannel читает JPEG-кадры из сканера и передаёт их в frameCh до EOF или отмены.
func (s *cameraSession) readFramesIntoChannel(subCtx context.Context, sc *capture.Scanner, frameCh chan []byte) {
	for {
		frame, err := sc.ReadFrameCtx(subCtx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			if errors.Is(err, io.EOF) {
				return
			}
			metrics.OperationErrors.WithLabelValues(MetricStageFrameRead).Inc()
			logSourceIssueThrottled(s.lastUpstreamLog, s.logger, "frame read (stream interrupted or paused)", "err", err)
			return
		}
		enqueueFrame(frameCh, frame)
	}
}

// handleFrame конвертирует JPEG в PNG, публикует кадр в Kafka для pusher и вызывает оба ML.
func (s *cameraSession) handleFrame(frame []byte) {
	handleStart := time.Now()
	defer func() {
		metrics.FrameHandleSeconds.Observe(time.Since(handleStart).Seconds())
	}()

	n := s.frameNo.Add(1)
	now := time.Now().UTC()
	pipelineStart := now.Format(time.RFC3339Nano)
	day := now.Format(frameKeyDateLayout)
	ts := now.UnixNano()
	key := fmt.Sprintf("%s/%s/%s/frame_%d%s", s.prefix, day, s.cam.CameraID, ts, frameObjectSuffix)

	pngBytes, err := image.JPEGBytesToPNG(frame)
	if err != nil {
		metrics.OperationErrors.WithLabelValues(MetricStageJpegPng).Inc()
		s.logger.Error().Err(err).Msg("jpeg to png")
		return
	}

	meta := domain.ProcessMeta{
		SegmentID:         s.cam.SegmentID,
		CameraID:          s.cam.CameraID,
		S3Key:             key,
		ObservedAt:        pipelineStart,
		PipelineStartedAt: pipelineStart,
	}

	var g errgroup.Group
	g.Go(func() error {
		if s.videoPub == nil {
			return nil
		}
		ev := domain.VideoIngestEvent{
			SegmentID:         s.cam.SegmentID,
			CameraID:          s.cam.CameraID,
			ObservedAt:        pipelineStart,
			S3Key:             key,
			PipelineStartedAt: pipelineStart,
		}
		b, jerr := json.Marshal(ev)
		if jerr != nil {
			metrics.KafkaVideoPublishErrors.WithLabelValues(MetricKafkaPublishStageJSON).Inc()
			return jerr
		}
		if err := s.videoPub.Publish(s.ctx, []byte(s.cam.SegmentID), b); err != nil {
			metrics.KafkaVideoPublishErrors.WithLabelValues(MetricKafkaPublishStageWrite).Inc()
			s.logger.Warn().Err(err).Msg("kafka video ingest")
			return err
		}
		return nil
	})
	g.Go(func() error {
		if s.framePub == nil {
			return nil
		}
		ev := domain.FrameIngestEvent{
			SegmentID:         s.cam.SegmentID,
			CameraID:          s.cam.CameraID,
			ObservedAt:        pipelineStart,
			PipelineStartedAt: pipelineStart,
			S3Key:             key,
			ContentBase64:     base64.StdEncoding.EncodeToString(pngBytes),
			ContentType:       framePNGContentType,
		}
		if err := s.framePub.PublishFrame(s.ctx, ev); err != nil {
			metrics.KafkaFramesPublishErrors.WithLabelValues(MetricKafkaPublishStageWrite).Inc()
			s.logger.Warn().Err(err).Str("key", key).Msg("kafka frames ingest")
			return err
		}
		metrics.KafkaFrameBytes.Add(float64(len(pngBytes)))
		return nil
	})
	g.Go(func() error {
		if s.mlPub == nil {
			return nil
		}
		metrics.KafkaMLFrameBytes.Add(float64(len(frame)) * 2)
		mlStart := time.Now()
		err := s.mlPub.PublishBoth(s.ctx, frame, meta)
		metrics.KafkaMLPublishDurationSeconds.Observe(time.Since(mlStart).Seconds())
		if err != nil {
			metrics.OperationErrors.WithLabelValues(MetricStageKafkaMLPublish).Inc()
			metrics.KafkaMLPublishErrors.WithLabelValues(MetricStageKafkaMLPublish).Inc()
			metrics.FramesProcessed.WithLabelValues(MetricFrameOutcomeKafkaMLError).Inc()
			logSourceIssueThrottled(s.lastMLLog, s.logger, "kafka ml publish", "err", err)
			return err
		}
		metrics.FramesProcessed.WithLabelValues(MetricFrameOutcomeKafkaMLOk).Inc()
		return nil
	})
	_ = g.Wait()

	if n%frameLogEveryN == 0 {
		s.logger.Info().Uint64("count", n).Str("last_key", key).Msg("frames")
	}
}
