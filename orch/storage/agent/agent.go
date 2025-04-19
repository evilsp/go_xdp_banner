package agent

import (
	"context"
	"xdp-banner/pkg/etcd"
)

const EtcdDir = "/agent/"

type Storage struct {
	client etcd.Client
}

func New(client etcd.Client) Storage {
	return Storage{
		client: client,
	}
}

func (s Storage) DeleteDir(ctx context.Context) error {
	return s.client.DeleteWithPrefix(ctx, EtcdDir)
}
