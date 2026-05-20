package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrMissingEnvVariable = errors.New("missing environment variable")
)

const (
	EnvKey = "ENV"
)

// Config полная конфигурация pusher.
type Config struct {
	Env        string
	Server     ServerConfig
	Metrics    MetricsConfig
	Kafka      KafkaConfig
	Clickhouse ClickhouseConfig
	S3         S3Config
}

// LoadConfig читает окружение процесса (в Docker задаёт compose).
func LoadConfig() (*Config, error) {
	env := os.Getenv(EnvKey)
	if env == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingEnvVariable, EnvKey)
	}

	serverCfg, err := loadServerConfig()
	if err != nil {
		return nil, fmt.Errorf("server config: %w", err)
	}

	metricsCfg, err := loadMetricsConfig()
	if err != nil {
		return nil, fmt.Errorf("metrics config: %w", err)
	}

	kafkaCfg, err := loadKafkaConfig()
	if err != nil {
		return nil, fmt.Errorf("kafka config: %w", err)
	}

	chCfg, err := loadClickhouseConfig()
	if err != nil {
		return nil, fmt.Errorf("clickhouse config: %w", err)
	}
	if chCfg.Addr == "" {
		return nil, fmt.Errorf("%w: CLICKHOUSE_ADDR", ErrMissingEnvVariable)
	}

	s3Cfg, err := loadS3Config()
	if err != nil {
		return nil, fmt.Errorf("s3 config: %w", err)
	}

	return &Config{
		Env:        env,
		Server:     serverCfg,
		Metrics:    metricsCfg,
		Kafka:      kafkaCfg,
		Clickhouse: chCfg,
		S3:         s3Cfg,
	}, nil
}
