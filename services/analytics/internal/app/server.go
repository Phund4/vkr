package app

import (
	"context"
	"fmt"
	"net"

	zlog "github.com/rs/zerolog/log"
	"net/http"

	httpx "traffic-analytics/internal/adapters/http"
)

// RunHTTPServer поднимает HTTP-сервер с маршрутами до отмены rootCtx или ошибки Listen.
func RunHTTPServer(rootCtx context.Context, deps *Deps) error {
	mux := http.NewServeMux()
	httpx.Register(mux, deps.Ingest)

	srv := &http.Server{
		Addr:              deps.Config.ListenAddr,
		Handler:           mux,
		BaseContext:       func(net.Listener) context.Context { return rootCtx },
		ReadHeaderTimeout: httpReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		zlog.Info().
			Str("listen", deps.Config.ListenAddr).
			Str("kafka_persist_topic", deps.Config.KafkaTopicPersist).
			Str("congestion_persist_interval", deps.Config.CongestionPersistInterval.String()).
			Msg("analytics starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("%w: %v", ErrHTTPListen, err)
	case <-rootCtx.Done():
	}

	zlog.Info().Msg("analytics shutdown signal received")
	shCtx, cancel := context.WithTimeout(context.Background(), httpServerShutdown)
	defer cancel()
	if err := srv.Shutdown(shCtx); err != nil {
		zlog.Warn().Err(err).Msg("graceful shutdown")
	}
	zlog.Info().Msg("analytics stopped")
	return nil
}
