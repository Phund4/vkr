package handlers

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog"

	"data-service/internal/adapters/http/dto"
)

// HTTPErrorHandler создает обработчик HTTP-ошибок с единым JSON-форматом ответа.
func HTTPErrorHandler(_ *zerolog.Logger) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		status := http.StatusInternalServerError
		message := http.StatusText(http.StatusInternalServerError)

		if httpErr, ok := errors.AsType[*echo.HTTPError](err); ok {
			status = httpErr.Code
			message = http.StatusText(status)
		}

		_ = c.JSON(status, dto.ErrorResponse{
			Status:  dto.StatusError,
			Message: message,
		})
	}
}
