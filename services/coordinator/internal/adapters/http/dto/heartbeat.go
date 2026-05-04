package dto

import (
	"strings"
	"time"

	"coordinator/internal/core/domain"
)

// WorkerHeartbeatIn тело POST /v1/workers/heartbeat.
type WorkerHeartbeatIn struct {
	ZoneID      string    `json:"zone_id"`
	ClusterID   string    `json:"cluster_id"`
	InstanceID  string    `json:"instance_id"`
	Load        float64   `json:"load"`
	ObservedAt  time.Time `json:"observed_at"`
	Assignments int       `json:"assignments"`
}

func (in *WorkerHeartbeatIn) ToDomain() domain.WorkerHeartbeat {
	return domain.WorkerHeartbeat{
		ZoneID:      in.ZoneID,
		ClusterID:   in.ClusterID,
		InstanceID:  in.InstanceID,
		Load:        in.Load,
		ObservedAt:  in.ObservedAt,
		Assignments: in.Assignments,
	}
}

func (in *WorkerHeartbeatIn) Validate() bool {
	return strings.TrimSpace(in.ZoneID) != "" && strings.TrimSpace(in.ClusterID) != "" && strings.TrimSpace(in.InstanceID) != ""
}
