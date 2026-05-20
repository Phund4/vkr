package domain

import "encoding/json"

// RoadEvent JSON из its.ml.accident.out / its.ml.congestion.out (и опционально POST /v1/ingest для отладки).
type RoadEvent struct {
	// SegmentID логический сегмент дороги / линии.
	SegmentID string `json:"segment_id"`

	// CameraID идентификатор камеры.
	CameraID string `json:"camera_id"`

	// ObservedAt время события RFC3339.
	ObservedAt string `json:"observed_at"`

	// PipelineStartedAt RFC3339Nano — момент старта конвейера в router (e2e до БД).
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`

	// S3Key ключ кадра в S3 при видео-контуре.
	S3Key string `json:"s3_key,omitempty"`

	// ML сырой JSON ответа ML (инцидент/загруженность).
	ML json.RawMessage `json:"ml,omitempty"`
}

// IncidentBlock поддерево ml.incident.
type IncidentBlock struct {
	// HasIncident явный флаг наличия аварии (источник истины при наличии поля).
	HasIncident *bool `json:"has_incident,omitempty"`

	// CrashProbability оценка вероятности ДТП [0, 1].
	CrashProbability float64 `json:"crash_probability"`

	// Label класс события (например crash).
	Label string `json:"label"`
}

// CongestionBlock поддерево ml.congestion.
type CongestionBlock struct {
	// CongestionScore степень загруженности [0, 1].
	CongestionScore float64 `json:"congestion_score"`
}

// MLParsed разбор поля ml для метрик и публикации в Kafka (pusher).
type MLParsed struct {
	// Incident блок инцидента.
	Incident IncidentBlock `json:"incident"`

	// Congestion блок загруженности.
	Congestion CongestionBlock `json:"congestion"`
}
