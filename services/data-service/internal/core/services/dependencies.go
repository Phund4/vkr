package services

import (
	"context"

	"data-service/internal/core/domain"
)

// InputRepository контракт репозитория, используемого сервисным слоем.
type InputRepository interface {
	Ping(ctx context.Context) error
	Close(ctx context.Context) error
	GetProductCreation(ctx context.Context, productID int64) (*domain.ProductCreation, error)
}
