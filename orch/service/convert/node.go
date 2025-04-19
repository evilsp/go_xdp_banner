package convert

import (
	"fmt"
	"xdp-banner/api/orch/v1/agent/control"
	model "xdp-banner/orch/model/node"

	"google.golang.org/protobuf/types/known/structpb"
)

func RegistrationItemsToDto(items model.RegistrationItems) (map[string]*control.RegisterResponse, error) {
	dto := make(map[string]*control.RegisterResponse, len(items))
	for name, item := range items {
		dto[name] = &control.RegisterResponse{
			Token: item.Token,
		}
	}

	return dto, nil
}

func AgentStatusToDto(status *model.AgentStatus) (*structpb.Struct, error) {
	if status == nil {
		return nil, NewErrInvalidField("status", "empty status")
	}

	dtojson := status.Marshal()

	dto := &structpb.Struct{}
	err := dto.UnmarshalJSON(dtojson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to pb struct: %w", err)
	}

	return dto, nil
}

func AgentInfoToDto(info *model.AgentInfo) (*structpb.Struct, error) {
	if info == nil {
		return nil, NewErrInvalidField("info", "empty info")
	}

	dtojson := info.Marshal()

	dto := &structpb.Struct{}
	err := dto.UnmarshalJSON(dtojson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to pb struct: %w", err)
	}

	return dto, nil
}

func AgentItemsToDto(agents model.AgentItems) (map[string]*control.Agent, error) {
	dto := make(map[string]*control.Agent, len(agents))
	for name, agent := range agents {
		ad := &control.Agent{}

		if agent.Info != nil {
			info, err := AgentInfoToDto(agent.Info)
			if err != nil {
				return nil, fmt.Errorf("failed to convert agent info to dto: %w", err)
			}
			ad.Info = info
		}

		if agent.Status != nil {
			status, err := AgentStatusToDto(agent.Status)
			if err != nil {
				return nil, fmt.Errorf("failed to convert agent status to dto: %w", err)
			}
			ad.Status = status
		}

		dto[name] = ad
	}

	return dto, nil
}

func OrchInfoToDto(info *model.OrchInfo) (*structpb.Struct, error) {
	if info == nil {
		return nil, NewErrInvalidField("info", "empty info")
	}

	dtojson := info.Marshal()

	dto := &structpb.Struct{}
	err := dto.UnmarshalJSON(dtojson)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to pb struct: %w", err)
	}

	return dto, nil
}

func OrchInfoItemsToDto(items model.OrchInfoItems) (map[string]*structpb.Struct, error) {
	dto := make(map[string]*structpb.Struct, len(items))
	for name, item := range items {
		itemDto, err := OrchInfoToDto(item)
		if err != nil {
			return nil, fmt.Errorf("failed to convert orch info to dto: %w", err)
		}
		dto[name] = itemDto
	}

	return dto, nil
}
