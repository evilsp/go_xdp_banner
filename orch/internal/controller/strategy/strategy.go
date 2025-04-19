package strategy

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"xdp-banner/orch/logic/agent/control"
	"xdp-banner/orch/logic/strategy"

	strategym "xdp-banner/orch/model/strategy"

	"xdp-banner/orch/storage/agent/applied"

	"xdp-banner/orch/internal/otlp"
	"xdp-banner/pkg/controller"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/informer"
	"xdp-banner/pkg/log"

	"go.opentelemetry.io/otel/metric"
)

// gateway
var _ controller.ControllerImpl = (*strategyController)(nil)

type StrategyController struct {
	*controller.Controller
}

func New(i informer.Informer, agent *control.Control, applied *strategy.Applied) StrategyController {
	sc := newStrategyController(agent, applied)

	controller := controller.New(i, controller.ControllerOption{
		Name: "strategy",
		Impl: sc,
	})

	return StrategyController{
		Controller: controller,
	}
}

type strategyController struct {
	agent   *control.Control
	applied *strategy.Applied
	metrics metrics
}

func newStrategyController(agent *control.Control, applied *strategy.Applied) *strategyController {
	return &strategyController{
		agent:   agent,
		applied: applied,
		metrics: newMetrics(),
	}
}

func (sc *strategyController) ListenPrefix() []string {
	return []string{applied.EtcdDirRunning}
}

func (sc *strategyController) KeyProcessor(key etcd.Key) etcd.Key {
	return key
}

func (sc *strategyController) SyncHandler(key string, informer informer.Informer) error {
	raw, exist, err := informer.Get(key)
	if err != nil {
		return fmt.Errorf("get applied failed %w", err)
	} else if !exist {
		return nil
	}

	rawStr, ok := raw.(string)
	if !ok {
		return fmt.Errorf("applied data is not string")
	}

	applied := new(strategym.Applied)
	if err := applied.UnmarshalStr(rawStr); err != nil {
		return fmt.Errorf("unmarshal applied data failed %w", err)
	}

	if skipHistoryApplied(applied.Status) {
		log.Debug("skip history applied", log.StringField("applied", applied.Name))
		return nil
	}

	err = sc.reconcile(applied)
	if err != nil {
		sc.metrics.SyncError.Add(context.Background(), 1)
		return err
	}

	return sc.applied.MoveToHistory(context.Background(), applied)
}

func skipHistoryApplied(status strategym.AppliedStatus) bool {
	switch status {
	case strategym.AppliedStatusPending, strategym.AppliedStatusRunning:
		return false
	default:
		return true
	}
}

func (sc *strategyController) reconcile(applied *strategym.Applied) error {
	switch applied.Action {
	case strategym.StrategyActionConfig:
		return sc.handleConfig(applied)
	case strategym.StrategyActionEnable:
		return sc.handleEnable(applied)
	default:
		return sc.handleUnknown(applied)
	}
}

func (sc *strategyController) handleConfig(applied *strategym.Applied) error {
	config := applied.Value
	allSuccess := true
	for _, agent := range applied.Agents {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := sc.agent.SetConfig(ctx, agent, config)
		if err != nil {
			allSuccess = false
			applied.Error = append(applied.Error, fmt.Sprintf("set config to agent %s failed %s", agent, err))
			continue
		}
	}

	if allSuccess {
		applied.Status = strategym.AppliedStatusSuccess
	} else {
		applied.Status = strategym.AppliedStatusFailed
	}

	return nil
}

func (sc *strategyController) handleEnable(applied *strategym.Applied) error {
	enable, err := strconv.ParseBool(applied.Value)
	if err != nil {
		applied.Status = strategym.AppliedStatusFailed
		applied.Error = append(applied.Error, fmt.Sprintf("parse enable value %s failed %s", applied.Value, err))
		return nil
	}

	allSuccess := true
	for _, agent := range applied.Agents {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		err := sc.agent.Enable(ctx, agent, enable)
		if err != nil {
			allSuccess = false
			continue
		}
	}

	if allSuccess {
		applied.Status = strategym.AppliedStatusSuccess
	} else {
		applied.Status = strategym.AppliedStatusFailed
	}

	return nil
}

func (sc *strategyController) handleUnknown(applied *strategym.Applied) error {
	applied.Status = strategym.AppliedStatusFailed
	applied.Error = []string{fmt.Sprintf("unknown action %s", applied.Action)}
	return nil
}

type metrics struct {
	SyncSuccess metric.Int64Counter
	SyncError   metric.Int64Counter
}

func newMetrics() metrics {
	meter := otlp.Meter()

	SyncSuccess, _ := meter.Int64Counter("sync_success", metric.WithDescription("success number of syncs"))
	syncError, _ := meter.Int64Counter("sync_error", metric.WithDescription("error number of syncs"))

	return metrics{
		SyncSuccess: SyncSuccess,
		SyncError:   syncError,
	}
}
