package retryclients

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"data-service/internal/utils/retry"
)

const (
	PingRetries = 3
)

var (
	ErrClientIsNil         = errors.New("client is nil")
	ErrRetryObjResultIsNil = errors.New("retryObjResult is nil")
	ErrRetryObjRowIsNil    = errors.New("retryObjRow is nil")
	ErrRetryObjRowsIsNil   = errors.New("retryObjRows is nil")
	ErrInvalidInitParams   = errors.New("invalid initialization parameters")
)

type ClickhouseClientAdapter struct {
	client         *sql.DB
	retryObjResult *retry.RetryObj[sql.Result]
	retryObjRow    *retry.RetryObj[sql.Row]
	retryObjRows   *retry.RetryObj[sql.Rows]
}

func NewClickhouseClientAdapter(
	client *sql.DB,
	retryObjResult *retry.RetryObj[sql.Result],
	retryObjRow *retry.RetryObj[sql.Row],
	retryObjRows *retry.RetryObj[sql.Rows],
) (*ClickhouseClientAdapter, error) {
	if err := validateClickhouseClientAdapter(client, retryObjResult, retryObjRow, retryObjRows); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidInitParams, err)
	}
	return &ClickhouseClientAdapter{
		client:         client,
		retryObjResult: retryObjResult,
		retryObjRow:    retryObjRow,
		retryObjRows:   retryObjRows,
	}, nil
}

func validateClickhouseClientAdapter(
	client *sql.DB,
	retryObjResult *retry.RetryObj[sql.Result],
	retryObjRow *retry.RetryObj[sql.Row],
	retryObjRows *retry.RetryObj[sql.Rows],
) error {
	switch {
	case client == nil:
		return ErrClientIsNil
	case retryObjResult == nil:
		return ErrRetryObjResultIsNil
	case retryObjRow == nil:
		return ErrRetryObjRowIsNil
	case retryObjRows == nil:
		return ErrRetryObjRowsIsNil
	default:
		return nil
	}
}

func (c *ClickhouseClientAdapter) Ping(ctx context.Context) error {
	_, err := c.retryObjResult.DoWithRetry(ctx, func(ctx context.Context) (*sql.Result, error) {
		return nil, c.client.PingContext(ctx)
	})

	return err
}

func (c *ClickhouseClientAdapter) Close(_ context.Context) error {
	return c.client.Close()
}

func (c *ClickhouseClientAdapter) Execute(ctx context.Context, query string, args ...any) (*sql.Result, error) {
	return c.retryObjResult.DoWithRetry(ctx, func(ctx context.Context) (*sql.Result, error) {
		result, err := c.client.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return &result, nil
	})
}

func (c *ClickhouseClientAdapter) ExecContext(ctx context.Context, query string, args ...any) (*sql.Result, error) {
	return c.retryObjResult.DoWithRetry(ctx, func(ctx context.Context) (*sql.Result, error) {
		result, err := c.client.ExecContext(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		return &result, nil
	})
}

func (c *ClickhouseClientAdapter) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	row, _ := c.retryObjRow.DoWithRetry(ctx, func(ctx context.Context) (*sql.Row, error) {
		return c.client.QueryRowContext(ctx, query, args...), nil
	})
	return row
}

func (c *ClickhouseClientAdapter) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return c.retryObjRows.DoWithRetry(ctx, func(ctx context.Context) (*sql.Rows, error) {
		return c.client.QueryContext(ctx, query, args...)
	})
}
