package service

import (
	"context"
	"notifications/internal/models"

	"go.uber.org/zap"
)

type NotificationService interface {
	HandleProductEvent(ctx context.Context, event *models.ProductEvent) error
}

type notificationService struct {
	logger *zap.Logger
}

func NewNotificationService(logger *zap.Logger) NotificationService {
	return &notificationService{
		logger: logger.Named("NotificationService"),
	}
}

func (s *notificationService) HandleProductEvent(ctx context.Context, pEvent *models.ProductEvent) error {

	switch pEvent.EventType {
	case models.ProductCreated:
		s.logger.Info("PRODUCT CREATED",
			zap.String("name", pEvent.Product.Name),
			zap.String("id", pEvent.Product.ID),
			zap.Int("price", pEvent.Product.Price),
			zap.Time("created_at", pEvent.Product.CreatedAt),
		)

	case models.ProductDeleted:
		s.logger.Info("PRODUCT DELETED",
			zap.String("name", pEvent.Product.Name),
			zap.String("id", pEvent.Product.ID),
			zap.Time("created_at", pEvent.Product.CreatedAt),
		)

	default:
		s.logger.Warn("UNKNOWN EVENT TYPE",
			zap.String("event_type", string(pEvent.EventType)),
		)
	}

	return nil
}
