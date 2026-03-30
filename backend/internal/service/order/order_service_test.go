package order

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	mock_repository "github.com/avraam311/order-service/backend/internal/mocks/repository"
	"github.com/avraam311/order-service/backend/internal/models"
)

func TestService_SaveOrder(t *testing.T) {
	t.Helper()
	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*Service, *mock_repository.MockorderRepository)
		wantID  uuid.UUID
		wantErr bool
	}{
		{
			name: "успешное сохранение заказа",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository) {
				mockRepo := mock_repository.NewMockorderRepository(ctrl)
				mockRepo.EXPECT().SaveOrder(gomock.Any(), gomock.Any()).Return(uuid.New(), nil)
				srv := New(nil, mockRepo)
				return srv, mockRepo
			},
			wantErr: false,
		},
		{
			name: "ошибка репозитория",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository) {
				mockRepo := mock_repository.NewMockorderRepository(ctrl)
				mockRepo.EXPECT().SaveOrder(gomock.Any(), gomock.Any()).Return(uuid.Nil, errors.New("db error"))
				srv := New(nil, mockRepo)
				return srv, mockRepo
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			srv, _ := tt.setup(ctrl)

			order := &models.Order{OrderID: uuid.New()}
			id, err := srv.SaveOrder(context.Background(), order)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, uuid.Nil, id)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, uuid.Nil, id)
			}
		})
	}
}

func TestService_GetOrderByID(t *testing.T) {
	t.Helper()
	orderID := uuid.New()
	sampleOrder := &models.Order{
		OrderID:     orderID,
		TrackNumber: "test-track",
		Items:       []models.Item{{ChrtID: 1, Name: "test item"}},
	}
	sampleItems := []models.Item{{ChrtID: 1, Name: "test item"}}

	tests := []struct {
		name    string
		setup   func(*gomock.Controller) (*Service, *mock_repository.MockorderRepository, *mock_repository.MockorderCache)
		want    *models.Order
		wantErr bool
	}{
		{
			name: "cache hit",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository, *mock_repository.MockorderCache) {
				mockCache := mock_repository.NewMockorderCache(ctrl)
				mockCache.EXPECT().Get(orderID).Return(sampleOrder, true)
				srv := New(mockCache, nil)
				return srv, nil, mockCache
			},
			want: sampleOrder,
		},
		{
			name: "cache miss, repo success",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository, *mock_repository.MockorderCache) {
				mockRepo := mock_repository.NewMockorderRepository(ctrl)
				mockCache := mock_repository.NewMockorderCache(ctrl)
				mockCache.EXPECT().Get(orderID).Return((*models.Order)(nil), false)
				mockRepo.EXPECT().GetOrderById(gomock.Any(), orderID).Return(&models.Order{OrderID: orderID}, nil)
				mockRepo.EXPECT().GetItemsByOrderID(gomock.Any(), orderID).Return(sampleItems, nil)
				mockCache.EXPECT().Set(orderID, gomock.Any())
				srv := New(mockCache, mockRepo)
				return srv, mockRepo, mockCache
			},
			want: &models.Order{OrderID: orderID, Items: sampleItems},
		},
		{
			name: "repo GetOrderById error",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository, *mock_repository.MockorderCache) {
				mockRepo := mock_repository.NewMockorderRepository(ctrl)
				mockCache := mock_repository.NewMockorderCache(ctrl)
				mockCache.EXPECT().Get(orderID).Return(nil, false)
				mockRepo.EXPECT().GetOrderById(gomock.Any(), orderID).Return(nil, errors.New("not found"))
				srv := New(mockCache, mockRepo)
				return srv, mockRepo, mockCache
			},
			wantErr: true,
		},
		{
			name: "repo GetItemsByOrderID error",
			setup: func(ctrl *gomock.Controller) (*Service, *mock_repository.MockorderRepository, *mock_repository.MockorderCache) {
				mockRepo := mock_repository.NewMockorderRepository(ctrl)
				mockCache := mock_repository.NewMockorderCache(ctrl)
				mockCache.EXPECT().Get(orderID).Return(nil, false)
				mockRepo.EXPECT().GetOrderById(gomock.Any(), orderID).Return(&models.Order{OrderID: orderID}, nil)
				mockRepo.EXPECT().GetItemsByOrderID(gomock.Any(), orderID).Return(nil, errors.New("items error"))
				srv := New(mockCache, mockRepo)
				return srv, mockRepo, mockCache
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			srv, _, _ := tt.setup(ctrl)

			order, err := srv.GetOrderByID(context.Background(), orderID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want.OrderID, order.OrderID)
			}
		})
	}
}
