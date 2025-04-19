package rule

import (
	"encoding/json"
	"fmt"
	"xdp-banner/orch/model/common"
	"xdp-banner/pkg/rule"
)

type Rule struct {
	RuleMeta rule.RuleMeta `json:"meta"`
	RuleInfo rule.RuleInfo `json:"info"`
}

func (c *Rule) Marshal() []byte {
	b, err := json.Marshal(c)
	if err != nil {
		panic(fmt.Errorf("marshal config failed: %v", err))
	}
	return b
}

// MarshalStr wraps c.Marshal and returns the result as a string.
func (c *Rule) MarshalStr() string {
	return string(c.Marshal())
}

func (c *Rule) Unmarshal(b []byte) error {
	if err := json.Unmarshal(b, c); err != nil {
		return fmt.Errorf("unmarshal config failed: %v", err)
	}
	return nil
}

// UnmarshalStr wraps c.Unmarshal and accepts a string.
func (c *Rule) UnmarshalStr(s string) error {
	return c.Unmarshal([]byte(s))
}

type RuleItems = map[name]RuleItem
type RuleItem = []Rule

// RuleList is a list of configs.
type RuleList struct {
	common.List `json:",inline"`

	// Items is the list of configs identified by their names.
	Items RuleItems `json:"items"`
}

type name = string
