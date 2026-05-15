package handlers

import (
	"net/http"
	"strings"

	"coordinator/internal/adapters/http/dto"
)

func (h *Handler) Sources(w http.ResponseWriter, r *http.Request) {
	zoneID := strings.TrimSpace(r.URL.Query().Get("zone_id"))
	items := dto.SourcesListFromDomain(h.svc.Sources(zoneID))
	WriteJSON(w, http.StatusOK, dto.SourcesResponse{Items: items})
}
