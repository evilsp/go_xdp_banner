package orch

import (
	"context"
	nodem "xdp-banner/orch/model/node"
	"xdp-banner/orch/storage/orch/node"

	"xdp-banner/pkg/errors"
)

type Orch struct {
	infos node.InfoStorage
}

func New(is node.InfoStorage) *Orch {
	return &Orch{
		infos: is,
	}
}

func (c *Orch) GetInfo(ctx context.Context, name string) (*nodem.OrchInfo, error) {
	info, err := c.infos.Get(ctx, name)
	if err != nil {
		if err == node.ErrInfoNotFound {
			return nil, errors.NewInputError("info not found")
		}

		return nil, errors.NewServiceErrorf("get info failed, %v", err)
	}

	return info, nil
}

func (c *Orch) ListInfo(ctx context.Context, pageSize int64, nextCursor string) (nodem.OrchInfoList, error) {
	il, err := c.infos.List(ctx, pageSize, nextCursor)
	if err != nil {
		return nodem.OrchInfoList{}, errors.NewServiceErrorf("list orch info failed, %v", err)
	}

	return il, nil
}
