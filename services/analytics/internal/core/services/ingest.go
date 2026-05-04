package services

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/domain"
)

// errIngestValidation обязательные поля JSON отсутствуют.
var errIngestValidation = errors.New("segment_id, camera_id, observed_at required")

// EventStore — хранение инцидентов и загруженности (реализует адаптер clickhouse).
type EventStore interface {
	InsertIncident(ctx context.Context, observedAt time.Time, segmentID, cameraID, s3Key string, crashProb float64, label, rawML string) error
	InsertCongestion(ctx context.Context, observedAt time.Time, segmentID, cameraID, s3Key string, congestionScore float64, rawML string) error
	Close() error
}

// IngestService use-case приёма событий дороги.
type IngestService struct {
	// store запись инцидентов и congestion в ClickHouse.
	store EventStore

	// cfg настройки порогов и таймаутов.
	cfg config.Config

	// appCtx отмена при остановке процесса.
	appCtx context.Context

	// mu защита lastCongest.
	mu sync.Mutex

	// lastCongest время последней записи congestion по ключу segment|camera.
	lastCongest map[string]time.Time

	// clickhouseTO таймаут запросов к CH из ingest.
	clickhouseTO time.Duration
}

// NewIngestService создаёт сервис приёма.
func NewIngestService(store EventStore, cfg config.Config, appCtx context.Context) *IngestService {
	return &IngestService{
		store:        store,
		cfg:          cfg,
		appCtx:       appCtx,
		lastCongest:  make(map[string]time.Time),
		clickhouseTO: clickHouseQueryTimeout * time.Second,
	}
}

// HandleIngest обрабатывает POST /v1/ingest.
func (s *IngestService) HandleIngest(w http.ResponseWriter, r *http.Request) {
	reqCtx, reqCancel := withAppShutdown(r.Context(), s.appCtx)
	defer reqCancel()

	body, err := io.ReadAll(io.LimitReader(r.Body, maxIngestBodyBytes))
	if err != nil {
		metrics.IngestErrors.WithLabelValues("read_body").Inc()
		http.Error(w, "read body", http.StatusBadRequest)
		return
	}
	if err := s.ProcessIngest(reqCtx, body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ProcessIngest разбирает JSON тела дорожного события (как POST /v1/ingest) и пишет метрики/CH.
func (s *IngestService) ProcessIngest(ctx context.Context, body []byte) error {
	var ev domain.RoadEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		metrics.IngestErrors.WithLabelValues("json_decode").Inc()
		return err
	}
	seg := sanitizeLabel(strings.TrimSpace(ev.SegmentID))
	cam := sanitizeLabel(strings.TrimSpace(ev.CameraID))
	atStr := strings.TrimSpace(ev.ObservedAt)
	s3k := strings.TrimSpace(ev.S3Key)
	if seg == "" || cam == "" || atStr == "" {
		metrics.IngestErrors.WithLabelValues("validate").Inc()
		return errIngestValidation
	}

	hasML := len(ev.ML) > 0 && string(ev.ML) != "null"

	var ml domain.MLParsed
	if hasML {
		if err := json.Unmarshal(ev.ML, &ml); err != nil {
			slog.Warn("ml parse", "err", err)
		}
	}

	crashP := ml.Incident.CrashProbability
	cong := ml.Congestion.CongestionScore
	lbl := strings.TrimSpace(ml.Incident.Label)

	var alert bool
	if hasML {
		metrics.CongestionScore.WithLabelValues(seg, cam).Set(cong)
		metrics.CrashProbability.WithLabelValues(seg, cam).Set(crashP)
		// Presence-first semantics:
		// 1) explicit has_incident when producer provides it,
		// 2) then label-based legacy signal,
		// 3) probability threshold only as compatibility fallback.
		if ml.Incident.HasIncident != nil {
			alert = *ml.Incident.HasIncident
		} else if lbl != "" {
			alert = strings.EqualFold(lbl, "crash")
		} else {
			alert = crashP >= s.cfg.CrashAlertThreshold
		}
		alertVal := 0.0
		if alert {
			alertVal = 1.0
		}
		metrics.CrashAlert.WithLabelValues(seg, cam).Set(alertVal)
	}

	observedAt, err := parseObservedAt(atStr)
	if err != nil {
		observedAt = time.Now().UTC()
	}

	raw := string(ev.ML)
	if raw == "" {
		raw = "{}"
	}

	chCtx, chCancel := context.WithTimeout(ctx, s.clickhouseTO)
	defer chCancel()

	if hasML && s.shouldPersistCongestion(seg, cam) {
		if err := s.store.InsertCongestion(chCtx, observedAt, seg, cam, s3k, cong, raw); err != nil {
			slog.Warn("clickhouse congestion insert", "err", err)
		} else {
			s.markCongestionWritten(seg, cam)
			metrics.CongestionRecorded.WithLabelValues(seg, cam).Inc()
		}
	}

	if hasML && alert {
		if err := s.store.InsertIncident(chCtx, observedAt, seg, cam, s3k, crashP, lbl, raw); err != nil {
			slog.Warn("clickhouse incident insert", "err", err)
		} else {
			metrics.IncidentsRecorded.WithLabelValues(seg, cam).Inc()
		}
	}

	return nil
}

func sanitizeLabel(s string) string {
	if len(s) > maxLabelLen {
		return s[:maxLabelLen]
	}
	return s
}

func parseObservedAt(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
		return t.UTC(), nil
	}
	return time.Parse(time.RFC3339, s)
}

func (s *IngestService) shouldPersistCongestion(seg, cam string) bool {
	if s.cfg.CongestionPersistInterval <= 0 {
		return true
	}
	now := time.Now()
	k := seg + "\x00" + cam
	s.mu.Lock()
	defer s.mu.Unlock()
	if t, ok := s.lastCongest[k]; ok && now.Sub(t) < s.cfg.CongestionPersistInterval {
		return false
	}
	return true
}

func (s *IngestService) markCongestionWritten(seg, cam string) {
	k := seg + "\x00" + cam
	s.mu.Lock()
	s.lastCongest[k] = time.Now()
	s.mu.Unlock()
}
