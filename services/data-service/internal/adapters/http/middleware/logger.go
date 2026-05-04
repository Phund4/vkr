package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog"
)

// Logger возвращает middleware логирования HTTP-запросов.
func Logger(logger *zerolog.Logger) echo.MiddlewareFunc {
	return echoMiddleware.RequestLoggerWithConfig(echoMiddleware.RequestLoggerConfig{
		LogURI:       true,
		LogMethod:    true,
		LogStatus:    true,
		LogLatency:   true,
		LogRemoteIP:  true,
		LogRequestID: true,
		LogError:     true,
		LogValuesFunc: func(_ echo.Context, v echoMiddleware.RequestLoggerValues) error {
			logEvent := logger.Info()

			if v.Error != nil || v.Status >= http.StatusInternalServerError {
				logEvent = logger.Error().Err(v.Error)
			} else if v.Status >= http.StatusBadRequest {
				logEvent = logger.Warn()
			}

			logEvent.
				Str("request_id", v.RequestID).
				Str("remote_ip", v.RemoteIP).
				Str("method", v.Method).
				Str("uri", v.URI).
				Int("status", v.Status).
				Dur("latency", v.Latency).
				Msg("http request completed")

			return nil
		},
	})
}
