package http

import (
	"encoding/json"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, resp *KVResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}
