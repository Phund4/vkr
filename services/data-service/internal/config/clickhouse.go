package config

import (
	"net/url"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

type ClickhouseConfig struct {
	Host     string `envconfig:"HOST"`
	User     string `envconfig:"USER"`
	Password string `envconfig:"PASSWORD"`
	Database string `envconfig:"DATABASE"`
}

// loadClickhouseConfig загружает конфигурацию ClickHouse из переменных окружения
func loadClickhouseConfig() (ClickhouseConfig, error) {
	var cfg ClickhouseConfig
	if err := envconfig.Process("CLICKHOUSE", &cfg); err != nil {
		return ClickhouseConfig{}, err
	}
	return cfg, nil
}

func (c ClickhouseConfig) EffectiveDSN() string {
	host := strings.TrimSpace(c.Host)
	if strings.HasPrefix(host, "http://") || strings.HasPrefix(host, "https://") {
		parsedURL, err := url.Parse(host)
		if err == nil && parsedURL.Host != "" {
			host = parsedURL.Host
		}
	}

	dsnURL := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   "/" + c.Database,
		User:   url.UserPassword(c.User, c.Password),
	}

	return dsnURL.String()
}
