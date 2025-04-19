package convert

import (
	api "xdp-banner/api/orch/v1/strategy"
	"xdp-banner/orch/model/strategy"
)

func StrategyDtoToModel(dto *api.Strategy) (*strategy.Strategy, error) {
	if dto == nil {
		return nil, NewErrInvalidField("strategy", "empty strategy")
	}

	action := strategy.StrategyActionConfig
	switch dto.Action {
	case string(strategy.StrategyActionConfig):
		action = strategy.StrategyActionConfig
	case string(strategy.StrategyActionEnable):
		action = strategy.StrategyActionEnable
	default:
		return nil, NewErrInvalidField("action", "invalid action")
	}

	return &strategy.Strategy{
		Name:          dto.Name,
		NameSelector:  dto.NameSelector,
		LabelSelector: dto.LabelSelector,
		Action:        action,
		Value:         dto.Value,
	}, nil
}

func StrategyModelToDto(model *strategy.Strategy) (*api.Strategy, error) {
	if model == nil {
		return nil, NewErrInvalidField("strategy", "empty strategy")
	}

	return &api.Strategy{
		Name:          model.Name,
		NameSelector:  model.NameSelector,
		LabelSelector: model.LabelSelector,
		Action:        string(model.Action),
		Value:         model.Value,
	}, nil
}

func StrategyListItemToDto(models strategy.StrategyItems) (map[string]*api.Strategy, error) {
	dtos := make(map[string]*api.Strategy, len(models))
	for name, strategy := range models {
		dto, err := StrategyModelToDto(strategy)
		if err != nil {
			return nil, NewErrInvalidField("strategy", "failed to convert strategy model to dto")
		}
		dtos[name] = dto
	}
	return dtos, nil
}

func AppliedDtoToModel(dto *api.Applied) (*strategy.Applied, error) {
	if dto == nil {
		return nil, NewErrInvalidField("applied strategy", "empty applied strategy")
	}

	status := strategy.AppliedStatusPending
	switch dto.Status {
	case api.AppliedStatus_APPLIED_STATUS_PENDING:
		status = strategy.AppliedStatusPending
	case api.AppliedStatus_APPLIED_STATUS_RUNNING:
		status = strategy.AppliedStatusRunning
	case api.AppliedStatus_APPLIED_STATUS_SUCCESS:
		status = strategy.AppliedStatusSuccess
	case api.AppliedStatus_APPLIED_STATUS_FAILED:
		status = strategy.AppliedStatusFailed
	default:
		return nil, NewErrInvalidField("status", "invalid status")
	}

	return &strategy.Applied{
		Name:   dto.Name,
		Agents: dto.Agents,
		Action: strategy.StrategyAction(dto.Action),
		Status: status,
		Value:  dto.Value,
		Error:  dto.Error,
	}, nil
}

func AppliedModelToDto(model *strategy.Applied) (*api.Applied, error) {
	if model == nil {
		return nil, NewErrInvalidField("applied strategy", "empty applied strategy")
	}

	status := api.AppliedStatus_APPLIED_STATUS_PENDING
	switch model.Status {
	case strategy.AppliedStatusPending:
		status = api.AppliedStatus_APPLIED_STATUS_PENDING
	case strategy.AppliedStatusRunning:
		status = api.AppliedStatus_APPLIED_STATUS_RUNNING
	case strategy.AppliedStatusSuccess:
		status = api.AppliedStatus_APPLIED_STATUS_SUCCESS
	case strategy.AppliedStatusFailed:
		status = api.AppliedStatus_APPLIED_STATUS_FAILED
	default:
		return nil, NewErrInvalidField("status", "invalid status")
	}

	return &api.Applied{
		Name:   model.Name,
		Agents: model.Agents,
		Action: string(model.Action),
		Status: status,
		Value:  model.Value,
		Error:  model.Error,
	}, nil
}

func AppliedListItemToDto(models strategy.AppliedItems) (map[string]*api.Applied, error) {
	dtos := make(map[string]*api.Applied, len(models))
	for name, applied := range models {
		dto, err := AppliedModelToDto(applied)
		if err != nil {
			return nil, NewErrInvalidField("applied strategy", "failed to convert applied model to dto")
		}
		dtos[name] = dto
	}
	return dtos, nil
}
