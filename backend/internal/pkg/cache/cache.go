package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	"github.com/avraam311/order-service/backend/internal/models"
)

var (
	ErrCachePreload = errors.New("ошибка загрузки кэша")
)

type orderRepository interface {
	GetLastOrders(ctx context.Context, limit int) ([]models.Order, error)
}

type GoCache struct {
	c      *cache.Cache
	logger *zap.Logger
	repo   orderRepository
}

func New(defaultExpiration, cleanupInterval time.Duration, l *zap.Logger, r orderRepository) *GoCache {
	return &GoCache{
		c:      cache.New(defaultExpiration, cleanupInterval),
		logger: l,
		repo:   r,
	}
}

func (g *GoCache) Get(orderID uuid.UUID) (*models.Order, bool) {
	val, found := g.c.Get(orderID.String())
	if !found {
		return nil, false
	}

	order, ok := val.(*models.Order)

	return order, ok
}

func (g *GoCache) Set(orderID uuid.UUID, order *models.Order) {
	g.c.Set(orderID.String(), order, cache.DefaultExpiration)
}

func (g *GoCache) Preload(ctx context.Context, limit int) error {
	orders, err := g.repo.GetLastOrders(ctx, limit)
	if err != nil {
		g.logger.Error("ошибка загрузки кэша", zap.Error(err))
		return fmt.Errorf("backend/internal/pkg/cache/cache.go, ошибка загрузки кэша: %w", ErrCachePreload)
	}

	if len(orders) == 0 {
		g.logger.Info("нет заказов для загрузки в кэш")
		return nil
	}

	for _, order := range orders {
		o := order
		g.Set(order.OrderID, &o)
	}

	g.logger.Info("кэш загружен успешно", zap.Int("orders_count", len(orders)))
	return nil
}
