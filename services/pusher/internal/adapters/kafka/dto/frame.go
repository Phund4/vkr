package dto

// FrameIngestEvent кадр от router: pusher кладёт PNG в S3 по s3_key.
type FrameIngestEvent struct {
	SegmentID         string `json:"segment_id"`
	CameraID          string `json:"camera_id"`
	ObservedAt        string `json:"observed_at"`
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`
	S3Key             string `json:"s3_key"`
	ContentBase64     string `json:"content_base64"`
	ContentType       string `json:"content_type,omitempty"`
}
