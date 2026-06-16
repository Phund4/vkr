package app

import (
	"context"
	"net/http"
	"time"

	zlog "github.com/rs/zerolog/log"

	httpmetrics "router/internal/adapters/http"
)

const reloadHandlerTimeout = 60 * time.Second

// HandleReload POST /v1/reload — push от coordinator на перестройку RTSP-воркеров.
func (a *App) HandleReload(w http.ResponseWriter, r *http.Request) {
	if a.assign == nil {
		http.Error(w, "not ready", http.StatusServiceUnavailable)
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), reloadHandlerTimeout)
	defer cancel()
	if err := a.assign.ReloadFromCoordinator(ctx); err != nil {
		zlog.Warn().Err(err).Msg("reload assignments failed")
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	rev, n := a.assign.snapshot()
	httpmetrics.WriteReloadOK(w, rev, n)
}
