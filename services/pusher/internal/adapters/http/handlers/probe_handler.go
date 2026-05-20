package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"pusher/internal/adapters/http/dto"
)

// ProbeHandler readiness / liveness.
type ProbeHandler struct{}

// NewProbeHandler конструктор.
func NewProbeHandler() *ProbeHandler {
	return &ProbeHandler{}
}

func (h *ProbeHandler) ok(c echo.Context) error {
	return c.JSON(http.StatusOK, dto.ProbeResponse{
		Status:  dto.StatusOK,
		Message: dto.MessageOK,
	})
}

// Ready probe.
func (h *ProbeHandler) Ready(c echo.Context) error {
	return h.ok(c)
}

// Health liveness.
func (h *ProbeHandler) Health(c echo.Context) error {
	return h.ok(c)
}

// ProbeHealth алиас.
func (h *ProbeHandler) ProbeHealth(c echo.Context) error {
	return h.ok(c)
}
