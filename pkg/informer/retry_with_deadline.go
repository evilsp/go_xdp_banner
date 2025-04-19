package informer

import (
	"time"

	"xdp-banner/pkg/utils/clock"
)

type RetryWithDeadline interface {
	After(error)
	ShouldRetry() bool
}

type retryWithDeadlineImpl struct {
	firstErrorTime   time.Time
	lastErrorTime    time.Time
	maxRetryDuration time.Duration
	minResetPeriod   time.Duration
	isRetryable      func(error) bool
	clock            clock.Clock
}

func NewRetryWithDeadline(maxRetryDuration, minResetPeriod time.Duration, isRetryable func(error) bool, clock clock.Clock) RetryWithDeadline {
	return &retryWithDeadlineImpl{
		firstErrorTime:   time.Time{},
		lastErrorTime:    time.Time{},
		maxRetryDuration: maxRetryDuration,
		minResetPeriod:   minResetPeriod,
		isRetryable:      isRetryable,
		clock:            clock,
	}
}

func (r *retryWithDeadlineImpl) reset() {
	r.firstErrorTime = time.Time{}
	r.lastErrorTime = time.Time{}
}

func (r *retryWithDeadlineImpl) After(err error) {
	if r.isRetryable(err) {
		if r.clock.Now().Sub(r.lastErrorTime) >= r.minResetPeriod {
			r.reset()
		}

		if r.firstErrorTime.IsZero() {
			r.firstErrorTime = r.clock.Now()
		}
		r.lastErrorTime = r.clock.Now()
	}
}

func (r *retryWithDeadlineImpl) ShouldRetry() bool {
	if r.maxRetryDuration <= time.Duration(0) {
		return false
	}

	if r.clock.Now().Sub(r.firstErrorTime) <= r.maxRetryDuration {
		return true
	}

	r.reset()
	return false
}
