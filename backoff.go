package sdk

import (
	"math/rand"
	"time"
)

var (
	defaultMaxDelay  = 3.0 * time.Minute
	defaultBaseDelay = 1.0 * time.Second
	defaultFactor    = 1.6
	defaultJitter    = 0.2
)

type Backoff struct {
	MaxDelay  time.Duration
	baseDelay time.Duration
	factor    float64
	jitter    float64
}

func NewBackoff() *Backoff {
	var (
		factor    float64
		jitter    float64
		maxDelay  time.Duration
		baseDelay time.Duration
	)

	factor = defaultFactor
	jitter = defaultJitter
	maxDelay = defaultMaxDelay
	baseDelay = defaultBaseDelay

	return &Backoff{
		MaxDelay:  maxDelay,
		baseDelay: baseDelay,
		factor:    factor,
		jitter:    jitter,
	}
}

func (bc *Backoff) Duration(retries int) time.Duration {
	if retries <= 0 {
		return bc.baseDelay
	}

	backoff, max := float64(bc.baseDelay), float64(bc.MaxDelay)
	for backoff < max && retries > 0 {
		backoff *= bc.factor
		retries--
	}

	if backoff > max {
		backoff = max
	}

	backoff *= 1 + bc.jitter*(rand.Float64()*2-1)
	if backoff < 0 {
		return 0
	}

	return time.Duration(backoff)
}
