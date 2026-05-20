package domain

// FrameIngestEvent кадр для pusher: метаданные и PNG в base64 (загрузка в S3 — в pusher).
type FrameIngestEvent struct {
	SegmentID         string `json:"segment_id"`
	CameraID          string `json:"camera_id"`
	ObservedAt        string `json:"observed_at"`
	PipelineStartedAt string `json:"pipeline_started_at"`
	S3Key             string `json:"s3_key"`
	ContentBase64     string `json:"content_base64"`
	ContentType       string `json:"content_type"`
}
