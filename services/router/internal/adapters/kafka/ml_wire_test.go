package kafkapub

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"router/internal/core/domain"
)

// Поля совпадают с _decode_frame в services/ml-serving/api/kafka_worker.py.
type mlServingFrame struct {
	SegmentID         string `json:"segment_id"`
	CameraID          string `json:"camera_id"`
	ObservedAt        string `json:"observed_at"`
	PipelineStartedAt string `json:"pipeline_started_at"`
	S3Key             string `json:"s3_key"`
	JPEGBase64        string `json:"jpeg_base64"`
}

func TestMLPublisherPayload_MatchesMLServingDecoder(t *testing.T) {
	t.Parallel()
	jpeg := []byte{0xff, 0xd8, 0xff}
	meta := domain.ProcessMeta{
		SegmentID:         "seg-a",
		CameraID:          "cam-1",
		ObservedAt:        "2026-05-04T12:00:00Z",
		PipelineStartedAt: "2026-05-04T12:00:00.123456789Z",
		S3Key:             "its-ingest/2026-05-04/cam-1/f.png",
	}
	msg := domain.MLFrameMessage{
		SegmentID:         meta.SegmentID,
		CameraID:          meta.CameraID,
		ObservedAt:        meta.ObservedAt,
		PipelineStartedAt: meta.PipelineStartedAt,
		S3Key:             meta.S3Key,
		JPEGBase64:        base64.StdEncoding.EncodeToString(jpeg),
	}
	body, err := json.Marshal(msg)
	require.NoError(t, err)

	var wire mlServingFrame
	require.NoError(t, json.Unmarshal(body, &wire))
	require.Equal(t, meta.SegmentID, wire.SegmentID)
	require.Equal(t, meta.CameraID, wire.CameraID)
	require.Equal(t, meta.ObservedAt, wire.ObservedAt)
	require.Equal(t, meta.PipelineStartedAt, wire.PipelineStartedAt)
	require.Equal(t, meta.S3Key, wire.S3Key)
	decoded, err := base64.StdEncoding.DecodeString(wire.JPEGBase64)
	require.NoError(t, err)
	require.Equal(t, jpeg, decoded)
}
