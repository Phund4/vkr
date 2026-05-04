package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"data-service/internal/adapters/http/dto"
	"data-service/internal/adapters/http/handlers/mocks"
	"data-service/internal/core/domain"
	"data-service/internal/core/services"
)

func TestGetProduct(t *testing.T) {
	t.Parallel()

	e := echo.New()

	tests := []struct {
		name            string
		query           string
		setupMock       func(productGetterMock *mocks.MockProductGetter)
		expectedCode    int
		expectedStatus  string
		expectedMessage string
		expectedProduct *dto.ProductCreation
		assertErr       assert.ErrorAssertionFunc
	}{
		{
			name:            "missing product_id",
			query:           "",
			expectedCode:    http.StatusBadRequest,
			expectedStatus:  dto.StatusError,
			expectedMessage: dto.ErrProductIDQueryField,
			assertErr:       assert.NoError,
		},
		{
			name:            "invalid product_id",
			query:           "?product_id=abc",
			expectedCode:    http.StatusBadRequest,
			expectedStatus:  dto.StatusError,
			expectedMessage: dto.ErrProductIDQueryFieldType,
			assertErr:       assert.NoError,
		},
		{
			name:  "success",
			query: "?product_id=12282384",
			setupMock: func(productGetterMock *mocks.MockProductGetter) {
				productID := int64(12282384)
				traceID := uuid.MustParse("48def697-16e8-4ab4-9a46-83d695585c59")
				id := uuid.MustParse("7beeae6a-4da2-4f6a-bdb6-d80342889193")
				now := time.Now().UTC().Truncate(time.Second)

				productGetterMock.EXPECT().
					GetProductCreation(mock.Anything, productID).
					RunAndReturn(func(_ context.Context, _ int64) (*domain.ProductCreation, error) {
						return &domain.ProductCreation{
							ID:        id,
							ProductID: productID,
							TraceID:   traceID,
							Timestamp: now,
						}, nil
					})
			},
			expectedCode: http.StatusOK,
			expectedProduct: &dto.ProductCreation{
				ID:        "7beeae6a-4da2-4f6a-bdb6-d80342889193",
				ProductID: 12282384,
				TraceID:   "48def69716e84ab49a4683d695585c59",
			},
			assertErr: assert.NoError,
		},
		{
			name:  "product id not found",
			query: "?product_id=12282384",
			setupMock: func(productGetterMock *mocks.MockProductGetter) {
				productGetterMock.EXPECT().
					GetProductCreation(mock.Anything, int64(12282384)).
					Return(nil, services.ErrProductCreationNotFound)
			},
			expectedCode:    http.StatusNotFound,
			expectedStatus:  dto.StatusError,
			expectedMessage: dto.ErrProductCreationNotFound,
			assertErr:       assert.NoError,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			productGetterMock := mocks.NewMockProductGetter(t)
			if tc.setupMock != nil {
				tc.setupMock(productGetterMock)
			}
			handler := NewProductHandler(productGetterMock)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/get_product"+tc.query, nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handler.GetProduct(c)
			tc.assertErr(t, err)
			assert.Equal(t, tc.expectedCode, rec.Code)

			if tc.expectedProduct != nil {
				var resp dto.ProductCreation
				assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
				assert.Equal(t, tc.expectedProduct.ProductID, resp.ProductID)
				assert.Equal(t, tc.expectedProduct.ID, resp.ID)
				assert.Equal(t, tc.expectedProduct.TraceID, resp.TraceID)
				return
			}

			var resp dto.GetProductResponse
			assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
			assert.Equal(t, tc.expectedStatus, resp.Status)
			assert.Equal(t, tc.expectedMessage, resp.Message)
		})
	}
}
