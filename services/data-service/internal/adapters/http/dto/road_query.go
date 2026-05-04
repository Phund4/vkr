package dto

import "data-service/internal/core/domain"

// RoadListParamsFromQuery собирает доменные параметры списка из query HTTP.
func RoadListParamsFromQuery(segmentID, cameraID string, limit int) domain.RoadListParams {
	return domain.RoadListParams{
		SegmentID: segmentID,
		CameraID:  cameraID,
		Limit:     limit,
	}
}
