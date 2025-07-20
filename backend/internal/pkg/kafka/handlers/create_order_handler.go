package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrCreateOrder = errors.New("error creating order")
	ErrInvalidJSON = errors.New("invalid JSON format")
	ErrNilOrder    = errors.New("order is nil")
)

type orderService interface {
	SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error)
}

type CreateHandler struct {
	logger       *zap.Logger
	orderService orderService
}

func NewCreateHandler(l *zap.Logger, s orderService) *CreateHandler {
	return &CreateHandler{
		logger:       l,
		orderService: s,
	}
}

func (h *CreateHandler) HandleMessage(ctx context.Context, msg []byte) error {
	var order *models.Order
	if err := json.Unmarshal(msg, &order); err != nil {
		return fmt.Errorf("backend/internal/pkg/kafka/handlers/create_order_handler.go, %w: %v", ErrInvalidJSON, err)
	}

	if order == nil {
		return fmt.Errorf("backend/internal/pkg/kafka/handlers/create_order_handler.go, %w", ErrInvalidJSON)
	}

	if _, err := h.orderService.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("%w: %v", ErrCreateOrder, err)
	}

	return nil
}
