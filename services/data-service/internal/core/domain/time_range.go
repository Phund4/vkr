package domain

import "time"

// RoadTimeRangeParams фильтр по времени и идентификаторам контура/камеры.
type RoadTimeRangeParams struct {
	SegmentID string
	CameraID  string
	From      time.Time
	To        time.Time
	Limit     int
}

// FrameRef кадр в S3, привязанный ко времени наблюдения.
type FrameRef struct {
	ObservedAt time.Time
	SegmentID  string
	CameraID   string
	S3Key      string
}

// CongestionAverageResult средняя загруженность за интервал.
type CongestionAverageResult struct {
	SegmentID       string
	From            time.Time
	To              time.Time
	AvgScore        float64
	SampleCount     uint64
}
