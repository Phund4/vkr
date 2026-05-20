package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"pusher/internal/config"
)

// Server HTTP endpoint /metrics (default registry, включая promauto-счётчики).
type Server struct {
	httpServer *http.Server
	config     config.PrometheusConfig
	logger     *zerolog.Logger
}

const httpMetricServerAddrPattern = ":%d"

// NewServer регистрирует опциональные collectors и mux с promhttp.Handler().
func NewServer(cfg config.MetricsConfig, logger *zerolog.Logger) (*Server, error) {
	reg := prometheus.DefaultRegisterer

	if cfg.Prometheus.EnableGoCollector {
		if err := reg.Register(collectors.NewGoCollector()); err != nil {
			var ar prometheus.AlreadyRegisteredError
			if !errors.As(err, &ar) {
				return nil, fmt.Errorf("register go collector: %w", err)
			}
		}
	}
	if cfg.Prometheus.EnableProcessCollector {
		if err := reg.Register(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{})); err != nil {
			var ar prometheus.AlreadyRegisteredError
			if !errors.As(err, &ar) {
				return nil, fmt.Errorf("register process collector: %w", err)
			}
		}
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Prometheus.Path, promhttp.Handler())

	return &Server{
		httpServer: &http.Server{
			Addr:    fmt.Sprintf(httpMetricServerAddrPattern, cfg.Prometheus.Port),
			Handler: mux,
		},
		config: cfg.Prometheus,
		logger: logger,
	}, nil
}

// Start ListenAndServe в фоне.
func (s *Server) Start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("metrics server error")
		}
	}()
	s.logger.Info().Msgf("metrics server on %s path %s", s.httpServer.Addr, s.config.Path)
}

// Stop graceful shutdown.
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
