package node

import (
	"xdp-banner/orch/storage/orch"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDir = etcd.Join(orch.EtcdDir, "node/")
)
