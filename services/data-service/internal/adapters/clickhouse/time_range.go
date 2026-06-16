package clickhouse

import (
	"context"
	"database/sql"
	"fmt"

	"data-service/internal/core/domain"
)

// ListIncidentsByCamera возвращает инциденты камеры за интервал времени.
func (r *Repository) ListIncidentsByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.RoadIncident, error) {
	lp := domain.RoadListParams{
		SegmentID: p.SegmentID,
		CameraID:  p.CameraID,
		From:      p.From,
		To:        p.To,
		Limit:     p.Limit,
	}
	return r.listIncidents(ctx, lp)
}

// AvgCongestionBySegment средняя загруженность по segment_id за интервал.
func (r *Repository) AvgCongestionBySegment(ctx context.Context, p domain.RoadTimeRangeParams) (domain.CongestionAverageResult, error) {
	table := r.qualifiedTable(tableRoadCongestion)
	q, args := buildAvgCongestionQuery(table, p)
	var avg sql.NullFloat64
	var cnt uint64
	row := r.client.QueryRowContext(ctx, q, args...)
	if err := row.Scan(&avg, &cnt); err != nil {
		return domain.CongestionAverageResult{}, fmt.Errorf("avg congestion: %w", err)
	}
	out := domain.CongestionAverageResult{
		SegmentID:   p.SegmentID,
		From:        p.From,
		To:          p.To,
		SampleCount: cnt,
	}
	if avg.Valid {
		out.AvgScore = avg.Float64
	}
	return out, nil
}

// ListFramesByCamera уникальные кадры (s3_key) камеры за интервал из обеих таблиц.
func (r *Repository) ListFramesByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.FrameRef, error) {
	inc := r.qualifiedTable(tableRoadIncidents)
	cong := r.qualifiedTable(tableRoadCongestion)
	q, args := buildFramesByCameraQuery(inc, cong, p)
	rows, err := r.client.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("query frames: %w", err)
	}
	defer rows.Close()

	var out []domain.FrameRef
	for rows.Next() {
		var fr domain.FrameRef
		if err := rows.Scan(&fr.ObservedAt, &fr.SegmentID, &fr.CameraID, &fr.S3Key); err != nil {
			return nil, fmt.Errorf("scan frame: %w", err)
		}
		out = append(out, fr)
	}
	return out, rows.Err()
}

func (r *Repository) listIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error) {
	table := r.qualifiedTable(tableRoadIncidents)
	q, args := buildIncidentsListQuery(table, p)
	rows, err := r.client.QueryContext(ctx, q, args...)
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
	return out, rows.Err()
}

func (r *Repository) listCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error) {
	table := r.qualifiedTable(tableRoadCongestion)
	q, args := buildCongestionListQuery(table, p)
	rows, err := r.client.QueryContext(ctx, q, args...)
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
	return out, rows.Err()
}
