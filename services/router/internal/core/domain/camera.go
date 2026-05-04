package domain

// Camera назначенный или статически заданный RTSP-источник.
type Camera struct {
	SegmentID string
	CameraID  string
	RTSPURL   string
}
