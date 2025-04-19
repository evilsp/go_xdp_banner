package node

import (
	"encoding/json"
	"xdp-banner/orch/model/common"
)

type Registration struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}

func (r *Registration) Marshal() []byte {
	return common.MustMarshal(r)
}

func (r *Registration) MarshalStr() string {
	return string(r.Marshal())
}
func (r *Registration) Unmarshal(data []byte) error {
	return json.Unmarshal(data, r)
}

func (r *Registration) UnmarshalStr(data string) error {
	return r.Unmarshal([]byte(data))
}

type RegistrationItems = map[string]Registration

type RegistrationList struct {
	common.List `json:",inline"`
	Items       RegistrationItems `json:"items"`
}
