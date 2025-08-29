package handlers

import (
	middleware "products/internal/middlewares"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func SetupRoutes(productsHandler *ProductsHandler, logger *zap.Logger) *gin.Engine {
	router := gin.New()
	router.Use(middleware.ZapLoggerMiddleware(logger))
	router.Use(middleware.ZapRecoveryMiddleware(logger, true))

	router.GET("/products", productsHandler.List)
	router.POST("/products", productsHandler.Create)
	router.DELETE("/products/:id", productsHandler.Delete)

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	return router
}
