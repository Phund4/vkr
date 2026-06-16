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

	// FramesProcessed исход обработки кадра (Kafka ML publish).
	FramesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_frames_processed_total",
			Help: "Frames processed by outcome (kafka_ml_ok / kafka_ml_error).",
		},
		[]string{"outcome"},
	)

	// FrameHandleSeconds полное время handleFrame (JPEG→PNG→Kafka).
	FrameHandleSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "router_frame_handle_duration_seconds",
			Help:    "Wall time for one frame pipeline inside router.",
			Buckets: []float64{0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60},
		},
	)

	// KafkaFrameBytes суммарный размер PNG, опубликованный в its.frames.ingest.
	KafkaFrameBytes = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "router_kafka_frame_bytes_total",
			Help: "Total PNG bytes published to Kafka frames ingest topic (decoded payload size).",
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

	KafkaFramesPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "router_kafka_frames_publish_errors_total",
			Help: "Failures publishing frame to its.frames.ingest.",
		},
		[]string{"stage"},
	)
)
