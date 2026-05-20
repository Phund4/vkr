package domain

import "time"

// RoadIncident одна запись таблицы road_incidents (событие после ML/analytics).
type RoadIncident struct {
	// ObservedAt время события в ClickHouse.
	ObservedAt time.Time
	// SegmentID сегмент дороги.
	SegmentID string
	// CameraID камера.
	CameraID string
	// S3Key ключ кадра в объектном хранилище.
	S3Key string
	// CrashProbability оценка вероятности ДТП.
	CrashProbability float64
	// IncidentLabel класс/метка инцидента.
	IncidentLabel string
	// RawML исходный JSON ML для аудита.
	RawML string
}

// RoadCongestion одна запись таблицы road_congestion.
type RoadCongestion struct {
	// ObservedAt время измерения.
	ObservedAt time.Time
	// SegmentID сегмент.
	SegmentID string
	// CameraID камера.
	CameraID string
	// S3Key ключ кадра при наличии.
	S3Key string
	// CongestionScore степень загруженности.
	CongestionScore float64
	// RawML сырой JSON ML.
	RawML string
}

// RoadListParams фильтры выборки списков дорожных событий для API аналитиков.
type RoadListParams struct {
	// SegmentID фильтр по сегменту (пусто — все).
	SegmentID string
	// CameraID фильтр по камере.
	CameraID string
	// From нижняя граница observed_at (UTC).
	From time.Time
	// To верхняя граница observed_at (UTC).
	To time.Time
	// Limit максимум строк в ответе.
	Limit int
}
