package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	hyperionv1 "github.com/andrearcaina/hyperion/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"resty.dev/v3"
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
		Value: string(e.Value),
	})
}

type Client interface {
	Put(ctx context.Context, key string, value []byte) (Entry, error)
	Get(ctx context.Context, key string) (Entry, error)
	Delete(ctx context.Context, key string) error
	List(ctx context.Context) ([]Entry, error)
	Join(ctx context.Context, nodeID, raftAddress string) error
	Close() error
}

type HTTPClient struct {
	client *resty.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

func NewHTTP(address string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: resty.New().
			SetBaseURL(address).
			SetTimeout(timeout).
			SetHeader("Accept", "application/json"),
	}
}

func (c *HTTPClient) Put(ctx context.Context, key string, value []byte) (Entry, error) {
	var wire struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := c.do(c.client.R().
		SetContext(ctx).
		SetBody(value).
		SetResult(&wire).
		Put("/hypr/kv/" + url.PathEscape(key)))
	return Entry{
		Key:   wire.Key,
		Value: []byte(wire.Value),
	}, err
}

func (c *HTTPClient) Get(ctx context.Context, key string) (Entry, error) {
	var wire struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	err := c.do(c.client.R().
		SetContext(ctx).
		SetResult(&wire).
		Get("/hypr/kv/" + url.PathEscape(key)))
	return Entry{
		Key:   wire.Key,
		Value: []byte(wire.Value),
	}, err
}

func (c *HTTPClient) Delete(ctx context.Context, key string) error {
	return c.do(c.client.R().
		SetContext(ctx).
		Delete("/hypr/kv/" + url.PathEscape(key)))
}

func (c *HTTPClient) List(ctx context.Context) ([]Entry, error) {
	var wire []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}

	if err := c.do(c.client.R().
		SetContext(ctx).
		SetResult(&wire).
		Get("/hypr/kv/")); err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(wire))
	for _, entry := range wire {
		entries = append(entries, Entry{
			Key:   entry.Key,
			Value: []byte(entry.Value),
		})
	}

	return entries, nil
}

func (c *HTTPClient) Join(ctx context.Context, nodeID, raftAddress string) error {
	body := map[string]string{"node_id": nodeID, "address": raftAddress}

	return c.do(c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post("/hypr/raft/join"))
}

func (c *HTTPClient) Close() error {
	return c.client.Close()
}

func (c *HTTPClient) do(response *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if !response.IsError() {
		return nil
	}

	var body errorResponse
	if err := json.Unmarshal(response.Bytes(), &body); err == nil && body.Error != "" {
		return fmt.Errorf("request failed with status %d: %s", response.StatusCode(), body.Error)
	}

	return fmt.Errorf("request failed with status %d", response.StatusCode())
}

type GRPCClient struct {
	connection *grpc.ClientConn
	client     hyperionv1.HyperionClient
	timeout    time.Duration
}

func NewGRPC(address string, timeout time.Duration) (*GRPCClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{connection: connection, client: hyperionv1.NewHyperionClient(connection), timeout: timeout}, nil
}

func (c *GRPCClient) Put(ctx context.Context, key string, value []byte) (Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.Put(ctx, &hyperionv1.PutRequest{Key: key, Value: value})
	if err != nil {
		return Entry{}, err
	}

	return entryFromProto(response.Entry), nil
}

func (c *GRPCClient) Get(ctx context.Context, key string) (Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.Get(ctx, &hyperionv1.GetRequest{Key: key})
	if err != nil {
		return Entry{}, err
	}

	return entryFromProto(response.Entry), nil
}

func (c *GRPCClient) Delete(ctx context.Context, key string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.Delete(ctx, &hyperionv1.DeleteRequest{Key: key})
	return err
}

func (c *GRPCClient) List(ctx context.Context) ([]Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.List(ctx, &hyperionv1.ListRequest{})
	if err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(response.Entries))
	for _, entry := range response.Entries {
		entries = append(entries, entryFromProto(entry))
	}

	return entries, nil
}

func (c *GRPCClient) Join(ctx context.Context, nodeID, raftAddress string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.Join(ctx, &hyperionv1.JoinRequest{NodeId: nodeID, RaftAddress: raftAddress})
	return err
}

func (c *GRPCClient) Close() error { return c.connection.Close() }

func entryFromProto(entry *hyperionv1.Entry) Entry {
	if entry == nil {
		return Entry{}
	}
	return Entry{Key: entry.Key, Value: entry.Value}
}
