package grpc

import (
	"context"
	"errors"
	"sort"

	"github.com/andrearcaina/hyperion/internal/store"
	hyperionv1 "github.com/andrearcaina/hyperion/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ForEach(func(key, value []byte) error) error
	Join(nodeID, nodeAddress string) error
}

type Handler struct {
	hyperionv1.UnimplementedHyperionServer
	store Store
}

func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

func (h *Handler) Put(_ context.Context, req *hyperionv1.PutRequest) (*hyperionv1.PutResponse, error) {
	if err := h.store.Set(req.GetKey(), req.GetValue()); err != nil {
		return nil, grpcError(err)
	}

	return &hyperionv1.PutResponse{Entry: &hyperionv1.Entry{Key: req.GetKey(), Value: req.GetValue()}}, nil
}

func (h *Handler) Get(_ context.Context, req *hyperionv1.GetRequest) (*hyperionv1.GetResponse, error) {
	value, err := h.store.Get(req.GetKey())
	if err != nil {
		return nil, grpcError(err)
	}

	return &hyperionv1.GetResponse{Entry: &hyperionv1.Entry{Key: req.GetKey(), Value: value}}, nil
}

func (h *Handler) Delete(_ context.Context, req *hyperionv1.DeleteRequest) (*hyperionv1.DeleteResponse, error) {
	if err := h.store.Delete(req.GetKey()); err != nil {
		return nil, grpcError(err)
	}

	return &hyperionv1.DeleteResponse{}, nil
}

func (h *Handler) List(_ context.Context, _ *hyperionv1.ListRequest) (*hyperionv1.ListResponse, error) {
	response := &hyperionv1.ListResponse{}
	if err := h.store.ForEach(func(key, value []byte) error {
		response.Entries = append(response.Entries, &hyperionv1.Entry{Key: string(key), Value: value})
		return nil
	}); err != nil {
		return nil, grpcError(err)
	}

	sort.Slice(response.Entries, func(i, j int) bool { return response.Entries[i].Key < response.Entries[j].Key })
	return response, nil
}

func (h *Handler) Join(_ context.Context, req *hyperionv1.JoinRequest) (*hyperionv1.JoinResponse, error) {
	if err := h.store.Join(req.GetNodeId(), req.GetRaftAddress()); err != nil {
		return nil, grpcError(err)
	}

	return &hyperionv1.JoinResponse{}, nil
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
