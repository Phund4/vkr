package config

import (
	"os"
	"strings"
)

// ApplyEnvOverrides перезаписывает поля конфига значениями из переменных окружения.
func ApplyEnvOverrides(c *Root) {
	if v := os.Getenv("ML_BASE_URL"); v != "" {
		c.ML.BaseURL = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("ML_PROCESS_PATH"); v != "" {
		c.ML.ProcessPath = v
	}
	if v := os.Getenv("S3_ENDPOINT"); v != "" {
		c.S3.Endpoint = strings.TrimRight(v, "/")
	}
	if v := os.Getenv("METRICS_LISTEN_ADDR"); v != "" {
		c.Metrics.ListenAddr = v
	}
}

// CoordinatorBaseURLFromEnv — адрес coordinator API (обязателен).
func CoordinatorBaseURLFromEnv() string {
	return strings.TrimRight(strings.TrimSpace(os.Getenv("COORDINATOR_BASE_URL")), "/")
}

func CoordinatorZoneIDFromEnv() string {
	return strings.TrimSpace(os.Getenv("COORDINATOR_ZONE_ID"))
}

func CoordinatorClusterIDFromEnv() string {
	return strings.TrimSpace(os.Getenv("COORDINATOR_CLUSTER_ID"))
}

func CoordinatorInstanceIDFromEnv() string {
	return strings.TrimSpace(os.Getenv("COORDINATOR_INSTANCE_ID"))
}
