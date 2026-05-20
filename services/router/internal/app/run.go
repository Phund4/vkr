package app

import (
	"context"
	"time"

	zlog "github.com/rs/zerolog/log"

	httpmetrics "router/internal/adapters/http"
	"router/internal/config"
)

// Run запускает HTTP (/metrics, /v1/reload), heartbeat и ждёт push назначений от coordinator.
func (a *App) Run(rootCtx context.Context) error {
	zoneID := config.CoordinatorZoneIDFromEnv()
	clusterID := config.CoordinatorClusterIDFromEnv()
	instanceID := config.CoordinatorInstanceIDFromEnv()
	if a.deps.coordinator == nil {
		return ErrCoordinatorBaseURL
	}
	if zoneID == "" || clusterID == "" || instanceID == "" {
		return ErrCoordinatorIdentity
	}

	assign := newAssignmentController(a, rootCtx, zoneID, clusterID, instanceID)
	a.assign = assign

	if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, 0); err != nil {
		zlog.Warn().Err(err).Msg("coordinator bootstrap heartbeat failed")
	}

	zlog.Info().
		Str("zone", zoneID).
		Str("cluster", clusterID).
		Str("instance", instanceID).
		Msg("router standby until coordinator POST /v1/assignments/reload")

	go func() {
		t := time.NewTicker(heartbeatTickerInterval)
		defer t.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-t.C:
				if err := a.deps.coordinator.SendHeartbeat(
					rootCtx, zoneID, clusterID, instanceID, assign.count(),
				); err != nil {
					zlog.Warn().Err(err).Msg("coordinator heartbeat failed")
				}
			}
		}
	}()

	startEvt := zlog.Info().
		Str("config", a.deps.cfg.ConfigFile).
		Str("metrics", a.deps.cfg.Metrics.ListenAddr)
	startEvt.Msg("router starting")

	srvDone := make(chan struct{})
	go func() {
		defer close(srvDone)
		if err := httpmetrics.RunServer(rootCtx, a.deps.cfg.Metrics.ListenAddr, a); err != nil {
			zlog.Error().Err(err).Msg("http server")
		}
	}()

	<-rootCtx.Done()
	assign.stopWorkers()
	<-srvDone
	zlog.Info().Msg("router stopped")
	return nil
}
