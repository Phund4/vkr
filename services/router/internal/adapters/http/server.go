package httpmetrics

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	pathReload          = "/v1/reload"
	readHeaderTimeout   = 10 * time.Second
	shutdownTimeout     = 10 * time.Second
)

// ReloadHandler вызывается coordinator (POST /v1/reload).
type ReloadHandler interface {
	HandleReload(w http.ResponseWriter, r *http.Request)
}

// RunServer /metrics и POST /v1/reload до отмены ctx.
func RunServer(ctx context.Context, listenAddr string, h ReloadHandler) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("POST "+pathReload, h.HandleReload)
	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		BaseContext:       func(net.Listener) context.Context { return ctx },
		ReadHeaderTimeout: readHeaderTimeout,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zlog.Error().Err(err).Msg("http server")
		}
	}()

	<-ctx.Done()
	zlog.Info().Msg("router shutdown signal, stopping http server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		zlog.Warn().Err(err).Msg("http server shutdown")
	}
	return nil
}

// WriteReloadOK JSON для успешного reload (опционально для coordinator).
func WriteReloadOK(w http.ResponseWriter, revision uint64, assignments int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"revision":    revision,
		"assignments": assignments,
	})
}
