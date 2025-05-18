package xdp

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"xdp-banner/agent/ebpf/xdp/types"
)

// ipv4PrefixLength 计算IPv4前缀长度
func ipv4PrefixLength(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

// ipv6PrefixLength 计算IPv6前缀长度
func ipv6PrefixLength(mask net.IPMask) int {
	ones, _ := mask.Size()
	return ones
}

// parseIPRuleKey 从 ruleKey 提取 CIDR、Protocol、Sport、Dport
func ParseIPRuleKey(ruleKey string) (cidr string, proto uint8, sport, dport uint16, err error) {
	// 1. 去掉末尾 "/" 避免最后出现空字符串
	ruleKey = strings.TrimSuffix(ruleKey, "/")
	parts := strings.Split(ruleKey, "/")

	// 期望至少 8 段: ["", "agent", "rule", "myrule", "192.168.0.1", "24", "6", "111-80"]
	if len(parts) < 8 {
		return "", 0, 0, 0, fmt.Errorf("ruleKey 分段不足, got: %v", parts)
	}

	// 2. 组合 CIDR
	ipPart := parts[4]             // "192.168.0.1"
	maskPart := parts[5]           // "24"
	cidr = ipPart + "/" + maskPart // "192.168.0.1/24"

	// 3. 协议
	protoStr := parts[6] // "6"
	switch protoStr {
	case "TCP", "tcp":
		proto = types.IPPROTO_TCP
	case "UDP", "udp":
		proto = types.IPPROTO_UDP
	case "ICMP", "icmp":
		proto = types.IPPROTO_ICMP
	default:
		return "", 0, 0, 0, fmt.Errorf("无法解析协议号 %q", protoStr)
	}

	// 4. 端口范围
	portRange := parts[7] // "111-80"
	portParts := strings.Split(portRange, "-")
	if len(portParts) != 2 {
		return "", 0, 0, 0, fmt.Errorf("端口范围格式不正确: %q", portRange)
	}
	sportInt, err := strconv.Atoi(portParts[0])
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("无法解析sport: %w", err)
	}
	dportInt, err := strconv.Atoi(portParts[1])
	if err != nil {
		return "", 0, 0, 0, fmt.Errorf("无法解析dport: %w", err)
	}
	sport = uint16(sportInt)
	dport = uint16(dportInt)

	return cidr, proto, sport, dport, nil
}

// WaitForInterrupt 等待中断信号
func (b *BannedIPXdpMap) WaitForInterrupt() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	log.Println("Received interrupt signal, shutting down...")
	b.Close()
}

func ClearMap() error {
	targets := []string{
		"identity_ipcache",
		"xdp_banner_banlist",
	}

	for _, name := range targets {
		fullPath := filepath.Join("/sys/fs/bpf/xdp_banner", name)

		if _, err := os.Stat(fullPath); err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("stat %s: %w", fullPath, err)
		}

		if err := os.Remove(fullPath); err != nil {
			return fmt.Errorf("remove %s: %w", fullPath, err)
		}
		log.Printf("removed %s", fullPath)
	}
	return nil
}
