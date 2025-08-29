package messaging

import (
	"context"
	"encoding/json"
	"notifications/internal/models"
	"notifications/internal/service"

	"go.uber.org/zap"
)

type MessageBroker interface {
	Consume(ctx context.Context, workerCount int, handler func(message []byte) error) error
	Close() error
}

type Consumer struct {
	broker  MessageBroker
	service service.NotificationService
	logger  *zap.Logger
}

func NewConsumer(broker MessageBroker, service service.NotificationService, logger *zap.Logger) *Consumer {
	return &Consumer{
		broker:  broker,
		service: service,
		logger:  logger.Named("Consumer"),
	}
}

func (c *Consumer) Start(ctx context.Context, workerCount int) error {
	return c.broker.Consume(ctx, workerCount, func(message []byte) error {
		var event models.ProductEvent
		if err := json.Unmarshal(message, &event); err != nil {
			c.logger.Error("Failed to unmarshal event",
				zap.Error(err),
			)
			return err
		}

		return c.service.HandleProductEvent(ctx, &event)
	})
}

func (c *Consumer) Stop() error {
	return c.broker.Close()
}
