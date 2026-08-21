package grpc

import (
	"context"
	"errors"
	"sort"

	"github.com/andrearcaina/hyperion/internal/store"
	hyperionv1 "github.com/andrearcaina/hyperion/proto/hyperion/v1"
	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ForEach(func(key, value []byte) error) error
	Join(nodeID, nodeAddress, httpAddress, grpcAddress string) error
	TransferLeadership(nodeID string) error
}

type Handler struct {
	hyperionv1.UnimplementedHyperionServiceServer
	store       Store
	dialOptions []googlegrpc.DialOption
}

func NewHandler(store Store) *Handler {
	return &Handler{
		store: store,
		dialOptions: []googlegrpc.DialOption{
			googlegrpc.WithTransportCredentials(insecure.NewCredentials()),
		},
	}
}

func (h *Handler) Put(ctx context.Context, req *hyperionv1.PutRequest) (*hyperionv1.PutResponse, error) {
	if err := h.store.Set(req.GetKey(), req.GetValue()); err != nil {
		if response, forwarded, forwardErr := h.forwardPut(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	return &hyperionv1.PutResponse{Entry: &hyperionv1.Entry{Key: req.GetKey(), Value: req.GetValue()}}, nil
}

func (h *Handler) Get(ctx context.Context, req *hyperionv1.GetRequest) (*hyperionv1.GetResponse, error) {
	value, err := h.store.Get(req.GetKey())
	if err != nil {
		if response, forwarded, forwardErr := h.forwardGet(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	return &hyperionv1.GetResponse{Entry: &hyperionv1.Entry{Key: req.GetKey(), Value: value}}, nil
}

func (h *Handler) Delete(ctx context.Context, req *hyperionv1.DeleteRequest) (*hyperionv1.DeleteResponse, error) {
	if err := h.store.Delete(req.GetKey()); err != nil {
		if response, forwarded, forwardErr := h.forwardDelete(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	return &hyperionv1.DeleteResponse{}, nil
}

func (h *Handler) List(ctx context.Context, req *hyperionv1.ListRequest) (*hyperionv1.ListResponse, error) {
	response := &hyperionv1.ListResponse{}
	if err := h.store.ForEach(func(key, value []byte) error {
		response.Entries = append(response.Entries, &hyperionv1.Entry{Key: string(key), Value: value})

		return nil
	}); err != nil {
		if response, forwarded, forwardErr := h.forwardList(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	sort.Slice(response.Entries, func(i, j int) bool { return response.Entries[i].Key < response.Entries[j].Key })
	return response, nil
}

func (h *Handler) Join(ctx context.Context, req *hyperionv1.JoinRequest) (*hyperionv1.JoinResponse, error) {
	if err := h.store.Join(req.GetNodeId(), req.GetRaftAddress(), req.GetHttpAddress(), req.GetGrpcAddress()); err != nil {
		if response, forwarded, forwardErr := h.forwardJoin(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	return &hyperionv1.JoinResponse{}, nil
}

func (h *Handler) TransferLeadership(ctx context.Context, req *hyperionv1.TransferLeadershipRequest) (*hyperionv1.TransferLeadershipResponse, error) {
	if err := h.store.TransferLeadership(req.GetNodeId()); err != nil {
		if response, forwarded, forwardErr := h.forwardTransferLeadership(ctx, req, err); forwarded {
			return response, forwardErr
		}

		return nil, grpcError(err)
	}

	return &hyperionv1.TransferLeadershipResponse{}, nil
}

func grpcError(err error) error {
	var notLeader *store.NotLeaderError

	switch {
	case errors.Is(err, store.ErrInvalidKey):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, store.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.As(err, &notLeader):
		return status.Error(codes.FailedPrecondition, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
