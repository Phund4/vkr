package domain

// MLFrameMessage кадр и метаданные для топиков its.ml.*.in (router → ml-serving).
type MLFrameMessage struct {
	SegmentID         string `json:"segment_id"`
	CameraID          string `json:"camera_id"`
	ObservedAt        string `json:"observed_at"`
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`
	S3Key             string `json:"s3_key,omitempty"`
	JPEGBase64        string `json:"jpeg_base64"`
}
