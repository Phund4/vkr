package domain

// PersistEvent нормализованное сообщение в топик Kafka для сервиса pusher (схема должна совпадать с pusher dto).
type PersistEvent struct {
	// SegmentID логический сегмент дороги.
	SegmentID string `json:"segment_id"`
	// CameraID идентификатор камеры.
	CameraID string `json:"camera_id"`

	// ObservedAt время события (RFC3339 или RFC3339Nano).
	ObservedAt string `json:"observed_at"`
	// PipelineStartedAt метка начала конвейера в router для e2e-метрик до ClickHouse.
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`
	// S3Key ключ кадра в объектном хранилище.
	S3Key string `json:"s3_key,omitempty"`
	// RawML сериализованный JSON блока ml для вставки в ClickHouse.
	RawML string `json:"raw_ml,omitempty"`
	// HasML признак наличия осмысленных ML-данных.
	HasML bool `json:"has_ml"`

	// PersistIncident писать ли строку в таблицу инцидентов.
	PersistIncident bool `json:"persist_incident"`
	// PersistCongestion писать ли строку в таблицу загруженности.
	PersistCongestion bool `json:"persist_congestion"`

	// CrashProbability вероятность ДТП из ML [0, 1].
	CrashProbability float64 `json:"crash_probability"`
	// IncidentLabel текстовый класс инцидента из ML.
	IncidentLabel string `json:"incident_label"`
	// CongestionScore степень загруженности [0, 1].
	CongestionScore float64 `json:"congestion_score"`

	// Files опциональные вложения для загрузки в S3 через pusher.
	Files []PersistFile `json:"files,omitempty"`
}

// PersistFile описание одного файла, передаваемого в pusher в base64.
type PersistFile struct {
	// Key целевой ключ объекта в S3.
	Key string `json:"key"`
	// ContentBase64 содержимое файла в Base64.
	ContentBase64 string `json:"content_base64"`
	// ContentType MIME-тип (например image/png).
	ContentType string `json:"content_type,omitempty"`
}
