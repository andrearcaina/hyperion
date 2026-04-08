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
	if status == http.StatusNoContent { // no content, no response body
		w.WriteHeader(status)
		return
	}

	data, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(data)
}
