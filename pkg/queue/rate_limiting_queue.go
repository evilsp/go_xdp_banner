package queue

import "xdp-banner/pkg/utils/clock"

// RateLimitingQueue is an interface that rate limits items being added to the queue.
type RateLimitingQueue[T comparable] interface {
	DelayingQueue[T]

	// AddRateLimited adds an item to the workqueue after the rate limiter says it's ok
	AddRateLimited(item T)

	// Forget indicates that an item is finished being retried.  Doesn't matter whether it's for perm failing
	// or for success, we'll stop the rate limiter from tracking it.  This only clears the `rateLimiter`, you
	// still have to call `Done` on the queue.
	Forget(item T)

	// NumRequeues returns back how many times the item was requeued
	NumRequeues(item T) int
}

// RateLimitingQueueConfig specifies optional configurations to customize a TypedRateLimitingInterface.
type RateLimitingQueueConfig[T comparable] struct {
	// Name for the queue. If unnamed, the metrics will not be registered.
	Name string

	// Clock optionally allows injecting a real or fake clock for testing purposes.
	Clock clock.WithTicker

	// DelayingQueue optionally allows injecting custom delaying queue DelayingInterface instead of the default one.
	DelayingQueue DelayingQueue[T]
}

// NewRateLimitingQueue constructs a new workqueue with rateLimited queuing ability
// Remember to call Forget!  If you don't, you may end up tracking failures forever.
func NewRateLimitingQueue[T comparable](rateLimiter RateLimiter[T]) RateLimitingQueue[T] {
	return NewRateLimitingQueueWithConfig(rateLimiter, RateLimitingQueueConfig[T]{})
}

// NewRateLimitingQueueWithName constructs a new workqueue with rateLimited queuing ability
// with a name. Remember to call Forget!  If you don't, you may end up tracking failures forever.
func NewRateLimitingQueueWithName[T comparable](rateLimiter RateLimiter[T], name string) RateLimitingQueue[T] {
	return NewRateLimitingQueueWithConfig(rateLimiter, RateLimitingQueueConfig[T]{Name: name})
}

// NewRateLimitingQueueWithConfig constructs a new workqueue with rateLimited queuing ability
// with options to customize different properties.
// Remember to call Forget!  If you don't, you may end up tracking failures forever.
func NewRateLimitingQueueWithConfig[T comparable](rateLimiter RateLimiter[T], config RateLimitingQueueConfig[T]) RateLimitingQueue[T] {
	if config.Clock == nil {
		config.Clock = clock.RealClock{}
	}

	if config.DelayingQueue == nil {
		config.DelayingQueue = NewDelayingQueueWithConfig(DelayingQueueConfig[T]{
			Name:  config.Name,
			Clock: config.Clock,
		})
	}

	return &rateLimitingQueue[T]{
		DelayingQueue: config.DelayingQueue,
		rateLimiter:   rateLimiter,
	}
}

// rateLimitingQueue wraps an Interface and provides rateLimited re-enquing
type rateLimitingQueue[T comparable] struct {
	DelayingQueue[T]

	rateLimiter RateLimiter[T]
}

// AddRateLimited AddAfter's the item based on the time when the rate limiter says it's ok
func (q *rateLimitingQueue[T]) AddRateLimited(item T) {
	q.DelayingQueue.AddAfter(item, q.rateLimiter.When(item))
}

func (q *rateLimitingQueue[T]) NumRequeues(item T) int {
	return q.rateLimiter.NumRequeues(item)
}

func (q *rateLimitingQueue[T]) Forget(item T) {
	q.rateLimiter.Forget(item)
}
