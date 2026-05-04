package dto

// ProbeResponse DTO ответа для liveness/readiness endpoint.
type ProbeResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}
