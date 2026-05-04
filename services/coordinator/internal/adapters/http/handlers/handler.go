package handlers

import "coordinator/internal/core/services"

// Handler HTTP-обработчики coordinator API.
type Handler struct {
	svc *services.CoordinatorService
}

// New создаёт обработчики поверх сервиса.
func New(svc *services.CoordinatorService) *Handler {
	return &Handler{svc: svc}
}
