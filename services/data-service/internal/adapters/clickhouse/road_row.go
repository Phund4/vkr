package clickhouse

import (
	"time"

	"data-service/internal/core/domain"
)

// RoadIncidentRow строка результата SELECT по road_incidents (адаптер ClickHouse).
type RoadIncidentRow struct {
	ObservedAt       time.Time
	SegmentID        string
	CameraID         string
	S3Key            string
	CrashProbability float64
	IncidentLabel    string
	RawML            string
}

func (r RoadIncidentRow) toDomain() domain.RoadIncident {
	return domain.RoadIncident{
		ObservedAt:       r.ObservedAt,
		SegmentID:        r.SegmentID,
		CameraID:         r.CameraID,
		S3Key:            r.S3Key,
		CrashProbability: r.CrashProbability,
		IncidentLabel:    r.IncidentLabel,
		RawML:            r.RawML,
	}
}

// RoadCongestionRow строка результата SELECT по road_congestion.
type RoadCongestionRow struct {
	ObservedAt      time.Time
	SegmentID       string
	CameraID        string
	S3Key           string
	CongestionScore float64
	RawML           string
}

func (r RoadCongestionRow) toDomain() domain.RoadCongestion {
	return domain.RoadCongestion{
		ObservedAt:      r.ObservedAt,
		SegmentID:       r.SegmentID,
		CameraID:        r.CameraID,
		S3Key:           r.S3Key,
		CongestionScore: r.CongestionScore,
		RawML:           r.RawML,
	}
}

// NewRoadIncidentRow домен → строка для подготовки INSERT/батчей (симметрия к toDomain).
func NewRoadIncidentRow(d domain.RoadIncident) RoadIncidentRow {
	return RoadIncidentRow{
		ObservedAt:       d.ObservedAt,
		SegmentID:        d.SegmentID,
		CameraID:         d.CameraID,
		S3Key:            d.S3Key,
		CrashProbability: d.CrashProbability,
		IncidentLabel:    d.IncidentLabel,
		RawML:            d.RawML,
	}
}

// NewRoadCongestionRow домен → строка для подготовки INSERT/батчей.
func NewRoadCongestionRow(d domain.RoadCongestion) RoadCongestionRow {
	return RoadCongestionRow{
		ObservedAt:      d.ObservedAt,
		SegmentID:       d.SegmentID,
		CameraID:        d.CameraID,
		S3Key:           d.S3Key,
		CongestionScore: d.CongestionScore,
		RawML:           d.RawML,
	}
}
