package queue

import (
	"sync"
	"testing"
)

func TestSafeQueue(t *testing.T) {
	q := NewSafeQueue[int]()

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		d := q.Pop()
		t.Logf("Pop 1: %v", d)
	}()

	go func() {
		defer wg.Done()
		d := q.Pop()
		t.Logf("Pop 2: %v", d)
	}()

	go func() {
		defer wg.Done()
		t.Log("Push 1")
		q.Push(1)
	}()

	go func() {
		defer wg.Done()
		t.Log("Push 2")
		q.Push(2)
	}()

	wg.Wait()
}
