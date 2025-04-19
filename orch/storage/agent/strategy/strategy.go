package strategy

import (
	"context"
	"errors"
	"xdp-banner/orch/model/common"
	model "xdp-banner/orch/model/strategy"
	"xdp-banner/orch/storage/agent"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDir etcd.Key = etcd.Join(agent.EtcdDir, "strategy/")
)

var (
	// ErrStrategyNotFound is returned when the strategy is not found.
	ErrStrategyNotFound = errors.New("strategy not found")
	// ErrStrategyAlreadyExists is returned when the strategy already exists.
	ErrStrategyAlreadyExists = errors.New("strategy already exists")
)

func StrategyKey(name string) etcd.Key {
	return etcd.Join(EtcdDir, name)
}

// Storage 提供对 etcd 的操作
type Storage struct {
	client etcd.Client
}

func New(client etcd.Client) Storage {
	return Storage{client: client}
}

// Add creates a new strategy in etcd.
func (s Storage) Add(ctx context.Context, name string, strategy *model.Strategy) error {
	key := StrategyKey(name)
	err := s.client.Create(ctx, key, strategy.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyExist {
			return ErrStrategyAlreadyExists
		}
	}

	return nil
}

// Delete deletes a strategy from etcd.
func (s Storage) Delete(ctx context.Context, name string) error {
	key := StrategyKey(name)
	_, err := s.client.Delete(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil
		}
	}
	return err
}

// Update updates a strategy in etcd.
func (s Storage) Update(ctx context.Context, name string, strategy *model.Strategy) error {
	key := StrategyKey(name)
	err := s.client.Update(ctx, key, strategy.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return ErrStrategyNotFound
		}
	}

	return nil
}

// Get retrieves a strategy from etcd.
func (s Storage) Get(ctx context.Context, name string) (*model.Strategy, error) {
	key := StrategyKey(name)
	raw, err := s.client.GetMustExist(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrStrategyNotFound
		}
		return nil, err
	}

	strategy := new(model.Strategy)
	if err := strategy.Unmarshal(raw.Kvs[0].Value); err != nil {
		return nil, err
	}
	return strategy, nil
}

func (s Storage) List(ctx context.Context, size int64, nextCursor string) (model.StrategyList, error) {
	rawStrats, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDir,
		Size:   size,
		Cursor: nextCursor,
	})
	if err != nil {
		return model.StrategyList{}, err
	}

	strategies := make(map[string]*model.Strategy)
	for key, r := range rawStrats.Items.Iterator() {
		strategy := new(model.Strategy)
		if err := strategy.UnmarshalStr(r.(string)); err != nil {
			return model.StrategyList{}, err
		}
		strategies[etcd.Base(key)] = strategy
	}

	return model.StrategyList{
		List: common.List{
			TotalCount:  rawStrats.TotalCount,
			TotalPage:   rawStrats.TotalPage,
			CurrentPage: rawStrats.CurrentPage,
			HasNext:     rawStrats.More(),
			NextCursor:  rawStrats.NextCursor,
		},
		Items: strategies,
	}, nil
}

func (s Storage) DeleteDir(ctx context.Context) error {
	return s.client.DeleteWithPrefix(ctx, EtcdDir)
}
