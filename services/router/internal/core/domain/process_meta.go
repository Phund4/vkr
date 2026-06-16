package domain

// ProcessMeta поля multipart-формы при POST к ML-сервису.
type ProcessMeta struct {
	// SegmentID сегмент дороги.
	SegmentID string

	// CameraID идентификатор камеры.
	CameraID string

	// S3Key ключ объекта кадра в объектном хранилище.
	S3Key string

	// ObservedAt время кадра (обычно RFC3339Nano), совпадает с началом конвейера.
	ObservedAt string

	// PipelineStartedAt RFC3339Nano — начало конвейера в router (e2e до БД).
	PipelineStartedAt string
}
