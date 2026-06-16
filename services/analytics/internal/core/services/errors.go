package services

import "errors"

var (
	// ErrIngestValidation в теле запроса нет обязательных полей segment_id, camera_id, observed_at.
	ErrIngestValidation = errors.New("segment_id, camera_id, observed_at required")

	// ErrPublishKafka не удалось записать сообщение в топик persist для pusher.
	ErrPublishKafka = errors.New("kafka publish")
)
