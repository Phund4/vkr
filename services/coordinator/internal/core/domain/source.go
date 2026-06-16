package domain

import "time"

// Replica описывает один инстанс router в пуле зоны (порядок в zone_workers — tie-break при равной нагрузке).
type Replica struct {
	// ClusterID идентификатор кластера внутри зоны.
	ClusterID string
	// InstanceID уникальный идентификатор инстанса router.
	InstanceID string
	// URL базовый HTTP URL инстанса для проксирования или проверок.
	URL string
}

// IngestionInstance строка каталога зарегистрированных инстансов приёма (ingestion) по зоне.
type IngestionInstance struct {
	// ZoneID логическая зона.
	ZoneID string
	// ClusterID кластер.
	ClusterID string
	// InstanceID инстанс.
	InstanceID string
	// URL адрес HTTP API инстанса.
	URL string
}

// Source описывает один входной источник данных (камера, поток), привязанный к зоне.
type Source struct {
	// SourceID суррогатный ключ источника в БД.
	SourceID string
	// DataClass тип источника (см. ValidDataClasses).
	DataClass string
	// ZoneID зона размещения.
	ZoneID string
	// SegmentID дорожный сегмент.
	SegmentID string
	// CameraID идентификатор камеры.
	CameraID string
	// RTSPURL URL RTSP-потока.
	RTSPURL string
	// Enabled участвует ли источник в назначениях.
	Enabled bool
}

// WorkerHeartbeat снимок состояния живого воркера router для алгоритма балансировки.
type WorkerHeartbeat struct {
	// ZoneID зона.
	ZoneID string
	// ClusterID кластер.
	ClusterID string
	// InstanceID инстанс.
	InstanceID string
	// Load условная нагрузка [0, 1] или иная метрика.
	Load float64
	// ObservedAt время фиксации heartbeat.
	ObservedAt time.Time
	// Assignments число активных назначений на инстансе.
	Assignments int
}
