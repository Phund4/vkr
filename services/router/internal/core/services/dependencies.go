package services

import (
	"context"

	"router/internal/core/domain"
)

// S3Uploader загрузка объектов в S3.
type S3Uploader interface {
	PutPNG(ctx context.Context, key string, png []byte) error
}

// MLProcessor вызов ML по кадру.
type MLProcessor interface {
	PostProcess(ctx context.Context, jpeg []byte, filename string, meta domain.ProcessMeta) error
}
