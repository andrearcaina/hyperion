package http

import (
	"fmt"
	"io"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/andrearcaina/hyperion/internal/logger"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db     *db.DB
	logger *logger.Logger
}

func NewHandler(db *db.DB, logger *logger.Logger) *Handler {
	return &Handler{
		db:     db,
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

	return r
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Debug(r.Context(), "failed to read request body", "error", err)
		writeError(w, http.StatusBadRequest, fmt.Errorf("failed to read request body: %v", err))
		return
	}

	if err := h.db.Set([]byte(key), body); err != nil {
		h.logger.Error(r.Context(), "failed to set key", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to set key: %s", key))
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{
		Key:   key,
		Value: string(body),
	})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	val, err := h.db.Get([]byte(key))
	if err != nil {
		h.logger.Error(r.Context(), "failed to get key", "error", err)
		writeError(w, http.StatusNotFound, fmt.Errorf("key not found: %s", key))
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{
		Key:   key,
		Value: string(val),
	})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	if err := h.db.Delete([]byte(key)); err != nil {
		h.logger.Error(r.Context(), "failed to delete key", "error", err)
		writeError(w, http.StatusNotFound, fmt.Errorf("key not found: %s", key))
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}

func (h *Handler) ForEach(w http.ResponseWriter, r *http.Request) {
	results := []KVResponse{}

	err := h.db.ForEach(func(key, value []byte) error {
		results = append(results, KVResponse{
			Key:   string(key),
			Value: string(value),
		})

		return nil
	})
	if err != nil {
		h.logger.Error(r.Context(), "failed to iterate over key-value pairs", "error", err)
		writeError(w, http.StatusInternalServerError, fmt.Errorf("failed to iterate over key-value pairs: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, results)
}
