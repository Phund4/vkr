package coordinatorclient

import (
	"strings"

	"router/internal/core/domain"
)

// assignmentItemJSON один элемент ответа GET /v1/assignments (внутренний DTO).
type assignmentItemJSON struct {
	DataClass string `json:"data_class"`
	SegmentID string `json:"segment_id"`
	CameraID  string `json:"camera_id"`
	RTSPURL   string `json:"rtsp_url"`
}

// assignmentsRespJSON тело ответа coordinator со списком назначений.
type assignmentsRespJSON struct {
	Revision uint64               `json:"revision"`
	Items    []assignmentItemJSON `json:"items"`
}

// assignmentItemsToDomain отфильтровывает пустые поля и мапит JSON в domain.Camera.
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
