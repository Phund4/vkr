package config

import "github.com/kelseyhightower/envconfig"

// KafkaConfig consumer (KAFKA_BOOTSTRAP_SERVERS, KAFKA_CONSUMER_GROUP, topics).
type KafkaConfig struct {
	Bootstrap     string `envconfig:"BOOTSTRAP_SERVERS"`
	ConsumerGroup string `envconfig:"CONSUMER_GROUP" default:"pusher"`
	TopicFrames   string `envconfig:"TOPIC_FRAMES" default:"its.frames.ingest"`
	TopicPersist  string `envconfig:"TOPIC_PERSIST" default:"its.persist.events"`
}

func loadKafkaConfig() (KafkaConfig, error) {
	var cfg KafkaConfig
	if err := envconfig.Process("KAFKA", &cfg); err != nil {
		return KafkaConfig{}, err
	}
	return cfg, nil
}
