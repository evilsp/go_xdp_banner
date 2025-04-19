package etcd

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"xdp-banner/pkg/log"
)

// CacheItem 表示缓存中单个条目的结构体
type CacheItem struct {
	Value        string    // etcd中的值
	Revision     int64     // Key在etcd中的修订版本
	LastAccess   time.Time // 最近一次访问或更新的时间
	RefreshCount int64     // 连续相同更新的次数（用于触发强制刷新）
}

type WriteThroughCache interface {
	Get(ctx context.Context, key Key, force bool) (string, int64, error)
	Update(ctx context.Context, key, newValue string) error

	Close() error
}

// NewEtcdCache 创建并初始化EtcdCache
//
// 参数说明：
//   - endpoints: etcd集群节点的地址数组，例如[]string{"127.0.0.1:2379"}
//   - forceRefresh: 连续多少次相同值更新后，仍执行一次强制Put
//   - ttl: 缓存条目在无访问后可存活的时间
//   - cleanupPeriod: 执行过期扫描的周期
func NewWriteThroughCache(cli Client, forceRefresh int64, ttl, cleanupPeriod time.Duration) WriteThroughCache {
	c := &writeThroughCache{
		client:        cli,
		forceRefresh:  forceRefresh,
		ttl:           ttl,
		cleanupPeriod: cleanupPeriod,
		cache:         make(map[string]*CacheItem),
		closeCh:       make(chan struct{}),
	}

	go c.cleanupExpiredLoop()

	return c
}

// writeThroughCache
// 主要用于缓存etcd中数据，减少对etcd的频繁访问
// 非一致性缓存，不保证缓存数据与etcd中完全一致
// 提供了过期清理机制，定期清理本地缓存。
// 当缓存不存在时，自动从etcd获取并写入缓存。
// 适用于对数据一致性要求不高的场景
type writeThroughCache struct {
	client Client

	// 强制刷新阈值：当连续多次相同更新时，仍然执行一次Put
	forceRefresh int64
	// 缓存每条记录的过期时间
	ttl time.Duration
	// 过期清理执行周期
	cleanupPeriod time.Duration

	// 缓存本身 + 读写锁
	mu    sync.RWMutex
	cache map[string]*CacheItem

	// 用于停止后台协程
	closeCh   chan struct{}
	closeOnce sync.Once
}

// Close 用于在程序退出前关闭相关资源
func (c *writeThroughCache) Close() error {
	// 保证只执行一次
	c.closeOnce.Do(func() {
		close(c.closeCh) // 结束后台协程
	})
	return nil
}

// cleanupExpiredLoop 定期清理过期的缓存数据
func (c *writeThroughCache) cleanupExpiredLoop() {
	ticker := time.NewTicker(c.cleanupPeriod)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			now := time.Now()
			c.mu.Lock()
			for key, item := range c.cache {
				if now.Sub(item.LastAccess) > c.ttl {
					log.Info(fmt.Sprintf("delete %s from cache", key))
					delete(c.cache, key)
				}
			}
			c.mu.Unlock()

		case <-c.closeCh:
			return
		}
	}
}

// GetValue 尝试先从缓存获取对应key的value和revision；
// 若缓存中无则从etcd获取，并写入缓存。 force 强制刷新
func (c *writeThroughCache) Get(ctx context.Context, key Key, force bool) (string, int64, error) {
	item, err := c.get(ctx, key, force)
	if err != nil {
		return "", 0, err
	}

	if item == nil {
		return "", 0, nil
	}

	return item.Value, item.Revision, nil
}

func (c *writeThroughCache) get(ctx context.Context, key Key, force bool) (item *CacheItem, err error) {
	if !force {
		item = c.getFromCache(key)
		if item != nil {
			return item, nil
		}
	}

	// not in cache => get from etcd
	item, err = c.getFromEtcd(ctx, key)
	if err != nil {
		return item, err
	}

	if item == nil {
		return nil, nil
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = item

	return item, nil
}

func (c *writeThroughCache) getFromCache(key Key) *CacheItem {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if item, ok := c.cache[key]; ok {
		item.LastAccess = time.Now()

		return item
	}

	return nil

}

// getFromEtcd 从etcd获取数据, 如果不存在返回默认值
func (c *writeThroughCache) getFromEtcd(ctx context.Context, key Key) (*CacheItem, error) {
	resp, err := c.client.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("etcd get error: %v", err)
	}

	if len(resp.Kvs) == 0 {
		return nil, nil
	}

	kv := resp.Kvs[0]
	valStr := string(kv.Value)
	rev := kv.ModRevision

	return &CacheItem{
		Value:      valStr,
		Revision:   rev,
		LastAccess: time.Now(),
	}, nil

}

func (c *writeThroughCache) skipUpdate(item *CacheItem, newValue string) bool {
	if item == nil {
		return false
	} else if item.Revision == 0 {
		return false
	} else if item.Value != newValue {
		return false
	}

	return item.RefreshCount <= c.forceRefresh
}

// Update 对key设置新值；
// 若新值与缓存中的相同，则仅在连续多次相同值后执行强制刷新；
// 若值有变，则立即写入etcd并更新缓存。
func (c *writeThroughCache) Update(ctx context.Context, key, newValue string) error {
	item, err := c.get(ctx, key, false)
	if err != nil {
		return err
	}

	if c.skipUpdate(item, newValue) {
		atomic.AddInt64(&item.RefreshCount, 1)
		return nil
	}

	// 准备Put
	putResp, err := c.client.Put(ctx, key, newValue)
	if err != nil {
		return fmt.Errorf("etcd put error: %v", err)
	}

	if item == nil {
		item = &CacheItem{}
	}

	// 更新本地缓存
	c.mu.Lock()
	defer c.mu.Unlock()

	item.Value = newValue
	item.Revision = putResp.Header.Revision
	item.LastAccess = time.Now()
	item.RefreshCount = 0

	c.cache[key] = item

	return nil
}
