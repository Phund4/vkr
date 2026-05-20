// Package config читает настройки analytics из переменных окружения.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config заполняется из окружения функцией Load.
type Config struct {
	// ListenAddr адрес HTTP (ingest, metrics, health).
	ListenAddr string

	// CrashAlertThreshold порог crash_probability для алерта и записи инцидента.
	CrashAlertThreshold float64

	// CongestionPersistInterval минимальный интервал записи congestion на пару (segment, camera).
	CongestionPersistInterval time.Duration

	// KafkaBootstrap серверы брокера (через запятую); пусто — консьюмер Kafka не запускается.
	KafkaBootstrap string

	// KafkaConsumerGroup группа для чтения топиков ingest.
	KafkaConsumerGroup string

	// KafkaTopicVideo метаданные кадра (router → analytics), без ожидания ML.
	KafkaTopicVideo string

	// KafkaTopicMLAccidentOut результаты accident-модели (ml-serving → analytics).
	KafkaTopicMLAccidentOut string

	// KafkaTopicMLCongestionOut результаты congestion-модели (отдельное событие).
	KafkaTopicMLCongestionOut string

	// KafkaTopicPersist топик нормализованных событий для pusher (ClickHouse/S3).
	KafkaTopicPersist string
}

// Load читает переменные окружения и возвращает Config с дефолтами.
func Load() Config {
	c := Config{
		ListenAddr:                ":8093",
		CrashAlertThreshold:       0.8,
		CongestionPersistInterval: 2 * time.Second,
	}
	if v := strings.TrimSpace(os.Getenv("LISTEN_ADDR")); v != "" {
		c.ListenAddr = v
	}
	if v := strings.TrimSpace(os.Getenv("CRASH_ALERT_THRESHOLD")); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			c.CrashAlertThreshold = f
		}
	}
	if v := strings.TrimSpace(os.Getenv("CONGESTION_PERSIST_INTERVAL_SEC")); v != "" {
		if sec, err := strconv.ParseFloat(v, 64); err == nil && sec >= 0 {
			c.CongestionPersistInterval = time.Duration(sec * float64(time.Second))
		}
	}
	c.KafkaBootstrap = strings.TrimSpace(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	c.KafkaConsumerGroup = strings.TrimSpace(os.Getenv("KAFKA_CONSUMER_GROUP"))
	if c.KafkaConsumerGroup == "" {
		c.KafkaConsumerGroup = "analytics-ingest"
	}
	c.KafkaTopicVideo = strings.TrimSpace(os.Getenv("KAFKA_TOPIC_VIDEO"))
	if c.KafkaTopicVideo == "" {
		c.KafkaTopicVideo = "its.video.ingest"
	}
	c.KafkaTopicPersist = strings.TrimSpace(os.Getenv("KAFKA_TOPIC_PERSIST"))
	if c.KafkaTopicPersist == "" {
		c.KafkaTopicPersist = "its.persist.events"
	}
	c.KafkaTopicMLAccidentOut = strings.TrimSpace(os.Getenv("KAFKA_TOPIC_ML_ACCIDENT_OUT"))
	if c.KafkaTopicMLAccidentOut == "" {
		c.KafkaTopicMLAccidentOut = "its.ml.accident.out"
	}
	c.KafkaTopicMLCongestionOut = strings.TrimSpace(os.Getenv("KAFKA_TOPIC_ML_CONGESTION_OUT"))
	if c.KafkaTopicMLCongestionOut == "" {
		c.KafkaTopicMLCongestionOut = "its.ml.congestion.out"
	}
	return c
}
