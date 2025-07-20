package order

import (
	"context"

	"github.com/google/uuid"

	"github.com/avraam311/order-service/backend/internal/models"
)

type orderRepository interface {
	SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error)
}

type Service struct {
	repo   orderRepository
}

func New(repo orderRepository) *Service {
	return &Service{
		repo:   repo,
	}
}

func (s *Service) SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error) {
	orderID, err := s.repo.SaveOrder(ctx, order)
	if err != nil {
		return uuid.Nil, err
	}

	return orderID, nil
}
