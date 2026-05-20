package main

import (
	"context"
	"os"

	zlog "github.com/rs/zerolog/log"
	"os/signal"
	"syscall"

	"coordinator/internal/app"
	"coordinator/internal/logging"
)

func main() {
	logging.InitGlobal("coordinator")
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("init")
		os.Exit(1)
	}
	if err := a.Run(ctx); err != nil {
		zlog.Error().Err(err).Msg("run")
		os.Exit(1)
	}
}
