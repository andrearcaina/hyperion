package http

import (
	"encoding/json"
	"net/http"
)

func writeError(w http.ResponseWriter, status int, err error) {
	resp := &KVResponse{
		Error: err.Error(),
	}
	writeJSON(w, status, resp)
}

func writeJSON(w http.ResponseWriter, status int, resp interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
