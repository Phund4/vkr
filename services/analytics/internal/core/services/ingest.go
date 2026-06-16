package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"
	"time"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/config"
)

// PersistPublisher отправка нормализованного события в Kafka.
type PersistPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}

// IngestService use-case приёма событий дороги (раздельные ветки accident / congestion).
type IngestService struct {
	publisher PersistPublisher
	cfg       config.Config
	appCtx    context.Context

	mu          sync.Mutex
	lastCongest map[string]time.Time

	publishTO time.Duration
}

// NewIngestService конструирует IngestService с publisher Kafka и таймаутом публикации.
func NewIngestService(pub PersistPublisher, cfg config.Config, appCtx context.Context) *IngestService {
	return &IngestService{
		publisher:   pub,
		cfg:         cfg,
		appCtx:      appCtx,
		lastCongest: make(map[string]time.Time),
		publishTO:   publishRequestTimeout * time.Second,
	}
}

// HandleIngest обрабатывает HTTP POST /v1/ingest (отладка / ручной push; основной путь — Kafka).
func (s *IngestService) HandleIngest(w http.ResponseWriter, r *http.Request) {
	reqCtx, reqCancel := withAppShutdown(r.Context(), s.appCtx)
	defer reqCancel()

	body, err := io.ReadAll(io.LimitReader(r.Body, maxIngestBodyBytes))
	if err != nil {
		metrics.ProcessErrors.WithLabelValues(ingestMetricStageReadBody).Inc()
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
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
		http.Error(w, err.Error(), st)
		return
	}
	w.WriteHeader(http.StatusNoContent)
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
