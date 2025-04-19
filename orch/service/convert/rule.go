package convert

import (
	"encoding/json"
	"fmt"
	"time"
	model "xdp-banner/orch/model/rule"
	"xdp-banner/pkg/rule"

	"google.golang.org/protobuf/types/known/structpb"
)

// dto to model
func RuleDtoToModel(dto *structpb.Struct) (*model.Rule, error) {
	if dto == nil {
		return nil, NewErrInvalidField("rule", "empty rule")
	}

	ruleinfo := rule.RuleInfo{}
	dtojson, err := dto.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal pb sturct to json: %w", err)
	}

	err = json.Unmarshal(dtojson, &ruleinfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal rule: %w", err)
	}

	if ruleinfo.Cidr == "" {
		return nil, NewErrInvalidField("cidr", "missing")
	}
	if ruleinfo.Protocol == "" {
		return nil, NewErrInvalidField("protocol", "missing")
	}

	createdAt := time.Now()
	if ruleinfo.Duration == "" {
		ruleinfo.Duration = "300s"
	}
	d, err := time.ParseDuration(ruleinfo.Duration)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal durations input: %w", err)
	}
	expiresAt := createdAt.Add(d)

	rule := &model.Rule{
		RuleMeta: rule.RuleMeta{
			Comment:   ruleinfo.Comment,
			CreatedAt: createdAt,
			ExpiresAt: expiresAt,
			// Init when needed
			Identity: "0",
		},
		RuleInfo: ruleinfo,
	}
	return rule, nil
}

func RuleModelToDto(model *model.Rule) (*structpb.Struct, error) {
	if model == nil {
		return nil, fmt.Errorf("invalid model: config is nil")
	}

	dtoJson, err := json.Marshal(model)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	dto := &structpb.Struct{}
	err = json.Unmarshal(dtoJson, dto)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal json to pb struct: %w", err)
	}

	return dto, nil
}

func RuleListToDto(ruleitems model.RuleItems) (map[string]*structpb.Struct, error) {
	dtos := make(map[string]*structpb.Struct, len(ruleitems))

	for name, ruleitem := range ruleitems {
		combined := &structpb.Struct{
			Fields: make(map[string]*structpb.Value, len(ruleitem)),
		}

		for _, item := range ruleitem {
			respEntry, err := RuleModelToDto(&item)
			if err != nil {
				return nil, fmt.Errorf("failed to convert rule %s: %w", item.RuleInfo.Key(), err)
			}

			// 使用规则名称作为字段名
			combined.Fields[item.RuleInfo.Key()] = &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: respEntry,
				},
			}
		}
		dtos[name] = combined
	}
	return dtos, nil
}
