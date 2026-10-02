package store

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Store represents a restaurant or merchant storefront in Foodeli.
type Store struct {
	ID         uuid.UUID `json:"id"`
	MerchantID uuid.UUID `json:"merchant_id"`
	Name       string    `json:"name"`
	Address    string    `json:"address"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MenuItem represents a product or food dish belonging to a specific store.
type MenuItem struct {
	ID          uuid.UUID `json:"id"`
	StoreID     uuid.UUID `json:"store_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateStoreRequest defines the payload for creating a new store.
type CreateStoreRequest struct {
	Name    string `json:"name" binding:"required,min=2,max=100"`
	Address string `json:"address" binding:"required,min=3"`
}

// Validate sanitizes and validates the store creation request.
func (r *CreateStoreRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Address = strings.TrimSpace(r.Address)

	if len(r.Name) < 2 {
		return errors.New("store name must be at least 2 characters long")
	}
	if len(r.Address) < 3 {
		return errors.New("store address must be at least 3 characters long")
	}
	return nil
}

// CreateMenuItemRequest defines the payload for adding an item to a store's menu.
type CreateMenuItemRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=100"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,gt=0"`
}

// Validate sanitizes and validates the menu item creation request.
func (r *CreateMenuItemRequest) Validate() error {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)

	if len(r.Name) < 2 {
		return errors.New("menu item name must be at least 2 characters long")
	}
	if r.Price <= 0 {
		return errors.New("price must be greater than zero")
	}
	return nil
}
