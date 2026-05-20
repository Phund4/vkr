package domain

// VideoIngestEvent сообщение в топик its.video.ingest: метаданные кадра (без тела; ML — its.ml.*).
type VideoIngestEvent struct {
	SegmentID         string `json:"segment_id"`
	CameraID          string `json:"camera_id"`
	ObservedAt        string `json:"observed_at"`
	S3Key             string `json:"s3_key,omitempty"`
	PipelineStartedAt string `json:"pipeline_started_at,omitempty"`
}
