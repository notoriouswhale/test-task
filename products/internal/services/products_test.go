package services

import (
	"context"
	"errors"
	"products/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockProductsRepository struct {
	mock.Mock
}

func (m *MockProductsRepository) Create(ctx context.Context, createDTO *models.CreateProductDTO) (*models.Product, error) {
	args := m.Called(ctx, createDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductsRepository) Delete(ctx context.Context, id string) (*models.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Product), args.Error(1)
}

func (m *MockProductsRepository) List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, error) {
	args := m.Called(ctx, listDTO)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Product), args.Error(1)
}

func (m *MockProductsRepository) Count(ctx context.Context) (int, error) {
	args := m.Called(ctx)
	return args.Int(0), args.Error(1)
}

type MockMessageBroker struct {
	mock.Mock
}

func (m *MockMessageBroker) Send(ctx context.Context, topic string, message, key []byte) error {
	args := m.Called(ctx, topic, message, key)
	return args.Error(0)
}

func (m *MockMessageBroker) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestProductService(t *testing.T) {
	ctx := context.Background()
	mockRepo := new(MockProductsRepository)
	mockBroker := new(MockMessageBroker)
	service := NewProductsService(mockRepo, mockBroker, zap.NewNop())

	t.Run("CreateProduct", func(t *testing.T) {
		product := &models.Product{
			ID:          "uuid-1",
			Name:        "Test Product",
			Price:       100,
			Description: "A product for testing",
			CreatedAt:   time.Now(),
		}
		createDTO := &models.CreateProductDTO{
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
		}
		t.Run("Success", func(t *testing.T) {

			mockRepo.On("Create", ctx, createDTO).Return(product, nil).Once()
			mockBroker.On("Send", ctx, "", mock.Anything, []byte(product.ID)).Return(nil).Once()
			actualProduct, err := service.Create(ctx, createDTO)

			assert.NoError(t, err)
			assert.Equal(t, product, actualProduct)

			mockRepo.AssertExpectations(t)
			mockBroker.AssertExpectations(t)

		})

		t.Run("Error", func(t *testing.T) {
			repoErr := errors.New("repository error")
			mockRepo.On("Create", ctx, createDTO).Return(nil, repoErr).Once()

			actualProduct, err := service.Create(ctx, createDTO)

			assert.Nil(t, actualProduct)
			assert.Equal(t, repoErr, err)

			mockRepo.AssertExpectations(t)
			mockBroker.AssertNotCalled(t, "Send")
		})

	})

	t.Run("DeleteProduct", func(t *testing.T) {
		product := &models.Product{
			ID:          "uuid-1",
			Name:        "Test Product",
			Price:       100,
			Description: "A product for testing",
			CreatedAt:   time.Now(),
		}
		deleteDTO := &models.DeleteProductDTO{
			ID: product.ID,
		}
		t.Run("Success", func(t *testing.T) {

			mockRepo.On("Delete", ctx, deleteDTO.ID).Return(product, nil).Once()
			mockBroker.On("Send", ctx, "", mock.Anything, []byte(product.ID)).Return(nil).Once()
			actualProduct, err := service.Delete(ctx, deleteDTO.ID)

			assert.NoError(t, err)
			assert.Equal(t, product, actualProduct)

			mockRepo.AssertExpectations(t)
			mockBroker.AssertExpectations(t)

		})

		t.Run("Error", func(t *testing.T) {
			repoErr := errors.New("repository error")
			mockRepo.On("Delete", ctx, deleteDTO.ID).Return(nil, repoErr).Once()

			actualProduct, err := service.Delete(ctx, deleteDTO.ID)

			assert.Nil(t, actualProduct)
			assert.Equal(t, repoErr, err)

			mockRepo.AssertExpectations(t)
			mockBroker.AssertNotCalled(t, "Send")
		})
	})

	t.Run("ListProducts", func(t *testing.T) {
		products := []models.Product{
			{
				ID:          "uuid-1",
				Name:        "Test Product 1",
				Price:       100,
				Description: "A product for testing",
				CreatedAt:   time.Now(),
			},
			{
				ID:          "uuid-2",
				Name:        "Test Product 2",
				Price:       200,
				Description: "Another product for testing",
				CreatedAt:   time.Now(),
			},
		}
		listDTO := &models.ListProductsDTO{
			Page:  1,
			Limit: 10,
		}
		t.Run("Success", func(t *testing.T) {
			mockRepo.On("Count", ctx).Return(len(products), nil).Once()
			mockRepo.On("List", ctx, listDTO).Return(products, nil).Once()

			actualProducts, total, err := service.List(ctx, listDTO)

			assert.NoError(t, err)
			assert.Equal(t, products, actualProducts)
			assert.Equal(t, len(products), total)

			mockRepo.AssertExpectations(t)
		})

		t.Run("Error on Count", func(t *testing.T) {
			repoErr := errors.New("repository error")
			mockRepo.On("Count", ctx).Return(0, repoErr).Once()

			actualProducts, total, err := service.List(ctx, listDTO)

			assert.Nil(t, actualProducts)
			assert.Equal(t, 0, total)
			assert.Equal(t, repoErr, err)

			mockRepo.AssertExpectations(t)
		})

		t.Run("Error on List", func(t *testing.T) {
			repoErr := errors.New("repository error")
			mockRepo.On("Count", ctx).Return(len(products), nil).Once()
			mockRepo.On("List", ctx, listDTO).Return(nil, repoErr).Once()

			actualProducts, total, err := service.List(ctx, listDTO)

			assert.Nil(t, actualProducts)
			assert.Equal(t, 0, total)
			assert.Equal(t, repoErr, err)

			mockRepo.AssertExpectations(t)
			mockBroker.AssertNotCalled(t, "Send")
		})
	})

}
