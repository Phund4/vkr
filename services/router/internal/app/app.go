package app

import (
	"context"

	coordinatorclient "router/internal/adapters/coordinator"
	mlclient "router/internal/adapters/ml"
	s3store "router/internal/adapters/s3"
	"router/internal/config"
)

// App router: coordinator, опционально S3/ML после initVideoPipeline.
type App struct {
	deps deps
}

type deps struct {
	cfg         *config.Root
	coordinator *coordinatorclient.Client
	store       *s3store.Client
	ml          *mlclient.Client
}

// New загружает конфиг и клиент coordinator.
func New(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}
