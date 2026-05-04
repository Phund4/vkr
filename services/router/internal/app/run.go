package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	httpmetrics "router/internal/adapters/http"
	"router/internal/config"
)

// Run воркеры камер, heartbeat coordinator и /metrics.
func (a *App) Run(rootCtx context.Context) error {
	logArgs := []any{
		"config", a.deps.cfg.ConfigFile,
		"metrics", a.deps.cfg.Metrics.ListenAddr,
	}

	zoneID := config.CoordinatorZoneIDFromEnv()
	clusterID := config.CoordinatorClusterIDFromEnv()
	instanceID := config.CoordinatorInstanceIDFromEnv()
	if a.deps.coordinator == nil {
		return fmt.Errorf("set COORDINATOR_BASE_URL")
	}
	if zoneID == "" || clusterID == "" || instanceID == "" {
		return fmt.Errorf("set COORDINATOR_ZONE_ID, COORDINATOR_CLUSTER_ID, COORDINATOR_INSTANCE_ID")
	}

	if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, 0); err != nil {
		slog.Warn("coordinator bootstrap heartbeat failed", "err", err)
	}

	cameras, err := a.deps.coordinator.FetchCameraAssignments(rootCtx, zoneID, clusterID, instanceID)
	if err != nil {
		slog.Warn("coordinator camera assignments unavailable, starting in standby", "err", err)
		cameras = nil
	}
	if len(cameras) == 0 {
		slog.Info("no assignments yet, router is running in standby", "zone", zoneID, "cluster", clusterID, "instance", instanceID)
	}
	if len(cameras) > 0 {
		if err := a.initVideoPipeline(rootCtx); err != nil {
			return err
		}
		logArgs = append(logArgs, "rtsp_sources", len(cameras))
		slog.Info("coordinator camera assignments applied", "assigned_sources", len(cameras), "zone", zoneID, "cluster", clusterID, "instance", instanceID)
	}
	assignmentCount := len(cameras)
	if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, assignmentCount); err != nil {
		slog.Warn("coordinator heartbeat failed", "err", err)
	}
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-rootCtx.Done():
				return
			case <-t.C:
				if err := a.deps.coordinator.SendHeartbeat(rootCtx, zoneID, clusterID, instanceID, assignmentCount); err != nil {
					slog.Warn("coordinator heartbeat failed", "err", err)
				}
			}
		}
	}()
	slog.Info("router starting", logArgs...)

	var wg sync.WaitGroup
	if len(cameras) > 0 {
		a.startCameras(rootCtx, cameras, &wg)
	}

	srvDone := make(chan struct{})
	go func() {
		defer close(srvDone)
		if err := httpmetrics.RunMetricsServer(rootCtx, a.deps.cfg.Metrics.ListenAddr); err != nil {
			slog.Error("metrics server", "err", err)
		}
	}()

	<-rootCtx.Done()
	waitWorkers(&wg)
	<-srvDone
	slog.Info("router stopped")
	return nil
}
