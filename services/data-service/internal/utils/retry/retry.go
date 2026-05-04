package retry

import (
	"context"
	"fmt"
	"time"
)

type RetryFunc[R any] func(ctx context.Context) (*R, error)

type ShouldRetry[R any] func(*R, error) (bool, error)

func AlwaysRetry[R any](_ *R, _ error) (bool, error) { return true, nil }

func NeverRetry[R any](_ *R, _ error) (bool, error) { return false, nil }

func RetryOnError[R any](_ *R, err error) (bool, error) { return err != nil, err }

type retryConfig[R any] struct {
	maxAttempts int
	strategy    *expoStrategy
	shouldRetry ShouldRetry[R]
}

func NewRetryConfig[R any](maxAttempts int, shouldRetry ShouldRetry[R], opts ...ExpoOption) (*retryConfig[R], error) {
	strategy, err := NewExpoStrategy(opts...)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrCreateStrategy, err)
	}

	cfg := &retryConfig[R]{
		maxAttempts: maxAttempts,
		strategy:    strategy,
		shouldRetry: shouldRetry,
	}

	if err := validateConfig(cfg); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidRetriesInput, err)
	}

	return cfg, nil
}

type RetryObj[R any] struct {
	cfg *retryConfig[R]
}

func NewRetryObj[R any](cfg *retryConfig[R]) *RetryObj[R] {
	return &RetryObj[R]{
		cfg: cfg,
	}
}

func (r *RetryObj[R]) DoWithRetry(
	ctx context.Context,
	fn RetryFunc[R],
) (resp *R, err error) {
	if fn == nil {
		return nil, ErrEmptyOperation
	}

	r.cfg.strategy.Reset()

	for attempt := 1; attempt <= r.cfg.maxAttempts; attempt++ {
		resp, err = r.runAttempt(ctx, fn, attempt)
		if err == nil {
			return resp, nil
		}
	}

	return nil, err
}

func (r *RetryObj[R]) runAttempt(
	ctx context.Context,
	fn RetryFunc[R],
	attempt int,
) (*R, error) {
	var ok bool
	resp, err := fn(ctx)
	if ok, err = r.cfg.shouldRetry(resp, err); !ok && err == nil {
		return resp, err
	}

	if attempt == r.cfg.maxAttempts {
		return nil, fmt.Errorf("%w: %w", ErrAttemptsExhausted, err)
	}

	delay := r.cfg.strategy.Next(attempt)
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
		return nil, fmt.Errorf("%w: %w", ErrBackoffWaitCanceled, err)
	}
}

func validateConfig[R any](cfg *retryConfig[R]) error {
	switch {
	case cfg == nil:
		return ErrEmptyRetryConfig
	case cfg.maxAttempts <= 0:
		return ErrNonPositiveMaxAttempts
	case cfg.strategy == nil:
		return ErrEmptyStrategy
	case cfg.shouldRetry == nil:
		return ErrEmptyRetryPolicy
	default:
		return nil
	}
}
