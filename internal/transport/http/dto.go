package http

type KVResponse struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type JoinRequest struct {
	NodeID  string `json:"node_id"`
	Address string `json:"address"`
}
