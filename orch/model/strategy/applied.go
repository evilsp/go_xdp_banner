package strategy

import (
	"encoding/json"
	"xdp-banner/orch/model/common"
)

type Applied struct {
	// Name of the strategy
	Name string `json:"name" yaml:"name"`
	// Agents is the list of agents to which the strategy is applied
	Agents []string `json:"agents" yaml:"agents"`
	// Action is the action to be taken on the selected agents
	Action StrategyAction `json:"action" yaml:"action"`
	// Value depends on the action
	Value string `json:"value" yaml:"value"`
	// Status is the status of the strategy
	Status AppliedStatus `json:"status" yaml:"status"`
	// Error when the status is failed, list of errors
	Error []string `json:"error" yaml:"error"`
}

type AppliedStatus string

const (
	// AppliedStatusPending the strategy is pending
	AppliedStatusPending AppliedStatus = "pending"
	// AppliedStatusRunning the strategy is running
	AppliedStatusRunning AppliedStatus = "running"
	// AppliedStatusSuccess the strategy is success
	AppliedStatusSuccess AppliedStatus = "success"
	// AppliedStatusFailed the strategy is failed
	AppliedStatusFailed AppliedStatus = "failed"
)

func (a *Applied) Marshal() []byte {
	return common.MustMarshal(a)
}

func (a *Applied) MarshalStr() string {
	return string(a.Marshal())
}

func (a *Applied) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Applied) UnmarshalStr(data string) error {
	return a.Unmarshal([]byte(data))
}

type AppliedItems = map[name]*Applied

// AppliedList is a list of applied strategies.
type AppliedList struct {
	common.List `json:",inline"`
	// Items is the list of applied strategies identified by their names.
	Items AppliedItems `json:"items"`
}
