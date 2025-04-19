/*
encapsulate the basic etcd operations
*/

package etcd

import (
	"context"
	"fmt"
	"time"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func (c *client) Delete(ctx context.Context, key Key, opts ...clientv3.OpOption) (*clientv3.DeleteResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()
	return c.Client.Delete(ctx, key, opts...)
}

func (c *client) Get(ctx context.Context, key Key, opts ...clientv3.OpOption) (*clientv3.GetResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()
	return c.Client.Get(ctx, key, opts...)
}

func (c *client) Put(ctx context.Context, key Key, value string, opts ...clientv3.OpOption) (*clientv3.PutResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()

	return c.Client.Put(ctx, key, value, opts...)
}

func (c *client) Txn(ctx context.Context) (clientv3.Txn, context.CancelFunc) {
	ctx, cancel := c.WithTimeout(ctx)

	return c.Client.Txn(ctx), cancel
}

func noOpCancel() {}

func (c *client) WithTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.TimeOut > 0 {
		return context.WithTimeout(ctx, c.TimeOut)
	}

	// do nothing
	return ctx, noOpCancel
}

var ErrorLeaseTooShort = fmt.Errorf("lease duration must be at least 1 second")

func (c *client) Grant(ctx context.Context, dur time.Duration) (*clientv3.LeaseGrantResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()

	durFloat := dur.Seconds()
	ttl := int64(durFloat)

	if ttl < 1 {
		return nil, ErrorLeaseTooShort
	}

	return c.Client.Grant(ctx, ttl)
}

func (c *client) Revoke(ctx context.Context, leaseID clientv3.LeaseID) (*clientv3.LeaseRevokeResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()

	resp, err := c.Client.Revoke(ctx, leaseID)
	if err == rpctypes.ErrLeaseNotFound {
		return &clientv3.LeaseRevokeResponse{}, nil
	}

	return resp, err
}

func (c *client) KeepAliveOnce(ctx context.Context, leaseID clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()

	return c.Client.KeepAliveOnce(ctx, leaseID)
}
