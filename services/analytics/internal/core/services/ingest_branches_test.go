package services

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/domain"
)

type stubPublisher struct {
	payloads [][]byte
}

func (s *stubPublisher) Publish(_ context.Context, _ []byte, value []byte) error {
	s.payloads = append(s.payloads, append([]byte(nil), value...))
	return nil
}

func TestProcessAccidentResult_PublishesWithoutCongestion(t *testing.T) {
	t.Parallel()
	pub := &stubPublisher{}
	svc := NewIngestService(pub, config.Config{CrashAlertThreshold: 0.5}, context.Background())
	body, err := json.Marshal(domain.RoadEvent{
		SegmentID:  "seg",
		CameraID:   "cam",
		ObservedAt: "2026-05-04T12:00:00Z",
		ML: json.RawMessage(`{"incident":{"crash_probability":0.9,"label":"crash","has_incident":true}}`),
	})
	require.NoError(t, err)
	require.NoError(t, svc.ProcessAccidentResult(context.Background(), body))
	require.Len(t, pub.payloads, 1)
	var pe domain.PersistEvent
	require.NoError(t, json.Unmarshal(pub.payloads[0], &pe))
	require.True(t, pe.PersistIncident)
	require.False(t, pe.PersistCongestion)
}

func TestExtractMLHalves_Partial(t *testing.T) {
	t.Parallel()
	ev := domain.RoadEvent{ML: json.RawMessage(`{"congestion":{"congestion_score":0.3}}`)}
	hasI, hasC, _, _ := extractMLHalves(ev)
	require.False(t, hasI)
	require.True(t, hasC)
}
