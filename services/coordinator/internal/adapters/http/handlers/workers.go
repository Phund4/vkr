package handlers

import (
	"net/http"

	"coordinator/internal/adapters/http/dto"
)

func (h *Handler) Workers(w http.ResponseWriter, _ *http.Request) {
	items := dto.WorkerHeartbeatsFromDomain(h.svc.Heartbeats())
	WriteJSON(w, http.StatusOK, dto.WorkersResponse{Items: items})
}
