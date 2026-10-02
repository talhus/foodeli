package payment

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreatePayment(ctx context.Context, p *Payment) error
	GetById(ctx context.Context, paymentID uuid.UUID) (*Payment, error)
	GetByOrderId(ctx context.Context, orderID uuid.UUID) (*Payment, error)
	UpdateStatus(ctx context.Context, paymentID uuid.UUID, status string, failureReason *string) error
	UpdateTransactionAndStatus(ctx context.Context, paymentID uuid.UUID, transactionID string, status string) error
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &PostgresRepository{pool: pool}
}

//=================================
//METHODS
//=================================

func (r *PostgresRepository) CreatePayment(ctx context.Context, p *Payment) error {
	query := `INSERT INTO payments(id,order_id,user_id,amount,currency,payment_method,provider,transaction_id,status,failure_reason,created_at,updated_at)
	VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,NOW(),NOW()) RETURNING created_at,updated_at`
	return r.pool.QueryRow(ctx, query,
		p.ID,
		p.OrderID,
		p.UserID,
		p.Amount,
		p.Currency,
		p.PaymentMethod,
		p.Provider,
		p.TransactionID,
		p.Status,
		p.FailureReason,
	).Scan(&p.CreatedAt, &p.UpdatedAt)

}
func (r *PostgresRepository) GetById(ctx context.Context, paymentID uuid.UUID) (*Payment, error) {
	query := paymentSelectQuery + ` WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, paymentID)
	return r.scanPayment(row)
}
func (r *PostgresRepository) GetByOrderId(ctx context.Context, orderID uuid.UUID) (*Payment, error) {
	query := paymentSelectQuery + ` WHERE order_id = $1`
	row := r.pool.QueryRow(ctx, query, orderID)
	return r.scanPayment(row)
}
func (r *PostgresRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status string, failureReason *string) error {
	query := `UPDATE payments SET status = $1, failure_reason = $2, updated_at = NOW() WHERE id = $3`
	cmdTag, err := r.pool.Exec(ctx, query, status, failureReason, paymentID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

func (r *PostgresRepository) UpdateTransactionAndStatus(ctx context.Context, paymentID uuid.UUID, transactionID string, status string) error {
	query := `UPDATE payments SET transaction_id = $1, status = $2, updated_at = NOW() WHERE id = $3`
	cmdTag, err := r.pool.Exec(ctx, query, transactionID, status, paymentID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrPaymentNotFound
	}
	return nil
}

// helpers
const paymentSelectQuery = `SELECT id,order_id,user_id,amount,currency,payment_method,provider,transaction_id,status,failure_reason,created_at,updated_at FROM payments`

func (r *PostgresRepository) scanPayment(row pgx.Row) (*Payment, error) {
	var p Payment
	err := row.Scan(&p.ID, &p.OrderID, &p.UserID, &p.Amount, &p.Currency, &p.PaymentMethod, &p.Provider, &p.TransactionID, &p.Status, &p.FailureReason, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &p, nil
}
