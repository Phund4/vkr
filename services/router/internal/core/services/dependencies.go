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

// MLRunner отправляет один JPEG-кадр в два ML-эндпоинта (инцидент и загруженность).
type MLRunner interface {
	// PostBoth параллельные multipart POST; filename — имя поля файла (например frame.jpg).
	PostBoth(ctx context.Context, jpeg []byte, filename string, meta domain.ProcessMeta) error
}

// VideoMetaPublisher публикует JSON метаданных кадра в Kafka (топик видео-контура).
type VideoMetaPublisher interface {
	// Publish записывает сообщение с ключом partition key.
	Publish(ctx context.Context, key, value []byte) error
}
