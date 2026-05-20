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
	// frameObjectSuffix расширение файла кадра в S3.
	frameObjectSuffix = ".png"
	// frameJPEGUploadName имя файла в multipart к ML.
	frameJPEGUploadName = "frame.jpg"

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
	// MetricStageS3Put метка ошибки: загрузка в S3.
	MetricStageS3Put = "s3_put"
	// MetricStageMLProcess метка ошибки: вызов ML.
	MetricStageMLProcess = "ml_process"

	// MetricFrameOutcomeMLOk исход кадра: оба ML-вызова успешны.
	MetricFrameOutcomeMLOk = "ml_ok"
	// MetricFrameOutcomeMLError исход кадра: сбой хотя бы одного ML.
	MetricFrameOutcomeMLError = "ml_error"

	// MetricKafkaPublishStageJSON ошибка сериализации JSON для Kafka.
	MetricKafkaPublishStageJSON = "json"
	// MetricKafkaPublishStageWrite ошибка записи в Kafka.
	MetricKafkaPublishStageWrite = "write"
)

// reconnectBackoffDuration возвращает длительность паузы переподключения.
func reconnectBackoffDuration() time.Duration {
	return reconnectBackoffSec * time.Second
}
