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

	// FramesProcessed исход обработки кадра (после попытки S3+Kafka ML).
	FramesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_frames_processed_total",
			Help: "Frames processed by outcome (kafka_ml_ok / kafka_ml_error; s3 may have failed earlier).",
		},
		[]string{"outcome"},
	)

	// FrameHandleSeconds полное время handleFrame (JPEG→PNG→S3→Kafka ML).
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

	// KafkaMLFrameBytes суммарный размер JPEG, опубликованный в its.ml.*.in (×2: accident + congestion).
	KafkaMLFrameBytes = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "router_kafka_ml_frame_bytes_total",
			Help: "Total JPEG bytes published to Kafka ML input topics (accident + congestion per frame).",
		},
	)

	// KafkaMLPublishDurationSeconds время публикации в оба топика its.ml.*.in.
	KafkaMLPublishDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "router_kafka_ml_publish_duration_seconds",
			Help:    "Wall time for parallel publish to its.ml.accident.in and its.ml.congestion.in.",
			Buckets: []float64{0.02, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30},
		},
	)

	KafkaMLPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_kafka_ml_publish_errors_total",
			Help: "Failures publishing frame to Kafka ML input topics.",
		},
		[]string{"stage"},
	)

	KafkaVideoPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_kafka_video_publish_errors_total",
			Help: "Failures publishing frame metadata to its.video.ingest.",
		},
		[]string{"stage"},
	)
)
