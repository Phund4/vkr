package clickhouse

// Шаблоны INSERT для native batch API (плейсхолдеры %s — database, table).
const (
	insertIncidentSQLTemplate = `INSERT INTO %s.%s (observed_at, segment_id, camera_id, s3_key, crash_probability, incident_label, raw_ml) VALUES (?, ?, ?, ?, ?, ?, ?)`
	insertCongestionSQLTemplate = `INSERT INTO %s.%s (observed_at, segment_id, camera_id, s3_key, congestion_score, raw_ml) VALUES (?, ?, ?, ?, ?, ?)`
)
