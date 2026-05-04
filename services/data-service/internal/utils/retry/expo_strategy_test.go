package retry

import (
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateNoJitterDelay(t *testing.T) {
	t.Parallel()

	e := expoStrategy{
		baseDelay:  100 * time.Millisecond,
		maxDelay:   time.Second,
		factor:     2.0,
		jitterKind: NoJitter,
	}

	testCases := []struct {
		name      string
		baseDelay time.Duration
		want      time.Duration
	}{
		{
			name:      "base delay",
			baseDelay: 750 * time.Millisecond,
			want:      750 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			delay := e.calculateNoJitterDelay(tc.baseDelay)
			assert.Equal(t, tc.want, delay)
		})
	}
}

func TestCalculateEqualJitterDelay(t *testing.T) {
	t.Parallel()

	e := expoStrategy{
		baseDelay:  100 * time.Millisecond,
		maxDelay:   2 * time.Second,
		factor:     2.0,
		jitterKind: NoJitter,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	tests := []struct {
		name      string
		baseDelay time.Duration
		maxDelay  time.Duration
	}{
		{
			name:      "IN between [base/2, base]",
			baseDelay: 800 * time.Millisecond,
			maxDelay:  2 * time.Second,
		},
		{
			name:      "Capped by maxDelay",
			baseDelay: 900 * time.Millisecond,
			maxDelay:  500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			e.maxDelay = tt.maxDelay
			delay := e.calculateEqualJitterDelay(tt.baseDelay)

			half := tt.baseDelay / 2
			upper := min(tt.baseDelay, e.maxDelay)
			if delay < half || delay > upper {
				t.Fatalf("equal jitter out of range: got=%v; want in [%v, %v]", delay, half, upper)
			}
		})
	}
}

func TestCalculateFullJitterDelay(t *testing.T) {
	t.Parallel()

	e := expoStrategy{
		baseDelay:  100 * time.Millisecond,
		maxDelay:   time.Second,
		factor:     2.0,
		jitterKind: NoJitter,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}

	testCases := []struct {
		name      string
		baseDelay time.Duration
		maxDelay  time.Duration
	}{
		{
			name:      "In between [0, base]",
			baseDelay: 750 * time.Millisecond,
			maxDelay:  time.Second,
		},
		{
			name:      "Capped by maxDelay",
			baseDelay: 900 * time.Millisecond,
			maxDelay:  400 * time.Millisecond,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e.maxDelay = tc.maxDelay
			delay := e.calculateFullJitterDelay(tc.baseDelay)

			lower := time.Duration(0)
			upper := min(tc.baseDelay, e.maxDelay)
			if delay < lower || delay > upper {
				t.Fatalf("full jitter out of range: got=%v; want in [%v, %v]", delay, lower, upper)
			}
		})
	}
}

func TestCalculateDecorrelatedJitterDelay(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		iterations int
		checkFirst bool
	}{
		{
			name:       "first step in [base, base*factor] and updates prev",
			iterations: 1,
			checkFirst: true,
		},
		{
			name:       "subsequent steps in [base, max] and respect cap",
			iterations: 20,
			checkFirst: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := expoStrategy{
				baseDelay:  100 * time.Millisecond,
				maxDelay:   time.Second,
				factor:     2.0,
				jitterKind: DecorrelatedJitter,
				rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
			}

			for i := 0; i < tc.iterations; i++ {
				got := e.calculateDecorrelatedJitterDelay()

				if i == 0 && tc.checkFirst {
					lo := e.baseDelay
					hi := time.Duration(float64(e.baseDelay) * e.factor)
					if got < lo || got > hi {
						t.Fatalf("step1 out of range: got=%v; want in [%v, %v]", got, lo, hi)
					}
					if e.prevDelay != got {
						t.Fatalf("prevDelay not updated on step1: got=%v; prev=%v", got, e.prevDelay)
					}
					continue
				}

				if got < e.baseDelay || got > e.maxDelay {
					t.Fatalf("iter %d: out of range: got=%v; want in [%v, %v]", i, got, e.baseDelay, e.maxDelay)
				}
				if e.prevDelay != got {
					t.Fatalf("iter %d: prevDelay not updated: got=%v; prev=%v", i, got, e.prevDelay)
				}
			}
		})
	}
}
