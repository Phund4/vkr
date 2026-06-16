// Package dto — JSON сообщений топика persist (analytics → pusher).
package dto

// PersistEvent нормализованное событие: флаги записи в ClickHouse, метрики ML и опциональные файлы для S3.
type PersistEvent struct {
	SegmentID string `json:"segment_id"`
	CameraID  string `json:"camera_id"`

	ObservedAt string `json:"observed_at"`

	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`

	S3Key string `json:"s3_key,omitempty"`
	RawML string `json:"raw_ml,omitempty"`
	HasML bool   `json:"has_ml"`

	PersistIncident   bool `json:"persist_incident"`
	PersistCongestion bool `json:"persist_congestion"`

	CrashProbability float64 `json:"crash_probability"`
	IncidentLabel    string  `json:"incident_label"`
	CongestionScore  float64 `json:"congestion_score"`

	Files []PersistFile `json:"files,omitempty"`
}

// PersistFile вложение: ключ в S3 и полезная нагрузка в Base64.
type PersistFile struct {
	Key           string `json:"key"`
	ContentBase64 string `json:"content_base64"`
	ContentType   string `json:"content_type,omitempty"`
}
