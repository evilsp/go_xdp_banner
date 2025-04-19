package strategy

import (
	"encoding/json"
	"xdp-banner/orch/model/common"
)

type Strategy struct {
	// Name of the strategy
	Name string `json:"name" yaml:"name"`
	// NameSelector is a selector for the name of the strategy, regex
	NameSelector string `json:"nameSelector" yaml:"nameSelector"`
	// LabelSelector is a selector for the labels of the strategy, regex
	LabelSelector string `json:"labelSelector" yaml:"labelSelector"`
	// Action is the action to be taken on the selected agents
	Action StrategyAction `json:"action" yaml:"action"`
	// Value  depends on the action
	Value string `json:"value" yaml:"value"`
}

type StrategyAction string

const (
	// StrategyActionEnable the strategy is to enable agents
	StrategyActionEnable StrategyAction = "enable"
	// StrategyActionConfig the strategy is to configure agents
	StrategyActionConfig StrategyAction = "config"
)

func (s *Strategy) Marshal() []byte {
	return common.MustMarshal(s)
}

func (s *Strategy) MarshalStr() string {
	return string(s.Marshal())
}

func (s *Strategy) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *Strategy) UnmarshalStr(data string) error {
	return s.Unmarshal([]byte(data))
}

type StrategyItems = map[name]*Strategy

// StrategyList is a list of strategies.
type StrategyList struct {
	common.List `json:",inline"`

	// Items is the list of strategies identified by their names.
	Items StrategyItems `json:"items"`
}

type name = string
