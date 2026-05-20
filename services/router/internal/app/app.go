package app

import (
	"context"

	coordinatorclient "router/internal/adapters/coordinator"
	s3store "router/internal/adapters/s3"
	"router/internal/config"
	"router/internal/core/services"
)

// App точка входа router: coordinator, пайплайн видео после initVideoPipeline.
type App struct {
	deps   deps
	assign *assignmentController
}

// deps агрегирует зависимости рантайма.
type deps struct {
	cfg         *config.Root
	coordinator *coordinatorclient.Client
	store       *s3store.Client
	mlPub       services.MLFramePublisher
	videoPub    services.VideoMetaPublisher
}

// New загружает конфиг из окружения и инициализирует клиент coordinator.
func New(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}
