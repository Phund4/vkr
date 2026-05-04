package retry

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type expoStrategy struct {
	baseDelay   time.Duration
	maxDelay    time.Duration
	factor      float64
	jitterKind  Jitter
	prevDelayMu sync.RWMutex
	prevDelay   time.Duration
	rnd         *rand.Rand
}

func defaultExpoStrategy() *expoStrategy {
	return &expoStrategy{
		baseDelay:  baseDelayDefault,
		maxDelay:   maxDelayDefault,
		factor:     factorDefault,
		jitterKind: jitterKindDefault,
		rnd:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func NewExpoStrategy(opts ...ExpoOption) (*expoStrategy, error) {
	es := defaultExpoStrategy()

	for _, opt := range opts {
		if err := opt(es); err != nil {
			return nil, fmt.Errorf("%w: %w", ErrInvalidExpoOptions, err)
		}
	}

	if es.baseDelay > es.maxDelay {
		return nil, ErrBaseGreaterThanMax
	}

	return es, nil
}

func (e *expoStrategy) Reset() {
	e.prevDelayMu.Lock()
	e.prevDelay = 0
	e.prevDelayMu.Unlock()
}

func (e *expoStrategy) Next(attempt int) time.Duration {
	baseDelay := e.baseDelayForAttempt(attempt)

	switch e.jitterKind {
	case EqualJitter:
		return e.calculateEqualJitterDelay(baseDelay)
	case FullJitter:
		return e.calculateFullJitterDelay(baseDelay)
	case DecorrelatedJitter:
		return e.calculateDecorrelatedJitterDelay()
	default:
		return e.calculateNoJitterDelay(baseDelay)
	}
}

// baseDelayForAttempt вычисляет задержку для попытки с учётом экспоненты.
// Используется float64 -> time.Duration умножение, что теоретически может
// вносить неточность на больших значениях. Однако это безопасно, т.к.
// при достижении maxDelay рост обрезается и неточности не влияют на результат.
func (e *expoStrategy) baseDelayForAttempt(attempt int) time.Duration {
	delay := float64(e.baseDelay)
	for i := 1; i < attempt; i++ {
		delay *= e.factor
		if delay >= float64(e.maxDelay) {
			return e.maxDelay
		}
	}

	return time.Duration(delay)
}

func (e *expoStrategy) calculateNoJitterDelay(baseDelay time.Duration) time.Duration {
	return baseDelay
}

func (e *expoStrategy) calculateEqualJitterDelay(baseDelay time.Duration) time.Duration {
	half := baseDelay / 2
	j := time.Duration(e.rnd.Int63n(int64(half) + 1))
	return min(half+j, e.maxDelay)
}

func (e *expoStrategy) calculateFullJitterDelay(baseDelay time.Duration) time.Duration {
	j := time.Duration(e.rnd.Int63n(int64(baseDelay) + 1))
	return min(j, e.maxDelay)
}

func (e *expoStrategy) calculateDecorrelatedJitterDelay() time.Duration {
	e.prevDelayMu.Lock()
	defer e.prevDelayMu.Unlock()

	upper := e.prevDelay
	if upper <= 0 {
		upper = e.baseDelay
	}

	upper = time.Duration(float64(upper) * e.factor)
	if upper < e.baseDelay {
		upper = e.baseDelay
	}
	upper = min(upper, e.maxDelay)

	span := int64(upper - e.baseDelay)
	var next time.Duration
	if span <= 0 {
		next = e.baseDelay
	} else {
		next = e.baseDelay + time.Duration(e.rnd.Int63n(span+1))
	}

	e.prevDelay = next
	return next
}
