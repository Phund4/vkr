package clickhouse

import (
	"context"
	"database/sql"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

// Client контракт клиента для выполнения запросов в ClickHouse.
type Client interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	Execute(ctx context.Context, query string, args ...any) (*sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (*sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// NewConn создает SQL-подключение к ClickHouse по DSN.
func NewConn(dsn string) (*sql.DB, error) {
	return sql.Open("clickhouse", dsn)
}
