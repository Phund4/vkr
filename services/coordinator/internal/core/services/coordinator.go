package services

import (
	"context"
	"sort"
	"time"

	"coordinator/internal/core/domain"
)

// CoordinatorService назначения источников и учёт heartbeat.
type CoordinatorService struct {
	store            Store
	heartbeatTimeout time.Duration
}

// NewCoordinatorService создаёт сервис поверх стора.
func NewCoordinatorService(store Store, heartbeatTimeout time.Duration) *CoordinatorService {
	if heartbeatTimeout <= 0 {
		heartbeatTimeout = 30 * time.Second
	}
	return &CoordinatorService{
		store:            store,
		heartbeatTimeout: heartbeatTimeout,
	}
}

func (s *CoordinatorService) Sources(zoneID string) []domain.Source {
	items, err := s.store.Sources(context.Background(), zoneID)
	if err != nil {
		return nil
	}
	return items
}

func (s *CoordinatorService) Assignments(zoneID, clusterID, instanceID, dataClass string) []domain.Source {
	ctx := context.Background()
	candidates, err := s.store.Sources(ctx, zoneID)
	if err != nil {
		return nil
	}
	workersByZone, err := s.store.ZoneWorkers(ctx, zoneID)
	if err != nil {
		return nil
	}
	hbs, err := s.store.Heartbeats(ctx)
	if err != nil {
		return nil
	}
	heartbeats := make(map[string]domain.WorkerHeartbeat, len(hbs))
	for _, hb := range hbs {
		heartbeats[workerKey(hb.ZoneID, hb.ClusterID, hb.InstanceID)] = hb
	}
	filtered := make([]domain.Source, 0, len(candidates))
	for _, src := range candidates {
		if dataClass != "" && src.DataClass != dataClass {
			continue
		}
		filtered = append(filtered, src)
	}
	candidates = filtered
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].SourceID < candidates[j].SourceID
	})

	sim := s.initialSim(zoneID, workersByZone, heartbeats)
	owners := make(map[string]domain.Replica, len(candidates))

	for _, src := range candidates {
		pool := workersByZone[src.ZoneID]
		alive := s.aliveReplicas(src.ZoneID, pool, heartbeats)
		if len(alive) == 0 {
			continue
		}
		best := pickLeastLoaded(src.ZoneID, alive, pool, sim)
		owners[src.SourceID] = best
		k := workerKey(src.ZoneID, best.ClusterID, best.InstanceID)
		sim[k]++
	}

	out := make([]domain.Source, 0, len(candidates))
	for _, src := range candidates {
		br, ok := owners[src.SourceID]
		if !ok {
			continue
		}
		if clusterID != "" && br.ClusterID != clusterID {
			continue
		}
		if instanceID != "" && br.InstanceID != instanceID {
			continue
		}
		out = append(out, src)
	}
	return out
}

func (s *CoordinatorService) initialSim(zoneID string, workersByZone map[string][]domain.Replica, heartbeats map[string]domain.WorkerHeartbeat) map[string]float64 {
	sim := make(map[string]float64)
	for zid, pool := range workersByZone {
		if zoneID != "" && zid != zoneID {
			continue
		}
		for _, r := range pool {
			k := workerKey(zid, r.ClusterID, r.InstanceID)
			if _, ok := sim[k]; ok {
				continue
			}
			if !s.isAlive(zid, r.ClusterID, r.InstanceID, heartbeats) {
				continue
			}
			hb := heartbeats[k]
			sim[k] = float64(hb.Assignments) + hb.Load
		}
	}
	return sim
}

func (s *CoordinatorService) aliveReplicas(zoneID string, pool []domain.Replica, heartbeats map[string]domain.WorkerHeartbeat) []domain.Replica {
	out := make([]domain.Replica, 0, len(pool))
	for _, r := range pool {
		if s.isAlive(zoneID, r.ClusterID, r.InstanceID, heartbeats) {
			out = append(out, r)
		}
	}
	return out
}

func (s *CoordinatorService) UpsertHeartbeat(hb domain.WorkerHeartbeat) {
	hb.ObservedAt = time.Now().UTC()
	_ = s.store.UpsertHeartbeat(context.Background(), hb)
}

func (s *CoordinatorService) Heartbeats() []domain.WorkerHeartbeat {
	out, err := s.store.Heartbeats(context.Background())
	if err != nil {
		return nil
	}
	return out
}

func (s *CoordinatorService) IngestionInstances(zoneID string) []domain.IngestionInstance {
	workersByZone, err := s.store.ZoneWorkers(context.Background(), zoneID)
	if err != nil {
		return nil
	}
	out := make([]domain.IngestionInstance, 0)
	for zid, pool := range workersByZone {
		for _, r := range pool {
			out = append(out, domain.IngestionInstance{
				ZoneID:     zid,
				ClusterID:  r.ClusterID,
				InstanceID: r.InstanceID,
				URL:        r.URL,
			})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].ZoneID != out[j].ZoneID {
			return out[i].ZoneID < out[j].ZoneID
		}
		if out[i].ClusterID != out[j].ClusterID {
			return out[i].ClusterID < out[j].ClusterID
		}
		return out[i].InstanceID < out[j].InstanceID
	})
	return out
}

func workerKey(zoneID, clusterID, instanceID string) string {
	return zoneID + "|" + clusterID + "|" + instanceID
}

func (s *CoordinatorService) isAlive(zoneID, clusterID, instanceID string, heartbeats map[string]domain.WorkerHeartbeat) bool {
	if clusterID == "" || instanceID == "" {
		return false
	}
	hb, ok := heartbeats[workerKey(zoneID, clusterID, instanceID)]
	if !ok {
		return false
	}
	return time.Since(hb.ObservedAt) <= s.heartbeatTimeout
}

func pickLeastLoaded(zoneID string, alive []domain.Replica, order []domain.Replica, sim map[string]float64) domain.Replica {
	index := make(map[string]int, len(order))
	for i, r := range order {
		k := r.ClusterID + "|" + r.InstanceID
		if _, ok := index[k]; !ok {
			index[k] = i
		}
	}
	best := alive[0]
	bestK := workerKey(zoneID, best.ClusterID, best.InstanceID)
	bestScore := sim[bestK]
	bestIdx := index[best.ClusterID+"|"+best.InstanceID]

	for _, r := range alive[1:] {
		k := workerKey(zoneID, r.ClusterID, r.InstanceID)
		sc := sim[k]
		idx := index[r.ClusterID+"|"+r.InstanceID]
		if sc < bestScore || (sc == bestScore && idx < bestIdx) {
			best = r
			bestScore = sc
			bestIdx = idx
		}
	}
	return best
}
