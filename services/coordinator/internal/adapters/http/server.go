package httpserver

import (
	"context"
	"net/http"

	zlog "github.com/rs/zerolog/log"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"coordinator/internal/adapters/http/handlers"
	"coordinator/internal/core/services"
)

// Run HTTP API и /metrics до отмены ctx.
func Run(ctx context.Context, listenAddr string, svc *services.CoordinatorService) error {
	h := handlers.New(svc)
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", h.Health)
	mux.Handle("GET /metrics", promhttp.Handler())
	mux.HandleFunc("GET /v1/sources", h.Sources)
	mux.HandleFunc("GET /v1/assignments", h.Assignments)
	mux.HandleFunc("POST /v1/assignments/reload", h.ReloadAssignments)
	mux.HandleFunc("POST /v1/workers/heartbeat", h.WorkerHeartbeat)
	mux.HandleFunc("GET /v1/workers", h.Workers)
	mux.HandleFunc("GET /v1/ingestion_instances", h.IngestionInstances)

	srv := &http.Server{
		Addr:              listenAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shCtx)
	}()
	zlog.Info().Str("listen", listenAddr).Msg("coordinator starting")
	err := srv.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
