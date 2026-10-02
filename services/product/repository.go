package product

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	Create(ctx context.Context, product *Product) error
	GetById(ctx context.Context, id uuid.UUID) (*Product, error)
	ListByStoreId(ctx context.Context, storeId uuid.UUID) ([]Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) Repository {
	return &PostgresRepository{pool: db}
}

// CREATE PRODUCT
func (r *PostgresRepository) Create(ctx context.Context, product *Product) error {
	query := `INSERT INTO products(id,store_id,name,description,price,stock,created_at,updated_at) VALUES($1, $2, $3, $4, $5, $6, NOW(), NOW()) returning created_at,updated_at`
	return r.pool.QueryRow(ctx, query, product.ID, product.StoreID, product.Name, product.Description, product.Price, product.Stock).Scan(&product.CreatedAt, &product.UpdatedAt)
}

// GET BY ID
func (r *PostgresRepository) GetById(ctx context.Context, id uuid.UUID) (*Product, error) {
	query := `SELECT id,store_id,name,description,price,stock,created_at,updated_at,deleted_at FROM products WHERE id = $1 and deleted_at is null`
	var product Product
	err := r.pool.QueryRow(ctx, query, id).Scan(&product.ID, &product.StoreID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return &product, nil
}

// LIST BY STORE ID
func (r *PostgresRepository) ListByStoreId(ctx context.Context, storeId uuid.UUID) ([]Product, error) {
	query := `SELECT id,store_id,name,description,price,stock,created_at,updated_at,deleted_at FROM products WHERE store_id = $1 and deleted_at is null`
	rows, err := r.pool.Query(ctx, query, storeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.StoreID, &product.Name, &product.Description, &product.Price, &product.Stock, &product.CreatedAt, &product.UpdatedAt, &product.DeletedAt)
		if err != nil {
			return nil, err
		}
		products = append(products, product)
	}
	return products, nil
}

// UPDATE
func (r *PostgresRepository) Update(ctx context.Context, product *Product) error {
	query := `UPDATE products SET name = $1, description = $2, price = $3, stock = $4, updated_at = NOW() WHERE id = $5 and deleted_at is null returning updated_at`
	err := r.pool.QueryRow(ctx, query, product.Name, product.Description, product.Price, product.Stock, product.ID).Scan(&product.UpdatedAt)
	if err != nil {
		return ErrProductNotFound
	}
	return nil
}

// DELETE
func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE products SET deleted_at=NOW() WHERE id = $1 and deleted_at is null`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrProductNotFound
	}
	return nil
}
