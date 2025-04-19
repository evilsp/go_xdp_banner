package rulecenter

import (
	"context"
	"xdp-banner/orch/logic/rulecenter/validation"
	model "xdp-banner/orch/model/rule"
	ruleStorage "xdp-banner/orch/storage/agent/rule"
	"xdp-banner/pkg/errors"
)

type RuleCenter struct {
	storage   ruleStorage.Storage
	validator validation.Validator
}

func New(rs ruleStorage.Storage) *RuleCenter {
	// TODO: validator should be configurable
	validator := new(validation.MockValidator)

	return &RuleCenter{
		storage:   rs,
		validator: validator,
	}
}

// AddRule adds a new rule.
func (r *RuleCenter) AddRule(ctx context.Context, name string, rule *model.Rule) error {
	if err := r.validator.Validate(ctx, rule); err != nil {
		return errors.NewInputErrorf("valid rule failed: %v", err)
	}

	err := r.storage.Add(ctx, name, rule)
	if err != nil {
		if err == ruleStorage.ErrRuleAlreadyExists {
			return errors.NewInputError("rule already exists, please use another name")
		}

		return errors.NewServiceErrorf("failed to add rule: %v", err)
	}

	return nil
}

// DeleteRule deletes a rule.
func (r *RuleCenter) DeleteRule(ctx context.Context, name string, rule *model.Rule) error {
	//TODO: check Strategy is using the Rule, if is true, we should prevent deleting the Rule.
	if err := r.storage.Delete(ctx, name, rule); err != nil {
		return errors.NewServiceErrorf("failed to delete rule: %v", err)
	}
	return nil
}

// UpdateRule updates a rule.
func (r *RuleCenter) UpdateRule(ctx context.Context, name string, rule *model.Rule) error {
	if err := r.validator.Validate(ctx, rule); err != nil {
		return errors.NewInputErrorf("valid rule failed: %v", err)
	}

	err := r.storage.Update(ctx, name, rule)
	if err != nil {
		if err == ruleStorage.ErrRuleNotFound {
			return errors.NewInputError("rule not found")
		}

		return errors.NewServiceErrorf("failed to update rule: %v", err)
	}
	return nil
}

// GetRule gets a new rule.
func (c *RuleCenter) GetRule(ctx context.Context, name string) (model.RuleItem, error) {
	cfg, err := c.storage.GetRule(ctx, name)
	if err != nil {
		if err == ruleStorage.ErrRuleNotFound {
			return nil, errors.NewInputError("rule not found")
		}

		return nil, errors.NewServiceErrorf("failed to get rule: %v", err)
	}

	return cfg, nil
}

// ListRule lists rule.
func (c *RuleCenter) ListRule(ctx context.Context, pageSize int64, nextCursor string) (*model.RuleList, error) {
	list, err := c.storage.List(ctx, pageSize, nextCursor)
	if err != nil {
		return &model.RuleList{}, errors.NewServiceErrorf("failed to list rules: %v", err)
	}

	return list, nil
}
