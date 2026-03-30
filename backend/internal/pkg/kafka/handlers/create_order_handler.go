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
	ErrCreateOrder = errors.New("ошибка создания заказа")
	ErrInvalidJSON = errors.New("неправильный json")
	ErrValidation  = errors.New("ошибка валидации")
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
		return fmt.Errorf("неправильный json: %w", err)
	}

	if order == nil {
		return errors.New("пустой заказ")
	}

	if err := h.validator.Validate(order); err != nil {
		return fmt.Errorf("ошибка валидации: %w", err)
	}

	if _, err := h.orderService.SaveOrder(ctx, order); err != nil {
		return fmt.Errorf("ошибка создания заказа: %w", err)
	}

	return nil
}
