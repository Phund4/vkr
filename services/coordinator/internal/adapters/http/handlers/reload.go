package handlers

import (
	"net/http"
	"strings"

	"coordinator/internal/adapters/http/dto"
	"coordinator/internal/core/services"
)

// ReloadAssignments увеличивает revision и шлёт POST /v1/reload живым router из ingestion_instances.
func (h *Handler) ReloadAssignments(w http.ResponseWriter, r *http.Request) {
	zoneID := strings.TrimSpace(r.URL.Query().Get("zone_id"))
	rev := services.BumpAssignmentsRevision()
	results := h.svc.NotifyRoutersReload(r.Context(), zoneID, rev)
	notified := make([]dto.RouterNotifyResultItem, 0, len(results))
	for _, res := range results {
		notified = append(notified, dto.RouterNotifyResultItem{
			ZoneID:     res.ZoneID,
			ClusterID:  res.ClusterID,
			InstanceID: res.InstanceID,
			URL:        res.URL,
			OK:         res.OK,
			Error:      res.Error,
		})
	}
	WriteJSON(w, http.StatusOK, dto.ReloadAssignmentsResponse{
		Revision: rev,
		Notified: notified,
	})
}
