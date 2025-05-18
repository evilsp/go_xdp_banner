package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"xdp-banner/agent/ebpf"
	"xdp-banner/agent/ebpf/xdp"
	"xdp-banner/agent/internal/client"
	"xdp-banner/api/orch/v1/rule"
	"xdp-banner/pkg/log"
	model "xdp-banner/pkg/rule"

	"github.com/looplab/fsm"
	"go.uber.org/zap"
)

type controller struct {
	client    client.Client
	ctx       context.Context
	cancelCtx context.CancelFunc
	xdpMap    *xdp.BannedIPXdpMap
	xdpProg   *xdp.XdpProgManager
	attached  bool       // 标记是否已附加到接口
	attachIf  []string   // 记录附加的接口名
	mu        sync.Mutex // 保护并发访问
	wg        sync.WaitGroup
}

// Global Controller Ctx
var controllerCtx controller

func initControllerCtx(client client.Client) *controller {

	ctx, cancel := context.WithCancel(context.Background())
	controllerCtx = controller{
		client:    client,
		ctx:       ctx,
		cancelCtx: cancel,
		xdpMap:    nil,
		xdpProg:   nil,
		attached:  false,
		attachIf:  nil,
	}
	return &controllerCtx
}

func (c *controller) Start(ctx context.Context, e *fsm.Event) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	configName := e.Args[0].(string)
	log.Info("Starting XDP controller", log.StringField("config", configName))
	defer func() {
		if err == nil {
			client.SetConfigName(configName)
		}
	}()

	if c.cancelCtx != nil {
		c.cancelCtx()
	}

	c.ctx, c.cancelCtx = context.WithCancel(context.Background())

	if c.xdpMap == nil || c.xdpProg == nil {
		c.xdpMap, c.xdpProg, c.attachIf, err = ebpf.Init()
		if err != nil {
			return fmt.Errorf("ebpf init failed: %w", err)
		}
	}
	c.attached = true

	c.wg.Add(1)
	go c.watchRules(configName)

	return nil
}

func (c *controller) Stop(ctx context.Context, e *fsm.Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Info("Stopping XDP controller")

	if !c.attached {
		return nil // 已经处于停止状态
	}

	//	for _, ifaceName := range c.attachIf {
	//		if err := c.xdpProg.Detach(ifaceName); err != nil {
	//			return fmt.Errorf("detach failed: %w", err)
	//		}
	//	}

	// cancel 掉 watchRules goroutine
	if c.cancelCtx != nil {
		c.cancelCtx()
	}

	if err := c.xdpProg.Close(); err != nil {
		return err
	}

	if err := xdp.ClearMap(); err != nil {
		return err
	}

	c.xdpProg = nil
	c.xdpMap = nil
	c.attached = false

	return nil
}

func (c *controller) Reload(ctx context.Context, e *fsm.Event) (err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	log.Info("Stopping XDP controller")

	if c.attached {
		// 清理阶段
		for _, iface := range c.attachIf {
			_ = c.xdpProg.Detach(iface) // 忽略 error 或累计报告
		}
		if err := c.xdpProg.Close(); err != nil {
			c.mu.Unlock()
			return err
		}
	}

	//	for _, ifaceName := range c.attachIf {
	//		if err := c.xdpProg.Detach(ifaceName); err != nil {
	//			return fmt.Errorf("detach failed: %w", err)
	//		}
	//	}

	// cancel 掉 watchRules goroutine
	if c.cancelCtx != nil {
		c.cancelCtx()
	}

	if err := c.xdpProg.Close(); err != nil {
		return err
	}

	if err := xdp.ClearMap(); err != nil {
		return err
	}

	c.xdpProg = nil
	c.xdpMap = nil
	c.attached = false

	configName := e.Args[0].(string)
	log.Info("Reloading XDP controller", log.StringField("config", configName))
	defer func() {
		if err == nil {
			client.SetConfigName(configName)
		}
	}()

	c.ctx, c.cancelCtx = context.WithCancel(context.Background())
	c.xdpMap, c.xdpProg, c.attachIf, err = ebpf.Init()
	if err != nil {
		return fmt.Errorf("ebpf init failed: %w", err)
	}
	c.attached = true

	c.wg.Add(1)
	go c.watchRules(configName)

	return nil
}

// watchRules 根据给定的 configName, 用 c.ctx 监听服务器下发的 WatchRuleResponse
func (c *controller) watchRules(configName string) {
	defer c.wg.Done()

	ruleChan := make(chan *rule.WatchRuleResponse, 100)
	// 第一次拉取并开启 stream

	go func() {
		if err := c.client.GetRule(c.ctx, configName, ruleChan); err != nil {
			log.Error("GetRule failed", zap.Error(err))
		}
		close(ruleChan)
	}()

	for {
		select {
		case resp, ok := <-ruleChan:
			if !ok {
				// 服务器 stream 关闭
				return
			}
			c.handleRuleEvent(resp)
		case <-c.ctx.Done():
			// 上层 cancelContext() 被调用
			return
		}
	}
}

// handleRuleEvent 负责把 WatchRuleResponse 转成 IPRule 并打到 eBPF map
func (c *controller) handleRuleEvent(resp *rule.WatchRuleResponse) {
	cidr, proto, sport, dport, err := xdp.ParseIPRuleKey(resp.RuleKey)
	if err != nil {
		log.Error("parse RuleKey failed", zap.Error(err))
		return
	}

	// RuleVal 反序列成本地模型
	m := resp.RuleVal.AsMap()
	data, _ := json.Marshal(m)
	var meta model.RuleMeta
	if err := json.Unmarshal(data, &meta); err != nil {
		log.Error("unmarshal RuleVal failed", zap.Error(err))
		return
	}

	ipRule := xdp.IPRule{
		CIDR:           cidr,
		BannedProtocol: proto,
		Sport:          sport,
		Dport:          dport,
		Identity:       meta.Identity,
	}

	switch resp.EventType {
	case 0:
		if err := c.xdpMap.AddCIDRRule(ipRule); err != nil {
			log.Error("AddCIDRRule failed", zap.Error(err))
		}
	case 1:
		if err := c.xdpMap.RemoveCIDRRule(ipRule); err != nil {
			log.Error("RemoveCIDRRule failed", zap.Error(err))
		}
	}
}

func ErrorWrapper(op func(context.Context, *fsm.Event) error) func(context.Context, *fsm.Event) {
	return func(ctx context.Context, e *fsm.Event) {
		if err := op(ctx, e); err != nil {
			e.Cancel(err)
		}
	}
}
