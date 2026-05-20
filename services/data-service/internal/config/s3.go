package config

import (
	"fmt"
	"os"
	"strings"
)

// S3Config параметры MinIO/S3 для presigned URL кадров.
type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
	Enabled   bool
}

func loadS3Config() (S3Config, error) {
	endpoint := strings.TrimSpace(os.Getenv("S3_ENDPOINT"))
	bucket := strings.TrimSpace(os.Getenv("S3_BUCKET"))
	ak := strings.TrimSpace(os.Getenv("AWS_ACCESS_KEY_ID"))
	sk := strings.TrimSpace(os.Getenv("AWS_SECRET_ACCESS_KEY"))
	region := strings.TrimSpace(os.Getenv("S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	enabled := endpoint != "" && bucket != "" && ak != "" && sk != ""
	if endpoint != "" && bucket == "" {
		return S3Config{}, fmt.Errorf("%w: S3_BUCKET required when S3_ENDPOINT is set", ErrMissingEnvVariable)
	}
	return S3Config{
		Endpoint:  endpoint,
		Region:    region,
		Bucket:    bucket,
		AccessKey: ak,
		SecretKey: sk,
		Enabled:   enabled,
	}, nil
}
