package http

import "encoding/base64"

type KVResponse struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
	Error string `json:"error,omitempty"`
}

type JoinRequest struct {
	NodeID      string `json:"node_id"`
	Address     string `json:"address"`
	HTTPAddress string `json:"http_address,omitempty"`
	GRPCAddress string `json:"grpc_address,omitempty"`
}

func newKVResponse(key string, value []byte) *KVResponse {
	return &KVResponse{
		Key:   key,
		Value: base64.StdEncoding.EncodeToString(value),
	}
}
