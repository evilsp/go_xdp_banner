package logic

import (
	"xdp-banner/orch/logic/agent/control"
	"xdp-banner/orch/logic/agent/report"
	"xdp-banner/orch/logic/orch"
	"xdp-banner/orch/logic/rulecenter"
	"xdp-banner/orch/logic/strategy"
	"xdp-banner/orch/storage"
)

type Logic struct {
	Control      *control.Control
	Report       *report.Report
	ConfigCenter *rulecenter.RuleCenter
	Orch         *orch.Orch
	Strategy     *strategy.Strategy
	Applied      *strategy.Applied
}

func New(s storage.Storage) *Logic {
	cc := rulecenter.New(s.Rule)
	ctrl := control.New(s.AgentRegisteration, s.AgentInfo, s.AgentStatus)
	report := report.New(s.AgentStatus)
	orch := orch.New(s.OrchInfo)
	applied := strategy.NewApplied(s.Strategy, s.AgentInfo, s.Applied)
	strategy := strategy.NewStrategy(s.Strategy)

	return &Logic{
		Control:      ctrl,
		Report:       report,
		ConfigCenter: cc,
		Orch:         orch,
		Strategy:     strategy,
		Applied:      applied,
	}
}
