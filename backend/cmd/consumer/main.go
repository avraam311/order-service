package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/avraam311/order-service/backend/internal/config"
	"github.com/avraam311/order-service/backend/internal/pkg/kafka"
	"github.com/avraam311/order-service/backend/internal/pkg/kafka/handlers"
	"github.com/avraam311/order-service/backend/internal/pkg/logger"
	"github.com/avraam311/order-service/backend/internal/pkg/validator"
	orderRepo "github.com/avraam311/order-service/backend/internal/repository/order"
	orderService "github.com/avraam311/order-service/backend/internal/service/order"
)

func main() {
	var wg sync.WaitGroup

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Logger.Env, cfg.Logger.LogFilePath)
	defer log.Sync()

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("backend/cmd/consumer/main.go, ошибка при создании пула соединений", zap.Error(err))
	}

	repo := orderRepo.New(dbpool)
	orderService := orderService.New(nil, repo)
	val := validator.New()

	orderCreatedHandler := handlers.NewCreateHandler(val, orderService)

	reader := kafka.NewReader(cfg.Kafka.GroupID, cfg.Kafka.Topic, cfg.Kafka.Brokers)
	consumer := kafka.NewConsumer(reader, log, orderCreatedHandler)
	wg.Add(1)
	go consumer.ConsumeMessage(ctx, &wg)

	log.Info("кафка консьюмер запущен")

	<-ctx.Done()
	log.Info("получен сигнал shutdown")

	wg.Wait()

	log.Info("закрытие консьюмера кафки")
	err = consumer.Close()
	if err != nil {
		log.Error("backend/cmd/consumer/main.go, ошибка при закрытии консьюмера кафки: %v", zap.Error(err))
	}

	log.Info("закрытие пула соединений бд")
	dbpool.Close()
}
