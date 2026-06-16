package config

import "github.com/kelseyhightower/envconfig"

// ServerConfig HTTP (probe).
type ServerConfig struct {
	Port string `envconfig:"PORT" default:":8094"`
}

func loadServerConfig() (ServerConfig, error) {
	var cfg ServerConfig
	if err := envconfig.Process("SERVER", &cfg); err != nil {
		return ServerConfig{}, err
	}
	return cfg, nil
}
