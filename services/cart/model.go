package cart

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Domain Errors
var (
	ErrCartNotFound       = errors.New("cart is empty or not found")
	ErrCartEmpty          = errors.New("cannot checkout an empty cart")
	ErrCrossStoreConflict = errors.New("cart contains items from another store; clear cart first")
	ErrProductNotFound    = errors.New("product not found")
	ErrInsufficientStock  = errors.New("insufficient stock for product")
	ErrItemNotFoundInCart = errors.New("product not found in cart")
	ErrInvalidQuantity    = errors.New("quantity must be greater than zero")
)

// CartItem represents an item stored in the customer's ephemeral cart.
type CartItem struct {
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   int64     `json:"unit_price"` // minor units (cents)
	Quantity    int       `json:"quantity"`
	Subtotal    int64     `json:"subtotal"` // unit_price * quantity
}

// Cart represents the customer's active shopping cart in Redis.
type Cart struct {
	UserID      uuid.UUID  `json:"user_id"`
	StoreID     *uuid.UUID `json:"store_id,omitempty"` // nil if cart is empty
	StoreName   string     `json:"store_name,omitempty"`
	Items       []CartItem `json:"items"`
	TotalAmount int64      `json:"total_amount"` // minor units (cents)
	ItemCount   int        `json:"item_count"`   // sum of quantities
	UpdatedAt   time.Time  `json:"updated_at"`
}

// RecalculateTotals recalculates ItemCount and TotalAmount based on current items.
func (c *Cart) RecalculateTotals() {
	var total int64
	var count int
	for i := range c.Items {
		c.Items[i].Subtotal = c.Items[i].UnitPrice * int64(c.Items[i].Quantity)
		total += c.Items[i].Subtotal
		count += c.Items[i].Quantity
	}
	c.TotalAmount = total
	c.ItemCount = count
	c.UpdatedAt = time.Now().UTC()
}

// Request DTOs

type AddToCartRequest struct {
	StoreID   uuid.UUID `json:"store_id" binding:"required"`
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
	Force     bool      `json:"force"` // If true and cart has items from another store, clears old cart
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,gt=0"`
}

type CheckoutRequest struct {
	DeliveryAddress string `json:"delivery_address" binding:"required,min=5"`
}
