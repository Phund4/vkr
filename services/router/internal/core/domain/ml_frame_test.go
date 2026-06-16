package domain

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestMLFrameMessage_JSONRoundTrip(t *testing.T) {
	t.Parallel()
	raw := []byte("jpeg-bytes")
	msg := MLFrameMessage{
		SegmentID:  "seg-1",
		CameraID:   "cam-1",
		ObservedAt: "2026-05-04T12:00:00Z",
		S3Key:      "its-ingest/2026-05-04/cam-1/frame_1.png",
		JPEGBase64: base64.StdEncoding.EncodeToString(raw),
	}
	b, err := json.Marshal(msg)
	require.NoError(t, err)
	var decoded MLFrameMessage
	require.NoError(t, json.Unmarshal(b, &decoded))
	require.Equal(t, msg.SegmentID, decoded.SegmentID)
	out, err := base64.StdEncoding.DecodeString(decoded.JPEGBase64)
	require.NoError(t, err)
	require.Equal(t, raw, out)
}
