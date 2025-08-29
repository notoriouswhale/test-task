package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"products/internal/config"
	"products/internal/handlers"
	loggerPkg "products/internal/logger"
	"products/internal/messaging"
	"products/internal/repository/pg"
	"products/internal/services"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	_ "github.com/lib/pq"
)

func main() {
	logger := loggerPkg.NewLogger("development", zap.InfoLevel)

	defer logger.Sync()
	cfg := config.Load()

	broker, err := messaging.NewKafkaBroker(messaging.Config{
		Endpoint:     cfg.MessageBroker.Endpoint,
		BaseClientID: cfg.MessageBroker.ClientID,
		Topic:        cfg.MessageBroker.Topic,
	}, logger)

	if err != nil {
		logger.Fatal("Failed to initialize broker", zap.Error(err))
	}
	defer broker.Close()

	db, err := sqlx.Connect("postgres", fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=%s sslmode=%s", cfg.DB.User, cfg.DB.DBName, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.SSLMode))
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	ProductsRepository := pg.NewProductsRepository(db)
	productsService := services.NewProductsService(ProductsRepository, broker, logger)
	productsHandler := handlers.NewProductsHandler(productsService, logger)

	router := handlers.SetupRoutes(productsHandler, logger)
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTP.Port),
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to run server", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server Shutdown:", zap.Error(err))
	}

	logger.Info("Server exited gracefully")
}
