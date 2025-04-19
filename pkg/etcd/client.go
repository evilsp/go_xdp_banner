package etcd

import (
	"context"
	"time"
	"xdp-banner/pkg/log"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Client interface {
	RawClient() *clientv3.Client
	Close() error

	Get(ctx context.Context, key Key, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Put(ctx context.Context, key Key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error)
	Delete(ctx context.Context, key Key, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error)
	Txn(ctx context.Context) (clientv3.Txn, context.CancelFunc)
	Grant(ctx context.Context, dur time.Duration) (*clientv3.LeaseGrantResponse, error)
	Revoke(ctx context.Context, leaseID clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error)
	KeepAliveOnce(ctx context.Context, leaseID clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error)
	Watch(opt WatchOption) (WatchController, error)
	List(ctx context.Context, opt ListOption) (PagedList, error)

	GetMustExist(ctx context.Context, key Key, opts ...clientv3.OpOption) (*clientv3.GetResponse, error)
	Exist(ctx context.Context, key Key, opts ...clientv3.OpOption) (bool, error)
	DeleteWithPrefix(ctx context.Context, key Key) error
	Create(ctx context.Context, key Key, value string) error
	Update(ctx context.Context, key Key, value string) error

	Endpoints() []string
}

type client struct {
	*clientv3.Client

	TimeOut time.Duration
}

func New(config clientv3.Config) (Client, error) {
	if config.Logger != nil {
		config.Logger = log.GlobalLogger()
	}

	cli, err := clientv3.New(config)
	if err != nil {
		return nil, err
	}

	return &client{Client: cli, TimeOut: config.DialTimeout}, nil
}

func (c *client) RawClient() *clientv3.Client {
	return c.Client
}

func (c *client) Close() error {
	return c.Client.Close()
}
