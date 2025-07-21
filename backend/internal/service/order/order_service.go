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

type orderCache interface {
	Get(orderID uuid.UUID) (*models.Order, bool)
	Set(orderID uuid.UUID, order *models.Order)
}

type Service struct {
	cache orderCache
	repo  orderRepository
}

func New(c orderCache, repo orderRepository) *Service {
	return &Service{
		cache: c,
		repo:  repo,
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
	if s.cache != nil {
		if order, found := s.cache.Get(orderID); found {
			return order, nil
		}
	}

	order, err := s.repo.GetOrderById(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items, err := s.repo.GetItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	order.Items = items

	s.cache.Set(orderID, order)

	return order, nil
}
