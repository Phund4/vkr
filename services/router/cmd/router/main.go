// Program router — захват кадров с RTSP, загрузка в S3 и вызов ML.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"router/internal/app"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo})))

	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := app.Run(rootCtx); err != nil {
		if errors.Is(err, app.ErrMissingAWSCredentials) {
			slog.Error("set AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY (e.g. minioadmin)")
			os.Exit(1)
		}
		slog.Error("run", "err", err)
		os.Exit(1)
	}
}
