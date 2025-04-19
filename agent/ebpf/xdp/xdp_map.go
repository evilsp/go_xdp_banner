package xdp

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"xdp-banner/agent/ebpf/xdp/types"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
)

// Reference： https://pkg.go.dev/github.com/cilium/ebpf#Map

func htons(v uint16) uint16 {
	return (v<<8)&0xff00 | v>>8
}

type IPRule struct {
	CIDR           string
	Identity       string
	BannedProtocol uint8 // IPPROTO_ICMP：1 IPPROTO_TCP：6 IPPROTO_UDP：17
	Sport          uint16
	Dport          uint16
}

// BannedIPXdpMap 主结构体
type BannedIPXdpMap struct {
	maps *xdpMaps
	// Current Config
	mu sync.Mutex
}

const (
	// LPM-Trie key 的静态前缀长度（bits），对于 ipcache_key 是 8*(4 byte pad) == 32 bits
	IPCACHE_STATIC_PREFIX_BITS = 32

	BANLIST_L3_FULL  = 64 // protocol + identity
	BANLIST_L4_SPORT = 80 // + sport
	BANLIST_L4_DPORT = 96 // + dport
	BANLIST_L4_FULL  = 96 // = protocol+identity+sport+dport
)

// NewBannedIPXdpMap 创建新的封禁管理器
func NewBannedIPXdpMap() (*BannedIPXdpMap, error) {
	// 1) 解除 memlock 限制
	if err := rlimit.RemoveMemlock(); err != nil {
		return nil, fmt.Errorf("failed to remove memlock limit: %w", err)
	}

	// 2) 确保 BPF FS 目录存在
	pinDir := filepath.Join("/sys/fs/bpf", "xdp_banner")
	if err := os.MkdirAll(pinDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create pin directory %q: %w", pinDir, err)
	}

	// 3) 准备 pin 路径选项
	opts := ebpf.CollectionOptions{
		Maps: ebpf.MapOptions{
			PinPath: pinDir,
		},
	}

	// 4) 加载并 pin eBPF maps
	maps := xdpMaps{}
	if err := loadXdpObjects(&maps, &opts); err != nil {
		return nil, fmt.Errorf("failed to load eBPF maps: %w", err)
	}

	// 5) 返回封装好的管理器
	return &BannedIPXdpMap{
		maps: &maps,
	}, nil
}

// addCIDRRule 添加/更新 CIDR 规则
// type xdpBanruleKey struct {
//	Prefixlen uint32
//	Pad1     uint16
//	Pad2     uint8
//	Protocol uint8
//	Identity uint32
//	Sport    uint16
//	Dport    uint16
//}

// type xdpIdentityInfo struct{ Identity uint32 }

