package domain

// Camera описывает один RTSP-источник: назначение от coordinator или запись из статического конфига.
type Camera struct {
	// SegmentID логический идентификатор дорожного сегмента.
	SegmentID string
	// CameraID идентификатор камеры внутри сегмента.
	CameraID string
	// RTSPURL полный URL потока (например rtsp://host:8554/path).
	RTSPURL string
}
