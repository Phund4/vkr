package services

import (
	"context"
	"errors"

	"data-service/internal/core/domain"
)

var ErrProductCreationNotFound = errors.New("product creation not found")

// ProductService сервисный слой для сценариев получения product creation.
type ProductService struct {
	clickhouseRepo InputRepository
}

// NewProductService создает сервис для работы с данными о создании продукта.
func NewProductService(clickhouseRepo InputRepository) *ProductService {
	return &ProductService{clickhouseRepo: clickhouseRepo}
}

// Close закрывает внешние ресурсы сервиса.
func (s *ProductService) Close(ctx context.Context) error {
	return s.clickhouseRepo.Close(ctx)
}

// GetProductCreation получает данные о создании продукта по productID.
func (s *ProductService) GetProductCreation(ctx context.Context, productID int64) (*domain.ProductCreation, error) {
	productCreation, err := s.clickhouseRepo.GetProductCreation(ctx, productID)
	if err != nil {
		return nil, err
	}

	if productCreation == nil || productCreation.ProductID == 0 {
		return nil, ErrProductCreationNotFound
	}

	return productCreation, nil
}
