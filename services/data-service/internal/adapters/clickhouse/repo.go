package clickhouse

import (
	"context"
	"strings"
)

// Repository чтение ИТС-таблиц из ClickHouse.
type Repository struct {
	client   Client
	database string
}

// NewRepository создает новый репозиторий для работы с ClickHouse.
// database — имя БД из DSN (квалификация запросов как database.table).
func NewRepository(_ context.Context, client Client, database string) (*Repository, error) {
	return &Repository{client: client, database: strings.TrimSpace(database)}, nil
}

func (r *Repository) qualifiedTable(table string) string {
	if r.database == "" {
		return table
	}
	return r.database + "." + table
}

// Ping проверяет, что соединение с ClickHouse работает
func (r *Repository) Ping(ctx context.Context) error {
	return r.client.Ping(ctx)
}

// Close закрывает соединение с ClickHouse
func (r *Repository) Close(ctx context.Context) error {
	return r.client.Close(ctx)
}
