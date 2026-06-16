package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

func TestParseRFC3339Time(t *testing.T) {
	t.Parallel()
	ts := time.Date(2026, 5, 4, 12, 30, 0, 0, time.UTC)
	got, err := parseRFC3339Time(ts.Format(time.RFC3339))
	require.NoError(t, err)
	require.Equal(t, ts, got)
}

func TestParseTimeRangeQuery_OK(t *testing.T) {
	t.Parallel()
	e := echo.New()
	from := time.Date(2026, 5, 1, 10, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/?from="+from.Format(time.RFC3339)+"&to="+to.Format(time.RFC3339), nil)
	c := e.NewContext(req, httptest.NewRecorder())
	gotFrom, gotTo, err := parseTimeRangeQuery(c)
	require.NoError(t, err)
	require.Equal(t, from, gotFrom)
	require.Equal(t, to, gotTo)
}

func TestParseTimeRangeQuery_InvalidOrder(t *testing.T) {
	t.Parallel()
	e := echo.New()
	from := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	req := httptest.NewRequest(http.MethodGet, "/?from="+from.Format(time.RFC3339)+"&to="+to.Format(time.RFC3339), nil)
	c := e.NewContext(req, httptest.NewRecorder())
	_, _, err := parseTimeRangeQuery(c)
	require.ErrorIs(t, err, errInvalidTimeRange)
}
