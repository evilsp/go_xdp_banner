package control

import (
	"context"
	nodem "xdp-banner/orch/model/node"
	"xdp-banner/orch/storage/agent/node"

	"xdp-banner/pkg/errors"
)

func (c *Control) GetConfig(ctx context.Context, name string) (string, error) {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return "", errors.NewInputError("agent not found")
		}

		return "", errors.NewServiceErrorf("get info failed, %v", err)
	}

	return info.Config, nil
}

func (c *Control) SetConfig(ctx context.Context, name string, configName string) error {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return errors.NewInputError("agent not found")
		}

		return errors.NewServiceErrorf("get info failed, %v", err)
	}

	if info.Config == configName {
		// no need to update
		return nil
	}

	info.Config = configName
	err = c.infos.Update(ctx, info)
	if err != nil {
		return errors.NewServiceErrorf("update info failed, %v", err)
	}

	return nil
}

func (c *Control) GetLabels(ctx context.Context, name string) ([]string, error) {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return nil, errors.NewInputError("agent not found")
		}

		return nil, errors.NewServiceErrorf("get info failed, %v", err)
	}

	return info.Labels, nil
}

func (c *Control) SetLabels(ctx context.Context, name string, labels []string) error {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return errors.NewInputError("agent not found")
		}

		return errors.NewServiceErrorf("get info failed, %v", err)
	}

	if info.Labels == nil && labels == nil {
		// no need to update
		return nil
	}

	info.Labels = labels

	err = c.infos.Update(ctx, info)
	if err != nil {
		return errors.NewServiceErrorf("update info failed, %v", err)
	}

	return nil
}

func (c *Control) GetInfo(ctx context.Context, name string) (*nodem.AgentInfo, error) {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return nil, errors.NewInputError("agent not found")
		}

		return nil, errors.NewServiceErrorf("get info failed, %v", err)
	}

	return info, nil
}

func (c *Control) GetStatus(ctx context.Context, name string) (*nodem.AgentStatus, error) {
	status, err := c.statuss.Get(ctx, name, false)
	if err != nil {
		if err == node.ErrStatusNotFound {
			return nil, errors.NewInputError("agent not found")
		}

		return nil, errors.NewServiceErrorf("get status failed, %v", err)
	}

	return status, nil
}

func (c *Control) GetAgent(ctx context.Context, name string) (*nodem.Agent, error) {
	info, ierr := c.infos.Get(ctx, name)
	if ierr != nil && ierr != node.ErrInfoNotFound {
		return nil, errors.NewServiceErrorf("get info failed, %v", ierr)
	}

	status, serr := c.statuss.Get(ctx, name, false)
	if serr != nil && serr != node.ErrStatusNotFound {
		return nil, errors.NewServiceErrorf("get status failed, %v", serr)
	}

	if ierr == node.ErrInfoNotFound && serr == node.ErrStatusNotFound {
		return nil, errors.NewInputError("agent not found")
	}

	agent := &nodem.Agent{
		Info:   info,
		Status: status,
	}

	return agent, nil
}

func (c *Control) ListAgents(ctx context.Context, pageSize int64, nextCursor string) (nodem.AgentList, error) {
	il, err := c.infos.List(ctx, pageSize, nextCursor)
	if err != nil {
		return nodem.AgentList{}, errors.NewServiceErrorf("failed to list agents: %v", err)
	}

	list := make(nodem.AgentItems, len(il.Items))
	for _, info := range il.Items {
		status, err := c.statuss.Get(ctx, info.CommonInfo.Name, false)
		if err != nil && err != node.ErrStatusNotFound {
			return nodem.AgentList{}, errors.NewServiceErrorf("get status failed, %v", err)
		}

		list[info.CommonInfo.Name] = &nodem.Agent{
			Info:   info,
			Status: status,
		}
	}

	return nodem.AgentList{
		List:  il.List,
		Items: list,
	}, nil
}
