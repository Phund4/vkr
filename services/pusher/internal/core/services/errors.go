package services

import "errors"

var (
	// ErrObservedAtEmpty строка observed_at пуста после trim.
	ErrObservedAtEmpty = errors.New("observed_at empty")

	// ErrSegmentOrCameraRequired не заданы segment_id или camera_id.
	ErrSegmentOrCameraRequired = errors.New("segment_id and camera_id required")

	// ErrFileKeyOrContentRequired для вложения не задан key или content_base64.
	ErrFileKeyOrContentRequired = errors.New("file key and content_base64 required")
)
