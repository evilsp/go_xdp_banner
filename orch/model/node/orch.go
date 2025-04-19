package node

import (
	"encoding/json"
	"xdp-banner/orch/model/common"
)

type OrchInfo struct {
	CommonInfo `json:",inline"`
}

func (i *OrchInfo) Marshal() []byte {
	return common.MustMarshal(i)
}

func (i *OrchInfo) MarshalStr() string {
	return string(i.Marshal())
}

func (i *OrchInfo) Unmarshal(data []byte) error {
	return json.Unmarshal(data, i)
}

func (i *OrchInfo) UnmarshalStr(data string) error {
	return i.Unmarshal([]byte(data))
}

type OrchInfoItems = map[string]*OrchInfo

type OrchInfoList struct {
	common.List `json:",inline"`
	Items       OrchInfoItems `json:"items"`
}
