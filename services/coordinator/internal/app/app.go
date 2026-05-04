package app

import (
	"context"

	"coordinator/internal/config"
	"coordinator/internal/core/services"
)

// App корень приложения coordinator.
type App struct {
	Cfg  *config.Root
	deps deps
}

type deps struct {
	svc        *services.CoordinatorService
	closeStore func() error
}

// New собирает зависимости и сервис.
func New(ctx context.Context) (*App, error) {
	a := &App{}
	if err := a.initDeps(ctx); err != nil {
		return nil, err
	}
	return a, nil
}
