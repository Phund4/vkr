package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"router/internal/adapters/capture"
	"router/internal/adapters/metrics"
	s3store "router/internal/adapters/s3"
	"router/internal/core/domain"
)

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

func frameChanCapacity(processWorkers int) int {
	c := processWorkers * 2
	if c < 4 {
		return 4
	}
	return c
}

// RunCamera цикл RTSP → S3 → ML для одной камеры.
func RunCamera(
	ctx context.Context,
	cam domain.Camera,
	store S3Uploader,
	mlc MLProcessor,
	s3Prefix string,
	ffmpegPath string,
	targetFPS float64,
	processWorkers int,
) {
	var frameNo atomic.Uint64
	var lastUpstreamLog, lastMLLog time.Time
	backoff := time.Duration(reconnectBackoffSec) * time.Second

	s := &cameraSession{
		ctx:              ctx,
		cam:              cam,
		store:            store,
		mlc:              mlc,
		prefix:           strings.Trim(s3Prefix, "/"),
		ffmpegPath:       ffmpegPath,
		targetFPS:        targetFPS,
		processWorkers:   processWorkers,
		log:              slog.With("segment", cam.SegmentID, "camera", cam.CameraID),
		frameNo:          &frameNo,
		lastUpstreamLog:  &lastUpstreamLog,
		lastMLLog:        &lastMLLog,
	}

	for ctx.Err() == nil {
		s.runOnce(backoff)
		if ctx.Err() != nil {
			return
		}
	}
}

type cameraSession struct {
	ctx             context.Context
	cam             domain.Camera
	store           S3Uploader
	mlc             MLProcessor
	prefix          string
	ffmpegPath      string
	targetFPS       float64
	processWorkers  int
	log             *slog.Logger
	frameNo         *atomic.Uint64
	lastUpstreamLog *time.Time
	lastMLLog       *time.Time
}

func (s *cameraSession) runOnce(backoff time.Duration) {
	subCtx, cancel := context.WithCancel(s.ctx)
	pipe, err := capture.FFmpegPipe(subCtx, s.ffmpegPath, s.cam.RTSPURL, s.targetFPS)
	if err != nil {
		metrics.OperationErrors.WithLabelValues("ffmpeg_start").Inc()
		logSourceIssueThrottled(s.lastUpstreamLog, s.log, "ffmpeg start (source not ready or invalid URL)", "err", err)
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
	logSourceIssueThrottled(s.lastUpstreamLog, s.log, "source stopped delivering frames, reconnecting", "rtsp_url", s.cam.RTSPURL)
	sleepBackoff(s.ctx, backoff)
}

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
			metrics.OperationErrors.WithLabelValues("frame_read").Inc()
			logSourceIssueThrottled(s.lastUpstreamLog, s.log, "frame read (stream interrupted or paused)", "err", err)
			return
		}
		enqueueFrame(frameCh, frame)
	}
}

func (s *cameraSession) handleFrame(frame []byte) {
	n := s.frameNo.Add(1)
	now := time.Now().UTC()
	day := now.Format("2006-01-02")
	ts := now.UnixNano()
	key := fmt.Sprintf("%s/%s/%s/frame_%d.png", s.prefix, day, s.cam.CameraID, ts)

	pngBytes, err := s3store.JPEGBytesToPNG(frame)
	if err != nil {
		metrics.OperationErrors.WithLabelValues("s3_put").Inc()
		s.log.Error("jpeg to png", "err", err)
		return
	}
	if err := s.store.PutPNG(s.ctx, key, pngBytes); err != nil {
		metrics.OperationErrors.WithLabelValues("s3_put").Inc()
		s.log.Error("s3 put", "key", key, "err", err)
	}

	meta := domain.ProcessMeta{
		SegmentID:  s.cam.SegmentID,
		CameraID:   s.cam.CameraID,
		S3Key:      key,
		ObservedAt: now.Format(time.RFC3339Nano),
	}
	if err := s.mlc.PostProcess(s.ctx, frame, "frame.jpg", meta); err != nil {
		metrics.OperationErrors.WithLabelValues("ml_process").Inc()
		logSourceIssueThrottled(s.lastMLLog, s.log, "ml process", "err", err)
	}

	if n%frameLogEveryN == 0 {
		s.log.Info("frames", "count", n, "last_key", key)
	}
}
