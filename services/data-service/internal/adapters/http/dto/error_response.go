package dto

// ErrorResponse DTO стандартного ответа с ошибкой.
type ErrorResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}
