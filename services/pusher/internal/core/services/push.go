package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"pusher/internal/adapters/kafka/dto"
	"pusher/internal/adapters/metrics"
	chwriter "pusher/internal/adapters/clickhouse"
	s3store "pusher/internal/adapters/s3"
)

// PushService обрабатывает сообщения Kafka: кадры в S3 и persist → ClickHouse.
type PushService struct {
	log *zerolog.Logger
	s3  *s3store.Client
	ch  *chwriter.Writer
}

// NewPushService создаёт сервис; s3 может быть nil — тогда блок Files пропускается.
func NewPushService(log *zerolog.Logger, s3 *s3store.Client, ch *chwriter.Writer) *PushService {
	return &PushService{log: log, s3: s3, ch: ch}
}

// ProcessFrameMessage загружает PNG из its.frames.ingest в S3 по s3_key.
func (s *PushService) ProcessFrameMessage(ctx context.Context, payload []byte) error {
	start := time.Now()
	defer func() {
		metrics.FrameProcessDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	var ev dto.FrameIngestEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageJSON).Inc()
		return fmt.Errorf("json: %w", err)
	}
	metrics.BytesKafkaConsumed.Add(float64(len(payload)))

	key := strings.TrimSpace(ev.S3Key)
	if key == "" || ev.ContentBase64 == "" {
		metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageValidate).Inc()
		return ErrFileKeyOrContentRequired
	}
	if s.s3 == nil {
		s.log.Warn().Str("key", key).Msg("skip frame S3: client not configured")
		return nil
	}

	raw, err := base64.StdEncoding.DecodeString(ev.ContentBase64)
	if err != nil {
		metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageFileB64).Inc()
		return fmt.Errorf("base64: %w", err)
	}
	ct := strings.TrimSpace(ev.ContentType)
	if ct == "" {
		ct = "image/png"
	}
	putCtx, cancel := context.WithTimeout(ctx, clickhouseOpTimeout)
	err = s.s3.PutObject(putCtx, key, raw, ct)
	cancel()
	if err != nil {
		metrics.S3Errors.Inc()
		return err
	}
	metrics.S3BytesUploaded.Add(float64(len(raw)))
	return nil
}

// ProcessMessage десериализует dto.PersistEvent, выполняет PutObject для Files и Insert* в ClickHouse при флагах.
func (s *PushService) ProcessMessage(ctx context.Context, payload []byte) error {
	start := time.Now()
	defer func() {
		metrics.ProcessDurationSeconds.Observe(time.Since(start).Seconds())
	}()

	var ev dto.PersistEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageJSON).Inc()
		return fmt.Errorf("json: %w", err)
	}
	metrics.BytesKafkaConsumed.Add(float64(len(payload)))

	seg := strings.TrimSpace(ev.SegmentID)
	cam := strings.TrimSpace(ev.CameraID)
	if seg == "" || cam == "" {
		metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageValidate).Inc()
		return ErrSegmentOrCameraRequired
	}

	if s.s3 != nil && len(ev.Files) > 0 {
		for _, f := range ev.Files {
			key := strings.TrimSpace(f.Key)
			if key == "" || f.ContentBase64 == "" {
				metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageFileMeta).Inc()
				return ErrFileKeyOrContentRequired
			}
			raw, err := base64.StdEncoding.DecodeString(f.ContentBase64)
			if err != nil {
				metrics.ProcessErrors.WithLabelValues(metrics.ProcessStageFileB64).Inc()
				return fmt.Errorf("base64: %w", err)
			}
			putCtx, cancel := context.WithTimeout(ctx, clickhouseOpTimeout)
			err = s.s3.PutObject(putCtx, key, raw, strings.TrimSpace(f.ContentType))
			cancel()
			if err != nil {
				metrics.S3Errors.Inc()
				return err
			}
			metrics.S3BytesUploaded.Add(float64(len(raw)))
		}
	} else if len(ev.Files) > 0 && s.s3 == nil {
		s.log.Warn().Int("files", len(ev.Files)).Msg("skip S3 files: client not configured")
	}

	observedAt, err := parseObservedAt(ev.ObservedAt)
	if err != nil {
		observedAt = time.Now().UTC()
	}
	s3Key := strings.TrimSpace(ev.S3Key)
	rawML := ev.RawML
	if rawML == "" {
		rawML = "{}"
	}

	chCtx, chCancel := context.WithTimeout(ctx, clickhouseOpTimeout)
	defer chCancel()

	var wroteCH bool
	var e2eStart time.Time
	if t, err := parseObservedAt(ev.PipelineStartedAt); err == nil {
		e2eStart = t
	}

	if ev.PersistCongestion && ev.HasML {
		if err := s.ch.InsertCongestion(chCtx, observedAt, seg, cam, s3Key, ev.CongestionScore, rawML); err != nil {
			return err
		}
		wroteCH = true
		metrics.BytesClickHouseWritten.Add(float64(len(rawML)))
	}
	if ev.PersistIncident && ev.HasML {
		if err := s.ch.InsertIncident(chCtx, observedAt, seg, cam, s3Key, ev.CrashProbability, strings.TrimSpace(ev.IncidentLabel), rawML); err != nil {
			return err
		}
		wroteCH = true
		metrics.BytesClickHouseWritten.Add(float64(len(rawML)))
	}

	if wroteCH && !e2eStart.IsZero() {
		metrics.PipelineE2ESeconds.Observe(time.Since(e2eStart).Seconds())
	}

	return nil
}

// parseObservedAt парсит RFC3339Nano или RFC3339 в UTC; пустая строка даёт ErrObservedAtEmpty.
func parseObservedAt(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, ErrObservedAtEmpty
	}
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, err
	}
	return t.UTC(), nil
}
