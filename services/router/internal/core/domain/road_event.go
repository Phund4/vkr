package domain

// VideoIngestEvent сообщение в топик its.video.ingest: метаданные кадра без ML (инференс шлёт analytics по HTTP отдельно).
type VideoIngestEvent struct {
	// SegmentID логический сегмент дороги.
	SegmentID string `json:"segment_id"`
	// CameraID идентификатор камеры.
	CameraID string `json:"camera_id"`
	// ObservedAt время наблюдения кадра (RFC3339Nano).
	ObservedAt string `json:"observed_at"`
	// S3Key ключ PNG-объекта в бакете.
	S3Key string `json:"s3_key,omitempty"`
	// PipelineStartedAt момент старта конвейера в router для e2e-метрик.
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`
}
