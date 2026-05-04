package domain

// Классификатор источника: что приходит и куда ведёт пайплайн router.
const (
	// DataClassRoadSegmentVideo — RTSP с камер дорожного участка → S3 + ML.
	DataClassRoadSegmentVideo = "road_segment_video"
)

// ValidDataClasses допустимые значения data_class в sources.yaml.
func ValidDataClasses() []string {
	return []string{
		DataClassRoadSegmentVideo,
	}
}
