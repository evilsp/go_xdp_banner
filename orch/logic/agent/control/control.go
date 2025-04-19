package control

import "xdp-banner/orch/storage/agent/node"

type Control struct {
	registers node.RegisterStorage
	infos     node.InfoStorage
	statuss   node.StatusStorage
}

func New(rs node.RegisterStorage, is node.InfoStorage, ss node.StatusStorage) *Control {
	return &Control{
		registers: rs,
		infos:     is,
		statuss:   ss,
	}
}
