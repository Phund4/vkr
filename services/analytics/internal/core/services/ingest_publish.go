package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	zlog "github.com/rs/zerolog/log"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/core/domain"
)

// publishPersist публикует PersistEvent в Kafka для pusher; флаги задают, что писать в ClickHouse.
func (s *IngestService) publishPersist(
	ctx context.Context,
	seg, cam, atStr, s3k, pipeAt string,
	ml domain.MLParsed,
	rawML string,
	hasML bool,
	persistIncident, persistCongestion bool,
	updateIncidentMetrics, updateCongestionMetrics bool,
) error {
	crashP := ml.Incident.CrashProbability
	cong := ml.Congestion.CongestionScore
	lbl := strings.TrimSpace(ml.Incident.Label)

	if hasML && updateIncidentMetrics {
		metrics.CrashProbability.WithLabelValues(seg, cam).Set(crashP)
		metrics.CrashAlert.WithLabelValues(seg, cam).Set(crashAlertMetricValue(crashP, s.cfg.CrashAlertThreshold))
	}
	if hasML && updateCongestionMetrics {
		metrics.CongestionScore.WithLabelValues(seg, cam).Set(cong)
	}

	if persistCongestion && updateCongestionMetrics {
		if !s.shouldPersistCongestion(seg, cam) {
			persistCongestion = false
		}
	}

	if rawML == "" {
		rawML = rawMLJSONEmpty
	}

	pe := domain.PersistEvent{
		SegmentID:         seg,
		CameraID:          cam,
		ObservedAt:        atStr,
		PipelineStartedAt: pipeAt,
		S3Key:             s3k,
		RawML:             rawML,
		HasML:             hasML,
		PersistIncident:   persistIncident,
		PersistCongestion: persistCongestion,
		CrashProbability:  crashP,
		IncidentLabel:     lbl,
		CongestionScore:   cong,
	}

	payload, err := json.Marshal(pe)
	if err != nil {
		metrics.ProcessErrors.WithLabelValues(ingestMetricStageJSONEncode).Inc()
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

// parseRoadEvent разбирает и валидирует входящий JSON.
func parseRoadEvent(body []byte) (domain.RoadEvent, string, string, string, string, string, error) {
	var ev domain.RoadEvent
	if err := json.Unmarshal(body, &ev); err != nil {
		metrics.ProcessErrors.WithLabelValues(ingestMetricStageJSONDecode).Inc()
		return ev, "", "", "", "", "", err
	}
	seg := sanitizeLabel(strings.TrimSpace(ev.SegmentID))
	cam := sanitizeLabel(strings.TrimSpace(ev.CameraID))
	atStr := strings.TrimSpace(ev.ObservedAt)
	s3k := strings.TrimSpace(ev.S3Key)
	pipeAt := strings.TrimSpace(ev.PipelineStartedAt)
	if seg == "" || cam == "" || atStr == "" {
		metrics.ProcessErrors.WithLabelValues(ingestMetricStageValidate).Inc()
		return ev, "", "", "", "", "", ErrIngestValidation
	}
	return ev, seg, cam, atStr, s3k, pipeAt, nil
}

// crashAlertMetricValue: 1 если crash_probability >= порога, иначе 0 (Prometheus analytics_road_crash_alert).
func crashAlertMetricValue(crashProbability, threshold float64) float64 {
	if crashProbability >= threshold {
		return 1
	}
	return 0
}

// incidentAlert вычисляет, нужно ли писать инцидент в CH.
func (s *IngestService) incidentAlert(ml domain.MLParsed, hasML bool) bool {
	if !hasML {
		return false
	}
	lbl := strings.TrimSpace(ml.Incident.Label)
	if ml.Incident.HasIncident != nil {
		return *ml.Incident.HasIncident
	}
	if lbl != "" {
		return strings.EqualFold(lbl, incidentLabelCrash)
	}
	return ml.Incident.CrashProbability >= s.cfg.CrashAlertThreshold
}

// parseMLParsed разбирает поле ml; при частичном JSON заполняет только переданные блоки.
func parseMLParsed(ev domain.RoadEvent) (domain.MLParsed, string, bool) {
	raw := string(ev.ML)
	if len(ev.ML) == 0 || raw == "null" {
		return domain.MLParsed{}, rawMLJSONEmpty, false
	}
	hasI, hasC, inc, cong := extractMLHalves(ev)
	if !hasI && !hasC {
		var ml domain.MLParsed
		if err := json.Unmarshal(ev.ML, &ml); err != nil {
			zlog.Warn().Err(err).Msg("ml parse")
			return domain.MLParsed{}, raw, true
		}
		return ml, raw, true
	}
	var ml domain.MLParsed
	if hasI {
		_ = json.Unmarshal(inc, &ml.Incident)
	}
	if hasC {
		_ = json.Unmarshal(cong, &ml.Congestion)
	}
	return ml, raw, true
}
