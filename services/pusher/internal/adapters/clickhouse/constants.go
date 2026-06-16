package clickhouse

import "time"

const (
	// defaultDialTimeout таймаут установки соединения с ClickHouse.
	defaultDialTimeout = 5 * time.Second

	// DefaultIncidentsTable имя таблицы инцидентов, если не задано в конфиге.
	DefaultIncidentsTable = "road_incidents"
	// DefaultCongestionTable имя таблицы загруженности, если не задано в конфиге.
	DefaultCongestionTable = "road_congestion"
)
