package app

import (
	"context"
	"fmt"

	"traffic-analytics/internal/adapters/clickhouse"
	"traffic-analytics/internal/config"
	"traffic-analytics/internal/core/services"
)

// Deps инициализированные адаптеры и сервис приложения.
type Deps struct {
	// Config снимок настроек из окружения.
	Config config.Config

	// Store клиент ClickHouse для OLAP и справочников.
	Store *clickhouse.Store

	// Ingest HTTP-обработчик приёма событий.
	Ingest *services.IngestService

	// CHAddr нормализованный адрес ClickHouse (для логов).
	CHAddr string
}

// InitializeDependencies загружает конфиг (включая .env), подключает ClickHouse и создаёт IngestService.
func InitializeDependencies(ctx context.Context) (*Deps, error) {
	cfg := config.Load()

	chAddr := clickhouse.NormalizeAddr(cfg.ClickHouseAddr)
	store, err := clickhouse.New(ctx, chAddr, cfg.ClickHouseDatabase, cfg.ClickHouseUser, cfg.ClickHousePassword, cfg.IncidentsTable, cfg.CongestionTable)
	if err != nil {
		return nil, fmt.Errorf("clickhouse: %w", err)
	}

	ingest := services.NewIngestService(store, cfg, ctx)
	return &Deps{
		Config: cfg,
		Store:  store,
		Ingest: ingest,
		CHAddr: chAddr,
	}, nil
}

// Close освобождает ресурсы зависимостей.
func (d *Deps) Close() error {
	if d.Store == nil {
		return nil
	}
	return d.Store.Close()
}
