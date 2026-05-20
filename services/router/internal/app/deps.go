package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"

	coordinatorclient "router/internal/adapters/coordinator"
	kafkapub "router/internal/adapters/kafka"
	s3store "router/internal/adapters/s3"
	"router/internal/config"
)

// initDeps загружает YAML/ENV и создаёт HTTP-клиент coordinator.
func (a *App) initDeps(ctx context.Context) error {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	base := config.CoordinatorBaseURLFromEnv()
	if base == "" {
		return ErrCoordinatorBaseURL
	}
	if config.CoordinatorZoneIDFromEnv() == "" || config.CoordinatorClusterIDFromEnv() == "" || config.CoordinatorInstanceIDFromEnv() == "" {
		return ErrCoordinatorIdentity
	}

	a.deps = deps{
		cfg:         cfg,
		coordinator: coordinatorclient.New(base, coordinatorHTTPTimeout),
	}
	return nil
}

// initVideoPipeline поднимает S3, пару ML-клиентов и опционально Kafka publisher.
func (a *App) initVideoPipeline(ctx context.Context) error {
	ak := os.Getenv("AWS_ACCESS_KEY_ID")
	sk := os.Getenv("AWS_SECRET_ACCESS_KEY")
	if ak == "" || sk == "" {
		return ErrMissingAWSCredentials
	}
	store, err := s3store.New(ctx, a.deps.cfg.S3.Endpoint, a.deps.cfg.S3.Region, a.deps.cfg.S3.Bucket, ak, sk)
	if err != nil {
		return fmt.Errorf("s3 client: %w", err)
	}
	if a.deps.cfg.Ingest.CreateBucketIfMissing {
		if err := store.EnsureBucket(ctx); err != nil {
			return fmt.Errorf("ensure bucket: %w", err)
		}
		zlog.Info().Str("bucket", a.deps.cfg.S3.Bucket).Msg("bucket ok")
	}
	a.deps.store = store

	brokers := strings.TrimSpace(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	if brokers == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required for ML frame pipeline")
	}
	accIn := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_ML_ACCIDENT_IN"))
	if accIn == "" {
		accIn = defaultKafkaTopicMLAccidentIn
	}
	congIn := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_ML_CONGESTION_IN"))
	if congIn == "" {
		congIn = defaultKafkaTopicMLCongestionIn
	}
	mlPub := kafkapub.NewMLPublisher(brokers, accIn, congIn)
	if mlPub == nil {
		return fmt.Errorf("kafka ML publishers: invalid topics or brokers")
	}
	a.deps.mlPub = mlPub
	zlog.Info().Str("accident_in", accIn).Str("congestion_in", congIn).Msg("router kafka ML ingest enabled")

	topic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_VIDEO"))
	if topic == "" {
		topic = defaultKafkaTopicVideo
	}
	if pub := kafkapub.NewPublisher(brokers, topic); pub != nil {
		a.deps.videoPub = pub
		zlog.Info().Str("topic", topic).Msg("router kafka video meta enabled")
	}
	return nil
}
