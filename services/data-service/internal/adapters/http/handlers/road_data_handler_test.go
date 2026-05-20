package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/adapters/http/handlers"
	"data-service/internal/core/domain"
)

type stubRoadData struct {
	incidents  []domain.RoadIncident
	congestion []domain.RoadCongestion
	err        error
}

func (s *stubRoadData) ListRoadIncidents(_ context.Context, _ domain.RoadListParams) ([]domain.RoadIncident, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.incidents, nil
}

func (s *stubRoadData) ListRoadCongestion(_ context.Context, _ domain.RoadListParams) ([]domain.RoadCongestion, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.congestion, nil
}

func (s *stubRoadData) ListIncidentsByCamera(_ context.Context, _ domain.RoadTimeRangeParams) ([]domain.RoadIncident, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.incidents, nil
}

func (s *stubRoadData) AvgCongestionBySegment(_ context.Context, _ domain.RoadTimeRangeParams) (domain.CongestionAverageResult, error) {
	return domain.CongestionAverageResult{AvgScore: 0.42, SampleCount: 3}, nil
}

func (s *stubRoadData) ListFramesByCamera(_ context.Context, _ domain.RoadTimeRangeParams) ([]domain.FrameWithURL, error) {
	return []domain.FrameWithURL{{FrameRef: domain.FrameRef{S3Key: "k/frame.png"}}}, nil
}

func TestRoadDataHandler_ListRoadIncidents_OK(t *testing.T) {
	t.Parallel()
	e := echo.New()
	at := time.Date(2026, 5, 4, 12, 0, 0, 0, time.UTC)
	stub := &stubRoadData{
		incidents: []domain.RoadIncident{
			{
				ObservedAt:       at,
				SegmentID:        "seg-1",
				CameraID:         "cam-1",
				S3Key:            "k1",
				CrashProbability: 0.7,
				IncidentLabel:    "crash",
				RawML:            `{}`,
			},
		},
	}
	h := handlers.NewRoadDataHandler(stub)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/road_incidents?limit=10", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.ListRoadIncidents(c))
	require.Equal(t, http.StatusOK, rec.Code)
	var body dto.RoadIncidentsListResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Len(t, body.Items, 1)
	require.Equal(t, "seg-1", body.Items[0].SegmentID)
}

func TestRoadDataHandler_ListRoadIncidents_BadLimit(t *testing.T) {
	t.Parallel()
	e := echo.New()
	h := handlers.NewRoadDataHandler(&stubRoadData{})
	req := httptest.NewRequest(http.MethodGet, "/api/v1/road_incidents?limit=abc", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	require.NoError(t, h.ListRoadIncidents(c))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
