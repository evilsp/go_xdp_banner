package strategy

import (
	"context"
	"regexp"
	strategym "xdp-banner/orch/model/strategy"
	"xdp-banner/pkg/errors"

	"xdp-banner/orch/storage/agent/strategy"
)

type Strategy struct {
	storage strategy.Storage
}

func NewStrategy(storage strategy.Storage) *Strategy {
	return &Strategy{
		storage: storage,
	}
}

func (s *Strategy) Add(ctx context.Context, new *strategym.Strategy) error {
	err := validateStrategy(new)
	if err != nil {
		return errors.NewInputErrorf("valid strategy failed: %v", err)
	}

	err = s.storage.Add(ctx, new.Name, new)
	if err != nil {
		if err == strategy.ErrStrategyAlreadyExists {
			return errors.NewInputError("strategy already exists, please use another name")
		}
		return errors.NewServiceErrorf("failed to add strategy: %v", err)
	}

	return nil
}

// Delete deletes a strategy.
func (s *Strategy) Delete(ctx context.Context, name string) error {
	if err := s.storage.Delete(ctx, name); err != nil {
		return errors.NewServiceErrorf("failed to delete strategy: %v", err)
	}
	return nil
}

// Update updates a strategy.
func (s *Strategy) Update(ctx context.Context, new *strategym.Strategy) error {
	err := validateStrategy(new)
	if err != nil {
		return errors.NewInputErrorf("valid strategy failed: %v", err)
	}

	err = s.storage.Update(ctx, new.Name, new)
	if err != nil {
		if err == strategy.ErrStrategyNotFound {
			return errors.NewInputError("strategy not found")
		}
		return errors.NewServiceErrorf("failed to update strategy: %v", err)
	}
	return nil
}

func (s *Strategy) Get(ctx context.Context, name string) (*strategym.Strategy, error) {
	new, err := s.storage.Get(ctx, name)
	if err != nil {
		if err == strategy.ErrStrategyNotFound {
			return nil, errors.NewInputError("strategy not found")
		}
		return nil, errors.NewServiceErrorf("failed to get strategy: %v", err)
	}
	return new, nil
}

func (s *Strategy) List(ctx context.Context, pagesize int64, cursor string) (strategym.StrategyList, error) {
	if pagesize < 0 {
		return strategym.StrategyList{}, errors.NewInputError("pagesize must be greater than 0")
	} else if pagesize == 0 {
		pagesize = 10
	}

	list, err := s.storage.List(ctx, pagesize, cursor)
	if err != nil {
		return strategym.StrategyList{}, errors.NewServiceErrorf("failed to list strategies: %v", err)
	}
	return list, nil
}

func validateStrategy(new *strategym.Strategy) error {
	// validate strategy name
	if new.Name == "" {
		return errors.NewInputError("strategy name is empty")
	}

	// validate strategy action
	switch new.Action {
	case strategym.StrategyActionConfig, strategym.StrategyActionEnable:
	default:
		return errors.NewInputErrorf("invalid strategy action: %s", new.Action)
	}

	if new.NameSelector == "" && new.LabelSelector == "" {
		return errors.NewInputError("strategy name selector and label selector are both empty")
	}

	// validate strategy name selector
	if new.NameSelector != "" {
		_, err := regexp.Compile(new.NameSelector)
		if err != nil {
			return errors.NewInputErrorf("strategy name selector is not a valid regular expression: %s", new.NameSelector)
		}
	}

	// validate strategy label selector
	if new.LabelSelector != "" {
		_, err := regexp.Compile(new.LabelSelector)
		if err != nil {
			return errors.NewInputErrorf("strategy label selector is not a valid regular expression: %s", new.LabelSelector)
		}
	}

	return nil
}
