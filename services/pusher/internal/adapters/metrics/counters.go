package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Counters метрик pusher (default Prometheus registry).
var (
	KafkaConsumeErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pusher_kafka_consume_errors_total",
			Help: "Kafka consumer failures by stage.",
		},
		[]string{"stage"},
	)
	KafkaMessagesProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pusher_kafka_messages_processed_total",
			Help: "Persist messages successfully processed.",
		},
		[]string{"topic"},
	)
	ClickHouseErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pusher_clickhouse_errors_total",
			Help: "ClickHouse write failures by operation.",
		},
		[]string{"op"},
	)
	S3Errors = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pusher_s3_errors_total",
			Help: "S3 put failures.",
		},
	)
	ProcessErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pusher_process_errors_total",
			Help: "Message processing failures by stage.",
		},
		[]string{"stage"},
	)
	// PipelineE2ESeconds задержка от router (pipeline_started_at) до успешной записи в ClickHouse.
	PipelineE2ESeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "traffic_pipeline_e2e_seconds",
			Help:    "End-to-end latency from router pipeline start to ClickHouse insert (persist path only).",
			Buckets: []float64{0.1, 0.25, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
		},
	)
	BytesKafkaConsumed = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pusher_kafka_message_bytes_total",
			Help: "Total raw Kafka message payload bytes consumed (persist topic).",
		},
	)
	BytesClickHouseWritten = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pusher_clickhouse_payload_bytes_total",
			Help: "Approximate bytes of ML JSON sent to ClickHouse inserts.",
		},
	)
	ProcessDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pusher_process_duration_seconds",
			Help:    "Wall time to process one persist message (through CH writes if any).",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
	)
	FrameProcessDurationSeconds = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "pusher_frame_process_duration_seconds",
			Help:    "Wall time to process one frames.ingest message (S3 put only).",
			Buckets: []float64{0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
		},
	)
	S3BytesUploaded = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "pusher_s3_bytes_uploaded_total",
			Help: "Bytes written to S3 (frames topic and persist files).",
		},
	)
)
