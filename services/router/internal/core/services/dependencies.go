package services

import (
	"context"

	"router/internal/core/domain"
)

// FramePublisher публикует кадр в новый топик для pusher (its.frames.ingest).
type FramePublisher interface {
	PublishFrame(ctx context.Context, ev domain.FrameIngestEvent) error
}

// MLFramePublisher отправляет кадр в Kafka-топики its.ml.accident.in и its.ml.congestion.in.
type MLFramePublisher interface {
	PublishBoth(ctx context.Context, jpeg []byte, meta domain.ProcessMeta) error
}

// VideoMetaPublisher публикует JSON метаданных кадра в Kafka (топик its.video.ingest).
type VideoMetaPublisher interface {
	Publish(ctx context.Context, key, value []byte) error
}
