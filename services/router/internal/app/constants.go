package app

import "time"

const (
	// coordinatorHTTPTimeout таймаут HTTP-запросов к coordinator.
	coordinatorHTTPTimeout = 10 * time.Second
	// heartbeatTickerInterval период фоновых heartbeat.
	heartbeatTickerInterval = 10 * time.Second
	// assignmentPollInterval повторный запрос назначений, если при старте coordinator был недоступен.
	assignmentPollInterval = 15 * time.Second

	// defaultKafkaTopicVideo топик метаданных кадра, если KAFKA_TOPIC_VIDEO не задан.
	defaultKafkaTopicVideo = "its.video.ingest"
)
