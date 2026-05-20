package services

import (
	"context"

	"router/internal/core/domain"
)

// S3Uploader сохраняет PNG-кадры в объектном хранилище.
type S3Uploader interface {
	// PutPNG загружает объект с ключом key и телом png.
	PutPNG(ctx context.Context, key string, png []byte) error
}

// MLFramePublisher отправляет кадр в Kafka-топики its.ml.accident.in и its.ml.congestion.in.
type MLFramePublisher interface {
	PublishBoth(ctx context.Context, jpeg []byte, meta domain.ProcessMeta) error
}

// VideoMetaPublisher публикует JSON метаданных кадра в Kafka (топик видео-контура).
type VideoMetaPublisher interface {
	// Publish записывает сообщение с ключом partition key.
	Publish(ctx context.Context, key, value []byte) error
}
