package app

import (
	"context"

	coordinatorclient "router/internal/adapters/coordinator"
	kafkapub "router/internal/adapters/kafka"
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
	framePub    *kafkapub.FramePublisher
	videoPub    services.VideoMetaPublisher
	mlPub       services.MLFramePublisher
}

// New загружает конфиг из окружения и инициализирует клиент coordinator.
func New(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}
