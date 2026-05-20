package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidDataClasses_ContainsRoadSegmentVideo(t *testing.T) {
	t.Parallel()
	classes := ValidDataClasses()
	require.Contains(t, classes, DataClassRoadSegmentVideo)
}

func TestDataClassRoadSegmentVideo_Value(t *testing.T) {
	t.Parallel()
	require.Equal(t, "road_segment_video", DataClassRoadSegmentVideo)
}
