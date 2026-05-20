package domain

// Классификатор источника: тип данных и целевой пайплайн (например видео → router).
const (
	// DataClassRoadSegmentVideo RTSP с камеры дорожного сегмента; обрабатывается router (S3 + Kafka ML).
	DataClassRoadSegmentVideo = "road_segment_video"
)

// ValidDataClasses возвращает допустимые значения поля data_class в таблице sources.
func ValidDataClasses() []string {
	return []string{
		DataClassRoadSegmentVideo,
	}
}
