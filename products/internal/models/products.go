package models

import "time"

type Product struct {
	ID          string `json:"id,omitempty" db:"id"`
	Name        string `json:"name,omitempty" db:"name"`
	Description string `json:"description,omitempty" db:"description"`
	// Price is stored in cents.
	Price     int       `json:"price,omitempty" db:"price"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type CreateProductDTO struct {
	Name        string `json:"name,omitempty" binding:"required,min=3,max=50"`
	Description string `json:"description,omitempty" binding:"max=200"`
	Price       int    `json:"price,omitempty" binding:"required,gt=0"`
}

type DeleteProductDTO struct {
	ID string `uri:"id" binding:"required,uuid"`
}

type ListProductsDTO struct {
	Page  int `form:"page"`
	Limit int `form:"limit"`
}
