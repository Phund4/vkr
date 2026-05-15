package config

import (
	"fmt"
	"os"
	"strings"
)

type Root struct {
	ListenAddr          string
	HeartbeatTimeoutSec int
	DatabaseURL         string
}

// LoadFromEnv загружает конфиг из переменных окружения.
func LoadFromEnv() (*Root, error) {
	if err := tryLoadDotEnv(); err != nil {
		return nil, err
	}
	listen := strings.TrimSpace(os.Getenv("LISTEN_ADDR"))
	if listen == "" {
		listen = ":8098"
	}
	heartbeatTimeoutSec := 30
	if v := strings.TrimSpace(os.Getenv("HEARTBEAT_TIMEOUT_SEC")); v != "" {
		var parsed int
		if _, err := fmt.Sscanf(v, "%d", &parsed); err == nil && parsed > 0 {
			heartbeatTimeoutSec = parsed
		}
	}
	databaseURL := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &Root{
		ListenAddr:          listen,
		HeartbeatTimeoutSec: heartbeatTimeoutSec,
		DatabaseURL:         databaseURL,
	}, nil
}
