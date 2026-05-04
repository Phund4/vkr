package handlers

import (
	"net/http"
	"strings"

	"coordinator/internal/adapters/http/dto"
)

func (h *Handler) IngestionInstances(w http.ResponseWriter, r *http.Request) {
	zoneID := strings.TrimSpace(r.URL.Query().Get("zone_id"))
	items := dto.IngestionInstancesFromDomain(h.svc.IngestionInstances(zoneID))
	WriteJSON(w, http.StatusOK, dto.IngestionInstancesResponse{Items: items})
}
