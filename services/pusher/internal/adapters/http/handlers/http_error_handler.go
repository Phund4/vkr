package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"pusher/internal/adapters/http/dto"
)

// HTTPErrorHandler единый JSON для ошибок Echo.
func HTTPErrorHandler(_ *zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		message := http.StatusText(http.StatusInternalServerError)

		var httpErr *echo.HTTPError
		if errors.As(err, &httpErr) {
			status = httpErr.Code
			message = http.StatusText(status)
		}

		_ = c.JSON(status, dto.ErrorResponse{
			Status:  "error",
			Message: message,
		})
	}
}
