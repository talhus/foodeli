package outbox

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// MockRepository provides an in-memory thread-safe implementation of Repository for testing.
type MockRepository struct {
	mu     sync.Mutex
	Events map[uuid.UUID]*OutboxEvent
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		Events: make(map[uuid.UUID]*OutboxEvent),
	}
}

func (m *MockRepository) CreateTx(ctx context.Context, tx pgx.Tx, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	id := uuid.New()
	m.Events[id] = &OutboxEvent{
		ID:            id,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Topic:         topic,
		Payload:       payloadBytes,
		Status:        StatusPending,
		RetryCount:    0,
		CreatedAt:     time.Now().UTC(),
	}
	return nil
}

func (m *MockRepository) Create(ctx context.Context, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error {
	return m.CreateTx(ctx, nil, aggregateType, aggregateID, topic, payload)
}

func (m *MockRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return nil, nil
}

func (m *MockRepository) FetchPendingTx(ctx context.Context, tx pgx.Tx, limit int) ([]OutboxEvent, error) {
	return m.FetchPending(ctx, limit)
}

func (m *MockRepository) MarkPublishedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	return m.MarkPublished(ctx, id)
}

func (m *MockRepository) IncrementRetryTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, maxRetries int) error {
	return m.IncrementRetry(ctx, id, maxRetries)
}

func (m *MockRepository) FetchPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var result []OutboxEvent
	for _, e := range m.Events {
		if e.Status == StatusPending {
			result = append(result, *e)
			if len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (m *MockRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.Events[id]
	if ok {
		e.Status = StatusPublished
		now := time.Now().UTC()
		e.PublishedAt = &now
	}
	return nil
}

func (m *MockRepository) IncrementRetry(ctx context.Context, id uuid.UUID, maxRetries int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	e, ok := m.Events[id]
	if ok {
		e.RetryCount++
		if e.RetryCount >= maxRetries {
			e.Status = StatusFailed
		}
	}
	return nil
}
