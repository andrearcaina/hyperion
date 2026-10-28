package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
)

type Entry struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (e Entry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}{
		Key:   e.Key,
		Value: base64.StdEncoding.EncodeToString(e.Value),
	})
}

type Client interface {
	Put(ctx context.Context, key string, value []byte) (Entry, error)
	Get(ctx context.Context, key string) (Entry, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]Entry, error)
	Join(ctx context.Context, nodeID, raftAddress, httpAddress, grpcAddress string) error
	MoveLeader(ctx context.Context, nodeID string) error
	Close() error
}
