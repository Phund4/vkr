package services

import (
	"context"

	"data-service/internal/core/domain"
)

// RoadDataService единый сервис чтения дорожных событий из ClickHouse.
type RoadDataService struct {
	repo RoadDataRepository
}

// NewRoadDataService создаёт сервис.
func NewRoadDataService(repo RoadDataRepository) *RoadDataService {
	return &RoadDataService{repo: repo}
}

// Close закрывает внешние ресурсы.
func (s *RoadDataService) Close(ctx context.Context) error {
	return s.repo.Close(ctx)
}

// ListRoadIncidents список инцидентов.
func (s *RoadDataService) ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error) {
	return s.repo.ListRoadIncidents(ctx, p)
}

// ListRoadCongestion список загруженности.
func (s *RoadDataService) ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error) {
	return s.repo.ListRoadCongestion(ctx, p)
}
