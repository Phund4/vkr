package clickhouse

import (
	"context"
	"fmt"

	"data-service/internal/core/domain"
)

const (
	roadIncidentsTable = "road_incidents"
	roadCongestionTable = "road_congestion"

	queryRoadIncidents = `
SELECT observed_at, segment_id, camera_id, s3_key, crash_probability, incident_label, raw_ml
FROM %s.%s
WHERE 1=1`

	queryRoadCongestion = `
SELECT observed_at, segment_id, camera_id, s3_key, congestion_score, raw_ml
FROM %s.%s
WHERE 1=1`
)

// ListRoadIncidents возвращает последние строки road_incidents (новые сверху).
func (r *Repository) ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error) {
	q := fmt.Sprintf(queryRoadIncidents, roadIncidentsTable, roadIncidentsTable)

	rows, err := r.client.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query road_incidents: %w", err)
	}
	defer rows.Close()

	var out []domain.RoadIncident
	for rows.Next() {
		var row RoadIncidentRow
		if err := rows.Scan(
			&row.ObservedAt,
			&row.SegmentID,
			&row.CameraID,
			&row.S3Key,
			&row.CrashProbability,
			&row.IncidentLabel,
			&row.RawML,
		); err != nil {
			return nil, fmt.Errorf("scan road_incidents: %w", err)
		}
		out = append(out, row.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListRoadCongestion возвращает последние строки road_congestion (новые сверху).
func (r *Repository) ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error) {
	q := fmt.Sprintf(queryRoadCongestion, roadCongestionTable, roadCongestionTable)

	rows, err := r.client.QueryContext(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("query road_congestion: %w", err)
	}
	defer rows.Close()

	var out []domain.RoadCongestion
	for rows.Next() {
		var row RoadCongestionRow
		if err := rows.Scan(
			&row.ObservedAt,
			&row.SegmentID,
			&row.CameraID,
			&row.S3Key,
			&row.CongestionScore,
			&row.RawML,
		); err != nil {
			return nil, fmt.Errorf("scan road_congestion: %w", err)
		}
		out = append(out, row.toDomain())
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
