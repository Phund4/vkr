package app

import (
	"context"

	"github.com/labstack/echo/v4"

	"pusher/internal/config"
)

// App pusher.
type App struct {
	deps       deps
	Cfg        *config.Config
	httpServer *echo.Echo
}

// New создаёт приложение.
func New(ctx context.Context) *App {
	app := &App{}
	app.initConfigAndDependencies(ctx)
	app.httpServer = echo.New()
	return app
}
