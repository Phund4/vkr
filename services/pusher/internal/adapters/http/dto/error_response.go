package dto

// ErrorResponse JSON при ошибке HTTP.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
