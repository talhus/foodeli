package outbox

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines database operations for the Transactional Outbox pattern.
type Repository interface {
	// CreateTx inserts an outbox event within an existing database transaction.
	CreateTx(ctx context.Context, tx pgx.Tx, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error

	// Create inserts an outbox event by wrapping it in a standalone transaction.
	Create(ctx context.Context, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error

	// BeginTx starts a new transaction for the outbox relay worker.
	BeginTx(ctx context.Context) (pgx.Tx, error)

	// FetchPendingTx fetches and locks pending outbox events using FOR UPDATE SKIP LOCKED.
	FetchPendingTx(ctx context.Context, tx pgx.Tx, limit int) ([]OutboxEvent, error)

	// MarkPublishedTx marks an outbox event as PUBLISHED within a transaction.
	MarkPublishedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error

	// IncrementRetryTx increments the retry count and marks as FAILED if maxRetries reached.
	IncrementRetryTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, maxRetries int) error

	// FetchPending fetches pending events directly using pool (convenience/fallback).
	FetchPending(ctx context.Context, limit int) ([]OutboxEvent, error)

	// MarkPublished marks an outbox event as PUBLISHED directly using pool.
	MarkPublished(ctx context.Context, id uuid.UUID) error

	// IncrementRetry increments retry count directly using pool.
	IncrementRetry(ctx context.Context, id uuid.UUID, maxRetries int) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateTx serializes payload and inserts a new pending outbox event inside the provided transaction.
func (r *PostgresRepository) CreateTx(ctx context.Context, tx pgx.Tx, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error {
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal outbox event payload: %w", err)
	}

	query := `INSERT INTO outbox_events (id, aggregate_type, aggregate_id, topic, payload, status, retry_count, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, 0, NOW())`

	_, err = tx.Exec(ctx, query, uuid.New(), aggregateType, aggregateID, topic, payloadBytes, StatusPending)
	if err != nil {
		return fmt.Errorf("failed to insert outbox event: %w", err)
	}

	return nil
}

// Create inserts an outbox event in a dedicated transaction using connection pool.
func (r *PostgresRepository) Create(ctx context.Context, aggregateType string, aggregateID uuid.UUID, topic string, payload any) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for outbox event: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.CreateTx(ctx, tx, aggregateType, aggregateID, topic, payload); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// BeginTx starts a transaction on the connection pool.
func (r *PostgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// FetchPendingTx selects pending events with FOR UPDATE SKIP LOCKED to prevent multi-worker concurrency collisions.
func (r *PostgresRepository) FetchPendingTx(ctx context.Context, tx pgx.Tx, limit int) ([]OutboxEvent, error) {
	query := `SELECT id, aggregate_type, aggregate_id, topic, payload, status, retry_count, created_at, published_at
	          FROM outbox_events
	          WHERE status = $1
	          ORDER BY created_at ASC
	          LIMIT $2
	          FOR UPDATE SKIP LOCKED`

	rows, err := tx.Query(ctx, query, StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox events: %w", err)
	}
	defer rows.Close()

	return scanOutboxEvents(rows)
}

// MarkPublishedTx updates the status of an event to PUBLISHED and sets published_at timestamp within a transaction.
func (r *PostgresRepository) MarkPublishedTx(ctx context.Context, tx pgx.Tx, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = $1, published_at = NOW() WHERE id = $2`
	_, err := tx.Exec(ctx, query, StatusPublished, id)
	return err
}

// IncrementRetryTx increments retry_count and sets status to FAILED if threshold is exceeded within a transaction.
func (r *PostgresRepository) IncrementRetryTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, maxRetries int) error {
	query := `UPDATE outbox_events 
	          SET retry_count = retry_count + 1,
	              status = CASE WHEN retry_count + 1 >= $1 THEN 'FAILED' ELSE status END
	          WHERE id = $2`
	_, err := tx.Exec(ctx, query, maxRetries, id)
	return err
}

// FetchPending fetches pending events directly from connection pool.
func (r *PostgresRepository) FetchPending(ctx context.Context, limit int) ([]OutboxEvent, error) {
	query := `SELECT id, aggregate_type, aggregate_id, topic, payload, status, retry_count, created_at, published_at
	          FROM outbox_events
	          WHERE status = $1
	          ORDER BY created_at ASC
	          LIMIT $2`

	rows, err := r.pool.Query(ctx, query, StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending outbox events: %w", err)
	}
	defer rows.Close()

	return scanOutboxEvents(rows)
}

// MarkPublished marks an event as PUBLISHED directly via pool.
func (r *PostgresRepository) MarkPublished(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE outbox_events SET status = $1, published_at = NOW() WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, StatusPublished, id)
	return err
}

// IncrementRetry increments retry count directly via pool.
func (r *PostgresRepository) IncrementRetry(ctx context.Context, id uuid.UUID, maxRetries int) error {
	query := `UPDATE outbox_events 
	          SET retry_count = retry_count + 1,
	              status = CASE WHEN retry_count + 1 >= $1 THEN 'FAILED' ELSE status END
	          WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, maxRetries, id)
	return err
}

func scanOutboxEvents(rows pgx.Rows) ([]OutboxEvent, error) {
	var events []OutboxEvent
	for rows.Next() {
		var e OutboxEvent
		if err := rows.Scan(
			&e.ID,
			&e.AggregateType,
			&e.AggregateID,
			&e.Topic,
			&e.Payload,
			&e.Status,
			&e.RetryCount,
			&e.CreatedAt,
			&e.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan outbox event row: %w", err)
		}
		events = append(events, e)
	}
	return events, rows.Err()
}
