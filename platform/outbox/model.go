package outbox

import (
	"time"

	"github.com/google/uuid"
)

// Outbox Event Status Constants
const (
	StatusPending   = "PENDING"
	StatusPublished = "PUBLISHED"
	StatusFailed    = "FAILED"
)

// Aggregate Type Constants for Domain Tracking
const (
	AggregateOrder   = "ORDER"
	AggregatePayment = "PAYMENT"
)

// OutboxEvent represents an atomic event record stored in PostgreSQL.
type OutboxEvent struct {
	ID            uuid.UUID  `json:"id"`
	AggregateType string     `json:"aggregate_type"`
	AggregateID   uuid.UUID  `json:"aggregate_id"`
	Topic         string     `json:"topic"`
	Payload       []byte     `json:"payload"`
	Status        string     `json:"status"`
	RetryCount    int        `json:"retry_count"`
	CreatedAt     time.Time  `json:"created_at"`
	PublishedAt   *time.Time `json:"published_at,omitempty"`
}
