package metrics

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/rs/zerolog"

	"data-service/internal/config"
)

var ErrRegisterPrometheusCollector = fmt.Errorf("error registering prometheus collector")

// Server представляет HTTP сервер для экспорта метрик Prometheus
type Server struct {
	httpServer *http.Server
	config     config.PrometheusConfig
	adapter    *PrometheusAdapter
	logger     *zerolog.Logger
}

const HttpMetricServerAddrPattern = ":%d"

func registerDefaultCollectors(cfg config.MetricsConfig, adapter *PrometheusAdapter) error {
	registerCollectorFunc := func(c prometheus.Collector) error {
		if err := adapter.registry.Register(c); err != nil {
			if alreadyRegisteredErr, ok := err.(prometheus.AlreadyRegisteredError); ok {
				_ = alreadyRegisteredErr.ExistingCollector
				return nil
			}
			return err
		}
		return nil
	}

	if cfg.Prometheus.EnableGoCollector {
		if err := registerCollectorFunc(collectors.NewGoCollector()); err != nil {
			return fmt.Errorf("%w: %w", ErrRegisterPrometheusCollector, err)
		}
	}

	if cfg.Prometheus.EnableProcessCollector {
		if err := registerCollectorFunc(
			collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		); err != nil {
			return fmt.Errorf("%w: %w", ErrRegisterPrometheusCollector, err)
		}
	}

	return nil
}

// NewServer создает новый сервер метрик
func NewServer(cfg config.MetricsConfig, adapter *PrometheusAdapter, logger *zerolog.Logger) (*Server, error) {
	mux := http.NewServeMux()
	mux.Handle(cfg.Prometheus.Path, adapter.PromHandler())

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(HttpMetricServerAddrPattern, cfg.Prometheus.Port),
		Handler: mux,
	}

	if err := registerDefaultCollectors(cfg, adapter); err != nil {
		return nil, err
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg.Prometheus,
		adapter:    adapter,
		logger:     logger,
	}, nil
}

// Start запускает сервер метрик
func (s *Server) Start() {
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error().Err(err).Msg("Error starting metrics server")
		}
	}()

	s.logger.Info().Msgf(
		"Metrics Server started on port %d with path %s",
		s.config.Port,
		s.config.Path,
	)
}

// Stop останавливает сервер метрик
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
