package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunAttempt(t *testing.T) {
	t.Parallel()

	type R = int
	mockResp := new(R)
	*mockResp = 42

	mockStrategy := func() *expoStrategy {
		return &expoStrategy{
			baseDelay:  1 * time.Millisecond,
			maxDelay:   1 * time.Millisecond,
			factor:     2.0,
			jitterKind: NoJitter,
		}
	}

	testCases := []struct {
		name string

		ctx     context.Context
		fn      RetryFunc[R]
		cfg     *retryConfig[R]
		attempt int

		wantResp *R
		wantErr  assert.ErrorAssertionFunc
	}{
		{
			name: "Successful run attempt",

			ctx: context.Background(),
			fn: func(ctx context.Context) (*R, error) {
				return mockResp, nil
			},
			cfg: &retryConfig[R]{
				maxAttempts: 3,
				strategy:    mockStrategy(),
				shouldRetry: NeverRetry[R],
			},
			attempt: 1,

			wantResp: mockResp,
			wantErr:  assert.NoError,
		},
		{
			name: "Attempts exhausted",

			ctx: context.Background(),
			fn: func(ctx context.Context) (*R, error) {
				return mockResp, ErrAttemptsExhausted
			},
			cfg: &retryConfig[R]{
				maxAttempts: 3,
				strategy:    mockStrategy(),
				shouldRetry: AlwaysRetry[R],
			},
			attempt: 3,

			wantResp: nil,
			wantErr:  assert.Error,
		},
		{
			name: "Context canceled",

			ctx: func() context.Context {
				c, cancel := context.WithCancel(context.Background())
				cancel()
				return c
			}(),
			fn: func(ctx context.Context) (*R, error) {
				return mockResp, context.Canceled
			},
			cfg: &retryConfig[R]{
				maxAttempts: 5,
				strategy:    mockStrategy(),
				shouldRetry: AlwaysRetry[R],
			},
			attempt: 1,

			wantResp: nil,
			wantErr:  assert.Error,
		},
		{
			name: "The waiting is over",

			ctx: context.Background(),
			fn: func(ctx context.Context) (*R, error) {
				return mockResp, ErrBackoffWaitCanceled
			},
			cfg: &retryConfig[R]{
				maxAttempts: 5,
				strategy:    mockStrategy(),
				shouldRetry: AlwaysRetry[R],
			},
			attempt: 1,

			wantResp: nil,
			wantErr:  assert.Error,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			r := NewRetryObj(tc.cfg)
			resp, err := r.runAttempt(tc.ctx, tc.fn, tc.attempt)

			assert.Equal(t, tc.wantResp, resp)
			tc.wantErr(t, err)
		})
	}
}
