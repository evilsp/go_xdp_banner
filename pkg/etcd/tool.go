package etcd

import (
	"context"

	clientv3 "go.etcd.io/etcd/client/v3"
)

// GetMustExist gets a key from etcd. If the key does not exist, it returns ErrKeyNotFound err.
func (c *client) GetMustExist(ctx context.Context, key Key, opts ...clientv3.OpOption) (resp *clientv3.GetResponse, err error) {
	resp, err = c.Get(ctx, key, opts...)
	if err != nil {
		return
	}

	if len(resp.Kvs) == 0 {
		return resp, ErrKeyNotFound
	}

	return
}

// Exists checks if the key exists in etcd.
func (c *client) Exist(ctx context.Context, key Key, opts ...clientv3.OpOption) (bool, error) {
	if _, err := c.GetMustExist(ctx, key, opts...); err != nil {
		if err == ErrKeyNotFound {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// DeleteWithPrefix deletes a key with a prefix in etcd.
func (c *client) DeleteWithPrefix(ctx context.Context, key Key) error {
	_, err := c.Delete(ctx, key, clientv3.WithPrefix())
	return err
}

// Create creates a key-value pair in etcd. If the key already exists, it returns an error.
func (c *client) Create(ctx context.Context, key Key, value string) error {
	ctx, cancel := c.WithTimeout(ctx)
	defer cancel()

	txn, cancel := c.Txn(ctx)
	defer cancel()

	// 检查键是否存在
	cond := clientv3.Compare(clientv3.CreateRevision(key), "=", 0)
	// 如果条件成立，创建键；否则什么都不做
	txnResp, err := txn.If(cond).
		Then(clientv3.OpPut(key, value)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		// 键已经存在
		return ErrKeyExist
	}
	return nil
}

// Update updates a key-value pair in etcd. If the key does not exist, it returns an error.
func (c *client) Update(ctx context.Context, key Key, value string) error {
	txn, canel := c.Txn(ctx)
	defer canel()

	cond := clientv3.Compare(clientv3.CreateRevision(key), ">", 0)
	txnResp, err := txn.If(cond).
		Then(clientv3.OpPut(key, value)).
		Else(clientv3.OpGet(key)).
		Commit()
	if err != nil {
		return err
	}

	if !txnResp.Succeeded {
		// 键不存在
		return ErrKeyNotFound
	}

	return nil
}
