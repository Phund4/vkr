package config

import (
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// ClickhouseConfig native-протокол (host:port).
type ClickhouseConfig struct {
	Addr            string `envconfig:"ADDR"`
	Database        string `envconfig:"DATABASE" default:"default"`
	User            string `envconfig:"USER" default:"default"`
	Password        string `envconfig:"PASSWORD"`
	IncidentsTable  string `envconfig:"INCIDENTS_TABLE" default:"road_incidents"`
	CongestionTable string `envconfig:"CONGESTION_TABLE" default:"road_congestion"`
}

func loadClickhouseConfig() (ClickhouseConfig, error) {
	var cfg ClickhouseConfig
	if err := envconfig.Process("CLICKHOUSE", &cfg); err != nil {
		return ClickhouseConfig{}, err
	}
	cfg.Addr = strings.TrimSpace(cfg.Addr)
	return cfg, nil
}
