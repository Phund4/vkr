package dto

import "coordinator/internal/core/domain"

type SourceItem struct {
	SourceID  string `json:"source_id"`
	DataClass string `json:"data_class"`
	ZoneID    string `json:"zone_id"`
	SegmentID string `json:"segment_id,omitempty"`
	CameraID  string `json:"camera_id,omitempty"`
	RTSPURL   string `json:"rtsp_url,omitempty"`
	Enabled   bool   `json:"enabled"`
}

func SourceItemFromDomain(s domain.Source) SourceItem {
	return SourceItem{
		SourceID:  s.SourceID,
		DataClass: s.DataClass,
		ZoneID:    s.ZoneID,
		SegmentID: s.SegmentID,
		CameraID:  s.CameraID,
		RTSPURL:   s.RTSPURL,
		Enabled:   s.Enabled,
	}
}

func SourcesListFromDomain(items []domain.Source) []SourceItem {
	out := make([]SourceItem, 0, len(items))
	for _, s := range items {
		out = append(out, SourceItemFromDomain(s))
	}
	return out
}

type AssignmentsResponse struct {
	Items []SourceItem `json:"items"`
}

type SourcesResponse struct {
	Items []SourceItem `json:"items"`
}

type WorkerHeartbeatItem struct {
	ZoneID      string  `json:"zone_id"`
	ClusterID   string  `json:"cluster_id"`
	InstanceID  string  `json:"instance_id"`
	Load        float64 `json:"load"`
	ObservedAt  string  `json:"observed_at"`
	Assignments int     `json:"assignments"`
}

func WorkerHeartbeatItemFromDomain(h domain.WorkerHeartbeat) WorkerHeartbeatItem {
	return WorkerHeartbeatItem{
		ZoneID:      h.ZoneID,
		ClusterID:   h.ClusterID,
		InstanceID:  h.InstanceID,
		Load:        h.Load,
		ObservedAt:  h.ObservedAt.Format("2006-01-02T15:04:05.999999999Z07:00"),
		Assignments: h.Assignments,
	}
}

func WorkerHeartbeatsFromDomain(items []domain.WorkerHeartbeat) []WorkerHeartbeatItem {
	out := make([]WorkerHeartbeatItem, 0, len(items))
	for _, h := range items {
		out = append(out, WorkerHeartbeatItemFromDomain(h))
	}
	return out
}

type WorkersResponse struct {
	Items []WorkerHeartbeatItem `json:"items"`
}

type IngestionInstanceItem struct {
	ZoneID     string `json:"zone_id"`
	ClusterID  string `json:"cluster_id"`
	InstanceID string `json:"instance_id"`
	URL        string `json:"url,omitempty"`
}

func IngestionInstanceFromDomain(i domain.IngestionInstance) IngestionInstanceItem {
	return IngestionInstanceItem{
		ZoneID:     i.ZoneID,
		ClusterID:  i.ClusterID,
		InstanceID: i.InstanceID,
		URL:        i.URL,
	}
}

func IngestionInstancesFromDomain(items []domain.IngestionInstance) []IngestionInstanceItem {
	out := make([]IngestionInstanceItem, 0, len(items))
	for _, i := range items {
		out = append(out, IngestionInstanceFromDomain(i))
	}
	return out
}

type IngestionInstancesResponse struct {
	Items []IngestionInstanceItem `json:"items"`
}
