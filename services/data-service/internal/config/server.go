package config

import "github.com/kelseyhightower/envconfig"

type ServerConfig struct {
	HTTPServerPort string `envconfig:"PORT"`
}

// loadServerConfig загружает конфигурацию сервера из переменных окружения
func loadServerConfig() (ServerConfig, error) {
	var cfg ServerConfig
	if err := envconfig.Process("SERVER", &cfg); err != nil {
		return ServerConfig{}, err
	}
	return cfg, nil
}
