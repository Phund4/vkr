package clickhouse

import (
	"fmt"
	"strings"
	"time"

	"data-service/internal/core/domain"
)

const (
	defaultListLimit = 100
	maxListLimit     = 500
)

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return defaultListLimit
	}
	if limit > maxListLimit {
		return maxListLimit
	}
	return limit
}

type whereBuilder struct {
	parts []string
	args  []any
}

func (b *whereBuilder) add(cond string, arg any) {
	b.parts = append(b.parts, cond)
	b.args = append(b.args, arg)
}

func (b *whereBuilder) addStringEq(column, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	b.add(column+" = ?", value)
}

func (b *whereBuilder) addTimeRange(column string, from, to time.Time) {
	if !from.IsZero() {
		b.add(column+" >= ?", from.UTC())
	}
	if !to.IsZero() {
		b.add(column+" <= ?", to.UTC())
	}
}

func (b *whereBuilder) sql() (string, []any) {
	if len(b.parts) == 0 {
		return "", nil
	}
	return " AND " + strings.Join(b.parts, " AND "), b.args
}

func listFilters(p domain.RoadListParams) (string, []any) {
	var b whereBuilder
	b.addStringEq("segment_id", p.SegmentID)
	b.addStringEq("camera_id", p.CameraID)
	if !p.From.IsZero() || !p.To.IsZero() {
		b.addTimeRange("observed_at", p.From, p.To)
	}
	return b.sql()
}

func timeRangeFilters(p domain.RoadTimeRangeParams) (string, []any) {
	var b whereBuilder
	b.addStringEq("segment_id", p.SegmentID)
	b.addStringEq("camera_id", p.CameraID)
	b.addTimeRange("observed_at", p.From, p.To)
	return b.sql()
}

func qualifiedSelect(table, database string) string {
	if database == "" {
		return table
	}
	return database + "." + table
}

func appendOrderLimit(base string, args []any, limit int) (string, []any) {
	q := base + " ORDER BY observed_at DESC LIMIT ?"
	return q, append(args, normalizeLimit(limit))
}

func buildIncidentsListQuery(table string, p domain.RoadListParams) (string, []any) {
	base := fmt.Sprintf(`
SELECT observed_at, segment_id, camera_id, s3_key, crash_probability, incident_label, raw_ml
FROM %s
WHERE 1=1`, table)
	where, args := listFilters(p)
	return appendOrderLimit(base+where, args, p.Limit)
}

func buildCongestionListQuery(table string, p domain.RoadListParams) (string, []any) {
	base := fmt.Sprintf(`
SELECT observed_at, segment_id, camera_id, s3_key, congestion_score, raw_ml
FROM %s
WHERE 1=1`, table)
	where, args := listFilters(p)
	return appendOrderLimit(base+where, args, p.Limit)
}

func buildAvgCongestionQuery(table string, p domain.RoadTimeRangeParams) (string, []any) {
	base := fmt.Sprintf(`
SELECT avg(congestion_score), count()
FROM %s
WHERE 1=1`, table)
	where, args := timeRangeFilters(p)
	return base + where, args
}

func buildFramesByCameraQuery(incTable, congTable string, p domain.RoadTimeRangeParams) (string, []any) {
	where, args := timeRangeFilters(p)
	inner := fmt.Sprintf(`
SELECT observed_at, segment_id, camera_id, s3_key
FROM %s
WHERE s3_key != ''%s
UNION ALL
SELECT observed_at, segment_id, camera_id, s3_key
FROM %s
WHERE s3_key != ''%s`, incTable, where, congTable, where)
	// duplicate args for both branches of UNION
	allArgs := append(append([]any{}, args...), args...)
	q := "SELECT observed_at, segment_id, camera_id, s3_key FROM (" + inner + ") ORDER BY observed_at DESC LIMIT ?"
	allArgs = append(allArgs, normalizeLimit(p.Limit))
	return q, allArgs
}
