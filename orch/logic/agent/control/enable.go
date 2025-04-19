package control

import (
	"context"
	"xdp-banner/orch/storage/agent/node"
	"xdp-banner/pkg/errors"
)

func (c *Control) Enable(ctx context.Context, name string, enable bool) (err error) {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return errors.NewInputError("agent not found")
		}

		return errors.NewServiceErrorf("get info failed, %v", err)
	}

	if info.Enable == enable {
		// no need to update
		return nil
	}

	info.Enable = enable
	err = c.infos.Update(ctx, info)
	if err != nil {
		return errors.NewServiceErrorf("update info failed, %v", err)
	}

	return nil
}
