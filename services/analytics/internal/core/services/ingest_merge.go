package services

import (
	"encoding/json"

	"traffic-analytics/internal/core/domain"
)

// extractMLHalves определяет, какие поддеревья incident/congestion присутствуют в ev.ML.
func extractMLHalves(ev domain.RoadEvent) (hasI, hasC bool, incident, congestion json.RawMessage) {
	if len(ev.ML) == 0 || string(ev.ML) == "null" {
		return false, false, nil, nil
	}
	var probe struct {
		Incident   json.RawMessage `json:"incident"`
		Congestion json.RawMessage `json:"congestion"`
	}
	if err := json.Unmarshal(ev.ML, &probe); err != nil {
		return false, false, nil, nil
	}
	hasI = len(probe.Incident) > 0
	hasC = len(probe.Congestion) > 0
	return hasI, hasC, probe.Incident, probe.Congestion
}
