package outbox_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"foodeli/platform/outbox"

	"github.com/google/uuid"
)

type MockBrokerPublisher struct {
	mu        sync.Mutex
	Messages  map[string][][]byte
	ShouldErr bool
}

func NewMockBrokerPublisher() *MockBrokerPublisher {
	return &MockBrokerPublisher{
		Messages: make(map[string][][]byte),
	}
}

func (m *MockBrokerPublisher) Publish(ctx context.Context, routingKey string, body []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.ShouldErr {
		return errors.New("connection reset by broker")
	}

	m.Messages[routingKey] = append(m.Messages[routingKey], body)
	return nil
}

func TestRelayWorker_ProcessBatch_Success(t *testing.T) {
	repo := outbox.NewMockRepository()
	broker := NewMockBrokerPublisher()

	cfg := outbox.WorkerConfig{
		PollInterval: 10 * time.Millisecond,
		BatchSize:    10,
		MaxRetries:   3,
	}
	worker := outbox.NewRelayWorker(repo, broker, cfg)

	ctx := context.Background()
	orderID := uuid.New()
	payload := map[string]any{"order_id": orderID.String(), "amount": 5000}

	// 1. Enqueue outbox event
	err := repo.CreateTx(ctx, nil, outbox.AggregateOrder, orderID, "order.created", payload)
	if err != nil {
		t.Fatalf("failed to create outbox event: %v", err)
	}

	// 2. Process batch
	worker.ProcessBatch(ctx)

	// 3. Verify message was received by broker
	broker.mu.Lock()
	messages, exists := broker.Messages["order.created"]
	broker.mu.Unlock()

	if !exists || len(messages) != 1 {
		t.Fatalf("expected 1 message published to 'order.created', got %d", len(messages))
	}

	// 4. Verify event in repo is marked as PUBLISHED
	pending, _ := repo.FetchPending(ctx, 10)
	if len(pending) != 0 {
		t.Fatalf("expected 0 pending events after relay processing, got %d", len(pending))
	}
}

func TestRelayWorker_ProcessBatch_BrokerFailure_Retries(t *testing.T) {
	repo := outbox.NewMockRepository()
	broker := NewMockBrokerPublisher()
	broker.ShouldErr = true // Broker is down

	cfg := outbox.WorkerConfig{
		PollInterval: 10 * time.Millisecond,
		BatchSize:    10,
		MaxRetries:   2,
	}
	worker := outbox.NewRelayWorker(repo, broker, cfg)

	ctx := context.Background()
	orderID := uuid.New()
	_ = repo.CreateTx(ctx, nil, outbox.AggregateOrder, orderID, "order.created", map[string]any{"order_id": orderID})

	// 1st attempt: fails, increments retry to 1, still PENDING
	worker.ProcessBatch(ctx)
	pending, _ := repo.FetchPending(ctx, 10)
	if len(pending) != 1 || pending[0].RetryCount != 1 {
		t.Fatalf("expected 1 pending event with retry_count 1, got %+v", pending)
	}

	// 2nd attempt: fails, exceeds MaxRetries (2), transitions to FAILED
	worker.ProcessBatch(ctx)
	pendingAfter, _ := repo.FetchPending(ctx, 10)
	if len(pendingAfter) != 0 {
		t.Fatalf("expected 0 pending events after exceeding max retries, got %d", len(pendingAfter))
	}
}

func TestRelayWorker_Lifecycle_StartStop(t *testing.T) {
	repo := outbox.NewMockRepository()
	broker := NewMockBrokerPublisher()

	cfg := outbox.WorkerConfig{
		PollInterval: 10 * time.Millisecond,
		BatchSize:    10,
		MaxRetries:   3,
	}
	worker := outbox.NewRelayWorker(repo, broker, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	worker.Start(ctx)
	time.Sleep(25 * time.Millisecond)
	worker.Stop()
}
