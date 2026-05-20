package app

import (
	"context"
	"sync"

	zlog "github.com/rs/zerolog/log"
	"time"

	httpmetrics "router/internal/adapters/http"
	"router/internal/config"
	"router/internal/core/domain"
)

// Run запускает воркеры камер, периодический heartbeat coordinator и HTTP /metrics до отменя rootCtx.
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

	if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, 0); err != nil {
		zlog.Warn().Err(err).Msg("coordinator bootstrap heartbeat failed")
	}

	var (
		assignMu          sync.Mutex
		workersStarted    bool
		assignmentCount   int
		wg                sync.WaitGroup
	)

	tryApplyAssignments := func() {
		assignMu.Lock()
		if workersStarted {
			assignMu.Unlock()
			return
		}
		assignMu.Unlock()

		cameras, err := a.deps.coordinator.FetchCameraAssignments(rootCtx, zoneID, clusterID, instanceID)
		if err != nil {
			zlog.Warn().Err(err).Msg("coordinator camera assignments unavailable")
			return
		}
		if len(cameras) == 0 {
			return
		}

		assignMu.Lock()
		if workersStarted {
			assignMu.Unlock()
			return
		}
		assignMu.Unlock()

		if err := a.initVideoPipeline(rootCtx); err != nil {
			zlog.Error().Err(err).Msg("video pipeline init failed")
			return
		}

		assignMu.Lock()
		workersStarted = true
		assignmentCount = len(cameras)
		assignMu.Unlock()

		zlog.Info().
			Int("assigned_sources", len(cameras)).
			Str("zone", zoneID).
			Str("cluster", clusterID).
			Str("instance", instanceID).
			Msg("coordinator camera assignments applied")
		a.startCameras(rootCtx, cameras, &wg)
	}

	tryApplyAssignments()

	assignMu.Lock()
	started := workersStarted
	assignMu.Unlock()
	if !started {
		zlog.Info().
			Str("zone", zoneID).
			Str("cluster", clusterID).
			Str("instance", instanceID).
			Dur("poll", assignmentPollInterval).
			Msg("no assignments yet, router in standby (will retry)")
	}

	if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, assignmentCount); err != nil {
		zlog.Warn().Err(err).Msg("coordinator heartbeat failed")
	}

	go func() {
		t := time.NewTicker(heartbeatTickerInterval)
		defer t.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-t.C:
				assignMu.Lock()
				n := assignmentCount
				assignMu.Unlock()
				if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, n); err != nil {
					zlog.Warn().Err(err).Msg("coordinator heartbeat failed")
				}
			}
		}
	}()

	go func() {
		t := time.NewTicker(assignmentPollInterval)
		defer t.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-t.C:
				assignMu.Lock()
				done := workersStarted
				assignMu.Unlock()
				if done {
					return
				}
				tryApplyAssignments()
			}
		}
	}()

	startEvt := zlog.Info().Str("config", a.deps.cfg.ConfigFile).Str("metrics", a.deps.cfg.Metrics.ListenAddr)
	assignMu.Lock()
	if assignmentCount > 0 {
		startEvt = startEvt.Int("rtsp_sources", assignmentCount)
	}
	assignMu.Unlock()
	startEvt.Msg("router starting")

	srvDone := make(chan struct{})
	go func() {
		defer close(srvDone)
		if err := httpmetrics.RunMetricsServer(rootCtx, a.deps.cfg.Metrics.ListenAddr); err != nil {
			zlog.Error().Err(err).Msg("metrics server")
		}
	}()

	<-rootCtx.Done()
	waitWorkers(&wg)
	<-srvDone
	zlog.Info().Msg("router stopped")
	return nil
}

// suppress unused import if domain only used via Fetch - actually domain.Camera is returned from fetch, not needed in run.go
var _ = domain.Camera{}
