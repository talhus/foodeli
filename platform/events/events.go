package events

import (
	"time"

	"github.com/google/uuid"
)

// Routing Key Constants
const (
	OrderCreatedRoutingKey     = "order.created"
	PaymentSucceededRoutingKey = "payment.succeeded"
)

// OrderCreatedPayload represents the data inside an order.created event.
type OrderCreatedPayload struct {
	OrderID         uuid.UUID `json:"order_id"`
	UserID          uuid.UUID `json:"user_id"`
	StoreID         uuid.UUID `json:"store_id"`
	TotalAmount     int64     `json:"total_amount"` 
	DeliveryAddress string    `json:"delivery_address"`
	ItemCount       int       `json:"item_count"`
	OccurredAt      time.Time `json:"occurred_at"`
}

// PaymentSucceededPayload represents the data inside a payment.succeeded event.
type PaymentSucceededPayload struct {
	PaymentID     uuid.UUID `json:"payment_id"`
	OrderID       uuid.UUID `json:"order_id"`
	UserID        uuid.UUID `json:"user_id"`
	Amount        int64     `json:"amount"` 
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id"`
	OccurredAt    time.Time `json:"occurred_at"`
}
