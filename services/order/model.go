package order

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Order Lifecycle Status Constants
const (
	StatusPending   = "PENDING"
	StatusConfirmed = "CONFIRMED"
	StatusPreparing = "PREPARING"
	StatusDelivered = "DELIVERED"
	StatusCancelled = "CANCELLED"
)

// Domain Errors
var (
	ErrOrderNotFound           = errors.New("order not found")
	ErrInsufficientStock       = errors.New("insufficient stock for product")
	ErrProductNotFound         = errors.New("product not found in store")
	ErrInvalidOrderStatus      = errors.New("invalid order status transition")
	ErrUnauthorizedOrderAccess = errors.New("unauthorized order access")
	ErrStoreNotFound           = errors.New("store not found")
	ErrEmptyOrder              = errors.New("order must contain at least one item")
)

// Order represents an order placed by a customer in a specific store.
type Order struct {
	ID              uuid.UUID   `json:"id"`
	UserID          uuid.UUID   `json:"user_id"`
	StoreID         uuid.UUID   `json:"store_id"`
	DeliveryAddress string      `json:"delivery_address"`
	TotalAmount     int64       `json:"total_amount"` // minor units (cents)
	Status          string      `json:"status"`
	Items           []OrderItem `json:"items,omitempty"`
	IdempotencyKey  string      `json:"idempotency_key"`
	CreatedAt       time.Time   `json:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"`
}

// OrderItem represents an immutable snapshot of an item inside an order.
type OrderItem struct {
	ID          uuid.UUID `json:"id"`
	OrderID     uuid.UUID `json:"order_id"`
	ProductID   uuid.UUID `json:"product_id"`
	ProductName string    `json:"product_name"`
	UnitPrice   int64     `json:"unit_price"`
	Quantity    int       `json:"quantity"`
	Subtotal    int64     `json:"subtotal"`
	CreatedAt   time.Time `json:"created_at"`
}

// OrderItemRequest represents an individual product line in order creation.
type OrderItemRequest struct {
	ProductID uuid.UUID `json:"product_id" binding:"required"`
	Quantity  int       `json:"quantity" binding:"required,gt=0"`
}

// CreateOrderRequest payload submitted by a customer.
type CreateOrderRequest struct {
	StoreID         uuid.UUID          `json:"store_id" binding:"required"`
	DeliveryAddress string             `json:"delivery_address" binding:"required,min=5,max=500"`
	Items           []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

// UpdateOrderStatusRequest payload submitted by a merchant.
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
}
