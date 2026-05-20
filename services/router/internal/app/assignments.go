package app

import (
	"context"
	"sync"

	zlog "github.com/rs/zerolog/log"

	"router/internal/core/domain"
)

type cameraWorkers struct {
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// assignmentController хранит RTSP-воркеры; обновляется только по push от coordinator.
type assignmentController struct {
	app *App

	mu              sync.Mutex
	rootCtx         context.Context
	zoneID          string
	clusterID       string
	instanceID      string
	workers         *cameraWorkers
	appliedRevision uint64
	appliedCameras  []domain.Camera
	assignmentCount int
	pipelineReady   bool
}

func newAssignmentController(app *App, rootCtx context.Context, zoneID, clusterID, instanceID string) *assignmentController {
	return &assignmentController{
		app:        app,
		rootCtx:    rootCtx,
		zoneID:     zoneID,
		clusterID:  clusterID,
		instanceID: instanceID,
	}
}

func (c *assignmentController) count() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.assignmentCount
}

func (c *assignmentController) snapshot() (revision uint64, assignments int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.appliedRevision, c.assignmentCount
}

func (c *assignmentController) stopWorkers() {
	c.mu.Lock()
	w := c.workers
	c.workers = nil
	c.mu.Unlock()
	if w == nil {
		return
	}
	w.cancel()
	w.wg.Wait()
}

// ReloadFromCoordinator перечитывает GET /v1/assignments и перестраивает воркеры.
func (c *assignmentController) ReloadFromCoordinator(ctx context.Context) error {
	cameras, revision, err := c.app.deps.coordinator.FetchCameraAssignments(
		ctx, c.zoneID, c.clusterID, c.instanceID,
	)
	if err != nil {
		return err
	}
	c.apply(revision, cameras)
	return nil
}

func (c *assignmentController) apply(revision uint64, cameras []domain.Camera) {
	c.mu.Lock()
	if c.pipelineReady && c.workers != nil && revision == c.appliedRevision && camerasEqual(cameras, c.appliedCameras) {
		c.mu.Unlock()
		return
	}
	needPipeline := !c.pipelineReady
	c.mu.Unlock()

	if needPipeline {
		if err := c.app.initVideoPipeline(c.rootCtx); err != nil {
			zlog.Error().Err(err).Msg("video pipeline init failed")
			return
		}
		c.mu.Lock()
		c.pipelineReady = true
		c.mu.Unlock()
	}

	c.stopWorkers()

	if len(cameras) == 0 {
		c.mu.Lock()
		c.appliedRevision = revision
		c.appliedCameras = nil
		c.assignmentCount = 0
		c.mu.Unlock()
		zlog.Info().Uint64("revision", revision).Msg("no camera assignments, workers stopped")
		return
	}

	workersCtx, cancel := context.WithCancel(c.rootCtx)
	w := &cameraWorkers{cancel: cancel}
	c.app.startCameras(workersCtx, cameras, &w.wg)

	c.mu.Lock()
	c.workers = w
	c.appliedRevision = revision
	c.appliedCameras = append([]domain.Camera(nil), cameras...)
	c.assignmentCount = len(cameras)
	c.mu.Unlock()

	zlog.Info().
		Uint64("revision", revision).
		Int("assigned_sources", len(cameras)).
		Str("zone", c.zoneID).
		Str("cluster", c.clusterID).
		Str("instance", c.instanceID).
		Msg("coordinator assignments applied")
}
