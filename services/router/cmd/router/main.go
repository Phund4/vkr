// Program router — захват кадров с RTSP, S3, Kafka и два вызова ML по назначениям coordinator.
package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	zlog "github.com/rs/zerolog/log"

	"router/internal/app"
	"router/internal/logging"
)

// main инициализирует приложение, подписывается на SIGINT/SIGTERM и блокируется до остановки.
func main() {
	logging.InitGlobal("router")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(rootCtx)
	if err != nil {
		zlog.Error().Err(err).Msg("init")
		os.Exit(1)
	}
	if err := a.Run(rootCtx); err != nil {
		switch {
		case errors.Is(err, app.ErrMissingAWSCredentials):
			zlog.Error().Msg("set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (e.g. minioadmin)")
		case errors.Is(err, app.ErrCoordinatorBaseURL), errors.Is(err, app.ErrCoordinatorIdentity):
			zlog.Error().Err(err).Msg("coordinator env")
		default:
			zlog.Error().Err(err).Msg("run")
		}
		os.Exit(1)
	}
}
