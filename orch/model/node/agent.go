package node

import (
	"encoding/json"
	"time"
	"xdp-banner/orch/model/common"
)

type Agent struct {
	Info   *AgentInfo   `json:"info"`
	Status *AgentStatus `json:"stats"`
}

func (a *Agent) Marshal() []byte {
	return common.MustMarshal(a)
}

func (a *Agent) MarshalStr() string {
	return string(a.Marshal())
}

func (a *Agent) Unmarshal(data []byte) error {
	return json.Unmarshal(data, a)
}

type AgentItems = map[string]*Agent

type AgentList struct {
	common.List `json:",inline"`
	Items       AgentItems `json:"items"`
}

type AgentInfo struct {
	CommonInfo `json:",inline"`
	Enable     bool   `json:"enable"`
	Config     string `json:"config"`
}

func (i *AgentInfo) Marshal() []byte {
	return common.MustMarshal(i)
}

func (i *AgentInfo) MarshalStr() string {
	return string(i.Marshal())
}

func (i *AgentInfo) Unmarshal(data []byte) error {
	return json.Unmarshal(data, i)
}

func (i *AgentInfo) UnmarshalStr(data string) error {
	return i.Unmarshal([]byte(data))
}

type AgentInfoItems = map[string]*AgentInfo

type AgentInfoList struct {
	common.List `json:",inline"`
	Items       AgentInfoItems `json:"items"`
}

type ErrorTime struct {
	Message string    `json:"message"`
	RetryAt time.Time `json:"retry_at"`
}

type AgentStatus struct {
	CommonStatus `json:",inline"`
	GrpcEndpoint string     `json:"grpc_endpoint"`
	HttpEndpoint string     `json:"http_endpoint"`
	Config       string     `json:"config"`
	Phase        string     `json:"phase"`
	Error        *ErrorTime `json:"error"`
}

func (s *AgentStatus) Marshal() []byte {
	return common.MustMarshal(s)
}

func (s *AgentStatus) MarshalStr() string {
	return string(s.Marshal())
}

func (s *AgentStatus) Unmarshal(data []byte) error {
	return json.Unmarshal(data, s)
}

func (s *AgentStatus) UnmarshalStr(data string) error {
	return s.Unmarshal([]byte(data))
}

type AgentStatusItems = map[string]*AgentStatus

type AgentStatusList struct {
	common.List `json:",inline"`
	Items       AgentStatusItems `json:"items"`
}
