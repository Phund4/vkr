package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/domain"
)

// Формат its.ml.*.out из services/ml-serving/api/kafka_worker.py (_result_envelope).
func TestProcessAccidentResult_MLServingKafkaEnvelope(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{}
	svc := NewIngestService(pub, config.Config{CrashAlertThreshold: 0.5}, context.Background())
	body := []byte(`{
		"segment_id":"seg",
		"camera_id":"cam",
		"observed_at":"2026-05-04T12:00:00Z",
		"s3_key":"k.png",
		"pipeline_started_at":"2026-05-04T12:00:00.1Z",
		"ml":{"incident":{"crash_probability":0.2,"label":"ok","has_incident":false}}
	}`)
	require.NoError(t, svc.ProcessAccidentResult(context.Background(), body))
	require.Len(t, pub.payloads, 1)
}

func TestProcessCongestionResult_MLServingKafkaEnvelope(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{}
	svc := NewIngestService(pub, config.Config{}, context.Background())
	body := []byte(`{
		"segment_id":"seg",
		"camera_id":"cam",
		"observed_at":"2026-05-04T12:00:00Z",
		"ml":{"congestion":{"congestion_score":0.42}}
	}`)
	require.NoError(t, svc.ProcessCongestionResult(context.Background(), body))
	require.Len(t, pub.payloads, 1)
	var pe domain.PersistEvent
	require.NoError(t, json.Unmarshal(pub.payloads[0], &pe))
	require.True(t, pe.PersistCongestion)
}
