package handlers

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/core/domain"
)

const (
	QueryFrom = "from"
	QueryTo   = "to"
)

var (
	errTimeRangeRequired = errors.New("from and to query params are required")
	errInvalidTimeRange  = errors.New("invalid time range")
)

// parseRFC3339Time парсит from/to в UTC.
func parseRFC3339Time(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, errTimeRangeRequired
	}
	if t, err := time.Parse(time.RFC3339Nano, raw); err == nil {
		return t.UTC(), nil
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}, errInvalidTimeRange
	}
	return t.UTC(), nil
}

func parseTimeRangeQuery(c echo.Context) (from, to time.Time, err error) {
	from, err = parseRFC3339Time(c.QueryParam(QueryFrom))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	to, err = parseRFC3339Time(c.QueryParam(QueryTo))
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if !to.After(from) {
		return time.Time{}, time.Time{}, errInvalidTimeRange
	}
	return from, to, nil
}

func parseLimitOptional(c echo.Context) (int, error) {
	raw := strings.TrimSpace(c.QueryParam(QueryLimit))
	if raw == "" {
		return 0, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 {
		return 0, errInvalidLimit
	}
	return n, nil
}

func timeRangeBadRequest(c echo.Context, err error) error {
	msg := dto.ErrTimeRangeInvalid
	if errors.Is(err, errTimeRangeRequired) {
		msg = dto.ErrTimeRangeRequired
	}
	return c.JSON(400, dto.ErrorResponse{Status: dto.StatusError, Message: msg})
}

func roadTimeParams(cameraID, segmentID string, from, to time.Time, limit int) domain.RoadTimeRangeParams {
	return domain.RoadTimeRangeParams{
		SegmentID: strings.TrimSpace(segmentID),
		CameraID:  strings.TrimSpace(cameraID),
		From:      from,
		To:        to,
		Limit:     limit,
	}
}

