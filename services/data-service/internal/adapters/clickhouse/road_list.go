package clickhouse

import (
	"context"

	"data-service/internal/core/domain"
)

// ListRoadIncidents возвращает строки road_incidents с фильтрами и сортировкой по времени.
func (r *Repository) ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error) {
	return r.listIncidents(ctx, p)
}

// ListRoadCongestion возвращает строки road_congestion с фильтрами.
func (r *Repository) ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error) {
	return r.listCongestion(ctx, p)
}
