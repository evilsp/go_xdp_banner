package queue

import (
	"sync"
	"time"

	"xdp-banner/pkg/utils/clock"
)

type WorkQueue[T comparable] interface {
	Add(item T)
	Len() int
	Get() (item T, shutdown bool)
	Done(item T)
	ShutDown()
	ShutDownWithDrain()
	ShuttingDown() bool
}

// DefaultQueue is a slice based FIFO queue.
func DefaultQueue[T comparable]() Queue[T] {
	return NewQueue[T]()
}

type WorkQueueConfig[T comparable] struct {
	// Name for the queue. If unnamed, the metrics will not be registered.
	Name string

	// Clock ability to inject real or fake clock for testing purposes.
	Clock clock.WithTicker

	// Queue provides the underlying queue to use. It is optional and defaults to slice based FIFO queue.
	Queue Queue[T]
}

// NewWorkQueue constructs a new work queue.
func NewWorkQueue[T comparable]() *workqueue[T] {
	return NewWorkQueueWithConfig(WorkQueueConfig[T]{
		Name: "",
	})
}

// NewTypedWithConfig constructs a new workqueue with ability to
// customize different properties.
func NewWorkQueueWithConfig[T comparable](config WorkQueueConfig[T]) *workqueue[T] {
	return newWorkQueueWithConfig(config, defaultUnfinishedWorkUpdatePeriod)
}

// newQueueWithConfig constructs a new named workqueue
// with the ability to customize different properties for testing purposes
func newWorkQueueWithConfig[T comparable](config WorkQueueConfig[T], updatePeriod time.Duration) *workqueue[T] {
	if config.Clock == nil {
		config.Clock = clock.RealClock{}
	}

	if config.Queue == nil {
		config.Queue = DefaultQueue[T]()
	}

	return newWorkQueue(
		config.Clock,
		config.Queue,
		updatePeriod,
	)
}

func newWorkQueue[T comparable](c clock.WithTicker, queue Queue[T], updatePeriod time.Duration) *workqueue[T] {
	t := &workqueue[T]{
		clock:                      c,
		queue:                      queue,
		dirty:                      set[T]{},
		processing:                 set[T]{},
		cond:                       sync.NewCond(&sync.Mutex{}),
		unfinishedWorkUpdatePeriod: updatePeriod,
	}

	return t
}

const defaultUnfinishedWorkUpdatePeriod = 500 * time.Millisecond

type workqueue[t comparable] struct {
	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue Queue[t]

	// dirty defines all of the items that need to be processed.
	dirty set[t]

	// Things that are currently being processed are in the processing set.
	// These things may be simultaneously in the dirty set. When we finish
	// processing something and remove it from this set, we'll check if
	// it's in the dirty set, and if so, add it to the queue.
	processing set[t]

	cond *sync.Cond

	shuttingDown bool
	drain        bool

	unfinishedWorkUpdatePeriod time.Duration
	clock                      clock.WithTicker
}

type empty struct{}
type t interface{}
type set[t comparable] map[t]empty

func (s set[t]) has(item t) bool {
	_, exists := s[item]
	return exists
}

func (s set[t]) insert(item t) {
	s[item] = empty{}
}

func (s set[t]) delete(item t) {
	delete(s, item)
}

func (s set[t]) len() int {
	return len(s)
}

// Add marks item as needing processing.
func (q *workqueue[T]) Add(item T) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return
	}
	if q.dirty.has(item) {
		// the same item is added again before it is processed, call the Touch
		// function if the queue cares about it (for e.g, reset its priority)
		if !q.processing.has(item) {
			q.queue.Touch(item)
		}
		return
	}

	q.dirty.insert(item)
	if q.processing.has(item) {
		return
	}

	q.queue.Push(item)
	q.cond.Signal()
}

// Len returns the current queue length, for informational purposes only. You
// shouldn't e.g. gate a call to Add() or Get() on Len() being a particular
// value, that can't be synchronized properly.
func (q *workqueue[T]) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.queue.Len()
}

// Get blocks until it can return an item to be processed. If shutdown = true,
// the caller should end their goroutine. You must call Done with item when you
// have finished processing it.
func (q *workqueue[T]) Get() (item T, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for q.queue.Len() == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if q.queue.Len() == 0 {
		// We must be shutting down.
		return *new(T), true
	}

	item = q.queue.Pop()

	q.processing.insert(item)
	q.dirty.delete(item)

	return item, false
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *workqueue[T]) Done(item T) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.processing.delete(item)
	if q.dirty.has(item) {
		q.queue.Push(item)
		q.cond.Signal()
	} else if q.processing.len() == 0 {
		q.cond.Signal()
	}
}

// ShutDown will cause q to ignore all new items added to it and
// immediately instruct the worker goroutines to exit.
func (q *workqueue[T]) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.drain = false
	q.shuttingDown = true
	q.cond.Broadcast()
}

// ShutDownWithDrain will cause q to ignore all new items added to it. As soon
// as the worker goroutines have "drained", i.e: finished processing and called
// Done on all existing items in the queue; they will be instructed to exit and
// ShutDownWithDrain will return. Hence: a strict requirement for using this is;
// your workers must ensure that Done is called on all items in the queue once
// the shut down has been initiated, if that is not the case: this will block
// indefinitely. It is, however, safe to call ShutDown after having called
// ShutDownWithDrain, as to force the queue shut down to terminate immediately
// without waiting for the drainage.
func (q *workqueue[T]) ShutDownWithDrain() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	q.drain = true
	q.shuttingDown = true
	q.cond.Broadcast()

	for q.processing.len() != 0 && q.drain {
		q.cond.Wait()
	}
}

func (q *workqueue[T]) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}
