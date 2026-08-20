package grpc

import (
	"context"
	"errors"

	"github.com/andrearcaina/hyperion/internal/store"
	hyperionv1 "github.com/andrearcaina/hyperion/proto/hyperion/v1"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const forwardedMetadataKey = "x-hyperion-forwarded"

func forwardRPC[T any](
	h *Handler,
	ctx context.Context,
	err error,
	call func(context.Context, hyperionv1.HyperionServiceClient) (T, error),
) (T, bool, error) {
	var zero T
	var leader *store.NotLeaderError
	if !errors.As(err, &leader) || leader.LeaderGRPCAddress == "" || wasForwarded(ctx) {
		return zero, false, nil
	}

	connection, dialErr := googlegrpc.NewClient(leader.LeaderGRPCAddress, h.dialOptions...)
	if dialErr != nil {
		return zero, true, status.Error(codes.Unavailable, "failed to connect to Raft leader")
	}
	defer connection.Close()

	response, callErr := call(forwardContext(ctx), hyperionv1.NewHyperionServiceClient(connection))
	return response, true, callErr
}

func (h *Handler) forwardPut(ctx context.Context, req *hyperionv1.PutRequest, err error) (*hyperionv1.PutResponse, bool, error) {
	return forwardRPC(h, ctx, err, func(ctx context.Context, client hyperionv1.HyperionServiceClient) (*hyperionv1.PutResponse, error) {
		return client.Put(ctx, req)
	})
}

func (h *Handler) forwardGet(ctx context.Context, req *hyperionv1.GetRequest, err error) (*hyperionv1.GetResponse, bool, error) {
	return forwardRPC(h, ctx, err, func(ctx context.Context, client hyperionv1.HyperionServiceClient) (*hyperionv1.GetResponse, error) {
		return client.Get(ctx, req)
	})
}

func (h *Handler) forwardDelete(ctx context.Context, req *hyperionv1.DeleteRequest, err error) (*hyperionv1.DeleteResponse, bool, error) {
	return forwardRPC(h, ctx, err, func(ctx context.Context, client hyperionv1.HyperionServiceClient) (*hyperionv1.DeleteResponse, error) {
		return client.Delete(ctx, req)
	})
}

func (h *Handler) forwardList(ctx context.Context, req *hyperionv1.ListRequest, err error) (*hyperionv1.ListResponse, bool, error) {
	return forwardRPC(h, ctx, err, func(ctx context.Context, client hyperionv1.HyperionServiceClient) (*hyperionv1.ListResponse, error) {
		return client.List(ctx, req)
	})
}

func (h *Handler) forwardJoin(ctx context.Context, req *hyperionv1.JoinRequest, err error) (*hyperionv1.JoinResponse, bool, error) {
	return forwardRPC(h, ctx, err, func(ctx context.Context, client hyperionv1.HyperionServiceClient) (*hyperionv1.JoinResponse, error) {
		return client.Join(ctx, req)
	})
}

func wasForwarded(ctx context.Context) bool {
	return len(metadata.ValueFromIncomingContext(ctx, forwardedMetadataKey)) != 0
}

func forwardContext(ctx context.Context) context.Context {
	return metadata.AppendToOutgoingContext(ctx, forwardedMetadataKey, "1")
}
