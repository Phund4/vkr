package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/core/domain"
)

const (
	QuerySegmentID = "segment_id"
	QueryCameraID  = "camera_id"
	QueryLimit     = "limit"
)

var errInvalidLimit = errors.New("invalid limit")

// RoadDataGetter контракт сервиса чтения дорожных данных из ClickHouse.
type RoadDataGetter interface {
	ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error)
	ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error)
}

// RoadDataHandler HTTP-обработчик списков road_incidents / road_congestion.
type RoadDataHandler struct {
	service RoadDataGetter
}

// NewRoadDataHandler создаёт обработчик.
func NewRoadDataHandler(service RoadDataGetter) *RoadDataHandler {
	return &RoadDataHandler{service: service}
}

// ListRoadIncidents godoc
// @Summary Список записей road_incidents
// @Param segment_id query string false "Фильтр по segment_id"
// @Param camera_id query string false "Фильтр по camera_id"
// @Param limit query int false "Лимит строк (по умолчанию 100, макс. 500)"
// @Success 200 {object} dto.RoadIncidentsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/road_incidents [get]
func (h *RoadDataHandler) ListRoadIncidents(c echo.Context) error {
	limit, err := parseLimitQuery(c.QueryParam(QueryLimit))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Status:  dto.StatusError,
			Message: dto.ErrLimitQueryInvalid,
		})
	}
	params := dto.RoadListParamsFromQuery(c.QueryParam(QuerySegmentID), c.QueryParam(QueryCameraID), limit)
	items, err := h.service.ListRoadIncidents(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Status:  dto.StatusError,
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, dto.NewRoadIncidentsListResponse(items))
}

// ListRoadCongestion godoc
// @Summary Список записей road_congestion
// @Param segment_id query string false "Фильтр по segment_id"
// @Param camera_id query string false "Фильтр по camera_id"
// @Param limit query int false "Лимит строк (по умолчанию 100, макс. 500)"
// @Success 200 {object} dto.RoadCongestionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/road_congestion [get]
func (h *RoadDataHandler) ListRoadCongestion(c echo.Context) error {
	limit, err := parseLimitQuery(c.QueryParam(QueryLimit))
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Status:  dto.StatusError,
			Message: dto.ErrLimitQueryInvalid,
		})
	}
	params := dto.RoadListParamsFromQuery(c.QueryParam(QuerySegmentID), c.QueryParam(QueryCameraID), limit)
	items, err := h.service.ListRoadCongestion(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Status:  dto.StatusError,
			Message: err.Error(),
		})
	}
	return c.JSON(http.StatusOK, dto.NewRoadCongestionListResponse(items))
}

func parseLimitQuery(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errInvalidLimit
	}
	return n, nil
}
