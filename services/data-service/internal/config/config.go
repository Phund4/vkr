package config

import (
	"errors"
	"fmt"
	"os"
)

var (
	ErrLoadMetricsConfig    = errors.New("failed to load metrics config")
	ErrMissingEnvVariable   = errors.New("missing environment variable")
	ErrLoadServerConfig     = errors.New("failed to load server config")
	errLoadClickhouseConfig = errors.New("failed to load clickhouse config")
)

const (
	EnvKey = "ENV"
)

// Config содержит всю конфигурацию приложения
type Config struct {
	Metrics    MetricsConfig
	Env        string
	Server     ServerConfig
	Clickhouse ClickhouseConfig
}

// LoadConfig загружает всю конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	env := os.Getenv(EnvKey)
	if env == "" {
		return nil, fmt.Errorf("%w: %s", ErrMissingEnvVariable, EnvKey)
	}

	serverCfg, err := loadServerConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadServerConfig, err)
	}

	metricsCfg, err := loadMetricsConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrLoadMetricsConfig, err)
	}

	clickhouseCfg, err := loadClickhouseConfig()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errLoadClickhouseConfig, err)
	}

	return &Config{
		Env:        env,
		Server:     serverCfg,
		Metrics:    metricsCfg,
		Clickhouse: clickhouseCfg,
	}, nil
}
