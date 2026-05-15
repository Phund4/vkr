// @title ITS Data Service API
// @version 1.0
// @description HTTP API: health/probe, чтение ClickHouse road_incidents и road_congestion, Swagger.
// @BasePath /

//go:generate sh -c "cd .. && go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/main.go -o docs -d . --parseInternal --parseDependency"

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	_ "data-service/docs"

	application "data-service/internal/app"
)

func main() {
	ctx, cancel := ctxWithSyscallHandler()
	defer cancel()

	app := application.New(ctx)
	if err := app.Run(ctx); err != nil {
		app.Logger().
			Fatal().Err(err).
			Msg("failed to run data service")
	}
}

func ctxWithSyscallHandler() (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(context.Background())

	sigChan := make(chan os.Signal, 1)
	signal.Notify(
		sigChan,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	go func() {
		<-sigChan
		cancel()
	}()

	return ctx, cancel
}
