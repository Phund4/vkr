// Package metrics — Prometheus для router.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OperationErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_operation_errors_total",
			Help: "Errors during video ingest pipeline by stage/type.",
		},
		[]string{"stage"},
	)

	// FramesProcessed исход обработки кадра (после попытки S3+ML).
	FramesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_frames_processed_total",
			Help: "Frames processed by outcome (ml_ok / ml_err; s3 may have failed earlier).",
		},
		[]string{"outcome"},
	)

	// FrameHandleSeconds полное время handleFrame (JPEG→PNG→S3→ML).
	FrameHandleSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "router_frame_handle_duration_seconds",
			Help:    "Wall time for one frame pipeline inside router.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		},
	)

	// BytesUploadedS3 успешные PutPNG (после успешного Put).
	BytesUploadedS3 = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "router_bytes_uploaded_s3_total",
			Help: "Total bytes written to S3 (PNG objects).",
		},
	)

	// BytesSentML размер JPEG в multipart к ML.
	BytesSentML = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "router_bytes_sent_ml_total",
			Help: "Total JPEG bytes sent to ML service.",
		},
	)

	// MLLatencySeconds время HTTP вызова ML.
	MLLatencySeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "router_ml_request_duration_seconds",
			Help:    "Wall time for parallel POSTs to ML accident + congestion endpoints.",
			Buckets: []float64{0.02, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		},
	)

	KafkaVideoPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_kafka_video_publish_errors_total",
			Help: "Failures publishing frame metadata to its.video.ingest.",
		},
		[]string{"stage"},
	)
)
