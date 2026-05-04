package retry

type Jitter int

const (
	NoJitter Jitter = iota
	EqualJitter
	FullJitter
	DecorrelatedJitter
)
