package app

import (
	"context"
	"log/slog"
	"time"

	"coordinator/internal/adapters/postgres"
	"coordinator/internal/config"
	"coordinator/internal/core/services"
)

func (a *App) initDeps(ctx context.Context) error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return err
	}
	a.Cfg = cfg

	pgs, err := postgres.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}

	a.deps = deps{
		svc:        services.NewCoordinatorService(pgs, time.Duration(cfg.HeartbeatTimeoutSec)*time.Second),
		closeStore: pgs.Close,
	}
	slog.Info("coordinator store", "backend", "postgres")
	return nil
}
