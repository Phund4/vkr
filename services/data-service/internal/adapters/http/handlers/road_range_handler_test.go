package handlers_test

import (
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

func TestRoadDataHandler_ListCameraIncidents_OK(t *testing.T) {
	t.Parallel()
	e := echo.New()
	stub := &stubRoadData{incidents: []domain.RoadIncident{{SegmentID: "s", CameraID: "cam-1"}}}
	h := handlers.NewRoadDataHandler(stub)
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(2 * time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/cameras/cam-1/incidents?from="+from.Format(time.RFC3339)+"&to="+to.Format(time.RFC3339), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("camera_id")
	c.SetParamValues("cam-1")
	require.NoError(t, h.ListCameraIncidents(c))
	require.Equal(t, http.StatusOK, rec.Code)
}

func TestRoadDataHandler_AvgSegmentCongestion_OK(t *testing.T) {
	t.Parallel()
	e := echo.New()
	h := handlers.NewRoadDataHandler(&stubRoadData{})
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/segments/zone-a/congestion/average?from="+from.Format(time.RFC3339)+"&to="+to.Format(time.RFC3339), nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("segment_id")
	c.SetParamValues("zone-a")
	require.NoError(t, h.AvgSegmentCongestion(c))
	require.Equal(t, http.StatusOK, rec.Code)
	var body dto.CongestionAverageResponse
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.InDelta(t, 0.42, body.AvgScore, 1e-6)
}
