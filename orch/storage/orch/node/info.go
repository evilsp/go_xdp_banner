package node

import (
	"context"
	"fmt"
	"xdp-banner/orch/model/common"
	"xdp-banner/orch/model/node"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDirInfo etcd.Key = etcd.Join(EtcdDir, "info/")

	ErrInfoExist    error = fmt.Errorf("orch info already exist with provided name")
	ErrInfoNotFound error = fmt.Errorf("orch info not found with provided name")
)

type InfoStorage struct {
	client etcd.Client
}

func NewInfoStorage(client etcd.Client) InfoStorage {
	return InfoStorage{
		client: client,
	}
}

func InfoKey(name string) etcd.Key {
	return etcd.Join(EtcdDirInfo, name)
}

func (s InfoStorage) Add(ctx context.Context, info *node.OrchInfo) error {
	key := InfoKey(info.CommonInfo.Name)
	err := s.client.Create(ctx, key, info.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyExist {
			return ErrInfoExist
		}
		return err
	}

	return nil
}

func (s InfoStorage) Delete(ctx context.Context, name string) error {
	key := InfoKey(name)
	_, err := s.client.Delete(ctx, key)

	return err
}

func (s InfoStorage) Update(ctx context.Context, info *node.OrchInfo) error {
	key := InfoKey(info.CommonInfo.Name)
	err := s.client.Create(ctx, key, info.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return ErrInfoNotFound
		}
		return err
	}

	return nil
}

func (s InfoStorage) Get(ctx context.Context, name string) (*node.OrchInfo, error) {
	key := InfoKey(name)
	resp, err := s.client.Get(ctx, key)
	if err != nil {
		return nil, err
	}

	if len(resp.Kvs) == 0 {
		return nil, ErrInfoNotFound
	}

	info := &node.OrchInfo{}
	err = info.Unmarshal(resp.Kvs[0].Value)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s InfoStorage) List(ctx context.Context, pagesize int64, cursor string) (node.OrchInfoList, error) {
	rawInfos, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDirInfo,
		Size:   pagesize,
		Cursor: cursor,
	})
	if err != nil {
		return node.OrchInfoList{}, err
	}

	infos := make(node.OrchInfoItems)
	for _, r := range rawInfos.Items.Iterator() {
		info := new(node.OrchInfo)
		if err := info.UnmarshalStr(r.(string)); err != nil {
			return node.OrchInfoList{}, err
		}

		infos[info.Name] = info
	}

	return node.OrchInfoList{
		List: common.List{
			TotalCount:  rawInfos.TotalCount,
			TotalPage:   rawInfos.TotalPage,
			CurrentPage: rawInfos.CurrentPage,
			HasNext:     rawInfos.More(),
			NextCursor:  rawInfos.NextCursor,
		},
		Items: infos,
	}, nil
}
