package config

import (
	"github.com/kelseyhightower/envconfig"
)

// MetricsConfig представляет конфигурацию для метрик
type MetricsConfig struct {
	// Enabled флаг включения/выключения сбора метрик
	Enabled bool `envconfig:"ENABLED" default:"true"`

	// Namespace пространство имен для метрик
	Namespace string `envconfig:"NAMESPACE" default:"tracing"`

	// Subsystem подсистема для метрик
	Subsystem string `envconfig:"SUBSYSTEM" default:"service"`

	// Prometheus конфигурация Prometheus
	Prometheus PrometheusConfig
}

// PrometheusConfig представляет конфигурацию для Prometheus
type PrometheusConfig struct {
	// Port порт для HTTP сервера Prometheus
	Port int `envconfig:"PORT" default:"8081"`

	// Path путь для эндпоинта метрик Prometheus
	Path string `envconfig:"PATH" default:"/metrics"`

	// EnableGoCollector флаг включения сбора метрик стандартной Go runtime
	EnableGoCollector bool `envconfig:"ENABLE_GO_COLLECTOR" default:"true"`

	// EnableProcessCollector флаг включения сбора метрик процесса
	EnableProcessCollector bool `envconfig:"ENABLE_PROCESS_COLLECTOR" default:"true"`
}

// loadMetricsConfig загружает конфигурацию метрик
func loadMetricsConfig() (MetricsConfig, error) {
	var cfg MetricsConfig
	if err := envconfig.Process("METRICS", &cfg); err != nil {
		return cfg, err
	}
	if err := envconfig.Process("PROMETHEUS", &cfg.Prometheus); err != nil {
		return cfg, err
	}
	return cfg, nil
}
