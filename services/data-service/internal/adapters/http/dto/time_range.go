package dto

import (
	"time"

	"data-service/internal/core/domain"
)

// FramesListResponse GET /api/v1/frames.
type FramesListResponse struct {
	Items []FrameItem `json:"items"`
}

// FrameItem кадр с метаданными и опциональной ссылкой S3.
type FrameItem struct {
	ObservedAt  time.Time `json:"observed_at"`
	SegmentID   string    `json:"segment_id"`
	CameraID    string    `json:"camera_id"`
	S3Key       string    `json:"s3_key"`
	DownloadURL string    `json:"download_url,omitempty"`
}

// CongestionAverageResponse GET /api/v1/segments/{segment_id}/congestion/average.
type CongestionAverageResponse struct {
	SegmentID   string    `json:"segment_id"`
	From        time.Time `json:"from"`
	To          time.Time `json:"to"`
	AvgScore    float64   `json:"avg_congestion_score"`
	SampleCount uint64    `json:"sample_count"`
}

// NewFramesListResponse домен → HTTP.
func NewFramesListResponse(items []domain.FrameWithURL) FramesListResponse {
	out := make([]FrameItem, 0, len(items))
	for _, it := range items {
		out = append(out, FrameItem{
			ObservedAt:  it.ObservedAt,
			SegmentID:   it.SegmentID,
			CameraID:    it.CameraID,
			S3Key:       it.S3Key,
			DownloadURL: it.DownloadURL,
		})
	}
	return FramesListResponse{Items: out}
}

// NewCongestionAverageResponse домен → HTTP.
func NewCongestionAverageResponse(r domain.CongestionAverageResult) CongestionAverageResponse {
	return CongestionAverageResponse{
		SegmentID:   r.SegmentID,
		From:        r.From,
		To:          r.To,
		AvgScore:    r.AvgScore,
		SampleCount: r.SampleCount,
	}
}
