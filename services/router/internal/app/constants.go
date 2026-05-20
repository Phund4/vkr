package app

import "time"

const (
	// coordinatorHTTPTimeout таймаут HTTP-запросов к coordinator.
	coordinatorHTTPTimeout = 10 * time.Second
	// heartbeatTickerInterval период фоновых heartbeat.
	heartbeatTickerInterval = 10 * time.Second

	// defaultKafkaTopicVideo топик метаданных кадра, если KAFKA_TOPIC_VIDEO не задан.
	defaultKafkaTopicVideo = "its.video.ingest"
	// defaultKafkaTopicMLAccidentIn вход accident-модели.
	defaultKafkaTopicMLAccidentIn = "its.ml.accident.in"
	// defaultKafkaTopicMLCongestionIn вход congestion-модели.
	defaultKafkaTopicMLCongestionIn = "its.ml.congestion.in"
)
