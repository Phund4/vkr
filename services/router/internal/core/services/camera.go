package services

import (
	"context"
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
	"router/internal/adapters/metrics"
	s3store "router/internal/adapters/s3"
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

// RunCamera бесконечный цикл: RTSP → S3 → параллельно Kafka (метаданные) и два вызова ML.
func RunCamera(
	ctx context.Context,
	cam domain.Camera,
	store S3Uploader,
	ml MLRunner,
	videoPub VideoMetaPublisher,
	s3Prefix string,
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
		store:           store,
		ml:              ml,
		videoPub:        videoPub,
		prefix:          strings.Trim(s3Prefix, "/"),
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
	store           S3Uploader
	ml              MLRunner
	videoPub        VideoMetaPublisher
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

// handleFrame конвертирует JPEG в PNG, кладёт в S3, публикует метаданные в Kafka и вызывает оба ML.
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

	pngBytes, err := s3store.JPEGBytesToPNG(frame)
	if err != nil {
		metrics.OperationErrors.WithLabelValues(MetricStageJpegPng).Inc()
		s.logger.Error().Err(err).Msg("jpeg to png")
		return
	}
	if err := s.store.PutPNG(s.ctx, key, pngBytes); err != nil {
		metrics.OperationErrors.WithLabelValues(MetricStageS3Put).Inc()
		s.logger.Error().Str("key", key).Err(err).Msg("s3 put")
	} else {
		metrics.BytesUploadedS3.Add(float64(len(pngBytes)))
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
		if s.ml == nil {
			return nil
		}
		metrics.BytesSentML.Add(float64(len(frame)) * 2)
		mlStart := time.Now()
		err := s.ml.PostBoth(s.ctx, frame, frameJPEGUploadName, meta)
		metrics.MLLatencySeconds.Observe(time.Since(mlStart).Seconds())
		if err != nil {
			metrics.OperationErrors.WithLabelValues(MetricStageMLProcess).Inc()
			metrics.FramesProcessed.WithLabelValues(MetricFrameOutcomeMLError).Inc()
			logSourceIssueThrottled(s.lastMLLog, s.logger, "ml process", "err", err)
			return err
		}
		metrics.FramesProcessed.WithLabelValues(MetricFrameOutcomeMLOk).Inc()
		return nil
	})
	_ = g.Wait()

	if n%frameLogEveryN == 0 {
		s.logger.Info().Uint64("count", n).Str("last_key", key).Msg("frames")
	}
}
