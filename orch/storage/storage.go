package storage

import (
	"context"
	"xdp-banner/orch/storage/agent"
	"xdp-banner/orch/storage/agent/applied"
	agentnode "xdp-banner/orch/storage/agent/node"
	"xdp-banner/orch/storage/agent/rule"
	"xdp-banner/orch/storage/agent/strategy"
	"xdp-banner/orch/storage/orch"
	"xdp-banner/orch/storage/orch/cert"
	orchnode "xdp-banner/orch/storage/orch/node"
	"xdp-banner/pkg/etcd"
)

type Storage struct {
	Rule rule.Storage
	Cert cert.Storage

	Orch     orch.Storage
	OrchInfo orchnode.InfoStorage

	Agent              agent.Storage
	AgentInfo          agentnode.InfoStorage
	AgentStatus        agentnode.StatusStorage
	AgentRegisteration agentnode.RegisterStorage

	Strategy strategy.Storage
	Applied  applied.Storage
}

func New(ctx context.Context, client etcd.Client) Storage {
	return Storage{
		Rule: rule.New(client),
		Cert: cert.New(client),

		Orch:     orch.New(client),
		OrchInfo: orchnode.NewInfoStorage(client),

		Agent:              agent.New(client),
		AgentInfo:          agentnode.NewInfoStorage(client),
		AgentStatus:        agentnode.NewStatusStorage(ctx, client),
		AgentRegisteration: agentnode.NewRegisterStorage(client),

		Strategy: strategy.New(client),
		Applied:  applied.New(client),
	}
}
