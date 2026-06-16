package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
)

var ErrRunHTTPServer = errors.New("failed to run HTTP server")

// Run HTTP, metrics и Kafka consumer до отмены ctx.
func (app *App) Run(ctx context.Context) error {
	if app.deps.metrics != nil {
		app.deps.metrics.Start()
	}

	go app.RunKafkaConsumer(ctx)

	httpShutdown, err := app.runHTTPServer(ctx)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrRunHTTPServer, err)
	}

	<-ctx.Done()

	app.Logger().Info().Msg("shutting down HTTP server")

	if app.deps.metrics != nil {
		app.Logger().Info().Msg("shutting down metrics server")
		_ = app.deps.metrics.Stop(ctx)
	}

	return httpShutdown(ctx)
}

func (app *App) runHTTPServer(_ context.Context) (func(context.Context) error, error) {
	app.initHandlers()
	app.configureHTTPServer()

	go func() {
		if err := app.httpServer.Start(app.Cfg.Server.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger().Error().Err(err).Msg("HTTP server")
		}
	}()

	app.Logger().Info().Str("addr", app.Cfg.Server.Port).Msg("HTTP server started")

	return app.httpServer.Shutdown, nil
}
