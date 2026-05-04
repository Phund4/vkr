package retry

import "errors"

var (
	ErrCreateStrategy         = errors.New("failed to create expo strategy")
	ErrInvalidRetriesInput    = errors.New("invalid retries function parameters")
	ErrEmptyRetryConfig       = errors.New("empty retry config")
	ErrEmptyOperation         = errors.New("empty operation")
	ErrEmptyRetryPolicy       = errors.New("empty retry policy")
	ErrNonPositiveMaxAttempts = errors.New("max attempts must be positive")
	ErrEmptyStrategy          = errors.New("empty retry strategy")
	ErrBaseGreaterThanMax     = errors.New("base delay must be lower than max delay")
	ErrProvideBody            = errors.New("failed to provide body")
	ErrEmptyBaseDelay         = errors.New("empty base delay")
	ErrEmptyMaxDelay          = errors.New("empty max delay")
	ErrInvalidFactor          = errors.New("factor must be greater than 1")
	ErrInvalidExpoOptions     = errors.New("error in exponential strategy options")
	ErrAttemptsExhausted      = errors.New("retry attempts exhausted")
	ErrBackoffWaitCanceled    = errors.New("backoff wait canceled")
)
