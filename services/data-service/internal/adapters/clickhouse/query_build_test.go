package clickhouse

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"data-service/internal/core/domain"
)

func TestNormalizeLimit(t *testing.T) {
	t.Parallel()
	require.Equal(t, 100, normalizeLimit(0))
	require.Equal(t, 500, normalizeLimit(1000))
	require.Equal(t, 10, normalizeLimit(10))
}

func TestBuildIncidentsListQuery_Filters(t *testing.T) {
	t.Parallel()
	from := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 5, 2, 0, 0, 0, 0, time.UTC)
	q, args := buildIncidentsListQuery("default.road_incidents", domain.RoadListParams{
		SegmentID: "seg-a",
		CameraID:  "cam-1",
		From:      from,
		To:        to,
		Limit:     5,
	})
	require.Contains(t, q, "segment_id = ?")
	require.Contains(t, q, "observed_at >=")
	require.Len(t, args, 5)
	require.Equal(t, 5, args[len(args)-1])
}
