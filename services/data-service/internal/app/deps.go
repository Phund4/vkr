package app

import (
	"context"

	"github.com/rs/zerolog"

	"data-service/internal/adapters/clickhouse"
	"data-service/internal/adapters/metrics"
	s3store "data-service/internal/adapters/s3"
	"data-service/internal/config"
	"data-service/internal/core/services"
	"data-service/internal/logging"
)

// deps содержит все зависимости, необходимые для приложения
type deps struct {
	log              *zerolog.Logger
	handlers         *handlers
	services         *services.RoadDataService
	metrics          *metrics.Server
	metricsAdapter   *metrics.PrometheusAdapter
}

// initConfigAndDependencies загружает конфиг и собирает зависимости приложения.
func (app *App) initConfigAndDependencies(ctx context.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app.Cfg = cfg

	logger := logging.NewZerolog("data-service")

	var metricsServer *metrics.Server
	var promAdapter *metrics.PrometheusAdapter
	if cfg.Metrics.Enabled {
		promAdapter = metrics.NewPrometheusAdapter(cfg.Metrics.Namespace, cfg.Metrics.Subsystem)
		srv, metricsErr := metrics.NewServer(cfg.Metrics, promAdapter, &logger)
		if metricsErr != nil {
			logger.Fatal().Err(metricsErr).Msg("failed to initialize metrics server")
		}
		metricsServer = srv
	}

	clickhouseRepo, err := app.initClickhouseDependency(ctx)
	if err != nil {
		panic(err)
	}

	var frameSigner services.FrameURLSigner
	if cfg.S3.Enabled {
		presigner, perr := s3store.NewPresigner(ctx, cfg.S3.Endpoint, cfg.S3.Region, cfg.S3.Bucket, cfg.S3.AccessKey, cfg.S3.SecretKey, 0)
		if perr != nil {
			panic(perr)
		}
		frameSigner = presigner
	}

	service := services.NewRoadDataService(clickhouseRepo, frameSigner)

	app.deps = deps{
		log:            &logger,
		handlers:       GetHandlers(service),
		services:       service,
		metrics:        metricsServer,
		metricsAdapter: promAdapter,
	}
}

// Logger возвращает инициализированный логгер приложения.
func (app *App) Logger() *zerolog.Logger {
	return app.deps.log
}

// initClickhouseDependency инициализирует репозиторий ClickHouse и проверяет соединение.
func (app *App) initClickhouseDependency(ctx context.Context) (*clickhouse.Repository, error) {
	db, err := clickhouse.NewConn(app.Cfg.Clickhouse.EffectiveDSN())
	if err != nil {
		return nil, err
	}

	clickhouseClient, err := app.newClickhouseClient(ctx, db)
	if err != nil {
		return nil, err
	}

	clickhouseRepo, err := clickhouse.NewRepository(ctx, clickhouseClient, app.Cfg.Clickhouse.Database)
	if err != nil {
		return nil, err
	}

	if err := clickhouseRepo.Ping(ctx); err != nil {
		return nil, err
	}

	return clickhouseRepo, nil
}
