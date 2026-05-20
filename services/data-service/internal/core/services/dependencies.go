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
	ListIncidentsByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.RoadIncident, error)
	AvgCongestionBySegment(ctx context.Context, p domain.RoadTimeRangeParams) (domain.CongestionAverageResult, error)
	ListFramesByCamera(ctx context.Context, p domain.RoadTimeRangeParams) ([]domain.FrameRef, error)
}

// FrameURLSigner опциональные presigned URL для кадров в S3.
type FrameURLSigner interface {
	PresignGetURL(ctx context.Context, key string) (string, error)
}
