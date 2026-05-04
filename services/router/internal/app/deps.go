package app

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	coordinatorclient "router/internal/adapters/coordinator"
	mlclient "router/internal/adapters/ml"
	s3store "router/internal/adapters/s3"
	"router/internal/config"
)

func (a *App) initDeps(ctx context.Context) error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	base := config.CoordinatorBaseURLFromEnv()
	if base == "" {
		return fmt.Errorf("set COORDINATOR_BASE_URL")
	}
	if config.CoordinatorZoneIDFromEnv() == "" || config.CoordinatorClusterIDFromEnv() == "" || config.CoordinatorInstanceIDFromEnv() == "" {
		return fmt.Errorf("set COORDINATOR_ZONE_ID, COORDINATOR_CLUSTER_ID, COORDINATOR_INSTANCE_ID")
	}

	a.deps = deps{
		cfg:         cfg,
		coordinator: coordinatorclient.New(base, 10*time.Second),
	}
	return nil
}

func (a *App) initVideoPipeline(ctx context.Context) error {
	ak := os.Getenv("AWS_ACCESS_KEY_ID")
	sk := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if ak == "" || sk == "" {
		return ErrMissingAWSCredentials
	}
	store, err := s3store.New(ctx, a.deps.cfg.S3.Endpoint, a.deps.cfg.S3.Region, a.deps.cfg.S3.Bucket, ak, sk)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}
	if a.deps.cfg.Ingest.CreateBucketIfMissing {
		if err := store.EnsureBucket(ctx); err != nil {
			return fmt.Errorf("ensure bucket: %w", err)
		}
		slog.Info("bucket ok", "bucket", a.deps.cfg.S3.Bucket)
	}
	a.deps.store = store
	a.deps.ml = mlclient.New(
		a.deps.cfg.ML.BaseURL,
		a.deps.cfg.ML.ProcessPath,
		time.Duration(a.deps.cfg.ML.TimeoutSeconds)*time.Second,
	)
	return nil
}
