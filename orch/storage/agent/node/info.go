package node

import (
	"context"
	"fmt"
	"xdp-banner/orch/model/common"
	"xdp-banner/orch/model/node"

	"xdp-banner/pkg/etcd"
)

var (
	EtcdDirInfo = etcd.Join(EtcdDir, "info/")

	ErrInfoExist    error = fmt.Errorf("agent info already exist with provided name")
	ErrInfoNotFound error = fmt.Errorf("agent info not found with provided name")
)

// InfoStorage store node info in etcd
type InfoStorage struct {
	client etcd.Client
}

func NewInfoStorage(cli etcd.Client) InfoStorage {
	return InfoStorage{
		client: cli,
	}
}

func InfoKey(name string) etcd.Key {
	return etcd.Join(EtcdDirInfo, name)
}

func (s InfoStorage) Add(ctx context.Context, info *node.AgentInfo) error {
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

func (s InfoStorage) Update(ctx context.Context, info *node.AgentInfo) error {
	key := InfoKey(info.CommonInfo.Name)
	err := s.client.Update(ctx, key, info.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return ErrInfoNotFound
		}
		return err
	}

	return nil
}

func (s InfoStorage) Get(ctx context.Context, name string) (*node.AgentInfo, error) {
	key := InfoKey(name)
	resp, err := s.client.GetMustExist(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrInfoNotFound
		}
		return nil, err
	}

	info := &node.AgentInfo{}
	err = info.Unmarshal(resp.Kvs[0].Value)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (s InfoStorage) List(ctx context.Context, size int64, nextCursor string) (node.AgentInfoList, error) {
	rawInfos, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDirInfo,
		Size:   size,
		Cursor: nextCursor,
	})
	if err != nil {
		return node.AgentInfoList{}, err
	}

	infos := make(node.AgentInfoItems)
	for _, r := range rawInfos.Items.Iterator() {
		info := new(node.AgentInfo)
		if err := info.UnmarshalStr(r.(string)); err != nil {
			return node.AgentInfoList{}, err
		}

		infos[info.Name] = info
	}

	return node.AgentInfoList{
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

func (s InfoStorage) CommonList(ctx context.Context, option etcd.ListOption) (etcd.PagedList, error) {
	option.Prefix = EtcdDirInfo
	return s.client.List(ctx, option)
}
