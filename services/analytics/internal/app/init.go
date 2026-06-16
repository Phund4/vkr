package app

import (
	"context"
	"fmt"
	"strings"

	ingestkafka "traffic-analytics/internal/adapters/kafka"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/services"
)

// Deps инициализированные адаптеры и сервис приложения.
type Deps struct {
	Config config.Config

	Ingest *services.IngestService

	// PersistKafka writer во второй топик (pusher).
	PersistKafka *ingestkafka.Publisher
}

// InitializeDependencies загружает конфиг, Kafka producer и IngestService.
func InitializeDependencies(ctx context.Context) (*Deps, error) {
	cfg := config.Load()

	brokers := splitBrokers(cfg.KafkaBootstrap)
	if len(brokers) == 0 || cfg.KafkaTopicPersist == "" {
		return nil, fmt.Errorf("kafka: KAFKA_BOOTSTRAP_SERVERS and KAFKA_TOPIC_PERSIST are required")
	}

	pub := ingestkafka.NewPublisher(brokers, cfg.KafkaTopicPersist)

	ingest := services.NewIngestService(pub, cfg, ctx)
	return &Deps{
		Config:       cfg,
		Ingest:       ingest,
		PersistKafka: pub,
	}, nil
}

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

// Close закрывает Kafka writer.
func (d *Deps) Close() error {
	if d.PersistKafka == nil {
		return nil
	}
	return d.PersistKafka.Close()
}
