package queue

import (
	"sync"
)

type Queue[T any] interface {
	// Touch can be hooked when an existing item is added again. This may be
	// useful if the implementation allows priority change for the given item.
	Touch(item T)
	// Push adds a new item.
	Push(item T)
	// Len tells the total number of items.
	Len() int
	// Pop retrieves an item.
	Pop() (item T)
}

type queue[T any] []T

func NewQueue[T any]() Queue[T] {
	return new(queue[T])
}

func (q *queue[T]) Touch(item T) {}

func (q *queue[T]) Push(item T) {
	*q = append(*q, item)
}

func (q *queue[T]) Len() int {
	return len(*q)
}

func (q *queue[T]) Pop() (item T) {
	item = (*q)[0]

	// The underlying array still exists and reference this object, so the object will not be garbage collected.
	(*q)[0] = *new(T)
	*q = (*q)[1:]

	return item
}

type safeQueue[T any] struct {
	queue    Queue[T]
	lock     sync.Mutex
	notEmpty *sync.Cond
}

func NewSafeQueue[T any]() Queue[T] {
	q := &safeQueue[T]{
		queue: NewQueue[T](),
	}
	q.notEmpty = sync.NewCond(&q.lock)

	return q
}

func (q *safeQueue[T]) Touch(item T) {}

func (q *safeQueue[T]) Push(item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.queue.Push(item)
	q.notEmpty.Signal()
}

func (q *safeQueue[T]) Pop() (item T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	for q.len_unsafe() == 0 {
		q.notEmpty.Wait()
	}

	return q.queue.Pop()
}

func (q *safeQueue[T]) Len() int {
	q.lock.Lock()
	defer q.lock.Unlock()

	return q.len_unsafe()
}

func (q *safeQueue[T]) len_unsafe() int {
	return q.queue.Len()
}
