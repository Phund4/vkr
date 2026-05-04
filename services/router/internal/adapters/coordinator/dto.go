package coordinatorclient

import (
	"strings"

	"router/internal/core/domain"
)

type assignmentItemJSON struct {
	DataClass string `json:"data_class"`
	SegmentID string `json:"segment_id"`
	CameraID  string `json:"camera_id"`
	RTSPURL   string `json:"rtsp_url"`
}

type assignmentsRespJSON struct {
	Items []assignmentItemJSON `json:"items"`
}

func assignmentItemsToDomain(items []assignmentItemJSON) []domain.Camera {
	out := make([]domain.Camera, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.SegmentID) == "" || strings.TrimSpace(it.CameraID) == "" || strings.TrimSpace(it.RTSPURL) == "" {
			continue
		}
		out = append(out, domain.Camera{
			SegmentID: it.SegmentID,
			CameraID:  it.CameraID,
			RTSPURL:   it.RTSPURL,
		})
	}
	return out
}