//	type xdpIpcacheKey struct {
//		Prefixlen uint32
//		Pad1   uint16
//		Pad2   uint8
//		Family uint8
//		// represents both IPv6 and IPv4 (in the lowest four bytes)
//		IP types.IPv6
//	}
func (b *BannedIPXdpMap) AddCIDRRule(rule IPRule) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1) parse CIDR
	ipAddr, ipNet, err := net.ParseCIDR(rule.CIDR)
	if err != nil {
		return fmt.Errorf("invalid CIDR %q: %w", rule.CIDR, err)
	}
	ones, _ := ipNet.Mask.Size()
	prefixLen := uint32(ones)

	// 2) 构造 ipcache_key
	var ipKey xdpIpcacheKey
	ipKey.Prefixlen = IPCACHE_STATIC_PREFIX_BITS + prefixLen

	if ip4 := ipAddr.To4(); ip4 != nil {
		ipKey.Family = types.AF_INET
		copy(ipKey.IP[:4], ip4)
	} else {
		ipKey.Family = types.AF_INET6
		// 比 ParseAddr(ip.String()) + FromAddr 更简洁：
		ip16 := ipAddr.To16()
		if ip16 == nil {
			return fmt.Errorf("bad IPv6 %q", ipAddr)
		}
		// 假设 xdpIpcacheKey.IP 底层是 [16]byte
		copy(ipKey.IP[:], ip16)
	}

	// 3) 解析 identity
	idVal, err := strconv.ParseUint(rule.Identity, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid identity %q: %w", rule.Identity, err)
	}
	identInfo := xdpIdentityInfo{Identity: uint32(idVal)}

	// 4) 更新 identity map
	if err := b.maps.IdentityIpcache.Update(ipKey, identInfo, ebpf.UpdateAny); err != nil {
		return fmt.Errorf("update identity_ipcache failed: %w", err)
	}

	// 5) 构造 banrule_key
	var banKey xdpBanruleKey
	banKey.Protocol = rule.BannedProtocol
	banKey.Identity = identInfo.Identity
	banKey.Sport = htons(rule.Sport)
	banKey.Dport = htons(rule.Dport)

	switch rule.BannedProtocol {
	case types.IPPROTO_TCP, types.IPPROTO_UDP:
		switch {
		case rule.Sport != 0 && rule.Dport != 0:
			banKey.Prefixlen = BANLIST_L4_FULL
		case rule.Sport == 0 && rule.Dport != 0:
			banKey.Prefixlen = BANLIST_L4_DPORT
		case rule.Sport != 0 && rule.Dport == 0:
			banKey.Prefixlen = BANLIST_L4_SPORT
		default:
			// 4) 粗粒度 L3，只按 protocol+identity
			banKey.Prefixlen = BANLIST_L3_FULL
		}
	default:
		// ICMP 之类用粗粒度 L3
		banKey.Prefixlen = BANLIST_L3_FULL
	}

	// 6) 更新 banlist map
	var banVal xdpBanruleVal
	if err := b.maps.XdpBannerBanlist.Update(banKey, banVal, ebpf.UpdateAny); err != nil {
		return fmt.Errorf("update xdp_banner_banlist failed: %w", err)
	}

	return nil
}

// removeCIDRRule 只删除与给定 IPRule 完全匹配的那一条 banlist 规则
func (b *BannedIPXdpMap) RemoveCIDRRule(rule IPRule) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. 先把 rule.Identity 转成 uint32
	idVal, err := strconv.ParseUint(rule.Identity, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid identity %q: %w", rule.Identity, err)
	}
	identity := uint32(idVal)

	// 2. 构造要删的 banrule key
	var banKey xdpBanruleKey
	banKey.Identity = identity
	banKey.Protocol = rule.BannedProtocol

	banKey.Sport = htons(rule.Sport)
	banKey.Dport = htons(rule.Dport)

	switch rule.BannedProtocol {
	case types.IPPROTO_TCP, types.IPPROTO_UDP:
		switch {
		case rule.Sport != 0 && rule.Dport != 0:
			banKey.Prefixlen = BANLIST_L4_FULL
		case rule.Sport == 0 && rule.Dport != 0:
			banKey.Prefixlen = BANLIST_L4_DPORT
		case rule.Sport != 0 && rule.Dport == 0:
			banKey.Prefixlen = BANLIST_L4_SPORT
		default:
			// 4) 粗粒度 L3，只按 protocol+identity
			banKey.Prefixlen = BANLIST_L3_FULL
		}
	default:
		// ICMP 之类用粗粒度 L3
		banKey.Prefixlen = BANLIST_L3_FULL
	}

	// 3. 调用 eBPF map 的 Delete
	if err := b.maps.XdpBannerBanlist.Delete(banKey); err != nil {
		return fmt.Errorf("failed to delete banlist rule for %+v: %w", banKey, err)
	}

	return nil
}

// clearAllMaps 把这两个 LPM‐Trie map 完全清空
func (b *BannedIPXdpMap) ClearAllMaps() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1) 清 empty identity map
	iter1 := b.maps.IdentityIpcache.Iterate()
	var k1 xdpIpcacheKey
	for iter1.Next(&k1, nil) {
		if err := b.maps.IdentityIpcache.Delete(k1); err != nil {
			return fmt.Errorf("clear identity_ipcache failed at key %+v: %w", k1, err)
		}
	}

	// 2) 清 empty banlist map
	iter2 := b.maps.XdpBannerBanlist.Iterate()
	var k2 xdpBanruleKey
	for iter2.Next(&k2, nil) {
		if err := b.maps.XdpBannerBanlist.Delete(k2); err != nil {
			return fmt.Errorf("clear xdp_banner_banlist failed at key %+v: %w", k2, err)
		}
	}

	return nil
}

func (b *BannedIPXdpMap) Close() error {

	return b.maps.Close()
}
