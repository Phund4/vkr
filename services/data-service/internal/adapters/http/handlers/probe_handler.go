package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"data-service/internal/adapters/http/dto"
)

// ProbeHandler HTTP-обработчик readiness/liveness endpointов.
type ProbeHandler struct{}

// NewProbeHandler создает обработчик probe endpointов.
func NewProbeHandler() *ProbeHandler {
	return &ProbeHandler{}
}

// ok возвращает стандартный успешный probe-ответ.
func (h *ProbeHandler) ok(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.ProbeResponse{
		Status:  dto.StatusOK,
		Message: dto.MessageOK,
	})
}

// @Summary Readiness probe
// @Success 200 {object} dto.ProbeResponse
// @Router /probe/ready [get]
func (h *ProbeHandler) Ready(c echo.Context) error {
	return h.ok(c)
}

// @Summary Liveness / health
// @Success 200 {object} dto.ProbeResponse
// @Router /health [get]
func (h *ProbeHandler) Health(c echo.Context) error {
	return h.ok(c)
}

// @Summary Health (probe path)
// @Success 200 {object} dto.ProbeResponse
// @Router /probe/health [get]
func (h *ProbeHandler) ProbeHealth(c echo.Context) error {
	return h.ok(c)
}
