package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrStoreNotFound              = errors.New("store not found")
	ErrMenuItemNotFound           = errors.New("menu item not found")
	ErrMerchantStoreAlreadyExists = errors.New("merchant already has a registered store")
)

// Repository defines database operations for stores and menu items.
type Repository interface {
	CreateStore(ctx context.Context, store *Store) error
	GetStoreById(ctx context.Context, id uuid.UUID) (*Store, error)
	ListStores(ctx context.Context) ([]Store, error)
	GetStoreByMerchantID(ctx context.Context, merchantID uuid.UUID) (*Store, error)

	CreateMenuItem(ctx context.Context, item *MenuItem) error
	GetMenuItemsByStoreID(ctx context.Context, storeID uuid.UUID) ([]MenuItem, error)
	GetMenuItemById(ctx context.Context, id uuid.UUID) (*MenuItem, error)
}

// PostgresRepository implements Repository using PostgreSQL pgxpool.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new instance of PostgresRepository for store management.
func NewRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// CreateStore inserts a new store with an application-generated UUID.
func (r *PostgresRepository) CreateStore(ctx context.Context, store *Store) error {
	store.ID = uuid.New()
	query := `
		INSERT INTO stores (id, merchant_id, name, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, store.ID, store.MerchantID, store.Name, store.Address).
		Scan(&store.CreatedAt, &store.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return ErrMerchantStoreAlreadyExists
		}
		return fmt.Errorf("failed to insert store: %w", err)
	}

	return nil
}

// GetStoreById retrieves a store by its unique ID.
func (r *PostgresRepository) GetStoreById(ctx context.Context, id uuid.UUID) (*Store, error) {
	query := `
		SELECT id, merchant_id, name, address, created_at, updated_at
		FROM stores
		WHERE id = $1
	`

	var store Store
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&store.ID,
		&store.MerchantID,
		&store.Name,
		&store.Address,
		&store.CreatedAt,
		&store.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		return nil, fmt.Errorf("failed to get store by id: %w", err)
	}

	return &store, nil
}

// ListStores returns all stores, sorted by creation date descending.
func (r *PostgresRepository) ListStores(ctx context.Context) ([]Store, error) {
	query := `
		SELECT id, merchant_id, name, address, created_at, updated_at
		FROM stores
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list stores: %w", err)
	}
	defer rows.Close()

	var stores []Store
	for rows.Next() {
		var s Store
		if err := rows.Scan(&s.ID, &s.MerchantID, &s.Name, &s.Address, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan store row: %w", err)
		}
		stores = append(stores, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return stores, nil
}

// GetStoreByMerchantID returns the single store owned by a merchant.
func (r *PostgresRepository) GetStoreByMerchantID(ctx context.Context, merchantID uuid.UUID) (*Store, error) {
	query := `
		SELECT id, merchant_id, name, address, created_at, updated_at
		FROM stores
		WHERE merchant_id = $1
	`

	var s Store
	err := r.pool.QueryRow(ctx, query, merchantID).Scan(
		&s.ID,
		&s.MerchantID,
		&s.Name,
		&s.Address,
		&s.CreatedAt,
		&s.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrStoreNotFound
		}
		return nil, fmt.Errorf("failed to get merchant store: %w", err)
	}

	return &s, nil
}

// CreateMenuItem inserts a new menu item with an application-generated UUID.
func (r *PostgresRepository) CreateMenuItem(ctx context.Context, item *MenuItem) error {
	item.ID = uuid.New()
	query := `
		INSERT INTO menu_items (id, store_id, name, description, price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING created_at, updated_at
	`

	err := r.pool.QueryRow(ctx, query, item.ID, item.StoreID, item.Name, item.Description, item.Price).
		Scan(&item.CreatedAt, &item.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to insert menu item: %w", err)
	}

	return nil
}

// GetMenuItemsByStoreID returns all items on a store's menu.
func (r *PostgresRepository) GetMenuItemsByStoreID(ctx context.Context, storeID uuid.UUID) ([]MenuItem, error) {
	query := `
		SELECT id, store_id, name, description, price, created_at, updated_at
		FROM menu_items
		WHERE store_id = $1
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to query menu items: %w", err)
	}
	defer rows.Close()

	var items []MenuItem
	for rows.Next() {
		var item MenuItem
		if err := rows.Scan(&item.ID, &item.StoreID, &item.Name, &item.Description, &item.Price, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %w", err)
		}
		items = append(items, item)
	}

	return items, nil
}

// GetMenuItemById retrieves a single menu item by ID.
func (r *PostgresRepository) GetMenuItemById(ctx context.Context, id uuid.UUID) (*MenuItem, error) {
	query := `
		SELECT id, store_id, name, description, price, created_at, updated_at
		FROM menu_items
		WHERE id = $1
	`

	var item MenuItem
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.StoreID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
		&item.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrMenuItemNotFound
		}
		return nil, fmt.Errorf("failed to get menu item: %w", err)
	}

	return &item, nil
}
