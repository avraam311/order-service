package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/avraam311/order-service/backend/internal/models"
)

type orderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error)
	GetOrderById(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
	GetItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]models.Item, error)
}

type Service struct {
	repo orderRepository
}

func New(repo orderRepository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error) {
	orderID, err := s.repo.SaveOrder(ctx, order)
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, nil
}

func (s *Service) GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error) {
	order, err := s.repo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	order.Items = items

	return order, nil
}
