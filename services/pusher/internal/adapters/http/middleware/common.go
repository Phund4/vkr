package middleware

import (
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

// Common стандартные middleware.
func Common() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		echoMiddleware.RequestID(),
		echoMiddleware.Recover(),
	}
}
