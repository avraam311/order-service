package order

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"net/http"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrDeliveryNotFound = errors.New("delivery not found")
	ErrPaymentNotFound  = errors.New("payment not found")
	ErrItemsNotFound    = errors.New("items not found")
	ErrScanRow          = errors.New("failed to scan row")
)

type orderService interface {
	GetOrderByID(ctx context.Context, orderID uuid.UUID) (*models.Order, error)
}

type GetHandler struct {
	logger       *zap.Logger
	orderService orderService
}

func NewGetHandler(l *zap.Logger, s orderService) *GetHandler {
	return &GetHandler{
		logger:       l,
		orderService: s,
	}
}

func (h *GetHandler) GetOrderByID(w http.ResponseWriter, r *http.Request) {
	orderStr := chi.URLParam(r, "id")
	orderID, err := uuid.Parse(orderStr)
	if err != nil {
		http.Error(w, "invalid UUID format", http.StatusBadRequest)
		return
	}

	if orderID == uuid.Nil {
		http.Error(w, "order ID is required", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByID(r.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			http.Error(w, "order not found", http.StatusNotFound)
		case errors.Is(err, ErrDeliveryNotFound):
			http.Error(w, "delivery not found", http.StatusNotFound)
		case errors.Is(err, ErrPaymentNotFound):
			http.Error(w, "payment not found", http.StatusNotFound)
		case errors.Is(err, ErrItemsNotFound):
			http.Error(w, "items not found", http.StatusNotFound)
		case errors.Is(err, ErrScanRow):
			http.Error(w, "failed to scan row", http.StatusInternalServerError)
		default:
			h.logger.Error("backend/internal/api/handlers/order/get_handler.go, failed to get order", zap.Error(err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("order received", zap.Any("order", order))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		h.logger.Error("backend/internal/api/handlers/order/get_handler.go, failed to encode order response", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
