package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/core/domain"
	"data-service/internal/core/services"
)

const ProductIDQueryParam = "product_id"

// ProductGetter контракт сервиса для получения данных о создании продукта.
//
//go:generate go run github.com/vektra/mockery/v2@v2.53.6
type ProductGetter interface {
	GetProductCreation(ctx context.Context, productID int64) (*domain.ProductCreation, error)
}

// ProductHandler HTTP-обработчик endpointов создания продукта.
type ProductHandler struct {
	service ProductGetter
}

// NewProductHandler создает ProductHandler с внедренным сервисом.
func NewProductHandler(service ProductGetter) *ProductHandler {
	return &ProductHandler{service: service}
}

// @Summary Get product creation by product ID
// @Param product_id query int true "Product ID"
// @Success 200 {object} dto.ProductCreation
// @Failure 400 {object} dto.GetProductResponse
// @Failure 404 {object} dto.GetProductResponse
// @Failure 500 {object} dto.GetProductResponse
// @Router /api/v1/get_product [get]
func (h *ProductHandler) GetProduct(c echo.Context) error {
	productIDStr := c.QueryParam(ProductIDQueryParam)
	if productIDStr == "" {
		return c.JSON(http.StatusBadRequest, dto.GetProductResponse{
			Status:  dto.StatusError,
			Message: dto.ErrProductIDQueryField,
		})
	}

	productID, err := strconv.ParseInt(productIDStr, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, dto.GetProductResponse{
			Status:  dto.StatusError,
			Message: dto.ErrProductIDQueryFieldType,
		})
	}

	productCreation, err := h.service.GetProductCreation(c.Request().Context(), productID)
	if err != nil {
		if errors.Is(err, services.ErrProductCreationNotFound) {
			return c.JSON(http.StatusNotFound, dto.GetProductResponse{
				Status:  dto.StatusError,
				Message: dto.ErrProductCreationNotFound,
			})
		}

		return c.JSON(http.StatusInternalServerError, dto.GetProductResponse{
			Status:  dto.StatusError,
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, dto.NewProductCreation(productCreation))
}
