package outbox

import (
	"context"
	"log"
	"time"
)

// BrokerPublisher defines the contract for sending raw message bytes to the event broker (RabbitMQ).
type BrokerPublisher interface {
	Publish(ctx context.Context, routingKey string, body []byte) error
}

// WorkerConfig holds configuration options for the Outbox Relay Worker.
type WorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int
	MaxRetries   int
}

// DefaultWorkerConfig returns production-grade defaults for the relay worker.
func DefaultWorkerConfig() WorkerConfig {
	return WorkerConfig{
		PollInterval: 500 * time.Millisecond,
		BatchSize:    50,
		MaxRetries:   5,
	}
}

// RelayWorker continuously polls pending outbox events from PostgreSQL and dispatches them to RabbitMQ.
type RelayWorker struct {
	repo      Repository
	publisher BrokerPublisher
	cfg       WorkerConfig
	stopCh    chan struct{}
}

// NewRelayWorker creates a new Outbox RelayWorker instance.
func NewRelayWorker(repo Repository, publisher BrokerPublisher, cfg WorkerConfig) *RelayWorker {
	if cfg.PollInterval <= 0 {
		cfg.PollInterval = 500 * time.Millisecond
	}
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = 50
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = 5
	}

	return &RelayWorker{
		repo:      repo,
		publisher: publisher,
		cfg:       cfg,
		stopCh:    make(chan struct{}),
	}
}

// Start launches the background relay worker goroutine.
func (w *RelayWorker) Start(ctx context.Context) {
	log.Printf("[Outbox Relay Worker] Started (polling interval: %v, batch size: %d, max retries: %d)",
		w.cfg.PollInterval, w.cfg.BatchSize, w.cfg.MaxRetries)

	go func() {
		ticker := time.NewTicker(w.cfg.PollInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("[Outbox Relay Worker] Context cancelled, shutting down...")
				return
			case <-w.stopCh:
				log.Println("[Outbox Relay Worker] Stop signal received, shutting down...")
				return
			case <-ticker.C:
				w.ProcessBatch(ctx)
			}
		}
	}()
}

// Stop signals the background worker to stop gracefully.
func (w *RelayWorker) Stop() {
	select {
	case <-w.stopCh:
		// already closed
	default:
		close(w.stopCh)
	}
}

// ProcessBatch fetches a batch of pending events under row locks and delivers them to the broker.
func (w *RelayWorker) ProcessBatch(ctx context.Context) {
	tx, err := w.repo.BeginTx(ctx)
	if err != nil {
		log.Printf("[Outbox Relay Worker] Failed to begin transaction: %v", err)
		return
	}
	if tx != nil {
		defer tx.Rollback(ctx)
	}

	events, err := w.repo.FetchPendingTx(ctx, tx, w.cfg.BatchSize)
	if err != nil {
		log.Printf("[Outbox Relay Worker] Failed to fetch pending events: %v", err)
		return
	}

	if len(events) == 0 {
		return
	}

	for _, event := range events {
		if err := w.publisher.Publish(ctx, event.Topic, event.Payload); err != nil {
			log.Printf("[Outbox Relay Worker] Failed to publish event %s to topic %s: %v", event.ID, event.Topic, err)
			_ = w.repo.IncrementRetryTx(ctx, tx, event.ID, w.cfg.MaxRetries)
			continue
		}

		if err := w.repo.MarkPublishedTx(ctx, tx, event.ID); err != nil {
			log.Printf("[Outbox Relay Worker] Failed to mark event %s as published: %v", event.ID, err)
			continue
		}
	}

	if tx != nil {
		if err := tx.Commit(ctx); err != nil {
			log.Printf("[Outbox Relay Worker] Failed to commit batch transaction: %v", err)
		}
	}
}
