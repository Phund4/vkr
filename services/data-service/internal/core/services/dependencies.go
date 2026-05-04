package services

import (
	"context"

	"data-service/internal/core/domain"
)

// RoadDataRepository контракт репозитория для чтения ИТС-таблиц в ClickHouse.
type RoadDataRepository interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	ListRoadIncidents(ctx context.Context, p domain.RoadListParams) ([]domain.RoadIncident, error)
	ListRoadCongestion(ctx context.Context, p domain.RoadListParams) ([]domain.RoadCongestion, error)
}
