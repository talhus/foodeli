package outbox_test

import (
	"context"
	"testing"

	"foodeli/platform/outbox"

	"github.com/google/uuid"
)

func TestMockRepository_Lifecycle(t *testing.T) {
	repo := outbox.NewMockRepository()
	ctx := context.Background()

	orderID := uuid.New()
	payload := map[string]any{"order_id": orderID.String(), "amount": 2500}

	// 1. Create event
	err := repo.CreateTx(ctx, nil, outbox.AggregateOrder, orderID, "order.created", payload)
	if err != nil {
		t.Fatalf("failed to create outbox event: %v", err)
	}

	// 2. Fetch pending
	pending, err := repo.FetchPending(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch pending events: %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending event, got %d", len(pending))
	}
	event := pending[0]
	if event.AggregateID != orderID || event.Topic != "order.created" {
		t.Errorf("unexpected event content: %+v", event)
	}

	// 3. Mark published
	err = repo.MarkPublished(ctx, event.ID)
	if err != nil {
		t.Fatalf("failed to mark published: %v", err)
	}

	// 4. Verify no pending remaining
	pendingAfter, err := repo.FetchPending(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch pending events: %v", err)
	}
	if len(pendingAfter) != 0 {
		t.Fatalf("expected 0 pending events after publish, got %d", len(pendingAfter))
	}
}
