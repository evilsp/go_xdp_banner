package node

import (
	"context"
	"fmt"
	"xdp-banner/orch/model/common"
	"xdp-banner/orch/model/node"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDirRegister     etcd.Key = etcd.Join(EtcdDir, "register/")
	ErrRegisterExist    error    = fmt.Errorf("provided name is already registered")
	ErrRegisterNotFound error    = fmt.Errorf("registration info not found with provided name")
)

// RegisterStorage store node registration in etcd
type RegisterStorage struct {
	client etcd.Client
}

func NewRegisterStorage(cli etcd.Client) RegisterStorage {
	return RegisterStorage{
		client: cli,
	}
}

func RegisterKey(name string) etcd.Key {
	return etcd.Join(EtcdDirRegister, name)
}

func (s RegisterStorage) Add(ctx context.Context, reg *node.Registration) error {
	key := RegisterKey(reg.Name)
	err := s.client.Create(ctx, key, reg.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyExist {
			return ErrRegisterExist
		}
	}
	return err
}

func (s RegisterStorage) Delete(ctx context.Context, name string) error {
	key := RegisterKey(name)
	_, err := s.client.Delete(ctx, key)
	return err
}

func (s RegisterStorage) Get(ctx context.Context, name string) (*node.Registration, error) {
	key := RegisterKey(name)
	resp, err := s.client.GetMustExist(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrRegisterNotFound
		}
		return nil, err
	}

	reg := &node.Registration{}
	err = reg.Unmarshal(resp.Kvs[0].Value)
	if err != nil {
		return nil, err
	}

	return reg, nil
}

func (s RegisterStorage) List(ctx context.Context, size int64, nextCursor string) (node.RegistrationList, error) {
	rawInfos, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDirRegister,
		Size:   size,
		Cursor: nextCursor,
	})
	if err != nil {
		return node.RegistrationList{}, err
	}

	items := make(node.RegistrationItems)
	for _, r := range rawInfos.Items.Iterator() {
		item := new(node.Registration)
		if err := item.UnmarshalStr(r.(string)); err != nil {
			return node.RegistrationList{}, err
		}

		items[item.Name] = *item
	}

	return node.RegistrationList{
		List: common.List{
			TotalCount:  rawInfos.TotalCount,
			TotalPage:   rawInfos.TotalPage,
			CurrentPage: rawInfos.CurrentPage,
			HasNext:     rawInfos.More(),
			NextCursor:  rawInfos.NextCursor,
		},
		Items: items,
	}, nil
}
