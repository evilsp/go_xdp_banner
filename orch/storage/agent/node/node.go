package node

import (
	"xdp-banner/orch/storage/agent"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDir = etcd.Join(agent.EtcdDir, "node/")
)
