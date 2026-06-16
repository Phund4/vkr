package kafka

import (
	"context"
	"time"

	kafkago "github.com/segmentio/kafka-go"
)

const producerBatchTimeout = 50 * time.Millisecond

// Publisher пишет в топик persist для pusher.
type Publisher struct {
	w *kafkago.Writer
}

// NewPublisher создаёт writer для указанных брокеров и имени топика.
func NewPublisher(brokers []string, topic string) *Publisher {
	return &Publisher{
		w: &kafkago.Writer{
			Addr:         kafkago.TCP(brokers...),
			Topic:        topic,
			BatchTimeout: producerBatchTimeout,
		},
	}
}

// Publish отправляет одно сообщение.
func (p *Publisher) Publish(ctx context.Context, key, value []byte) error {
	return p.w.WriteMessages(ctx, kafkago.Message{
		Key:   key,
		Value: value,
	})
}

// Close закрывает writer.
func (p *Publisher) Close() error {
	return p.w.Close()
}
