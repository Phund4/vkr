package app

import (
	"context"

	httpserver "coordinator/internal/adapters/http"
)

// Run поднимает HTTP до отмены ctx, затем закрывает стор при необходимости.
func (a *App) Run(ctx context.Context) error {
	err := httpserver.Run(ctx, a.Cfg.ListenAddr, a.deps.svc)
	if a.deps.closeStore != nil {
		_ = a.deps.closeStore()
	}
	return err
}
