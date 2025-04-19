package service

import (
	"xdp-banner/orch/logic"
	"xdp-banner/orch/service/agent/control"
	"xdp-banner/orch/service/agent/report"
	"xdp-banner/orch/service/auth"
	"xdp-banner/orch/service/orch"
	"xdp-banner/orch/service/rule"
	"xdp-banner/orch/service/strategy"
	"xdp-banner/pkg/server"
)

func New(logic *logic.Logic) server.Services {
	control := control.NewControlService(logic.Control)
	report := report.NewReportService(logic.Report)
	rule := rule.NewRuleService(logic.ConfigCenter)
	orch := orch.NewOrchService(logic.Orch)
	auth := auth.New()
	strategy := strategy.New(logic.Strategy, logic.Applied)

	return map[string]server.Service{
		"control":  control,
		"rule":     rule,
		"report":   report,
		"orch":     orch,
		"auth":     auth,
		"strategy": strategy,
	}
}
