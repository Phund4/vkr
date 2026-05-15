package handlers

import (
	"encoding/json"
	"net/http"
)

// WriteJSON пишет JSON-ответ; для 204 без тела.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	if status == http.StatusNoContent {
		w.WriteHeader(status)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
