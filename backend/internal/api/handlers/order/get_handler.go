package order

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrOrderNotFound     = errors.New("заказ не найден")
	ErrScanRow           = errors.New("ошибка сканирования строки")
	ErrItemScanFailed    = errors.New("ошибка сканирования items заказа")
	ErrGetItemsByOrderId = errors.New("ошибка получения items по orderID")
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
		http.Error(w, "неправильный формат uuid", http.StatusBadRequest)
		return
	}

	if orderID == uuid.Nil {
		http.Error(w, "нужно orderID", http.StatusBadRequest)
		return
	}

	order, err := h.orderService.GetOrderByID(r.Context(), orderID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			http.Error(w, "заказ не найден", http.StatusNotFound)
		case errors.Is(err, ErrItemScanFailed):
			http.Error(w, "ошибка сканирования items заказа", http.StatusNotFound)
		case errors.Is(err, ErrGetItemsByOrderId):
			http.Error(w, "ошибка получения items по orderID", http.StatusNotFound)
		case errors.Is(err, ErrScanRow):
			http.Error(w, "ошибка сканирования строки", http.StatusInternalServerError)
		default:
			h.logger.Error("backend/internal/api/handlers/order/get_handler.go, ошибка получения заказа", zap.Error(err))
			http.Error(w, "ошибка сервера", http.StatusInternalServerError)
		}

		return
	}

	h.logger.Info("заказ получен", zap.Any("order", order))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	err = json.NewEncoder(w).Encode(order)
	if err != nil {
		h.logger.Error("backend/internal/api/handlers/order/get_handler.go, ошибка кодироавния ответа для закака", zap.Error(err))
		http.Error(w, "ошибка кодирования ответа", http.StatusInternalServerError)
		return
	}
}
