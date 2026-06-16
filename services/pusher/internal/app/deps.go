package app

import (
	"context"
	"strings"

	"github.com/rs/zerolog"

	chwriter "pusher/internal/adapters/clickhouse"
	kafkaconsumer "pusher/internal/adapters/kafka"
	"pusher/internal/adapters/metrics"
	s3store "pusher/internal/adapters/s3"
	"pusher/internal/config"
	"pusher/internal/core/services"
	"pusher/internal/logging"
)

type deps struct {
	log      *zerolog.Logger
	handlers *handlers
	metrics  *metrics.Server
	push     *services.PushService
}

func (app *App) initConfigAndDependencies(ctx context.Context) {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	app.Cfg = cfg

	logger := logging.NewZerolog("pusher")

	var metricsServer *metrics.Server
	if cfg.Metrics.Enabled {
		srv, mErr := metrics.NewServer(cfg.Metrics, &logger)
		if mErr != nil {
			logger.Fatal().Err(mErr).Msg("metrics server")
		}
		metricsServer = srv
	}

	chw, err := chwriter.New(ctx,
		cfg.Clickhouse.Addr,
		cfg.Clickhouse.Database,
		cfg.Clickhouse.User,
		cfg.Clickhouse.Password,
		cfg.Clickhouse.IncidentsTable,
		cfg.Clickhouse.CongestionTable,
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("clickhouse")
	}

	var s3c *s3store.Client
	if ep := strings.TrimSpace(cfg.S3.Endpoint); ep != "" && strings.TrimSpace(cfg.S3.Bucket) != "" {
		s3c, err = s3store.New(ctx, ep, cfg.S3.Region, cfg.S3.Bucket, cfg.S3.AccessKey, cfg.S3.SecretKey)
		if err != nil {
			logger.Fatal().Err(err).Msg("s3 client")
		}
		if err := s3c.EnsureBucket(ctx); err != nil {
			logger.Fatal().Err(err).Msg("s3 bucket")
		}
	}

	push := services.NewPushService(&logger, s3c, chw)

	app.deps = deps{
		log:      &logger,
		handlers: newHandlers(),
		metrics:  metricsServer,
		push:     push,
	}
}

// Logger доступ к логгеру.
func (app *App) Logger() *zerolog.Logger {
	return app.deps.log
}

// RunKafkaConsumer фоновый consumer (вызывается из Run).
func (app *App) RunKafkaConsumer(ctx context.Context) {
	kafkaconsumer.RunConsumers(ctx, app.deps.push, app.Cfg, app.deps.log)
}
