package informer

import (
	"sync"
)

// ThreadSafeStore is an interface that allows concurrent indexed
// access to a storage backend.
type ThreadSafeStore interface {
	Add(key string, obj any) error
	Update(key string, obj any) error
	Delete(key string) error
	Get(key string) (item any, exists bool, err error)
	List() []any
	ListKeys() []string
	ListMap() map[string]any
	Range(yield func(key string, value any) bool)
	Replace(ListIter, string) error
	HasSynced() bool
	SyncDone()
}

type threadSafeMap struct {
	lock sync.RWMutex

	synced bool
	// store resource items
	items map[string]any
}

func NewThreadSafeStore() ThreadSafeStore {
	return &threadSafeMap{
		items: map[string]any{},
	}
}

// Add add object
func (c *threadSafeMap) Add(key string, obj any) error {
	c.Update(key, obj)
	return nil
}

// Update update object
func (c *threadSafeMap) Update(key string, obj any) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.items[key] = obj
	return nil
}

func (c *threadSafeMap) Delete(key string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.items, key)
	return nil
}

func (c *threadSafeMap) Get(key string) (item any, exists bool, err error) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	item, exists = c.items[key]
	return item, exists, nil
}

func (c *threadSafeMap) List() []any {
	c.lock.RLock()
	defer c.lock.RUnlock()
	list := make([]any, 0, len(c.items))
	for _, item := range c.items {
		list = append(list, item)
	}
	return list
}

func (c *threadSafeMap) ListKeys() []string {
	c.lock.RLock()
	defer c.lock.RUnlock()

	keys := make([]string, 0, len(c.items))
	for key := range c.items {
		keys = append(keys, key)
	}
	return keys
}

func (c *threadSafeMap) ListMap() map[string]any {
	c.lock.RLock()
	defer c.lock.RUnlock()

	copyItems := make(map[string]any, len(c.items))
	for k, v := range c.items {
		copyItems[k] = v
	}
	return copyItems
}

func (c *threadSafeMap) Replace(listiter ListIter, resourceVersion string) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	items := make(map[string]any)
	listiter(func(key string, value any) bool {
		items[key] = value
		return true
	})
	c.items = items

	return nil
}

func (c *threadSafeMap) SyncDone() {
	c.lock.Lock()
	defer c.lock.Unlock()
	c.synced = true
}

func (c *threadSafeMap) HasSynced() bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	return c.synced
}

func (c *threadSafeMap) Range(yield func(key string, value any) bool) {
	c.lock.RLock()
	defer c.lock.RUnlock()

	for key, value := range c.items {
		if !yield(key, value) {
			break
		}
	}
}
