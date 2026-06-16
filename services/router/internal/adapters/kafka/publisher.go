// Package kafkapub публикует JSON в топик видео-контура (Router → Kafka → Analytics).
package kafkapub

import (
	"context"
	"strings"

	kafkago "github.com/segmentio/kafka-go"
)

// Publisher синхронная запись сообщений (nil — no-op).
type Publisher struct {
	w *kafkago.Writer
}

// NewPublisher создаёт writer; brokers — список host:port через запятую.
func NewPublisher(brokersCSV, topic string) *Publisher {
	parts := splitBrokers(brokersCSV)
	if len(parts) == 0 || topic == "" {
		return nil
	}
	return &Publisher{
		w: &kafkago.Writer{
			Addr:                   kafkago.TCP(parts...),
			Topic:                  topic,
			Balancer:               &kafkago.Hash{},
			BatchTimeout:           writerBatchTimeout,
			RequiredAcks:           kafkago.RequireAll,
			AllowAutoTopicCreation: false,
		},
	}
}

// splitBrokers парсит CSV список брокеров Kafka.
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

// Publish key — partition key (например segment_id).
func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.WriteMessages(ctx, kafkago.Message{Key: key, Value: value})
}

// Close освобождает writer.
func (p *Publisher) Close() error {
	if p == nil || p.w == nil {
		return nil
	}
	return p.w.Close()
}
