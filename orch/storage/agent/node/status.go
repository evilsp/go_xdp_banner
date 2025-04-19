package node

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"xdp-banner/orch/model/node"
	"xdp-banner/pkg/etcd"
	"xdp-banner/pkg/log"

	"go.etcd.io/etcd/api/v3/v3rpc/rpctypes"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const StatusLeaseTime time.Duration = 30 * time.Second

var (
	EtcdDirStatus     etcd.Key = etcd.Join(EtcdDir, "status/")
	ErrStatusNotFound error    = fmt.Errorf("agent status not found with provided name")
)

// StatusStorage store node status in etcd
type StatusStorage struct {
	client etcd.Client
	cache  *CacheStore
	renew  *RenewStore
}

func NewStatusStorage(ctx context.Context, client etcd.Client) StatusStorage {
	cache := newCacheStore(ctx, client)

	return StatusStorage{
		client: client,
		cache:  cache,
		renew:  NewRenewStore(StatusLeaseTime, client, cache),
	}
}

func StatusKey(name string) etcd.Key {
	return etcd.Join(EtcdDirStatus, name)
}

func (s StatusStorage) Update(ctx context.Context, name string, status *node.AgentStatus) error {
	key := StatusKey(name)

	leaseID, err := s.renew.Renew(ctx, key)
	if err != nil {
		return err
	}

	oldStatusObj, exsit, err := s.cache.Get(ctx, key)
	if err != nil {
		return err
	}
	oldStatus := ""
	if exsit {
		oldStatus = oldStatusObj.(leaseValue).value
	}

	if status.MarshalStr() == oldStatus {
		// no status change, no need to update
		return nil
	}

	_, err = s.client.Put(ctx, key, status.MarshalStr(), clientv3.WithLease(leaseID))
	return err
}

// Get get node status by name
// if force is true, get status from etcd directly, otherwise get from cache
func (s StatusStorage) Get(ctx context.Context, name string, force bool) (*node.AgentStatus, error) {
	key := StatusKey(name)
	var statusStr string
	statusObj, exsit, err := s.cache.Get(ctx, key)
	if err != nil || !exsit {
		log.Debug("get status from cache failed, try to get from etcd", log.StringField("name", name), log.ErrorField(err))
		resp, err := s.client.GetMustExist(ctx, key)
		if err != nil {
			if err == etcd.ErrKeyNotFound {
				return nil, ErrStatusNotFound
			}
			return nil, err
		}

		statusStr = string(resp.Kvs[0].Value)
	} else {
		statusStr = statusObj.(leaseValue).value
	}

	status := &node.AgentStatus{}
	err = status.UnmarshalStr(statusStr)
	if err != nil {
		return nil, err
	}

	return status, nil
}

// RenewStore store lease in etcd
type RenewStore struct {
	// lease 过期时间
	ttl    time.Duration
	client etcd.Client
	cache  *CacheStore
}

func NewRenewStore(ttl time.Duration, client etcd.Client, cache *CacheStore) *RenewStore {
	return &RenewStore{
		ttl:    ttl,
		client: client,
		cache:  cache,
	}
}

func (r *RenewStore) Renew(ctx context.Context, key string) (leaseID clientv3.LeaseID, err error) {
	value, ok, err := r.cache.Get(ctx, key)
	if err != nil {
		return 0, err
	}

	if ok {
		lv := value.(leaseValue)
		if lv.leaseID == 0 {
			return 0, fmt.Errorf("leaseID is 0")
		}

		leaseID = clientv3.LeaseID(lv.leaseID)

		err, expired := r.tryRenew(ctx, leaseID)
		if err != nil {
			return 0, err
		}
		if !expired {
			return leaseID, nil
		}
	}

	leaseID, err = r.newLease(ctx)
	if err != nil {
		return 0, err
	}

	return leaseID, nil
}

func (r *RenewStore) tryRenew(ctx context.Context, leaseID clientv3.LeaseID) (err error, expired bool) {
	_, err = r.client.KeepAliveOnce(ctx, leaseID)
	if err != nil {
		if err == rpctypes.ErrLeaseNotFound {
			return nil, true
		}
		return err, false
	}

	return nil, false
}

func (r *RenewStore) newLease(ctx context.Context) (clientv3.LeaseID, error) {
	resp, err := r.client.Grant(ctx, r.ttl)
	if err != nil {
		return 0, err
	}

	return resp.ID, nil
}

func (r *RenewStore) Close(ctx context.Context) error {
	errs := make([]error, 0, 1)

	for n, v := range r.cache.Range {
		_, err := r.client.Revoke(ctx, v.(clientv3.LeaseID))
		if err != nil {
			errs = append(errs, fmt.Errorf("revoke lease %s error: %v", n, err))
		}
	}

	if len(errs) != 0 {
		var strBuilder strings.Builder
		strBuilder.WriteString("close lease error: ")
		for _, err := range errs {
			strBuilder.WriteString(err.Error() + "; ")
		}
		return errors.New(strBuilder.String())
	}

	return nil
}
