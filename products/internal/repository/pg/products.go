package pg

import (
	"context"
	"database/sql"
	"products/internal/apperrors"
	"products/internal/models"

	"github.com/jmoiron/sqlx"
)

type ProductsRepository struct {
	db *sqlx.DB
}

func NewProductsRepository(db *sqlx.DB) *ProductsRepository {
	return &ProductsRepository{
		db: db,
	}
}

func (r *ProductsRepository) Create(ctx context.Context, createDTO *models.CreateProductDTO) (*models.Product, error) {
	var query = `
		INSERT INTO products (name, description, price) 
		VALUES ($1, $2, $3) 
		RETURNING id, name, description, price, created_at
	`
	var product models.Product
	err := r.db.GetContext(ctx, &product, query, createDTO.Name, createDTO.Description, createDTO.Price)

	return &product, err
}

func (r *ProductsRepository) Delete(ctx context.Context, id string) (*models.Product, error) {
	var query = `
		DELETE FROM products 
		WHERE id = $1
		RETURNING id, name, description, price, created_at
	`
	var product models.Product
	err := r.db.GetContext(ctx, &product, query, id)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &apperrors.ErrorNotFound{ID: id}
		}

		return nil, err
	}
	return &product, err
}

func (r *ProductsRepository) List(ctx context.Context, listDTO *models.ListProductsDTO) ([]models.Product, error) {
	var query = `
		SELECT id, name, description, price, created_at FROM products 
		ORDER BY created_at
		LIMIT $1 OFFSET $2 
	`

	offset := (listDTO.Page - 1) * listDTO.Limit
	var products []models.Product
	err := r.db.SelectContext(ctx, &products, query, listDTO.Limit, offset)

	if err != nil {
		return nil, err
	}

	if products == nil {
		return []models.Product{}, nil
	}

	return products, err
}

func (r *ProductsRepository) Count(ctx context.Context) (int, error) {
	query := `
		SELECT COUNT(*) 
		FROM products 
	`

	var count int
	err := r.db.GetContext(ctx, &count, query)
	return count, err
}
