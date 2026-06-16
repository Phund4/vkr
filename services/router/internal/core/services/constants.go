package services

import "time"

const (
	// frameLogEveryN логировать прогресс каждые N обработанных кадров.
	frameLogEveryN uint64 = 30
	// reconnectBackoffSec пауза перед переподключением к RTSP после обрыва.
	reconnectBackoffSec = 1
	// sourceWaitLogIntervalSec минимальный интервал между предупреждениями о проблемах источника/ML.
	sourceWaitLogIntervalSec = 45

	// frameKeyDateLayout формат даты в ключе объекта S3 (YYYY-MM-DD).
	frameKeyDateLayout = "2006-01-02"
	// frameObjectSuffix расширение объекта кадра (ключ для pusher → S3).
	frameObjectSuffix = ".png"
	// framePNGContentType MIME для FrameIngestEvent.
	framePNGContentType = "image/png"
	// frameJPEGObjectName логическое имя JPEG-кадра в конвейере.
	frameJPEGObjectName = "frame.jpg"

	// minFrameChanCapacity минимальная ёмкость канала кадров между читателем и воркерами.
	minFrameChanCapacity = 4
	// frameChanCapacityMultiplier множитель ёмкости канала от числа воркеров (workers * mult).
	frameChanCapacityMultiplier = 2

	// MetricStageFfmpegStart метка ошибки: старт ffmpeg.
	MetricStageFfmpegStart = "ffmpeg_start"
	// MetricStageFrameRead метка ошибки: чтение кадра из потока.
	MetricStageFrameRead = "frame_read"
	// MetricStageJpegPng метка ошибки: JPEG → PNG.
	MetricStageJpegPng = "jpeg_png"
	// MetricStageKafkaFramesPublish метка ошибки: публикация кадра в its.frames.ingest.
	MetricStageKafkaFramesPublish = "kafka_frames_publish"
	// MetricStageKafkaMLPublish метка ошибки: публикация в Kafka ML in.
	MetricStageKafkaMLPublish = "kafka_ml_publish"

	// MetricFrameOutcomeKafkaMLOk исход кадра: публикация в оба топика ML успешна.
	MetricFrameOutcomeKafkaMLOk = "kafka_ml_ok"
	// MetricFrameOutcomeKafkaMLError исход кадра: сбой публикации в Kafka ML.
	MetricFrameOutcomeKafkaMLError = "kafka_ml_error"

	// MetricKafkaPublishStageJSON ошибка сериализации JSON для Kafka.
	MetricKafkaPublishStageJSON = "json"
	// MetricKafkaPublishStageWrite ошибка записи в Kafka.
	MetricKafkaPublishStageWrite = "write"
)

// reconnectBackoffDuration возвращает длительность паузы переподключения.
func reconnectBackoffDuration() time.Duration {
	return reconnectBackoffSec * time.Second
}
