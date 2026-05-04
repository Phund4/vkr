package app

import (
	"context"
	"log/slog"
	"sync"

	"router/internal/core/domain"
	"router/internal/core/services"
)

func (a *App) startCameras(ctx context.Context, cameras []domain.Camera, wg *sync.WaitGroup) {
	for _, cam := range cameras {
		cam := cam
		wg.Add(1)
		go func() {
			defer wg.Done()
			services.RunCamera(
				ctx,
				cam,
				a.deps.store,
				a.deps.ml,
				a.deps.cfg.S3.Prefix,
				a.deps.cfg.Ingest.FFmpegPath,
				a.deps.cfg.Ingest.TargetFPS,
				a.deps.cfg.Ingest.ProcessWorkers,
			)
		}()
	}
}

func waitWorkers(wg *sync.WaitGroup) {
	slog.Info("waiting for background workers")
	wg.Wait()
}
