package informer

import (
	"context"
	"fmt"
	"sync"
	"time"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"

	"go.uber.org/zap"
)

type InformerRegistry struct {
	mu  sync.RWMutex
	inf map[string]Informer
}

var GlobalInformerRegistry = &InformerRegistry{
	inf: make(map[string]Informer),
}

// Get 返回指定前缀的 informer（如果存在）
func (r *InformerRegistry) Get(prefix string) (Informer, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	i, ok := r.inf[prefix]
	return i, ok
}

// Add 保存 informer 到注册表
func (r *InformerRegistry) Add(prefix string, inf Informer) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.inf[prefix] = inf
}

// Add 保存 informer 到注册表
func (r *InformerRegistry) Delete(prefix string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.inf, prefix)
}

// createInformerForPrefix 根据指定 prefix 创建一个新的 informer 实例
func CreateInformerForPrefix(ctx context.Context, client etcd.Client, prefix string) (Informer, error) {

	// 一定要創建至少一個 Reflector
	deltaFIFO := NewDeltaFIFOWithWait(1)
	reflector := NewReflector(client, "rule_reflector", prefix, deltaFIFO)

	go reflector.Run(ctx)

	inf := New(deltaFIFO)

	go inf.Run(ctx)

	syncCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if !inf.WaitForCacheSync(syncCtx) {
		return nil, fmt.Errorf("informer cache sync 超时")
	}
	log.Info("createInformerForPrefix: 已启动 informer", zap.String("prefix", prefix))
	return inf, nil
}
