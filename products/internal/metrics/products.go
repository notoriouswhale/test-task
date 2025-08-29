package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ProductsCreated = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_created_total",
		Help: "Total number of created products",
	})

	ProductsDeleted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "products_deleted_total",
		Help: "Total number of deleted products",
	})
)
