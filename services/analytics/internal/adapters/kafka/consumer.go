// Package kafka — консьюмер событий ingest из Kafka (топик видео/ML).
package kafka

import (
	"context"
	"strings"

	zlog "github.com/rs/zerolog/log"
	"time"

	kafkago "github.com/segmentio/kafka-go"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/services"
)

// splitBrokers парсит CSV список адресов брокеров Kafka.
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

// RunIngestConsumer читает KAFKA_TOPIC_VIDEO до отмены ctx.
func RunIngestConsumer(ctx context.Context, ingest *services.IngestService, cfg config.Config) {
	brokers := splitBrokers(cfg.KafkaBootstrap)
	if len(brokers) == 0 {
		return
	}
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     cfg.KafkaConsumerGroup,
		GroupTopics: []string{cfg.KafkaTopicVideo},
		MinBytes:    1,
		MaxBytes:    10e6,
		MaxWait:     2 * time.Second,
	})
	defer func() {
		if err := r.Close(); err != nil {
			zlog.Warn().Err(err).Msg("kafka reader close")
		}
	}()
	zlog.Info().
		Strs("brokers", brokers).
		Str("group", cfg.KafkaConsumerGroup).
		Strs("topics", []string{cfg.KafkaTopicVideo}).
		Msg("analytics kafka consumer")
	for {
		m, err := r.ReadMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			metrics.KafkaConsumeErrors.WithLabelValues(metrics.KafkaConsumeStageRead).Inc()
			zlog.Warn().Err(err).Msg("kafka read")
			time.Sleep(time.Second)
			continue
		}
		if err := ingest.ProcessIngest(ctx, m.Value); err != nil {
			metrics.KafkaConsumeErrors.WithLabelValues(metrics.KafkaConsumeStageProcess).Inc()
			zlog.Warn().Str("topic", m.Topic).Err(err).Msg("kafka ingest process")
		} else {
			metrics.KafkaIngestProcessed.WithLabelValues(m.Topic).Inc()
		}
	}
}
