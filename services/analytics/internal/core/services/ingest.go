package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	zlog "github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"sync"
	"time"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/domain"
)

// PersistPublisher отправка нормализованного события в Kafka.
type PersistPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

// IngestService use-case приёма событий дороги.
type IngestService struct {
	publisher PersistPublisher
	cfg       config.Config
	appCtx    context.Context

	mu          sync.Mutex
	lastCongest map[string]time.Time

	mergeMu  sync.Mutex
	mergeBuf map[string]*mergeEntry

	publishTO time.Duration
}

// NewIngestService конструирует IngestService с publisher Kafka и таймаутом публикации.
func NewIngestService(pub PersistPublisher, cfg config.Config, appCtx context.Context) *IngestService {
	return &IngestService{
		publisher:   pub,
		cfg:         cfg,
		appCtx:      appCtx,
		lastCongest: make(map[string]time.Time),
		mergeBuf:    make(map[string]*mergeEntry),
		publishTO:   publishRequestTimeout * time.Second,
	}
}

// HandleIngest обрабатывает HTTP POST /v1/ingest: чтение тела, ProcessIngest, коды ответа и метрики.
func (s *IngestService) HandleIngest(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	reqCtx, reqCancel := withAppShutdown(r.Context(), s.appCtx)
	defer reqCancel()

	body, err := io.ReadAll(io.LimitReader(r.Body, maxIngestBodyBytes))
	if err != nil {
		metrics.IngestErrors.WithLabelValues(ingestMetricStageReadBody).Inc()
		metrics.IngestRequests.WithLabelValues(strconv.Itoa(http.StatusBadRequest)).Inc()
		metrics.IngestDuration.Observe(time.Since(start).Seconds())
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	metrics.IngestBodyBytes.Add(float64(len(body)))
	if err := s.ProcessIngest(reqCtx, body); err != nil {
		st := http.StatusBadRequest
		if errors.Is(err, ErrPublishKafka) {
			st = http.StatusBadGateway
		} else if errors.Is(err, ErrIngestValidation) {
			st = http.StatusBadRequest
		} else {
			var syn *json.SyntaxError
			var ut *json.UnmarshalTypeError
			if errors.As(err, &syn) || errors.As(err, &ut) {
				st = http.StatusBadRequest
			} else {
				st = http.StatusBadGateway
			}
		}
		metrics.IngestRequests.WithLabelValues(strconv.Itoa(st)).Inc()
		metrics.IngestDuration.Observe(time.Since(start).Seconds())
		http.Error(w, err.Error(), st)
		return
	}
	metrics.IngestRequests.WithLabelValues(strconv.Itoa(http.StatusNoContent)).Inc()
	metrics.IngestDuration.Observe(time.Since(start).Seconds())
	w.WriteHeader(http.StatusNoContent)
}

// ProcessIngest разбирает JSON; при частичном ML от двух сервисов — склеивает по s3_key / времени.
func (s *IngestService) ProcessIngest(ctx context.Context, body []byte) error {
	var ev domain.RoadEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		metrics.IngestErrors.WithLabelValues(ingestMetricStageJSONDecode).Inc()
		return err
	}
	seg := sanitizeLabel(strings.TrimSpace(ev.SegmentID))
	cam := sanitizeLabel(strings.TrimSpace(ev.CameraID))
	atStr := strings.TrimSpace(ev.ObservedAt)
	s3k := strings.TrimSpace(ev.S3Key)
	if seg == "" || cam == "" || atStr == "" {
		metrics.IngestErrors.WithLabelValues(ingestMetricStageValidate).Inc()
		return ErrIngestValidation
	}

	hasI, hasC, inc, cong := extractMLHalves(ev)
	if hasI && hasC {
		return s.processIngestDirect(ctx, ev)
	}
	if hasI || hasC || len(ev.ML) == 0 || string(ev.ML) == "null" {
		pipeAt := strings.TrimSpace(ev.PipelineStartedAt)
		return s.mergeIngest(ctx, seg, cam, atStr, s3k, pipeAt, hasI, hasC, inc, cong)
	}
	return s.processIngestDirect(ctx, ev)
}

