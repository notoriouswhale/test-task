package handlers

import (
	"context"
	"net/http"
	"products/internal/apperrors"
	"products/internal/models"
	"products/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductsHandler struct {
	pService ProductService
	logger   *zap.Logger
}

type ProductService interface {
	Create(ctx context.Context, productDTO *models.CreateProductDTO) (*models.Product, error)
	Delete(ctx context.Context, id string) (*models.Product, error)
	List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, int, error)
}

func NewProductsHandler(pService ProductService, logger *zap.Logger) *ProductsHandler {
	return &ProductsHandler{
		pService: pService,
		logger:   logger.Named("ProductsHandler"),
	}
}

func (h *ProductsHandler) Create(c *gin.Context) {
	var createDTO models.CreateProductDTO

	err := c.ShouldBindJSON(&createDTO)
	if err != nil {
		h.logger.Error("CreateProductDTO binding error:", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	product, err := h.pService.Create(c.Request.Context(), &createDTO)
	if err != nil {
		h.logger.Error("Error creating product:", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    product,
	})
}

func (h *ProductsHandler) Delete(c *gin.Context) {
	var deleteDTO models.DeleteProductDTO
	err := c.ShouldBindUri(&deleteDTO)
	if err != nil {
		h.logger.Error("DeleteProductDTO binding error:", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	product, err := h.pService.Delete(c.Request.Context(), deleteDTO.ID)

	if err != nil {
		if apperrors.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		h.logger.Error("Error deleting product:", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Internal Server Error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    product,
	})
}

func (h *ProductsHandler) List(c *gin.Context) {
	var listDTO models.ListProductsDTO
	err := c.ShouldBindQuery(&listDTO)
	if err != nil {
		h.logger.Error("ListProductsDTO binding error:", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if listDTO.Page < 1 {
		listDTO.Page = 1
	}

	if listDTO.Limit < 1 {
		listDTO.Limit = 20
	} else if listDTO.Limit > 100 {
		listDTO.Limit = 100
	}

	products, total, err := h.pService.List(c.Request.Context(), &listDTO)

	if err != nil {
		h.logger.Error("Error listing products:", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "InternalServerError",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    products,
		"total":   total,
		"page":    listDTO.Page,
		"size":    listDTO.Limit,
		"pages":   utils.CalculateTotalPages(total, listDTO.Limit),
	})
}
