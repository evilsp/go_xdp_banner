package xdp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

type XdpProgManager struct {
	program *ebpf.Program
	links   map[string]link.Link // 按接口名存储多个链接
	mu      sync.Mutex
}

func NewXdpProgManager() (*XdpProgManager, error) {
	// 1) 确保 pin 目录存在
	pinDir := filepath.Join("/sys/fs/bpf", "xdp_banner")
	if err := os.MkdirAll(pinDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create pin dir %q: %w", pinDir, err)
	}

	// 2) 加载 eBPF 程序 spec
	spec, err := loadXdp()
	if err != nil {
		return nil, fmt.Errorf("加载 eBPF 规范失败: %w", err)
	}

	// 3) 指定 map pin path，让 loader 去复用已经 pin 的 maps
	opts := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinDir,
		},
	}

	// 4) 只加载 cil_xdp_entry 程序
	var objs xdpPrograms
	if err := spec.LoadAndAssign(&objs, &opts); err != nil {
		return nil, fmt.Errorf("加载 XDP 程序失败: %w", err)
	}

	if objs.CilXdpEntry == nil {
		return nil, fmt.Errorf("找不到 cil_xdp_entry 程序")
	}

	return &XdpProgManager{
		program: objs.CilXdpEntry,
		links:   make(map[string]link.Link),
	}, nil
}

// Attach 将 XDP 程序附加到指定的网络设备
//
// 参数:
//   - ifaceName: 网络接口名 (如 "eth0")
//   - flags: 附加标志 (unix.XDP_FLAGS_*)
//   - XDPGenericMode 2
//   - XDPDriverMode 4
//   - XDPOffloadMode 8
//
// 返回:
//   - error: 错误信息
func (m *XdpProgManager) Attach(ifaceName string, flags link.XDPAttachFlags) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return err
	}

	l, err := link.AttachXDP(link.XDPOptions{
		Program:   m.program,
		Interface: iface.Index,
		Flags:     flags,
	})
	if err != nil {
		return err
	}

	// 关闭旧链接（如果存在）
	if old, ok := m.links[ifaceName]; ok {
		old.Close()
	}

	m.links[ifaceName] = l
	return nil
}

func (m *XdpProgManager) Detach(ifaceName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if l, ok := m.links[ifaceName]; ok {
		delete(m.links, ifaceName)
		return l.Close()
	}
	return nil
}

// Close 释放所有资源
func (m *XdpProgManager) Close() error {
	// 先分离程序

	for ifaceName, _ := range m.links {
		m.Detach(ifaceName)
	}

	// 关闭程序
	if m.program != nil {
		if err := m.program.Close(); err != nil {
			return fmt.Errorf("关闭 XDP 程序失败: %w", err)
		}
	}

	return nil
}
