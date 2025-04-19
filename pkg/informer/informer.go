package informer

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"
	"xdp-banner/pkg/wait"

	"go.uber.org/zap"
)

const (
	// syncedPollPeriod controls how often you look at the status of your sync funcs
	syncedPollPeriod = 100 * time.Millisecond
)

type Informer interface {
	Run(ctx context.Context)
	Register(keyPrefix []string, handler ResourceEventHandler)
	HasSynced() bool
	WaitForCacheSync(ctx context.Context) bool
	Get(key string) (any, bool, error)
	ListCacheStoreMap() map[string]any
	RegisterHandlerAndList(keyPrefix []string, handler ResourceEventHandler) (map[string]any, error)
}

type informer struct {
	stopChan                   chan struct{}
	deltaFIFO                  *DeltaFIFO
	cacheStore                 ThreadSafeStore
	dispatchByKeyPrefixHandler *dispatchByKeyPrefixHandler
	process                    PopProcessFunc

	// 新增：用于保护“注册 + list”临界区的互斥锁
	mu sync.Mutex
}

func New(deltaFIFO *DeltaFIFO) Informer {
	cacheStore := NewThreadSafeStore()
	h := NewDispatchByKeyPrefixHandler()

	inf := &informer{
		stopChan:                   make(chan struct{}),
		deltaFIFO:                  deltaFIFO,
		cacheStore:                 cacheStore,
		dispatchByKeyPrefixHandler: h,
	}

	// --- CHANGED: 用互斥锁包裹对缓存的实际写操作 ---
	inf.process = func(key string, obj any, isInInitialList bool) error {
		// 在这里加锁, 避免与 RegisterHandlerAndList 并发更新缓存
		inf.mu.Lock()
		defer inf.mu.Unlock()

		if deltas, ok := obj.(Deltas); ok {
			return processDeltas(h, cacheStore, deltas, isInInitialList)
		}
		return errors.New("object given as Process argument is not Deltas")
	}

	return inf

}

func (i *informer) ListCacheStoreMap() map[string]any {
	return i.cacheStore.ListMap()
}

// RegisterHandlerAndList 在加锁情况下执行：
//  1. 将指定的 keyPrefix 与 handler 进行 register
//  2. 在锁的保护下立刻 list 出当前 cacheStore 所有 KV
//  3. 返回 list 出的全部 KV
func (i *informer) RegisterHandlerAndList(keyPrefix []string, handler ResourceEventHandler) (map[string]any, error) {
	i.mu.Lock() // <-- 冻结缓存 & 事件处理
	defer i.mu.Unlock()

	// 注册 handler
	i.dispatchByKeyPrefixHandler.registerHandler(keyPrefix, handler)

	// 在锁定下读取缓存, 拿到一致快照
	items := i.cacheStore.ListMap()
	return items, nil
}

func (i *informer) Run(ctx context.Context) {
	go i.processLoop()

	<-ctx.Done()

	i.deltaFIFO.Close()
	log.Info("informer stopped")

}

func (i *informer) Register(keyPrefix []string, handler ResourceEventHandler) {
	i.dispatchByKeyPrefixHandler.registerHandler(keyPrefix, handler)
}

func (i *informer) HasSynced() bool {
	return i.deltaFIFO.HasSynced()
}

func (i *informer) WaitForCacheSync(ctx context.Context) bool {
	err := wait.PollImmediateUntilWithContext(ctx, syncedPollPeriod,
		func(_ context.Context) (bool, error) {
			return i.HasSynced(), nil
		},
	)

	return err == nil
}

func (i *informer) processLoop() {
	for {
		key, obj, err := i.deltaFIFO.Pop(PopProcessFunc(i.process))
		if err != nil {
			if err == ErrFIFOClosed {
				return
			}
			// This is the safe way to re-enqueue.
			log.Debug("Pop with error re-enqueueing item", log.ErrorField(err))
			i.deltaFIFO.AddIfNotPresent(key, obj)
		}
	}
}

