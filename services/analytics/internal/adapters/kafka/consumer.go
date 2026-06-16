// Package kafka — консьюмеры Kafka: meta кадра, ML accident/congestion out.
package kafka

import (
	"context"
	"strings"
	"time"

	zlog "github.com/rs/zerolog/log"

	kafkago "github.com/segmentio/kafka-go"

	"traffic-analytics/internal/adapters/metrics"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/services"
)

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

// RunConsumers читает video meta и выходы ML; accident и congestion обрабатываются независимо.
func RunConsumers(ctx context.Context, ingest *services.IngestService, cfg config.Config) {
	brokers := splitBrokers(cfg.KafkaBootstrap)
	if len(brokers) == 0 {
		return
	}
	topics := []string{
		cfg.KafkaTopicVideo,
		cfg.KafkaTopicMLAccidentOut,
		cfg.KafkaTopicMLCongestionOut,
	}
	r := kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		GroupID:     cfg.KafkaConsumerGroup,
		GroupTopics: topics,
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
		Strs("topics", topics).
		Msg("analytics kafka consumers")

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
		var procErr error
		switch m.Topic {
		case cfg.KafkaTopicMLAccidentOut:
			procErr = ingest.ProcessAccidentResult(ctx, m.Value)
		case cfg.KafkaTopicMLCongestionOut:
			procErr = ingest.ProcessCongestionResult(ctx, m.Value)
		default:
			procErr = ingest.ProcessVideoMeta(ctx, m.Value)
		}
		if procErr != nil {
			metrics.KafkaConsumeErrors.WithLabelValues(metrics.KafkaConsumeStageProcess).Inc()
			zlog.Warn().Str("topic", m.Topic).Err(procErr).Msg("kafka process")
		} else {
			metrics.KafkaIngestProcessed.WithLabelValues(m.Topic).Inc()
		}
	}
}
