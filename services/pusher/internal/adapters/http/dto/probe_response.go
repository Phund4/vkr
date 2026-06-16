package dto

// ProbeResponse ответ liveness/readiness.
type ProbeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}
