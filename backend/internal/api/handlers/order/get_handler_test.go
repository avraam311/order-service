package order

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	mock_service "github.com/avraam311/order-service/backend/internal/mocks/service"
	"github.com/avraam311/order-service/backend/internal/models"

	"go.uber.org/zap/zaptest"
)

type testOrderService interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
}

func TestGetHandler_GetOrderByID(t *testing.T) {
	t.Helper()

	logger := zaptest.NewLogger(t)

	orderID := uuid.New()
	sampleOrder := &models.Order{
		OrderID:     orderID,
		TrackNumber: "test-123",
	}

	tests := []struct {
		name                 string
		url                  string
		setupMock            func(*gomock.Controller) testOrderService
		expectedStatus       int
		expectedBodyContains string
	}{
		{
			name:                 "неправильный UUID",
			url:                  "/order/invalid",
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "неправильный формат uuid",
		},
		{
			name:                 "nil UUID",
			url:                  "/order/00000000-0000-0000-0000-000000000000",
			expectedStatus:       http.StatusBadRequest,
			expectedBodyContains: "нужно orderID",
		},
		{
			name: "успешный запрос",
			url:  "/order/" + orderID.String(),
			setupMock: func(ctrl *gomock.Controller) testOrderService {
				m := mock_service.NewMockorderService(ctrl)
				m.EXPECT().GetOrderByID(gomock.Any(), orderID).Return(sampleOrder, nil)
				return m
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name: "заказ не найден",
			url:  "/order/" + orderID.String(),
			setupMock: func(ctrl *gomock.Controller) testOrderService {
				m := mock_service.NewMockorderService(ctrl)
				m.EXPECT().GetOrderByID(gomock.Any(), orderID).Return(nil, ErrOrderNotFound)
				return m
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "ошибка сервера",
			url:  "/order/" + orderID.String(),
			setupMock: func(ctrl *gomock.Controller) testOrderService {
				m := mock_service.NewMockorderService(ctrl)
				m.EXPECT().GetOrderByID(gomock.Any(), orderID).Return(nil, errors.New("db error"))
				return m
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var svc testOrderService
			if tt.setupMock != nil {
				svc = tt.setupMock(ctrl)
			}

			h := NewGetHandler(logger, svc)

			r := httptest.NewRequest("GET", tt.url, nil)
			w := httptest.NewRecorder()

			router := chi.NewRouter()
			router.Get("/order/{id}", h.GetOrderByID)
			router.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBodyContains != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBodyContains)
			}
		})
	}
}
