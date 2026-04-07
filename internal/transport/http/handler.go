package http

import (
	"io"
	"net/http"

	"github.com/andrearcaina/hyperion/internal/db"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	db *db.DB
}

func NewHandler(db *db.DB) *Handler {
	return &Handler{db: db}
}

func (h *Handler) ServeRoutes() chi.Router {
	r := chi.NewRouter()

	r.Route("/kv", func(r chi.Router) {
		r.Put("/{key}", h.Set)
		r.Get("/{key}", h.Get)
		r.Delete("/{key}", h.Delete)
	})

	return r
}

func (h *Handler) Set(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, &KVResponse{Error: "failed to read body"})
		return
	}

	if err := h.db.Set([]byte(key), body); err != nil {
		writeJSON(w, http.StatusInternalServerError, &KVResponse{Error: "failed to set key"})
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{Key: key, Value: string(body)})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	val, err := h.db.Get([]byte(key))
	if err != nil {
		writeJSON(w, http.StatusNotFound, &KVResponse{Error: "key not found"})
		return
	}

	writeJSON(w, http.StatusOK, &KVResponse{Key: key, Value: string(val)})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")

	if err := h.db.Delete([]byte(key)); err != nil {
		writeJSON(w, http.StatusInternalServerError, &KVResponse{Error: "failed to delete key"})
		return
	}

	writeJSON(w, http.StatusNoContent, nil)
}
