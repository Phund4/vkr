package app

import (
	"context"

	"github.com/labstack/echo/v4"

	"data-service/internal/config"
)

type App struct {
	deps       deps
	Cfg        *config.Config
	httpServer *echo.Echo
}

// New создает и инициализирует экземпляр приложения.
func New(ctx context.Context) *App {
	app := &App{}
	app.initConfigAndDependencies(ctx)

	app.httpServer = echo.New()

	return app
}
