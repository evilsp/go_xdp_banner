package node

import (
	"fmt"
	"net"
	"os"
)

const (
	DefaultLocalNodeID int    = 0
	DefaultListenIP    string = "0.0.0.0"
	DefaultListenPort  int    = 6061
)

func Name() (string, error) {
	return os.Hostname()
}

func DefaultIP() (net.IP, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	for _, iface := range interfaces {
		// 忽略无效接口和非活动接口
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			return nil, fmt.Errorf("failed to get addresses for interface %s: %w", iface.Name, err)
		}

		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil {
					return ipNet.IP, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("no suitable IP address found")
}

func DefaultListenAddr() string {
	return fmt.Sprintf("%s:%d", DefaultListenIP, DefaultListenPort)
}
