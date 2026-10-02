package order

import (
	"context"
	"errors"
	"fmt"
	"foodeli/platform/events"
	"foodeli/platform/outbox"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	CreateOrder(ctx context.Context, order *Order) (*Order, error)

	ListByUserID(ctx context.Context, userID uuid.UUID) ([]Order, error)

	ListByStoreID(ctx context.Context, storeID uuid.UUID) ([]Order, error)

	GetById(ctx context.Context, orderID uuid.UUID) (*Order, error)

	UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error
}

func NewRepository(pool *pgxpool.Pool, outboxRepo outbox.Repository) Repository {
	return &PostgresRepository{pool: pool, outboxRepo: outboxRepo}
}

type PostgresRepository struct {
	pool       *pgxpool.Pool
	outboxRepo outbox.Repository
}

// CREATE ORDER
func (r *PostgresRepository) CreateOrder(ctx context.Context, order *Order) (*Order, error) {
	if order.IdempotencyKey != "" {
		var existingID uuid.UUID
		checkQuery := `SELECT id FROM orders where user_id=$1 AND idempotency_key=$2`
		err := r.pool.QueryRow(ctx, checkQuery, order.UserID, order.IdempotencyKey).Scan(&existingID)
		if err == nil {
			return r.GetById(ctx, existingID)
		} else if !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}
	}

	//start transaction

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	//query for check product, stock deduction and price snapshot
	deductStockQuery := `UPDATE products SET stock=stock-$1, updated_at=NOW() WHERE id=$2 AND stock>=$1 AND deleted_at IS NULL`

	for _, item := range order.Items {
		cmdTag, err := tx.Exec(ctx, deductStockQuery, item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}
		if cmdTag.RowsAffected() == 0 {
			return nil, ErrInsufficientStock
		}

	}

	//save the order
	insertOrderQuery := `INSERT INTO orders(id,user_id,store_id,delivery_address,total_amount,status,idempotency_key,created_at,updated_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW(),NOW()) RETURNING created_at,updated_at`

	var idempotencyVal *string
	if order.IdempotencyKey != "" {
		idempotencyVal = &order.IdempotencyKey
	}

	err = tx.QueryRow(ctx, insertOrderQuery, order.ID, order.UserID, order.StoreID, order.DeliveryAddress, order.TotalAmount, order.Status, idempotencyVal).Scan(&order.CreatedAt, &order.UpdatedAt)
	if err != nil {
		return nil, err
	}

	itemInsertQuery := `INSERT INTO order_items(id,order_id,product_id,product_name,unit_price,quantity,subtotal,created_at) VALUES($1,$2,$3,$4,$5,$6,$7,NOW()) RETURNING created_at`

	for i := range order.Items {
		item := &order.Items[i]
		err := tx.QueryRow(ctx, itemInsertQuery, item.ID, order.ID, item.ProductID, item.ProductName, item.UnitPrice, item.Quantity, item.Subtotal).Scan(&item.CreatedAt)
		if err != nil {
			return nil, err
		}
	}

	// Atomically enqueue order.created domain event in outbox_events within the SAME database transaction
	if r.outboxRepo != nil {
		eventPayload := events.OrderCreatedPayload{
			OrderID:         order.ID,
			UserID:          order.UserID,
			StoreID:         order.StoreID,
			TotalAmount:     order.TotalAmount,
			DeliveryAddress: order.DeliveryAddress,
			ItemCount:       len(order.Items),
			OccurredAt:      order.CreatedAt,
		}
		if err := r.outboxRepo.CreateTx(ctx, tx, outbox.AggregateOrder, order.ID, events.OrderCreatedRoutingKey, eventPayload); err != nil {
			return nil, fmt.Errorf("failed to enqueue outbox event for order %s: %w", order.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return order, nil
}

// LIST BY USER ID
func (r *PostgresRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]Order, error) {
	query := `SELECT id,user_id,store_id,delivery_address,total_amount,status,COALESCE(idempotency_key,''),created_at,updated_at 
	FROM orders WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrders(ctx, rows)
}

// LIST BY STORE ID
func (r *PostgresRepository) ListByStoreID(ctx context.Context, storeID uuid.UUID) ([]Order, error) {
	query := `SELECT id,user_id,store_id,delivery_address,total_amount,status,COALESCE(idempotency_key,''),created_at,updated_at FROM orders WHERE store_id=$1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return r.scanOrders(ctx, rows)
}

// GET BY ID
func (r *PostgresRepository) GetById(ctx context.Context, orderID uuid.UUID) (*Order, error) {
	//find the order
	query := `SELECT id,user_id,store_id,delivery_address,total_amount,status,COALESCE(idempotency_key,''),created_at,updated_at 
	FROM orders 
	WHERE id = $1`
	var o Order
	err := r.pool.QueryRow(ctx, query, orderID).Scan(
		&o.ID,
		&o.UserID,
		&o.StoreID,
		&o.DeliveryAddress,
		&o.TotalAmount,
		&o.Status,
		&o.IdempotencyKey,
		&o.CreatedAt,
		&o.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	//find items
	itemsQuery := `SELECT id,order_id,product_id,product_name,unit_price,quantity,subtotal,created_at 
	FROM order_items WHERE order_id = $1 ORDER BY created_at ASC`
	rows, err := r.pool.Query(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var oi OrderItem
		err = rows.Scan(
			&oi.ID,
			&oi.OrderID,
			&oi.ProductID,
			&oi.ProductName,
			&oi.UnitPrice,
			&oi.Quantity,
			&oi.Subtotal,
			&oi.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		o.Items = append(o.Items, oi)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &o, nil
}

// UPDATE STATUS
func (r *PostgresRepository) UpdateStatus(ctx context.Context, orderID uuid.UUID, status string) error {
	query := `UPDATE orders SET status=$1,updated_at=NOW() WHERE id=$2`
	cmdTag, err := r.pool.Exec(ctx, query, status, orderID)

	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

// helper
func (r *PostgresRepository) scanOrders(ctx context.Context, rows pgx.Rows) ([]Order, error) {
	orders := make([]Order, 0)
	orderIndexMap := make(map[uuid.UUID]int)

	for rows.Next() {
		var o Order
		err := rows.Scan(
			&o.ID,
			&o.UserID,
			&o.StoreID,
			&o.DeliveryAddress,
			&o.TotalAmount,
			&o.Status,
			&o.IdempotencyKey,
			&o.CreatedAt,
			&o.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		o.Items = []OrderItem{}
		orderIndexMap[o.ID] = len(orders)
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(orders) == 0 {
		return orders, nil
	}

	orderIDs := make([]uuid.UUID, len(orders))
	for i, o := range orders {
		orderIDs[i] = o.ID
	}

	itemsQuery := `SELECT id, order_id, product_id, product_name, unit_price, quantity, subtotal, created_at
	FROM order_items
	WHERE order_id = ANY($1)
	ORDER BY created_at ASC`

	itemRows, err := r.pool.Query(ctx, itemsQuery, orderIDs)
	if err != nil {
		return nil, err
	}
	defer itemRows.Close()

	for itemRows.Next() {
		var oi OrderItem
		if err := itemRows.Scan(
			&oi.ID,
			&oi.OrderID,
			&oi.ProductID,
			&oi.ProductName,
			&oi.UnitPrice,
			&oi.Quantity,
			&oi.Subtotal,
			&oi.CreatedAt,
		); err != nil {
			return nil, err
		}
		if idx, ok := orderIndexMap[oi.OrderID]; ok {
			orders[idx].Items = append(orders[idx].Items, oi)
		}
	}

	if err := itemRows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}
