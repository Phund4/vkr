package handlers

import (
	"net/http"
	"strings"

	"coordinator/internal/adapters/http/dto"
	"coordinator/internal/core/services"
)

func (h *Handler) Assignments(w http.ResponseWriter, r *http.Request) {
	zoneID := strings.TrimSpace(r.URL.Query().Get("zone_id"))
	clusterID := strings.TrimSpace(r.URL.Query().Get("cluster_id"))
	instanceID := strings.TrimSpace(r.URL.Query().Get("instance_id"))
	dataClass := strings.TrimSpace(r.URL.Query().Get("data_class"))
	items := dto.SourcesListFromDomain(h.svc.Assignments(zoneID, clusterID, instanceID, dataClass))
	WriteJSON(w, http.StatusOK, dto.AssignmentsResponse{
		Revision: services.CurrentAssignmentsRevision(),
		Items:    items,
	})
}
