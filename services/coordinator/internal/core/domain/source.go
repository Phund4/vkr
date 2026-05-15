package domain

import "time"

// Replica — инстанс router в пуле зоны (порядок в zone_workers = tie-break при равной загрузке).
type Replica struct {
	ClusterID  string
	InstanceID string
	URL        string
}

// IngestionInstance строка каталога инстансов по зоне.
type IngestionInstance struct {
	ZoneID     string
	ClusterID  string
	InstanceID string
	URL        string
}

// Source описывает один входной источник данных зоны.
type Source struct {
	SourceID  string
	DataClass string
	ZoneID    string
	SegmentID string
	CameraID  string
	RTSPURL   string
	Enabled   bool
}

// WorkerHeartbeat статус живого инстанса router.
type WorkerHeartbeat struct {
	ZoneID      string
	ClusterID   string
	InstanceID  string
	Load        float64
	ObservedAt  time.Time
	Assignments int
}
