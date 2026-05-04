package app

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"

	"data-service/internal/adapters/clickhouse"
	"data-service/internal/adapters/metrics"
	"data-service/internal/config"
	"data-service/internal/core/services"
)

const (
	LocalEnvFile = ".env.local"
	ConfigFile   = ".env"
	ConfigType   = "env"
)

// deps содержит все зависимости, необходимые для приложения
type deps struct {
	log      *zerolog.Logger
	handlers *handlers
	services *services.ProductService
	metrics  *metrics.Server
}

// initConfigAndDependencies загружает конфиг и собирает зависимости приложения.
func (app *App) initConfigAndDependencies(ctx context.Context) {
	config.SetupViper(ConfigFile, ConfigType)
	config.LoadAdditionalEnv(LocalEnvFile)

	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app.Cfg = cfg

	// Настройка логгера
	zerolog.TimeFieldFormat = time.RFC3339
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		adapter := metrics.NewPrometheusAdapter(cfg.Metrics.Namespace, cfg.Metrics.Subsystem)
		srv, metricsErr := metrics.NewServer(cfg.Metrics, adapter, &logger)
		if metricsErr != nil {
			logger.Fatal().Err(metricsErr).Msg("failed to initialize metrics server")
		}
		metricsServer = srv
	}

	clickhouseRepo, err := app.initClickhouseDependency(ctx)
	if err != nil {
		panic(err)
	}

	service := services.NewProductService(clickhouseRepo)

	app.deps = deps{
		log:      &logger,
		handlers: GetHandlers(service),
		services: service,
		metrics:  metricsServer,
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

	clickhouseRepo, err := clickhouse.NewRepository(ctx, clickhouseClient)
	if err != nil {
		return nil, err
	}

	if err := clickhouseRepo.Ping(ctx); err != nil {
		return nil, err
	}

	return clickhouseRepo, nil
}
