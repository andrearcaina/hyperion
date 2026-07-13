package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/andrearcaina/hyperion/internal/store"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	store  Store
	logger *logger.Logger
}

type Store interface {
	Set(key string, value []byte) error
	Get(key string) ([]byte, error)
	Delete(key string) error
	ForEach(func(key, value []byte) error) error
	Join(nodeID, nodeAddress string) error
}

func NewHandler(st Store, logger *logger.Logger) *Handler {
	return &Handler{
		store:  st,
		logger: logger,
	}
}

func (h *Handler) ServeRoutes() chi.Router {
	r := chi.NewRouter()

	r.Route("/kv", func(r chi.Router) {
		r.Put("/{key}", h.Set)
		r.Get("/{key}", h.Get)
		r.Delete("/{key}", h.Delete)
		r.Get("/", h.ForEach)
	})

	r.Route("/raft", func(r chi.Router) {
		r.Post("/join", h.Join)
	})

	return r
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	const maxValueSize = 4 << 20

	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxValueSize))
	if err != nil {
		h.logger.Debug(r.Context(), "failed to read request body", "error", err)
		writeError(w, http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err))
		return
	}

	if err := h.store.Set(key, body); err != nil {
		h.logger.Error(r.Context(), "failed to set key", "error", err)
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{
		Key:   key,
		Value: string(body),
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	val, err := h.store.Get(key)
	if err != nil {
		h.logger.Debug(r.Context(), "failed to get key", "key", key, "error", err)
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{
		Key:   key,
		Value: string(val),
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	if err := h.store.Delete(key); err != nil {
		h.logger.Error(r.Context(), "failed to delete key", "error", err)
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ForEach(w http.ResponseWriter, r *http.Request) {
	results := []KVResponse{}

	err := h.store.ForEach(func(key, value []byte) error {
		results = append(results, KVResponse{
			Key:   string(key),
			Value: string(value),
		})

		return nil
	})
	if err != nil {
		h.logger.Error(r.Context(), "failed to iterate over key-value pairs", "error", err)
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	var req JoinRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	if err := h.store.Join(req.NodeID, req.Address); err != nil {
		writeStoreError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"status": "joined",
	})
}

func writeStoreError(w http.ResponseWriter, err error) {
	var notLeader *store.NotLeaderError

	switch {
	case errors.Is(err, store.ErrInvalidKey):
		writeError(w, http.StatusBadRequest, err)
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, err)
	case errors.As(err, &notLeader):
		writeError(w, http.StatusConflict, err)
	default:
		writeError(w, http.StatusInternalServerError, errors.New("internal server error"))
	}
}
