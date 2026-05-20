package app

import (
	"context"

	zlog "github.com/rs/zerolog/log"

	ingestkafka "traffic-analytics/internal/adapters/kafka"
)

// Run инициализирует зависимости и HTTP до завершения rootCtx.
func Run(rootCtx context.Context) error {
	deps, err := InitializeDependencies(rootCtx)
	if err != nil {
		return err
	}
	defer func() {
		if err := deps.Close(); err != nil {
			zlog.Warn().Err(err).Msg("deps close")
		}
	}()

	if deps.Config.KafkaBootstrap != "" {
		go func() {
			ingestkafka.RunIngestConsumer(rootCtx, deps.Ingest, deps.Config)
		}()
	}

	return RunHTTPServer(rootCtx, deps)
}
