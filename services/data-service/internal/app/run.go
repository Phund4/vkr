package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	retryclients "data-service/internal/adapters/retry-clients"
	"data-service/internal/utils/retry"
)

var (
	ErrRunHTTPServer = errors.New("failed to run HTTP server")
)

// Run запускает HTTP и metrics серверы и ожидает завершения контекста.
func (app *App) Run(ctx context.Context) error {
	if app.deps.metrics != nil {
		app.deps.metrics.Start()
	}

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

// runHTTPServer поднимает HTTP сервер и возвращает функцию graceful shutdown.
func (app *App) runHTTPServer(_ context.Context) (httpShutdown func(ctx context.Context) error, err error) {
	app.initHandlers()
	app.configureHTTPServer()

	go func() {
		if err = app.httpServer.Start(app.Cfg.Server.HTTPServerPort); err != nil && !errors.Is(err, http.ErrServerClosed) {
			app.Logger().Error().Err(err).Msg("failed to serve HTTP")
		}
	}()

	app.Logger().Info().Msgf("started HTTP server on %s", app.Cfg.Server.HTTPServerPort)

	return app.httpServer.Shutdown, nil
}

// newClickhouseClient создает retry-обертку над sql.DB для ClickHouse.
func (a *App) newClickhouseClient(ctx context.Context, db *sql.DB) (*retryclients.ClickhouseClientAdapter, error) {
	clickhouseRetryResultCfg, err := retry.NewRetryConfig[sql.Result](retryclients.PingRetries, retry.RetryOnError)
	if err != nil {
		return nil, fmt.Errorf("clickhouse retry config: %w", err)
	}
	clickhouseRetryRowCfg, err := retry.NewRetryConfig[sql.Row](retryclients.PingRetries, retry.RetryOnError)
	if err != nil {
		return nil, fmt.Errorf("clickhouse retry config: %w", err)
	}
	clickhouseRetryRowsCfg, err := retry.NewRetryConfig[sql.Rows](retryclients.PingRetries, retry.RetryOnError)
	if err != nil {
		return nil, fmt.Errorf("clickhouse retry config: %w", err)
	}

	clickhouseRetryObjResult := retry.NewRetryObj(clickhouseRetryResultCfg)
	clickhouseRetryObjRow := retry.NewRetryObj(clickhouseRetryRowCfg)
	clickhouseRetryObjRows := retry.NewRetryObj(clickhouseRetryRowsCfg)

	clickhouseClient, err := retryclients.NewClickhouseClientAdapter(db, clickhouseRetryObjResult, clickhouseRetryObjRow, clickhouseRetryObjRows)
	if err != nil {
		return nil, fmt.Errorf("clickhouse client: %w", err)
	}

	if err := clickhouseClient.Ping(ctx); err != nil {
		return nil, fmt.Errorf("clickhouse ping: %w", err)
	}

	return clickhouseClient, nil
}
