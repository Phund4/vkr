package config

import (
	"os"
	"strings"

	"github.com/kelseyhightower/envconfig"
)

// S3Config MinIO / S3 (S3_*, плюс AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY как у router).
type S3Config struct {
	Endpoint  string `envconfig:"ENDPOINT"`
	Region    string `envconfig:"REGION" default:"us-east-1"`
	Bucket    string `envconfig:"BUCKET"`
	AccessKey string `envconfig:"ACCESS_KEY"`
	SecretKey string `envconfig:"SECRET_KEY"`
}

func loadS3Config() (S3Config, error) {
	var cfg S3Config
	if err := envconfig.Process("S3", &cfg); err != nil {
		return S3Config{}, err
	}
	if cfg.AccessKey == "" {
		cfg.AccessKey = strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	}
	if cfg.SecretKey == "" {
		cfg.SecretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	if r := strings.TrimSpace(os.Getenv("AWS_REGION")); r != "" {
		cfg.Region = r
	}
	return cfg, nil
}
