package report

import (
	"context"
	model "xdp-banner/orch/model/node"
	"xdp-banner/orch/storage/agent/node"
	"xdp-banner/pkg/errors"
)

type Report struct {
	storage node.StatusStorage
}

func New(ns node.StatusStorage) *Report {
	return &Report{
		storage: ns,
	}
}

// AddConfig adds a new config.
func (c *Report) UpdateStatus(ctx context.Context, name string, config *model.AgentStatus) error {
	err := c.storage.Update(ctx, name, config)
	if err != nil {
		return errors.NewInputError(err.Error())
	}

	return nil
}

// DeleteConfig deletes a config.
