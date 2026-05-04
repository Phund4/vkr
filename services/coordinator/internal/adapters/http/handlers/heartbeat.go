package handlers

import (
	"encoding/json"
	"net/http"

	"coordinator/internal/adapters/http/dto"
)

func (h *Handler) WorkerHeartbeat(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var in dto.WorkerHeartbeatIn
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		WriteJSON(w, http.StatusBadRequest, dto.ErrorBody{Error: "invalid json"})
		return
	}
	if !in.Validate() {
		WriteJSON(w, http.StatusBadRequest, dto.ErrorBody{Error: "zone_id, cluster_id, instance_id are required"})
		return
	}
	h.svc.UpsertHeartbeat(in.ToDomain())
	WriteJSON(w, http.StatusNoContent, nil)
}
