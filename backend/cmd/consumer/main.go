package main

import (
	"context"
	"os/signal"
	"sync"
	"syscall"

	"go.uber.org/zap"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/avraam311/order-service/backend/internal/config"
	"github.com/avraam311/order-service/backend/internal/pkg/kafka"
	"github.com/avraam311/order-service/backend/internal/pkg/kafka/handlers"
	"github.com/avraam311/order-service/backend/internal/pkg/logger"
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
		log.Fatal("backend/cmd/consumer/main.go, error creating connection pool", zap.Error(err))
	}

	repo := orderRepo.New(dbpool)
	orderService := orderService.New(repo)

	orderCreatedHandler := handlers.NewCreateHandler(log, orderService)

	reader := kafka.NewReader(cfg.Kafka.GroupID, cfg.Kafka.Topic, cfg.Kafka.Brokers)
	consumer := kafka.NewConsumer(reader, log, orderCreatedHandler)
	wg.Add(1)
	go consumer.ConsumeMessage(ctx, &wg)

	log.Info("Kafka consumer started...")

	<-ctx.Done()
	log.Info("shutdown signal received")

	wg.Wait()

	log.Info("closing kafka consumer...")
	err = consumer.Close()
	if err != nil {
		log.Error("backend/cmd/consumer/main.go, error closing kafka consumer: %v", zap.Error(err))
	}

	log.Info("closing database pool...")
	dbpool.Close()
}
