package dto

import (
	"time"

	"data-service/internal/core/domain"
)

// RoadIncidentItem тело элемента списка в HTTP-ответе (road_incidents).
type RoadIncidentItem struct {
	ObservedAt       time.Time `json:"observed_at"`
	SegmentID        string    `json:"segment_id"`
	CameraID         string    `json:"camera_id"`
	S3Key            string    `json:"s3_key"`
	CrashProbability float64   `json:"crash_probability"`
	IncidentLabel    string    `json:"incident_label"`
	RawML            string    `json:"raw_ml"`
}

// RoadCongestionItem тело элемента списка в HTTP-ответе (road_congestion).
type RoadCongestionItem struct {
	ObservedAt      time.Time `json:"observed_at"`
	SegmentID       string    `json:"segment_id"`
	CameraID        string    `json:"camera_id"`
	S3Key           string    `json:"s3_key"`
	CongestionScore float64   `json:"congestion_score"`
	RawML           string    `json:"raw_ml"`
}

// RoadIncidentsListResponse ответ GET /api/v1/road_incidents.
type RoadIncidentsListResponse struct {
	Items []RoadIncidentItem `json:"items"`
}

// RoadCongestionListResponse ответ GET /api/v1/road_congestion.
type RoadCongestionListResponse struct {
	Items []RoadCongestionItem `json:"items"`
}

// NewRoadIncidentItem домен → HTTP DTO.
func NewRoadIncidentItem(d domain.RoadIncident) RoadIncidentItem {
	return RoadIncidentItem{
		ObservedAt:       d.ObservedAt,
		SegmentID:        d.SegmentID,
		CameraID:         d.CameraID,
		S3Key:            d.S3Key,
		CrashProbability: d.CrashProbability,
		IncidentLabel:    d.IncidentLabel,
		RawML:            d.RawML,
	}
}

// ToDomain HTTP DTO → домен (для будущих входящих тел).
func (h RoadIncidentItem) ToDomain() domain.RoadIncident {
	return domain.RoadIncident{
		ObservedAt:       h.ObservedAt,
		SegmentID:        h.SegmentID,
		CameraID:         h.CameraID,
		S3Key:            h.S3Key,
		CrashProbability: h.CrashProbability,
		IncidentLabel:    h.IncidentLabel,
		RawML:            h.RawML,
	}
}

// NewRoadCongestionItem домен → HTTP DTO.
func NewRoadCongestionItem(d domain.RoadCongestion) RoadCongestionItem {
	return RoadCongestionItem{
		ObservedAt:      d.ObservedAt,
		SegmentID:       d.SegmentID,
		CameraID:        d.CameraID,
		S3Key:           d.S3Key,
		CongestionScore: d.CongestionScore,
		RawML:           d.RawML,
	}
}

// ToDomain HTTP DTO → домен.
func (h RoadCongestionItem) ToDomain() domain.RoadCongestion {
	return domain.RoadCongestion{
		ObservedAt:      h.ObservedAt,
		SegmentID:       h.SegmentID,
		CameraID:        h.CameraID,
		S3Key:           h.S3Key,
		CongestionScore: h.CongestionScore,
		RawML:           h.RawML,
	}
}

// NewRoadIncidentsListResponse список доменных сущностей → ответ API.
func NewRoadIncidentsListResponse(items []domain.RoadIncident) RoadIncidentsListResponse {
	out := make([]RoadIncidentItem, 0, len(items))
	for _, d := range items {
		out = append(out, NewRoadIncidentItem(d))
	}
	return RoadIncidentsListResponse{Items: out}
}

// NewRoadCongestionListResponse список доменных сущностей → ответ API.
func NewRoadCongestionListResponse(items []domain.RoadCongestion) RoadCongestionListResponse {
	out := make([]RoadCongestionItem, 0, len(items))
	for _, d := range items {
		out = append(out, NewRoadCongestionItem(d))
	}
	return RoadCongestionListResponse{Items: out}
}
