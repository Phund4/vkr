package domain

import "time"

// RoadIncident строка таблицы road_incidents (события после ML / analytics).
type RoadIncident struct {
	ObservedAt       time.Time
	SegmentID        string
	CameraID         string
	S3Key            string
	CrashProbability float64
	IncidentLabel    string
	RawML            string
}

// RoadCongestion строка таблицы road_congestion.
type RoadCongestion struct {
	ObservedAt      time.Time
	SegmentID       string
	CameraID        string
	S3Key           string
	CongestionScore float64
	RawML           string
}

// RoadListParams фильтры списков (сегмент/камера опциональны).
type RoadListParams struct {
	SegmentID string
	CameraID  string
	Limit     int
}
