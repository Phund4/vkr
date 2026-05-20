// Program analytics — HTTP-приём событий дороги, метрики Prometheus и публикация в Kafka для pusher.
package main

import (
	"context"
	"os"

	zlog "github.com/rs/zerolog/log"
	"os/signal"
	"syscall"

	"traffic-analytics/internal/app"
	"traffic-analytics/internal/logging"
)

func main() {
	logging.InitGlobal("analytics")

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(rootCtx); err != nil {
		zlog.Error().Err(err).Msg("run")
		os.Exit(1)
	}
}