func processDeltas(
	// Object which receives event notifications from the given deltas
	handler ResourceEventHandler,
	cacheStore Store,
	deltas Deltas,
	isInInitialList bool,
) error {
	// from oldest to newest
	for _, d := range deltas {
		key := etcd.Key(d.Key)
		obj := d.Object

		log.Info("Processing key:", zap.String("key", key))
		log.Info("Processing with:", zap.String("move", string(d.Type)))

		switch d.Type {
		case Sync, Added, Updated: // delta type
			// if object exists, update it
			if old, exists, err := cacheStore.Get(key); err == nil && exists {
				if err := cacheStore.Update(key, obj); err != nil {
					return err
				}

				// call eventHandler.OnUpdate
				handler.OnUpdate(key, old, obj)
			} else {
				// if object does not exist, add it
				if err := cacheStore.Add(key, obj); err != nil {
					return err
				}

				handler.OnAdd(key, obj, isInInitialList)
			}
		case Deleted:
			// 在 store 里删除对象
			if err := cacheStore.Delete(key); err != nil {
				return err
			}

			handler.OnDelete(key, obj)
		}
	}
	return nil
}

func (i *informer) Get(key string) (any, bool, error) {
	return i.cacheStore.Get(key)
}

type ResourceEventHandler interface {
	OnAdd(key etcd.Key, obj any, isInInitialList bool)
	OnUpdate(key etcd.Key, oldObj, newObj any)
	OnDelete(key etcd.Key, obj any)
}

type dispatchByKeyPrefixHandler struct {
	// keyPrefix -> ResourceEventHandler
	handlers map[string][]ResourceEventHandler
}

func NewDispatchByKeyPrefixHandler() *dispatchByKeyPrefixHandler {
	return &dispatchByKeyPrefixHandler{
		handlers: make(map[string][]ResourceEventHandler),
	}
}

func (d *dispatchByKeyPrefixHandler) registerHandler(keyPrefixes []string, handler ResourceEventHandler) {
	for _, prefix := range keyPrefixes {
		d.handlers[prefix] = append(d.handlers[prefix], handler)
	}
}

func (d *dispatchByKeyPrefixHandler) OnAdd(key etcd.Key, obj any, isInInitialList bool) {
	// prefix := etcd.Dir(key)
	//
	//	if hs, ok := d.handlers[prefix]; ok {
	//		for _, h := range hs {
	//			h.OnAdd(key, obj, isInInitialList)
	//		}
	//	}
	keyStr := string(key)
	for prefix, hs := range d.handlers {
		if strings.HasPrefix(keyStr, prefix) {
			for _, h := range hs {
				h.OnAdd(key, obj, isInInitialList)
			}
		}
	}
}

func (d *dispatchByKeyPrefixHandler) OnUpdate(key etcd.Key, oldObj, newObj any) {
	// prefix := etcd.Dir(key)
	//
	//	if hs, ok := d.handlers[prefix]; ok {
	//		for _, h := range hs {
	//			h.OnUpdate(key, oldObj, newObj)
	//		}
	//	}
	keyStr := string(key)
	for prefix, hs := range d.handlers {
		if strings.HasPrefix(keyStr, prefix) {
			for _, h := range hs {
				h.OnUpdate(key, oldObj, newObj)
			}
		}
	}
}

func (d *dispatchByKeyPrefixHandler) OnDelete(key etcd.Key, obj any) {
	// 删除同样用 etcd.Dir 保持一致
	//	prefix := etcd.Dir(key)
	//	if hs, ok := d.handlers[prefix]; ok {
	//		for _, h := range hs {
	//			h.OnDelete(key, obj)
	//		}
	//	}
	keyStr := string(key)
	for prefix, hs := range d.handlers {
		if strings.HasPrefix(keyStr, prefix) {
			for _, h := range hs {
				h.OnDelete(key, obj)
			}
		}
	}
}
