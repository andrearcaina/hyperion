package client

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"resty.dev/v3"
)

type HTTPClient struct {
	client *resty.Client
}

type errorResponse struct {
	Error string `json:"error"`
}

type httpEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func NewHTTPClient(address string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		client: resty.New().
			SetBaseURL(address).
			SetTimeout(timeout).
			SetHeader("Accept", "application/json"),
	}
}

func (c *HTTPClient) Put(ctx context.Context, key string, value []byte) (Entry, error) {
	var wire httpEntry

	err := c.do(c.client.R().
		SetContext(ctx).
		SetBody(value).
		SetResult(&wire).
		Put("/hypr/kv/" + url.PathEscape(key)))

	entry, decodeErr := entryFromHTTP(wire)
	if err != nil {
		return entry, err
	}

	return entry, decodeErr
}

func (c *HTTPClient) Get(ctx context.Context, key string) (Entry, error) {
	var wire httpEntry

	err := c.do(c.client.R().
		SetContext(ctx).
		SetResult(&wire).
		Get("/hypr/kv/" + url.PathEscape(key)))

	entry, decodeErr := entryFromHTTP(wire)
	if err != nil {
		return entry, err
	}

	return entry, decodeErr
}

func (c *HTTPClient) Delete(ctx context.Context, key string) error {
	return c.do(c.client.R().
		SetContext(ctx).
		Delete("/hypr/kv/" + url.PathEscape(key)))
}

func (c *HTTPClient) List(ctx context.Context) ([]Entry, error) {
	var wire []httpEntry

	if err := c.do(c.client.R().
		SetContext(ctx).
		SetResult(&wire).
		Get("/hypr/kv/")); err != nil {
		return nil, err
	}

	entries := make([]Entry, 0, len(wire))
	for _, entry := range wire {
		decoded, err := entryFromHTTP(entry)
		if err != nil {
			return nil, err
		}

		entries = append(entries, decoded)
	}

	return entries, nil
}

func entryFromHTTP(entry httpEntry) (Entry, error) {
	value, err := base64.StdEncoding.DecodeString(entry.Value)
	if err != nil {
		return Entry{}, fmt.Errorf("decode HTTP value: %w", err)
	}

	return Entry{
		Key:   entry.Key,
		Value: value,
	}, nil
}

func (c *HTTPClient) Join(ctx context.Context, nodeID, raftAddress, httpAddress, grpcAddress string) error {
	body := map[string]string{
		"node_id": nodeID, "address": raftAddress,
		"http_address": httpAddress, "grpc_address": grpcAddress,
	}

	return c.do(c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post("/hypr/raft/join"))
}

func (c *HTTPClient) MoveLeader(ctx context.Context, nodeID string) error {
	body := map[string]string{"node_id": nodeID}

	return c.do(c.client.R().
		SetContext(ctx).
		SetBody(body).
		Post("/hypr/raft/move-leader"))
}

func (c *HTTPClient) Close() error {
	return c.client.Close()
}

func (c *HTTPClient) do(response *resty.Response, err error) error {
	if err != nil {
		return err
	}

	if response.IsStatusSuccess() {
		return nil
	}

	var body errorResponse
	if err := json.Unmarshal(response.Bytes(), &body); err == nil && body.Error != "" {
		return fmt.Errorf("request failed with status %d: %s", response.StatusCode(), body.Error)
	}

	return fmt.Errorf("request failed with status %d", response.StatusCode())
}
