package services

import (
	"context"
	"strings"

	"coordinator/internal/adapters/routernotify"
	"coordinator/internal/core/domain"
)

// RouterNotifyResult исход push reload на один инстанс router.
type RouterNotifyResult struct {
	ZoneID     string
	ClusterID  string
	InstanceID string
	URL        string
	OK         bool
	Error      string
}

// NotifyRoutersReload шлёт POST /v1/reload всем живым ingestion_instances (опционально в zone_id).
func (s *CoordinatorService) NotifyRoutersReload(ctx context.Context, zoneID string, revision uint64) []RouterNotifyResult {
	ctxBg := context.Background()
	workersByZone, err := s.store.ZoneWorkers(ctxBg, zoneID)
	if err != nil {
		return nil
	}
	hbs, err := s.store.Heartbeats(ctxBg)
	if err != nil {
		return nil
	}
	heartbeats := make(map[string]domain.WorkerHeartbeat, len(hbs))
	for _, hb := range hbs {
		heartbeats[workerKey(hb.ZoneID, hb.ClusterID, hb.InstanceID)] = hb
	}

	var out []RouterNotifyResult
	for zid, pool := range workersByZone {
		if zoneID != "" && zid != zoneID {
			continue
		}
		for _, r := range pool {
			if strings.TrimSpace(r.URL) == "" {
				continue
			}
			if !s.isAlive(zid, r.ClusterID, r.InstanceID, heartbeats) {
				continue
			}
			res := RouterNotifyResult{
				ZoneID:     zid,
				ClusterID:  r.ClusterID,
				InstanceID: r.InstanceID,
				URL:        r.URL,
			}
			if err := routernotify.PostReload(ctx, r.URL, revision, 0); err != nil {
				res.Error = err.Error()
			} else {
				res.OK = true
			}
			out = append(out, res)
		}
	}
	return out
}
