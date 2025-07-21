package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrCreateOrder = errors.New("error creating order")
	ErrInvalidJSON = errors.New("invalid JSON format")
	ErrNilOrder    = errors.New("order is nil")
	ErrValidation  = errors.New("validation error")
)

type orderService interface {
	SaveOrder(ctx context.Context, order *models.Order) (uuid.UUID, error)
}

type validator interface {
	Validate(i interface{}) error
}

type CreateHandler struct {
	validator    validator
	orderService orderService
}

func NewCreateHandler(v validator, s orderService) *CreateHandler {
	return &CreateHandler{
		validator:    v,
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

	if err := h.validator.Validate(order); err != nil {
		return fmt.Errorf("%w: %v", ErrValidation, err)
	}

	if _, err := h.orderService.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("%w: %v", ErrCreateOrder, err)
	}

	return nil
}
