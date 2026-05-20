package mlclient

const (
	// formFieldImage имя multipart-поля с бинарным изображением.
	formFieldImage = "image"
	// formFieldSegmentID метаданные сегмента.
	formFieldSegmentID = "segment_id"
	// formFieldCameraID метаданные камеры.
	formFieldCameraID = "camera_id"
	// formFieldS3Key ключ объекта в S3.
	formFieldS3Key = "s3_key"
	// formFieldObservedAt время наблюдения.
	formFieldObservedAt = "observed_at"
	// formFieldPipelineStartedAt начало конвейера в router.
	formFieldPipelineStartedAt = "pipeline_started_at"

	// headerContentType заголовок Content-Type запроса.
	headerContentType = "Content-Type"

	// httpErrorBodyMaxBytes максимум байт тела ответа при ошибке ML для сообщения об ошибке.
	httpErrorBodyMaxBytes = 4096

	// defaultAccidentPath путь POST для модели инцидентов.
	defaultAccidentPath = "/v1/process/accident"
	// defaultCongestionPath путь POST для модели загруженности.
	defaultCongestionPath = "/v1/process/congestion"
)
