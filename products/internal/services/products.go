package services

import (
	"context"
	"encoding/json"
	"products/internal/metrics"
	"products/internal/models"
	"time"

	"go.uber.org/zap"
)

type MessageBroker interface {
	Send(ctx context.Context, topic string, message, key []byte) error
	Close() error
}

type ProductsRepository interface {
	Count(ctx context.Context) (int, error)
	Create(ctx context.Context, createDTO *models.CreateProductDTO) (*models.Product, error)
	Delete(ctx context.Context, id string) (*models.Product, error)
	List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, error)
}

type ProductsService struct {
	repo   ProductsRepository
	broker MessageBroker
	logger *zap.Logger
}

func NewProductsService(repo ProductsRepository, broker MessageBroker, logger *zap.Logger) *ProductsService {
	return &ProductsService{
		repo:   repo,
		broker: broker,
		logger: logger.Named("ProductsService"),
	}
}

func (p *ProductsService) Create(ctx context.Context, productDTO *models.CreateProductDTO) (*models.Product, error) {
	product, err := p.repo.Create(ctx, productDTO)

	if err != nil {
		return nil, err
	}

	metrics.ProductsCreated.Inc()
	p.trySendProductEvent(ctx, product, models.ProductCreated)

	return product, nil
}

func (p *ProductsService) Delete(ctx context.Context, id string) (*models.Product, error) {
	product, err := p.repo.Delete(ctx, id)

	if err != nil {
		return nil, err
	}

	metrics.ProductsDeleted.Inc()
	p.trySendProductEvent(ctx, product, models.ProductDeleted)

	return product, nil
}

func (p *ProductsService) List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, int, error) {
	total, err := p.repo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	products, err := p.repo.List(ctx, listDTO)
	if err != nil {
		return nil, 0, err
	}

	return products, total, nil
}

func (p *ProductsService) trySendProductEvent(ctx context.Context, product *models.Product, eventType models.ProductEventType) {
	event := &models.ProductEvent{
		EventType: eventType,
		Product:   product,
		Timestamp: time.Now(),
	}

	msg, err := json.Marshal(event)
	if err != nil {
		p.logger.Error(
			"failed to marshal product event",
			zap.Error(err),
			zap.String("product_id", product.ID),
			zap.String("event_type", string(eventType)))
		return
	}

	err = p.broker.Send(ctx, "", msg, []byte(product.ID))
	if err != nil {
		p.logger.Error(
			"failed to send product event",
			zap.Error(err),
			zap.String("product_id", product.ID),
			zap.String("event_type", string(eventType)),
		)
	}
}
