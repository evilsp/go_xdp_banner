package strategy

import (
	"context"
	"fmt"
	"regexp"
	"time"

	applieds "xdp-banner/orch/storage/agent/applied"
	"xdp-banner/orch/storage/agent/node"
	strategys "xdp-banner/orch/storage/agent/strategy"

	nodem "xdp-banner/orch/model/node"
	strategym "xdp-banner/orch/model/strategy"

	"xdp-banner/pkg/errors"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"
)

type Applied struct {
	strategyStorage strategys.Storage
	agentStorage    node.InfoStorage
	appliedStorage  applieds.Storage
}

func NewApplied(
	strategyStorage strategys.Storage,
	agentStorage node.InfoStorage,
	appliedStorage applieds.Storage,
) *Applied {
	return &Applied{
		strategyStorage: strategyStorage,
		agentStorage:    agentStorage,
		appliedStorage:  appliedStorage,
	}
}

func (a *Applied) Create(ctx context.Context, strategyName string) error {
	strategy, err := a.strategyStorage.Get(ctx, strategyName)
	if err != nil {
		return fmt.Errorf("failed to get strategy %s: %w", strategyName, err)
	}

	var nameReg *regexp.Regexp
	if strategy.NameSelector != "" {
		nameReg, err = regexp.Compile(strategy.NameSelector)
		if err != nil {
			return fmt.Errorf("failed to compile name selector %s: %w", strategy.NameSelector, err)
		}
	}

	var labelsReg *regexp.Regexp
	if strategy.LabelSelector != "" {
		labelsReg, err = regexp.Compile(strategy.LabelSelector)
		if err != nil {
			return fmt.Errorf("failed to compile label selector %s: %w", strategy.LabelSelector, err)
		}
	}

	options := etcd.ListOption{
		Size: 20,
	}

	pager := etcd.NewListPager(a.agentStorage.CommonList)
	list, err := pager.List(ctx, options)
	if err != nil {
		return fmt.Errorf("failed to list agents: %w", err)
	}

	rawitems := list.Items
	agents := make([]nodem.AgentInfo, 0, rawitems.Len())
	for _, r := range rawitems.Iterator() {
		agent := new(nodem.AgentInfo)
		if err := agent.UnmarshalStr(r.(string)); err != nil {
			log.Warn("failed to unmarshal agent info", log.ErrorField(err))
			continue
		}
		agents = append(agents, *agent)
	}

	selectedNode := make([]string, 0, 5)
	for _, agent := range agents {
		if nameReg != nil && nameReg.MatchString(agent.Name) {
			selectedNode = append(selectedNode, agent.Name)
			continue
		}

		for _, label := range agent.Labels {
			if labelsReg != nil && labelsReg.MatchString(label) {
				selectedNode = append(selectedNode, agent.Name)
				break
			}
		}
	}

	applied := &strategym.Applied{
		Name:   UniqueTimeName(strategy.Name),
		Agents: selectedNode,
		Action: strategy.Action,
		Value:  strategy.Value,
		Status: strategym.AppliedStatusPending,
	}

	return a.appliedStorage.AddRunning(ctx, applied)
}

// UniqueTimeName generates a unique name prefix based on the current time
// the prefix is fixed length and the newer generated name will be sorted
// before the older ones in alphabetical order
func UniqueTimeName(base string) string {
	const maxMs = int64(999999999999999) // 15 digits, this is max enough
	nowMs := time.Now().UnixNano() / 1e6
	revMs := maxMs - nowMs
	prefix := fmt.Sprintf("%014x", revMs)

	return fmt.Sprintf("%s-%s", prefix, base)
}

func (a *Applied) MoveToHistory(ctx context.Context, new *strategym.Applied) error {
	if err := a.appliedStorage.AddHistory(ctx, new); err != nil {
		return errors.NewServiceErrorf("failed to add config: %v", err)
	}

	if err := a.appliedStorage.DeleteRunning(ctx, new.Name); err != nil {
		return errors.NewServiceErrorf("failed to delete config: %v", err)
	}

	return nil
}

func (a *Applied) ListRunning(ctx context.Context, pagesize int64, cursor string) (strategym.AppliedList, error) {
	if pagesize < 0 {
		return strategym.AppliedList{}, errors.NewInputError("pagesize must be greater than 0")
	} else if pagesize == 0 {
		pagesize = 10
	}

	list, err := a.appliedStorage.ListRunning(ctx, pagesize, cursor)
	if err != nil {
		return strategym.AppliedList{}, errors.NewServiceErrorf("failed to list configs: %v", err)
	}

	return list, nil
}

func (a *Applied) GetRunning(ctx context.Context, name string) (*strategym.Applied, error) {
	applied, err := a.appliedStorage.GetRunning(ctx, name)
	if err != nil {
		if err == applieds.ErrAppliedNotFound {
			return nil, errors.NewInputError("applied not found")
		}

		return nil, errors.NewServiceErrorf("failed to get config: %v", err)
	}

	return applied, nil
}

func (a *Applied) ListHistory(ctx context.Context, pagesize int64, cursor string) (strategym.AppliedList, error) {
	list, err := a.appliedStorage.ListHistory(ctx, pagesize, cursor)
	if err != nil {
		return strategym.AppliedList{}, errors.NewServiceErrorf("failed to list configs: %v", err)
	}

	return list, nil
}

func (a *Applied) GetHistory(ctx context.Context, name string) (*strategym.Applied, error) {
	applied, err := a.appliedStorage.GetHistory(ctx, name)
	if err != nil {
		if err == applieds.ErrAppliedNotFound {
			return nil, errors.NewInputError("applied not found")
		}

		return nil, errors.NewServiceErrorf("failed to get config: %v", err)
	}

	return applied, nil
}

func (a *Applied) DeleteHistory(ctx context.Context, name string) error {
	if err := a.appliedStorage.DeleteHistory(ctx, name); err != nil {
		return errors.NewServiceErrorf("failed to delete applied: %v", err)
	}
	return nil
}
