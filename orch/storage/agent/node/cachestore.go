package node

import (
	"context"
	"sync/atomic"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/wait"

	"go.etcd.io/etcd/api/v3/mvccpb"
)

type CacheStore struct {
	cache     informer.ThreadSafeStore
	reflector *informer.Reflector
	synced    atomic.Bool
}

func newCacheStore(ctx context.Context, client etcd.Client) *CacheStore {
	store := informer.NewThreadSafeStore()

	lw := newLeaseConvertListerWatcher(client)
	reflector := informer.NewReflector(lw, "status_cache_store", EtcdDirStatus, store)

	c := &CacheStore{
		cache:     store,
		reflector: reflector,
	}

	go c.run(ctx)

	return c
}

func (c *CacheStore) run(ctx context.Context) {
	go c.reflector.Run(ctx)
	_ = wait.PollImmediateUntilWithContext(ctx, 100*time.Millisecond,
		func(_ context.Context) (bool, error) {
			return c.cache.HasSynced(), nil
		},
	)
	c.synced.Store(true)
}

func (c *CacheStore) waitForSync(ctx context.Context) {
	if c.synced.Load() {
		return
	}

	wait.PollUntilContextTimeout(ctx, 100*time.Millisecond, 5*time.Second, false, func(ctx context.Context) (bool, error) {
		return c.synced.Load(), nil
	})
}

func (c *CacheStore) Get(ctx context.Context, key etcd.Key) (any, bool, error) {
	c.waitForSync(ctx)
	return c.cache.Get(key)
}

func (c *CacheStore) Range(yield func(etcd.Key, any) bool) {
	for k, v := range c.cache.Range {
		if !yield(k, v) {
			return
		}
	}
}

type leaseValue struct {
	value   string
	leaseID int64
}

func leaseConvert(key etcd.Key, value *mvccpb.KeyValue) (etcd.Key, any) {
	lv := leaseValue{
		value:   string(value.Value),
		leaseID: int64(value.Lease),
	}

	return etcd.Key(key), lv
}

type leaseConvertListerWatcher struct {
	client  etcd.Client
	convert etcd.ItemConvert
}

func newLeaseConvertListerWatcher(client etcd.Client) etcd.ListerWatcher {
	return &leaseConvertListerWatcher{
		client:  client,
		convert: leaseConvert,
	}
}

func (l *leaseConvertListerWatcher) Watch(opt etcd.WatchOption) (etcd.WatchController, error) {
	opt.Convert = leaseConvert
	return l.client.Watch(opt)
}

func (l *leaseConvertListerWatcher) List(ctx context.Context, opt etcd.ListOption) (etcd.PagedList, error) {
	opt.Convert = leaseConvert
	return l.client.List(ctx, opt)
}
