package middleware

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// Common возвращает стандартный набор middleware для HTTP-сервера.
func Common() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		echoMiddleware.RequestID(),
		echoMiddleware.Recover(),
	}
}
