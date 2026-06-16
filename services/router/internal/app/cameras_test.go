package app

import (
	"testing"

	"router/internal/core/domain"
)

func TestCamerasEqual(t *testing.T) {
	t.Parallel()
	a := []domain.Camera{
		{SegmentID: "s", CameraID: "cam-02", RTSPURL: "rtsp://x/cam-02"},
		{SegmentID: "s", CameraID: "cam-01", RTSPURL: "rtsp://x/cam-01"},
	}
	b := []domain.Camera{
		{SegmentID: "s", CameraID: "cam-01", RTSPURL: "rtsp://x/cam-01"},
		{SegmentID: "s", CameraID: "cam-02", RTSPURL: "rtsp://x/cam-02"},
	}
	requireEqual(t, true, camerasEqual(a, b))
	requireEqual(t, false, camerasEqual(a, append(b, domain.Camera{SegmentID: "s", CameraID: "cam-03", RTSPURL: "rtsp://x/cam-03"})))
}

func requireEqual(t *testing.T, want, got bool) {
	t.Helper()
	if want != got {
		t.Fatalf("want %v got %v", want, got)
	}
}
