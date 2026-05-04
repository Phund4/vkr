package services

import (
	"context"

	"coordinator/internal/core/domain"
)

// Store персистентность источников, пула воркеров и heartbeat.
type Store interface {
	Sources(ctx context.Context, zoneID string) ([]domain.Source, error)
	ZoneWorkers(ctx context.Context, zoneID string) (map[string][]domain.Replica, error)
	UpsertHeartbeat(ctx context.Context, hb domain.WorkerHeartbeat) error
	Heartbeats(ctx context.Context) ([]domain.WorkerHeartbeat, error)
}
