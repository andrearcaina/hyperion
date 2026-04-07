package http

type KVResponse struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}
