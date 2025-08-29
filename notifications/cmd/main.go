package main

import (
	"context"
	"notifications/internal/config"
	loggerPkg "notifications/internal/logger"
	"notifications/internal/messaging"
	"notifications/internal/service"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger := loggerPkg.NewLogger("development", zap.InfoLevel)
	defer logger.Sync()

	cfg := config.Load()

	msgBroker, err := messaging.NewKafkaConsumer(messaging.Config{
		Endpoint: cfg.MessageBroker.Endpoint,
		Topic:    cfg.MessageBroker.Topic,
		GroupID:  cfg.MessageBroker.GroupID,
	}, logger)
	if err != nil {
		logger.Fatal("Failed to initialize message broker", zap.Error(err))
	}

	notificationService := service.NewNotificationService(logger)
	consumer := messaging.NewConsumer(msgBroker, notificationService, logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Start(ctx, runtime.NumCPU()); err != nil && err != context.Canceled {
			logger.Fatal("Failed to start consumer", zap.Error(err))
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	logger.Info("Shutdown signal received")
	cancel()
	consumer.Stop()
	logger.Info("Shutdown complete")
}
