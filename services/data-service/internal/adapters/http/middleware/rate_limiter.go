package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"

	"data-service/internal/adapters/http/dto"
)

// RateLimiterConfig конфигурация middleware ограничения частоты запросов.
type RateLimiterConfig struct {
	RPS                rate.Limit
	Burst              int
	ExpiresIn          time.Duration
	TooManyRequestsMsg string
}

// NewRateLimiter создает middleware rate limiter c единым JSON-ответом.
func NewRateLimiter(cfg RateLimiterConfig) echo.MiddlewareFunc {
	return echoMiddleware.RateLimiterWithConfig(echoMiddleware.RateLimiterConfig{
		Store: echoMiddleware.NewRateLimiterMemoryStoreWithConfig(echoMiddleware.RateLimiterMemoryStoreConfig{
			Rate:      cfg.RPS,
			Burst:     cfg.Burst,
			ExpiresIn: cfg.ExpiresIn,
		}),
		ErrorHandler: func(c echo.Context, _ error) error {
			return c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Status:  dto.StatusError,
				Message: dto.ErrInvalidRateLimitIdentifier,
			})
		},
		DenyHandler: func(c echo.Context, _ string, _ error) error {
			return c.JSON(http.StatusTooManyRequests, dto.ErrorResponse{
				Status:  dto.StatusError,
				Message: cfg.TooManyRequestsMsg,
			})
		},
	})
}

// ApiV1RateLimiter возвращает профиль лимитера.
func ApiV1RateLimiter() echo.MiddlewareFunc {
	return NewRateLimiter(RateLimiterConfig{
		RPS:                rate.Limit(dto.ApiV1RateLimitRPS),
		Burst:              dto.ApiV1RateLimitBurst,
		ExpiresIn:          dto.RateLimitExpiresInMinutes,
		TooManyRequestsMsg: dto.ErrApiRateLimitExceeded,
	})
}

// RoadDataRateLimiter лимитер для ручек road_incidents / road_congestion.
func RoadDataRateLimiter() echo.MiddlewareFunc {
	return NewRateLimiter(RateLimiterConfig{
		RPS:                rate.Limit(dto.RoadDataRateLimitRPS),
		Burst:              dto.RoadDataRateLimitBurst,
		ExpiresIn:          dto.RateLimitExpiresInMinutes,
		TooManyRequestsMsg: dto.ErrRoadDataRateLimitExceeded,
	})
}
