package node

// CommonInfo is a common info struct for xdp-banner and agent.
type CommonInfo struct {
	Name   string   `json:"name"`
	Labels []string `json:"labels"`
}

type CommonStatus struct {
	Name string `json:"name"`
}
