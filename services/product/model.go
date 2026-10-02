package product

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrProductNotFound = errors.New("product not found")
	ErrStoreNotFound   = errors.New("store not found for this merchant")
)

type Product struct {
	ID          uuid.UUID  `json:"id"`
	StoreID     uuid.UUID  `json:"store_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Price       int64      `json:"price"`
	Stock       int        `json:"stock"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

type CreateProductRequest struct {
	Name        string `json:"name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"required,min=2,max=1000"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Stock       int    `json:"stock" binding:"required,gt=0"`
}

type UpdateProductRequest struct {
	Name        string `json:"name" binding:"min=2,max=100"`
	Description string `json:"description" binding:"min=2,max=1000"`
	Price       int64  `json:"price" binding:"gt=0"`
	Stock       int    `json:"stock" binding:"gt=0"`
}

type ProductResponse struct{}

type DeleteResponse struct{}
