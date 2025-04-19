package applied

import (
	"context"
	"errors"
	"xdp-banner/orch/model/common"
	model "xdp-banner/orch/model/strategy"
	"xdp-banner/orch/storage/agent"
	"xdp-banner/pkg/etcd"
)

var (
	EtcdDir        etcd.Key = etcd.Join(agent.EtcdDir, "applied/")
	EtcdDirRunning etcd.Key = etcd.Join(EtcdDir, "running/")
	EtcdDirHistory etcd.Key = etcd.Join(EtcdDir, "history/")
)

func RunningKey(name string) etcd.Key {
	return etcd.Join(EtcdDirRunning, name)
}

func HistoryKey(name string) etcd.Key {
	return etcd.Join(EtcdDirHistory, name)
}

var (
	// ErrAppliedNotFound is returned when the applied is not found.
	ErrAppliedNotFound = errors.New("applied not found")
	// ErrAppliedAlreadyExists is returned when the applied already exists.
	ErrAppliedAlreadyExists = errors.New("applied already exists")
)

// Storage 提供对 etcd 的操作
type Storage struct {
	client etcd.Client
}

func New(client etcd.Client) Storage {
	return Storage{client: client}
}

// AddRunning creates a new running applied in etcd.
func (s Storage) AddRunning(ctx context.Context, applied *model.Applied) error {
	key := RunningKey(applied.Name)
	err := s.client.Create(ctx, key, applied.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyExist {
			return ErrAppliedAlreadyExists
		}
	}

	return nil
}

// DeleteRunning deletes a running applied from etcd.
func (s Storage) DeleteRunning(ctx context.Context, name string) error {
	key := RunningKey(name)
	_, err := s.client.Delete(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil
		}
	}
	return err
}

// UpdateRunning updates a running applied in etcd.
func (s Storage) UpdateRunning(ctx context.Context, name string, applied *model.Applied) error {
	key := RunningKey(name)
	err := s.client.Update(ctx, key, applied.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return ErrAppliedNotFound
		}
	}
	return nil
}

// GetRunning gets a running applied from etcd.
func (s Storage) GetRunning(ctx context.Context, name string) (*model.Applied, error) {
	key := RunningKey(name)
	raw, err := s.client.GetMustExist(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrAppliedNotFound
		}
		return nil, err
	}

	applied := new(model.Applied)
	if err := applied.Unmarshal(raw.Kvs[0].Value); err != nil {
		return nil, err
	}

	return applied, nil
}

// ListRunning lists all running applied in etcd.
func (s Storage) ListRunning(ctx context.Context, pagesize int64, cursor string) (model.AppliedList, error) {
	rawApplieds, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDirRunning,
		Size:   pagesize,
		Cursor: cursor,
	})
	if err != nil {
		return model.AppliedList{}, err
	}

	items := make(model.AppliedItems)
	for key, r := range rawApplieds.Items.Iterator() {
		applied := new(model.Applied)
		if err := applied.UnmarshalStr(r.(string)); err != nil {
			return model.AppliedList{}, err
		}
		items[key] = applied
	}

	return model.AppliedList{
		List: common.List{
			TotalCount:  rawApplieds.TotalCount,
			TotalPage:   rawApplieds.TotalPage,
			CurrentPage: rawApplieds.CurrentPage,
			HasNext:     rawApplieds.More(),
			NextCursor:  rawApplieds.NextCursor,
		},
		Items: items,
	}, nil
}

// AddHistory creates a new history applied in etcd.
func (s Storage) AddHistory(ctx context.Context, applied *model.Applied) error {
	key := HistoryKey(applied.Name)
	err := s.client.Create(ctx, key, applied.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyExist {
			return ErrAppliedAlreadyExists
		}
	}
	return nil
}

// DeleteHistory deletes a history applied from etcd.
func (s Storage) DeleteHistory(ctx context.Context, name string) error {
	key := HistoryKey(name)
	_, err := s.client.Delete(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil
		}
	}
	return err
}

// UpdateHistory updates a history applied in etcd.
func (s Storage) UpdateHistory(ctx context.Context, name string, applied *model.Applied) error {
	key := HistoryKey(name)
	err := s.client.Update(ctx, key, applied.MarshalStr())
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return ErrAppliedNotFound
		}
	}
	return nil
}

// GetHistory gets a history applied from etcd.
func (s Storage) GetHistory(ctx context.Context, name string) (*model.Applied, error) {
	key := HistoryKey(name)
	raw, err := s.client.GetMustExist(ctx, key)
	if err != nil {
		if err == etcd.ErrKeyNotFound {
			return nil, ErrAppliedNotFound
		}
		return nil, err
	}

	applied := new(model.Applied)
	if err := applied.Unmarshal(raw.Kvs[0].Value); err != nil {
		return nil, err
	}

	return applied, nil
}

// ListHistory lists all history applied in etcd.
func (s Storage) ListHistory(ctx context.Context, pagesize int64, cursor string) (model.AppliedList, error) {
	rawApplieds, err := s.client.List(ctx, etcd.ListOption{
		Prefix: EtcdDirHistory,
		Size:   pagesize,
		Cursor: cursor,
	})
	if err != nil {
		return model.AppliedList{}, err
	}

	items := make(model.AppliedItems)
	for key, r := range rawApplieds.Items.Iterator() {
		applied := new(model.Applied)
		if err := applied.UnmarshalStr(r.(string)); err != nil {
			return model.AppliedList{}, err
		}
		items[key] = applied
	}

	return model.AppliedList{
		List: common.List{
			TotalCount:  rawApplieds.TotalCount,
			TotalPage:   rawApplieds.TotalPage,
			CurrentPage: rawApplieds.CurrentPage,
			HasNext:     rawApplieds.More(),
			NextCursor:  rawApplieds.NextCursor,
		},
		Items: items,
	}, nil
}
