package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"

	"data-service/internal/core/domain"
)

// EchoMetricsMiddleware считает RPS и длительность HTTP (method, path Echo route, status).
func (p *PrometheusAdapter) EchoMetricsMiddleware() echo.MiddlewareFunc {
	var once sync.Once
	var req domain.CounterMetric
	var dur domain.HistogramMetric

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			once.Do(func() {
				req = p.Counter("http_requests_total", "HTTP API requests", "method", "path", "code")
				dur = p.Histogram(
					"http_request_duration_seconds",
					"HTTP request wall time in seconds",
					prometheus.DefBuckets,
					"method", "path",
				)
			})

			start := time.Now()
			err := next(c)

			status := c.Response().Status
			if status == 0 {
				status = 200
			}
			path := c.Path()
			if path == "" {
				path = c.Request().URL.Path
			}
			method := c.Request().Method
			code := strconv.Itoa(status)

			req.Inc(method, path, code)
			dur.Observe(time.Since(start).Seconds(), method, path)

			return err
		}
	}
}
