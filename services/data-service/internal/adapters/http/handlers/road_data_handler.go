package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/core/domain"
)

const (
	QuerySegmentID = "segment_id"
	QueryCameraID  = "camera_id"
	QueryLimit     = "limit"
	PathCameraID   = "camera_id"
	PathSegmentID  = "segment_id"
)

var errInvalidLimit = errors.New("invalid limit")

// RoadDataAPI контракт сервиса чтения дорожных данных.
type RoadDataAPI interface {
	ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error)
	ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error)
	ListIncidentsByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.RoadIncident, error)
	AvgCongestionBySegment(ctx context.Context, p domain.RoadTimeRangeParams) (domain.CongestionAverageResult, error)
	ListFramesByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.FrameWithURL, error)
}

// RoadDataHandler HTTP-обработчик дорожных данных и кадров.
type RoadDataHandler struct {
	service RoadDataAPI
}

// NewRoadDataHandler создаёт обработчик.
func NewRoadDataHandler(service RoadDataAPI) *RoadDataHandler {
	return &RoadDataHandler{service: service}
}

// ListRoadIncidents godoc
// @Summary Список записей road_incidents
// @Param segment_id query string false "Фильтр по segment_id"
// @Param camera_id query string false "Фильтр по camera_id"
// @Param from query string false "Начало интервала (RFC3339)"
// @Param to query string false "Конец интервала (RFC3339)"
// @Param limit query int false "Лимит строк (по умолчанию 100, макс. 500)"
// @Success 200 {object} dto.RoadIncidentsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/road_incidents [get]
func (h *RoadDataHandler) ListRoadIncidents(c echo.Context) error {
	params, err := h.parseListParams(c)
	if err != nil {
		return h.writeListParamsError(c, err)
	}
	items, err := h.service.ListRoadIncidents(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.NewRoadIncidentsListResponse(items))
}

// ListRoadCongestion godoc
// @Summary Список записей road_congestion
// @Param segment_id query string false "Фильтр по segment_id"
// @Param camera_id query string false "Фильтр по camera_id"
// @Param from query string false "Начало интервала (RFC3339)"
// @Param to query string false "Конец интервала (RFC3339)"
// @Param limit query int false "Лимит строк"
// @Success 200 {object} dto.RoadCongestionListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/road_congestion [get]
func (h *RoadDataHandler) ListRoadCongestion(c echo.Context) error {
	params, err := h.parseListParams(c)
	if err != nil {
		return h.writeListParamsError(c, err)
	}
	items, err := h.service.ListRoadCongestion(c.Request().Context(), params)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.NewRoadCongestionListResponse(items))
}

// ListCameraIncidents godoc
// @Summary Инциденты по камере за интервал времени
// @Param camera_id path string true "Идентификатор камеры"
// @Param from query string true "Начало интервала (RFC3339)"
// @Param to query string true "Конец интервала (RFC3339)"
// @Param limit query int false "Лимит строк"
// @Success 200 {object} dto.RoadIncidentsListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cameras/{camera_id}/incidents [get]
func (h *RoadDataHandler) ListCameraIncidents(c echo.Context) error {
	cameraID := strings.TrimSpace(c.Param(PathCameraID))
	if cameraID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrCameraIDRequired})
	}
	from, to, err := parseTimeRangeQuery(c)
	if err != nil {
		return timeRangeBadRequest(c, err)
	}
	limit, err := parseLimitOptional(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrLimitQueryInvalid})
	}
	p := roadTimeParams(cameraID, "", from, to, limit)
	items, err := h.service.ListIncidentsByCamera(c.Request().Context(), p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.NewRoadIncidentsListResponse(items))
}

// AvgSegmentCongestion godoc
// @Summary Средняя загруженность контура (segment) за интервал
// @Param segment_id path string true "Идентификатор контура/сегмента"
// @Param from query string true "Начало интервала (RFC3339)"
// @Param to query string true "Конец интервала (RFC3339)"
// @Success 200 {object} dto.CongestionAverageResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/segments/{segment_id}/congestion/average [get]
func (h *RoadDataHandler) AvgSegmentCongestion(c echo.Context) error {
	segmentID := strings.TrimSpace(c.Param(PathSegmentID))
	if segmentID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrSegmentIDRequired})
	}
	from, to, err := parseTimeRangeQuery(c)
	if err != nil {
		return timeRangeBadRequest(c, err)
	}
	p := roadTimeParams("", segmentID, from, to, 0)
	res, err := h.service.AvgCongestionBySegment(c.Request().Context(), p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.NewCongestionAverageResponse(res))
}

// ListFrames godoc
// @Summary Кадры камеры из S3 за интервал (ключи из ClickHouse + presigned URL)
// @Param camera_id query string true "Идентификатор камеры"
// @Param from query string true "Начало интервала (RFC3339)"
// @Param to query string true "Конец интервала (RFC3339)"
// @Param limit query int false "Лимит строк"
// @Success 200 {object} dto.FramesListResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/frames [get]
func (h *RoadDataHandler) ListFrames(c echo.Context) error {
	cameraID := strings.TrimSpace(c.QueryParam(QueryCameraID))
	if cameraID == "" {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrCameraIDRequired})
	}
	from, to, err := parseTimeRangeQuery(c)
	if err != nil {
		return timeRangeBadRequest(c, err)
	}
	limit, err := parseLimitOptional(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrLimitQueryInvalid})
	}
	p := roadTimeParams(cameraID, "", from, to, limit)
	items, err := h.service.ListFramesByCamera(c.Request().Context(), p)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, dto.NewFramesListResponse(items))
}

func (h *RoadDataHandler) parseListParams(c echo.Context) (domain.RoadListParams, error) {
	limit, err := parseLimitOptional(c)
	if err != nil {
		return domain.RoadListParams{}, errInvalidLimit
	}
	var from, to time.Time
	if c.QueryParam(QueryFrom) != "" || c.QueryParam(QueryTo) != "" {
		from, to, err = parseTimeRangeQuery(c)
		if err != nil {
			return domain.RoadListParams{}, err
		}
	}
	return domain.RoadListParams{
		SegmentID: c.QueryParam(QuerySegmentID),
		CameraID:  c.QueryParam(QueryCameraID),
		From:      from,
		To:        to,
		Limit:     limit,
	}, nil
}

func (h *RoadDataHandler) writeListParamsError(c echo.Context, err error) error {
	if errors.Is(err, errInvalidLimit) {
		return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: dto.ErrLimitQueryInvalid})
	}
	if errors.Is(err, errTimeRangeRequired) || errors.Is(err, errInvalidTimeRange) {
		return timeRangeBadRequest(c, err)
	}
	return c.JSON(http.StatusBadRequest, dto.ErrorResponse{Status: dto.StatusError, Message: err.Error()})
}
