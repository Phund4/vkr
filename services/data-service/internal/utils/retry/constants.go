package retry

import "time"

const (
	baseDelayDefault  = 200 * time.Millisecond
	maxDelayDefault   = 10 * time.Second
	factorDefault     = 2.0
	jitterKindDefault = FullJitter
)
