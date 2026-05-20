// Package metrics — Prometheus для analytics (Help на английском).
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CongestionScore = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_road_congestion_score",
			Help: "Last congestion score from ML per road segment and camera.",
		},
		[]string{"segment_id", "camera_id"},
	)
	CrashProbability = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_road_crash_probability",
			Help: "Last crash probability from ML per road segment and camera.",
		},
		[]string{"segment_id", "camera_id"},
	)
	CrashAlert = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "analytics_road_crash_alert",
			Help: "1 if crash_probability >= CRASH_ALERT_THRESHOLD, else 0.",
		},
		[]string{"segment_id", "camera_id"},
	)
	IncidentsRecorded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_road_incidents_recorded_total",
			Help: "Persist payloads with incident flag published to Kafka.",
		},
		[]string{"segment_id", "camera_id"},
	)
	CongestionRecorded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_road_congestion_recorded_total",
			Help: "Persist payloads with congestion flag published to Kafka.",
		},
		[]string{"segment_id", "camera_id"},
	)
	KafkaPublishErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_kafka_publish_errors_total",
			Help: "Kafka persist topic write failures by stage.",
		},
		[]string{"stage"},
	)
	ProcessErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_process_errors_total",
			Help: "Event processing errors by stage (Kafka ML results, video meta, optional debug handler).",
		},
		[]string{"stage"},
	)
	KafkaIngestProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_kafka_ingest_processed_total",
			Help: "Ingest payloads successfully processed from Kafka by topic.",
		},
		[]string{"topic"},
	)
	KafkaConsumeErrors = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "analytics_kafka_consume_errors_total",
			Help: "Kafka consumer failures by stage.",
		},
		[]string{"stage"},
	)
	KafkaPublishDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "analytics_kafka_publish_duration_seconds",
			Help:    "Time to publish one persist payload to Kafka.",
			Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2},
		},
	)
)
