package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

var (
	ErrCreateOrder = errors.New("ошибка создания закака")
	ErrInvalidJSON = errors.New("неправильный json")
	ErrNilOrder    = errors.New("пустой заказ")
	ErrValidation  = errors.New("ошибка валидации")
)

type messageHandler interface {
	HandleMessage(ctx context.Context, msg []byte) error
}

type Consumer struct {
	reader  *kafka.Reader
	logger  *zap.Logger
	handler messageHandler
}

func NewReader(groupID string, topic string, brokers []string) *kafka.Reader {
	return kafka.NewReader(kafka.ReaderConfig{
		Brokers:        brokers,
		GroupID:        groupID,
		Topic:          topic,
		StartOffset:    kafka.FirstOffset,
		CommitInterval: time.Second,
		MaxBytes:       10e6,
	})
}

func NewConsumer(r *kafka.Reader, l *zap.Logger, h messageHandler) *Consumer {
	return &Consumer{
		reader:  r,
		logger:  l,
		handler: h,
	}
}

func (c *Consumer) ConsumeMessage(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer c.Close()

	go func() {
		<-ctx.Done()
		c.logger.Info("получен сигнал shutdown, закрытие консьюмера")
		c.Close()
	}()

	c.logger.Info("читаем сообщения",
		zap.String("topic", c.reader.Config().Topic),
		zap.String("groupID", c.reader.Config().GroupID),
	)

	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				c.logger.Warn("backend/internal/pkg/kafka/consumer.go, контекст отменен или прошел дедлайн, закрытие консьюмера", zap.Error(err))
				break
			}

			c.logger.Error("backend/internal/pkg/kafka/consumer.go, ошибка чтения сообщения", zap.Error(err))
			continue
		}

		if err = c.handler.HandleMessage(ctx, m.Value); err != nil {
			c.handleMessageError(m, err)
			continue
		}

		c.logger.Info("сообщения успешно прочтено",
			zap.Int64("offset", m.Offset),
			zap.String("message", string(m.Value)),
		)
	}

	c.logger.Info("чтение сообщения завершено")
}

func (c *Consumer) handleMessageError(m kafka.Message, err error) {
	msgStr := string(m.Value)

	switch {
	case errors.Is(err, ErrInvalidJSON):
		c.logger.Warn("неправильный json",
			zap.Int64("offset", m.Offset),
			zap.String("message", msgStr),
			zap.Error(err),
		)
	case errors.Is(err, ErrNilOrder):
		c.logger.Warn("получен пустой заказ",
			zap.Int64("offset", m.Offset),
			zap.String("message", msgStr),
			zap.Error(err),
		)
	case errors.Is(err, ErrCreateOrder):
		c.logger.Warn("ошибка при создании заказа",
			zap.Int64("offset", m.Offset),
			zap.String("message", msgStr),
			zap.Error(err),
		)
	case errors.Is(err, ErrValidation):
		c.logger.Warn("ошибка валидации",
			zap.Int64("offset", m.Offset),
			zap.String("message", msgStr),
			zap.Error(err),
		)
	default:
		c.logger.Error("неожиданная ошибка при чтении сообщения",
			zap.Int64("offset", m.Offset),
			zap.String("message", msgStr),
			zap.Error(err),
		)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}
