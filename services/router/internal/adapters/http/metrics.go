package httpmetrics

import (
	"context"
	"net"

	zlog "github.com/rs/zerolog/log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	readHeaderTimeout = 10 * time.Second
	shutdownTimeout   = 10 * time.Second
)

// RunMetricsServer /metrics до отмены ctx.
func RunMetricsServer(ctx context.Context, listenAddr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Error().Err(err).Msg("metrics server")
		}
	}()

	<-ctx.Done()
	zlog.Info().Msg("router shutdown signal, stopping metrics server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Warn().Err(err).Msg("metrics server shutdown")
	}
	return nil
}
