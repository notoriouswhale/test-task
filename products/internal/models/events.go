package models

import (
	"time"
)

type ProductEventType string

const (
	ProductCreated ProductEventType = "product_created"
	ProductDeleted ProductEventType = "product_deleted"
)

type ProductEvent struct {
	EventType ProductEventType `json:"event_type"`
	Product   *Product         `json:"product"`
	Timestamp time.Time        `json:"timestamp"`
}
