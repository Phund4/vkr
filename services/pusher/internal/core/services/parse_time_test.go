package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestParseObservedAt_RFC3339Nano(t *testing.T) {
	t.Parallel()
	want := time.Date(2026, 5, 4, 12, 0, 0, 123000000, time.UTC)
	got, err := parseObservedAt("2026-05-04T12:00:00.123Z")
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestParseObservedAt_Empty(t *testing.T) {
	t.Parallel()
	_, err := parseObservedAt("")
	require.ErrorIs(t, err, ErrObservedAtEmpty)
}
