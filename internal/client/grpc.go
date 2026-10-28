package client

import (
	"context"
	"time"

	hyperionv1 "github.com/andrearcaina/hyperion/proto/hyperion/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClient struct {
	connection *grpc.ClientConn
	client     hyperionv1.HyperionServiceClient
	timeout    time.Duration
}

func NewGRPCClient(address string, timeout time.Duration) (*GRPCClient, error) {
	connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &GRPCClient{
		connection: connection,
		client:     hyperionv1.NewHyperionServiceClient(connection),
		timeout:    timeout,
	}, nil
}

func (c *GRPCClient) Put(ctx context.Context, key string, value []byte) (Entry, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	response, err := c.client.Put(ctx, &hyperionv1.PutRequest{
		Key:   key,
		Value: value,
	})
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

func (c *GRPCClient) Join(ctx context.Context, nodeID, raftAddress, httpAddress, grpcAddress string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.Join(ctx, &hyperionv1.JoinRequest{
		NodeId:      nodeID,
		RaftAddress: raftAddress,
		HttpAddress: httpAddress,
		GrpcAddress: grpcAddress,
	})
	return err
}

func (c *GRPCClient) MoveLeader(ctx context.Context, nodeID string) error {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	_, err := c.client.MoveLeader(ctx, &hyperionv1.MoveLeaderRequest{
		NodeId: nodeID,
	})
	return err
}

func (c *GRPCClient) Close() error {
	return c.connection.Close()
}

func entryFromProto(entry *hyperionv1.Entry) Entry {
	if entry == nil {
		return Entry{}
	}

	return Entry{
		Key:   entry.Key,
		Value: entry.Value,
	}
}
