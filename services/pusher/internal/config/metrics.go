package config

import "github.com/kelseyhightower/envconfig"

// MetricsConfig Prometheus.
type MetricsConfig struct {
	Enabled bool `envconfig:"ENABLED" default:"true"`

	Namespace string `envconfig:"NAMESPACE" default:"traffic"`
	Subsystem string `envconfig:"SUBSYSTEM" default:"pusher"`

	Prometheus PrometheusConfig
}

// PrometheusConfig endpoint метрик.
type PrometheusConfig struct {
	Port int    `envconfig:"PORT" default:"8082"`
	// Не использовать тег PATH: envconfig подхватывает системную переменную PATH.
	Path string `envconfig:"HTTP_PATH" default:"/metrics"`

	EnableGoCollector      bool `envconfig:"ENABLE_GO_COLLECTOR" default:"true"`
	EnableProcessCollector bool `envconfig:"ENABLE_PROCESS_COLLECTOR" default:"true"`
}

func loadMetricsConfig() (MetricsConfig, error) {
	var cfg MetricsConfig
	if err := envconfig.Process("METRICS", &cfg); err != nil {
		return MetricsConfig{}, err
	}
	if err := envconfig.Process("PROMETHEUS", &cfg.Prometheus); err != nil {
		return MetricsConfig{}, err
	}
	return cfg, nil
}