// processIngestDirect полный ML в одном сообщении (или иной путь без склейки).
func (s *IngestService) processIngestDirect(ctx context.Context, ev domain.RoadEvent) error {
	seg := sanitizeLabel(strings.TrimSpace(ev.SegmentID))
	cam := sanitizeLabel(strings.TrimSpace(ev.CameraID))
	atStr := strings.TrimSpace(ev.ObservedAt)
	s3k := strings.TrimSpace(ev.S3Key)
	if seg == "" || cam == "" || atStr == "" {
		metrics.IngestErrors.WithLabelValues(ingestMetricStageValidate).Inc()
		return ErrIngestValidation
	}

	hasML := len(ev.ML) > 0 && string(ev.ML) != "null"

	var ml domain.MLParsed
	if hasML {
		if err := json.Unmarshal(ev.ML, &ml); err != nil {
			zlog.Warn().Err(err).Msg("ml parse")
		}
	}

	crashP := ml.Incident.CrashProbability
	cong := ml.Congestion.CongestionScore
	lbl := strings.TrimSpace(ml.Incident.Label)

	var alert bool
	if hasML {
		metrics.CongestionScore.WithLabelValues(seg, cam).Set(cong)
		metrics.CrashProbability.WithLabelValues(seg, cam).Set(crashP)
		if ml.Incident.HasIncident != nil {
			alert = *ml.Incident.HasIncident
		} else if lbl != "" {
			alert = strings.EqualFold(lbl, incidentLabelCrash)
		} else {
			alert = crashP >= s.cfg.CrashAlertThreshold
		}
		alertVal := 0.0
		if alert {
			alertVal = 1.0
		}
		metrics.CrashAlert.WithLabelValues(seg, cam).Set(alertVal)
	}

	raw := string(ev.ML)
	if raw == "" {
		raw = rawMLJSONEmpty
	}

	persistCongestion := hasML && s.shouldPersistCongestion(seg, cam)
	persistIncident := hasML && alert

	pipeAt := strings.TrimSpace(ev.PipelineStartedAt)
	pe := domain.PersistEvent{
		SegmentID:         seg,
		CameraID:          cam,
		ObservedAt:        atStr,
		PipelineStartedAt: pipeAt,
		S3Key:             s3k,
		RawML:             raw,
		HasML:             hasML,
		PersistIncident:   persistIncident,
		PersistCongestion: persistCongestion,
		CrashProbability:  crashP,
		IncidentLabel:     lbl,
		CongestionScore:   cong,
	}

	payload, err := json.Marshal(pe)
	if err != nil {
		metrics.IngestErrors.WithLabelValues(ingestMetricStageJSONEncode).Inc()
		return err
	}

	pubCtx, pubCancel := context.WithTimeout(ctx, s.publishTO)
	defer pubCancel()

	pubStart := time.Now()
	if err := s.publisher.Publish(pubCtx, []byte(seg), payload); err != nil {
		metrics.KafkaPublishErrors.WithLabelValues(kafkaPublishErrorStageWrite).Inc()
		return fmt.Errorf("%w: %v", ErrPublishKafka, err)
	}
	metrics.KafkaPublishDuration.Observe(time.Since(pubStart).Seconds())

	if persistCongestion {
		s.markCongestionWritten(seg, cam)
		metrics.CongestionRecorded.WithLabelValues(seg, cam).Inc()
	}
	if persistIncident {
		metrics.IncidentsRecorded.WithLabelValues(seg, cam).Inc()
	}

	return nil
}

// sanitizeLabel обрезает строку-лейбл до maxLabelLen символов.
func sanitizeLabel(s string) string {
	if len(s) > maxLabelLen {
		return s[:maxLabelLen]
	}
	return s
}

// shouldPersistCongestion применяет интервал CongestionPersistInterval для пары (segment, camera).
func (s *IngestService) shouldPersistCongestion(seg, cam string) bool {
	if s.cfg.CongestionPersistInterval <= 0 {
		return true
	}
	now := time.Now()
	k := seg + congestionPairKeySep + cam
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.lastCongest[k]; ok && now.Sub(t) < s.cfg.CongestionPersistInterval {
		return false
	}
	return true
}

// markCongestionWritten фиксирует время последней записи congestion для пары (segment, camera).
func (s *IngestService) markCongestionWritten(seg, cam string) {
	k := seg + congestionPairKeySep + cam
	s.mu.Lock()
	s.lastCongest[k] = time.Now()
	s.mu.Unlock()
}
