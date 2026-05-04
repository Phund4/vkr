package retry

import "time"

type ExpoOption func(*expoStrategy) error

func WithBaseDelay(d time.Duration) ExpoOption {
	return func(es *expoStrategy) error {
		if d <= 0 {
			return ErrEmptyBaseDelay
		}

		es.baseDelay = d
		return nil
	}
}

func WithMaxDelay(d time.Duration) ExpoOption {
	return func(es *expoStrategy) error {
		if d <= 0 {
			return ErrEmptyMaxDelay
		}

		es.maxDelay = d
		return nil
	}
}

func WithFactor(f float64) ExpoOption {
	return func(es *expoStrategy) error {
		if f <= 1 {
			return ErrInvalidFactor
		}

		es.factor = f
		return nil
	}
}

func WithJitter(j Jitter) ExpoOption {
	return func(es *expoStrategy) error {
		es.jitterKind = j
		return nil
	}
}
