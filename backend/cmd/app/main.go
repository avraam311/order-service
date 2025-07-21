package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	orderHandler "github.com/avraam311/order-service/backend/internal/api/handlers/order"
	"github.com/avraam311/order-service/backend/internal/api/server"
	"github.com/avraam311/order-service/backend/internal/config"
	"github.com/avraam311/order-service/backend/internal/pkg/logger"
	"github.com/avraam311/order-service/backend/internal/pkg/cache"
	orderRepo "github.com/avraam311/order-service/backend/internal/repository/order"
	orderService "github.com/avraam311/order-service/backend/internal/service/order"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := config.MustLoad()
	log := logger.SetupLogger(cfg.Logger.Env, cfg.Logger.LogFilePath)

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL())
	if err != nil {
		log.Fatal("error creating connection pool", zap.Error(err))
	}

	repo := orderRepo.New(dbpool)
	cache := cache.New(cfg.Cache.DefaultExpiration, cfg.Cache.CleanupInterval, log, repo)

	err = cache.Preload(ctx, cfg.Cache.PreloadLimit)
	if err != nil {
		log.Fatal("error preloading cache", zap.Error(err))
	}

	orderService := orderService.New(cache, repo)
	orderGetHandler := orderHandler.NewGetHandler(log, orderService)

	r := server.NewRouter(orderGetHandler)
	server := server.NewServer(cfg.Server.HTTPPort, r)

	go func() {
		log.Info("starting HTTP server", zap.String("port", cfg.Server.HTTPPort))
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	log.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Info("shutting down HTTP server...")
	if err = server.Shutdown(shutdownCtx); err != nil {
		log.Error("could not shutdown HTTP server", zap.Error(err))
		os.Exit(1)
	}

	if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
		log.Fatal("timeout exceeded, forcing shutdown")
	}

	log.Info("closing database pool...")
	dbpool.Close()
}
