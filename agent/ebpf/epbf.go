package ebpf

import (
	"fmt"
	"log"
	"net"

	"xdp-banner/agent/ebpf/xdp"
)

func Init() (*xdp.BannedIPXdpMap, *xdp.XdpProgManager, []string, error) {
	// 1. 初始化 XDP 封禁映射
	mapInstance, err := xdp.NewBannedIPXdpMap()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create XDP map: %w", err)
	}

	// 2. 初始化 XDP 程序管理器
	progInstance, err := xdp.NewXdpProgManager()
	if err != nil {
		mapInstance.Close()
		return nil, nil, nil, fmt.Errorf("failed to create XDP program manager: %w", err)
	}

	// 3. 获取并过滤网络接口
	interfaces, err := getAttachableInterfaces()
	if err != nil {
		mapInstance.Close()
		progInstance.Close()
		return nil, nil, nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	if len(interfaces) == 0 {
		mapInstance.Close()
		progInstance.Close()
		return nil, nil, nil, fmt.Errorf("no suitable network interfaces found")
	}

	// 4. 在所有接口上挂载XDP程序
	attachedInterfaces := make([]string, 0, len(interfaces))
	for _, iface := range interfaces {
		// Use Generic Mode
		if err := progInstance.Attach(iface.Name, 2); err != nil {
			log.Printf("Failed to attach to interface %s: %v", iface.Name, err)
			continue
		}
		attachedInterfaces = append(attachedInterfaces, iface.Name)
	}

	if len(attachedInterfaces) == 0 {
		mapInstance.Close()
		progInstance.Close()
		return nil, nil, nil, fmt.Errorf("failed to attach to any interface")
	}

	log.Printf("Successfully attached to interfaces: %v", attachedInterfaces)

	return mapInstance, progInstance, attachedInterfaces, nil
}

func getAttachableInterfaces() ([]net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	var result []net.Interface
	for _, iface := range interfaces {
		// 跳过回环接口和未启用接口
		if iface.Flags&net.FlagLoopback != 0 || iface.Flags&net.FlagUp == 0 {
			continue
		}

		result = append(result, iface)
	}

	return result, nil
}
