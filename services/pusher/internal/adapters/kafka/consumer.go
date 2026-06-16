package kafka

import (
	"context"
	"strings"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/rs/zerolog"

	"pusher/internal/adapters/metrics"
	"pusher/internal/config"
	"pusher/internal/core/services"
)

// splitBrokers парсит CSV список брокеров из конфигурации.
func splitBrokers(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// RunConsumers читает its.frames.ingest и its.persist.events до отмены ctx.
func RunConsumers(ctx context.Context, push *services.PushService, cfg *config.Config, log *zerolog.Logger) {
	brokers := splitBrokers(cfg.Kafka.Bootstrap)
	if len(brokers) == 0 {
		log.Warn().Msg("KAFKA_BOOTSTRAP_SERVERS empty, consumer not started")
		return
	}
	topics := []string{cfg.Kafka.TopicFrames, cfg.Kafka.TopicPersist}
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     cfg.Kafka.ConsumerGroup,
		GroupTopics: topics,
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     2 * time.Second,
	})
	defer func() {
		if err := r.Close(); err != nil {
			log.Warn().Err(err).Msg("kafka reader close")
		}
	}()

	log.Info().
		Strs("brokers", brokers).
		Str("group", cfg.Kafka.ConsumerGroup).
		Strs("topics", topics).
		Msg("pusher kafka consumer started")

	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			metrics.KafkaConsumeErrors.WithLabelValues(metrics.KafkaConsumeStageRead).Inc()
			log.Warn().Err(err).Msg("kafka read")
			time.Sleep(time.Second)
			continue
		}
		var procErr error
		switch m.Topic {
		case cfg.Kafka.TopicFrames:
			procErr = push.ProcessFrameMessage(ctx, m.Value)
		default:
			procErr = push.ProcessMessage(ctx, m.Value)
		}
		if procErr != nil {
			metrics.KafkaConsumeErrors.WithLabelValues(metrics.KafkaConsumeStageProcess).Inc()
			log.Warn().Err(err).Str("topic", m.Topic).Msg("kafka process")
			continue
		}
		metrics.KafkaMessagesProcessed.WithLabelValues(m.Topic).Inc()
	}
}
