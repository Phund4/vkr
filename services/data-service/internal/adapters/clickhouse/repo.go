package clickhouse

import (
	"context"
)

// Repository предоставляет интерфейс для работы с ClickHouse
type Repository struct {
	client Client
}

// NewRepository создает новый репозиторий для работы с ClickHouse
func NewRepository(_ context.Context, client Client) (*Repository, error) {
	return &Repository{client: client}, nil
}

// Ping проверяет, что соединение с ClickHouse работает
func (r *Repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// Close закрывает соединение с ClickHouse
func (r *Repository) Close(ctx context.Context) error {
	return r.client.Close(ctx)
}
