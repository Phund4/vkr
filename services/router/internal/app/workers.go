package app

import (
	"context"
	"sync"

	zlog "github.com/rs/zerolog/log"

	"router/internal/core/domain"
	"router/internal/core/services"
)

// startCameras запускает по одной горутине RunCamera на каждую камеру.
func (a *App) startCameras(ctx context.Context, cameras []domain.Camera, wg *sync.WaitGroup) {
	for _, cam := range cameras {
		cam := cam
		wg.Add(1)
		go func() {
			defer wg.Done()
			services.RunCamera(
				ctx,
				cam,
				a.deps.framePub,
				a.deps.videoPub,
				a.deps.mlPub,
				a.deps.cfg.Storage.Prefix,
				a.deps.cfg.Ingest.FFmpegPath,
				a.deps.cfg.Ingest.TargetFPS,
				a.deps.cfg.Ingest.ProcessWorkers,
			)
		}()
	}
}

// waitWorkers блокируется до завершения всех воркеров камер.
func waitWorkers(wg *sync.WaitGroup) {
	zlog.Info().Msg("waiting for background workers")
	wg.Wait()
}
