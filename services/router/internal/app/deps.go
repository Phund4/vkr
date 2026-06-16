package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"

	coordinatorclient "router/internal/adapters/coordinator"
	kafkapub "router/internal/adapters/kafka"
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

// initVideoPipeline поднимает Kafka: its.video.ingest, its.frames.ingest (pusher), ML in-топики.
func (a *App) initVideoPipeline(ctx context.Context) error {
	_ = ctx
	brokers := strings.TrimSpace(os.Getenv("KAFKA_BOOTSTRAP_SERVERS"))
	if brokers == "" {
		return fmt.Errorf("KAFKA_BOOTSTRAP_SERVERS is required for frame pipeline")
	}

	videoTopic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_VIDEO"))
	if videoTopic == "" {
		videoTopic = defaultKafkaTopicVideo
	}
	if pub := kafkapub.NewPublisher(brokers, videoTopic); pub != nil {
		a.deps.videoPub = pub
		zlog.Info().Str("topic", videoTopic).Msg("router kafka video meta enabled")
	}

	framesTopic := strings.TrimSpace(os.Getenv("KAFKA_TOPIC_FRAMES"))
	if framesTopic == "" {
		framesTopic = defaultKafkaTopicFrames
	}
	framePub := kafkapub.NewFramePublisher(brokers, framesTopic)
	if framePub == nil {
		return fmt.Errorf("kafka frames publisher: invalid topics or brokers")
	}
	a.deps.framePub = framePub
	zlog.Info().Str("topic", framesTopic).Msg("router kafka frames for pusher enabled")

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
	return nil
}
