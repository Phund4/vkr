package services

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnqueueFrame_DropsOldestWhenFull(t *testing.T) {
	t.Parallel()
	ch := make(chan []byte, 1)
	enqueueFrame(ch, []byte("old"))
	enqueueFrame(ch, []byte("new"))
	require.Equal(t, []byte("new"), <-ch)
}

func TestFrameChanCapacity_Minimum(t *testing.T) {
	t.Parallel()
	require.Equal(t, minFrameChanCapacity, frameChanCapacity(0))
	require.GreaterOrEqual(t, frameChanCapacity(4), minFrameChanCapacity)
}
