package dto

import "time"

const (
	StatusOK    = true
	StatusError = "error"

	MessageOK = "ok"

	ErrInvalidRateLimitIdentifier = "invalid rate limit identifier"
	ErrApiRateLimitExceeded       = "api rate limit exceeded"
	ErrRoadDataRateLimitExceeded  = "road data rate limit exceeded"
	ErrTimeRangeRequired          = "from and to query parameters are required (RFC3339)"
	ErrTimeRangeInvalid           = "invalid from/to time range"
	ErrCameraIDRequired           = "camera_id is required"
	ErrSegmentIDRequired          = "segment_id is required"

	ErrLimitQueryInvalid = "query param limit must be a positive integer"

	ApiV1RateLimitRPS        = 40
	ApiV1RateLimitBurst      = 80
	RoadDataRateLimitRPS     = 20
	RoadDataRateLimitBurst   = 40
	RateLimitExpiresInMinutes = 3 * time.Minute
)
