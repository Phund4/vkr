package services

import "time"

const (
	// clickhouseOpTimeout таймаут контекста для запросов ClickHouse и S3 put в ProcessMessage.
	clickhouseOpTimeout = 10 * time.Second
)
