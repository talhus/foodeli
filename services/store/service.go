package store

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrUnauthorizedStoreAccess = errors.New("you do not have permission to manage this store")
)

// Service defines business operations for managing stores and menus.
type Service interface {
	CreateStore(ctx context.Context, merchantID uuid.UUID, req *CreateStoreRequest) (*Store, error)
	GetStore(ctx context.Context, id uuid.UUID) (*Store, error)
	ListStores(ctx context.Context) ([]Store, error)
	GetMerchantStore(ctx context.Context, merchantID uuid.UUID) (*Store, error)

	CreateMenuItem(ctx context.Context, merchantID uuid.UUID, storeID uuid.UUID, req *CreateMenuItemRequest) (*MenuItem, error)
	ListMenuItems(ctx context.Context, storeID uuid.UUID) ([]MenuItem, error)
	GetMenuItem(ctx context.Context, id uuid.UUID) (*MenuItem, error)
}

// StoreService implements Service.
type StoreService struct {
	repo Repository
}

// NewService creates a new instance of StoreService.
func NewService(repo Repository) *StoreService {
	return &StoreService{repo: repo}
}

// CreateStore validates input and creates a new store for the merchant (max 1 store per merchant).
func (s *StoreService) CreateStore(ctx context.Context, merchantID uuid.UUID, req *CreateStoreRequest) (*Store, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Enforce 1 merchant = 1 store rule
	existing, err := s.repo.GetStoreByMerchantID(ctx, merchantID)
	if err == nil && existing != nil {
		return nil, ErrMerchantStoreAlreadyExists
	}

	store := &Store{
		MerchantID: merchantID,
		Name:       req.Name,
		Address:    req.Address,
	}

	if err := s.repo.CreateStore(ctx, store); err != nil {
		return nil, err
	}

	return store, nil
}

// GetStore retrieves a store by ID.
func (s *StoreService) GetStore(ctx context.Context, id uuid.UUID) (*Store, error) {
	return s.repo.GetStoreById(ctx, id)
}

// ListStores returns all stores.
func (s *StoreService) ListStores(ctx context.Context) ([]Store, error) {
	stores, err := s.repo.ListStores(ctx)
	if err != nil {
		return nil, err
	}
	if stores == nil {
		stores = []Store{}
	}
	return stores, nil
}

// GetMerchantStore returns the single store owned by the merchant.
func (s *StoreService) GetMerchantStore(ctx context.Context, merchantID uuid.UUID) (*Store, error) {
	return s.repo.GetStoreByMerchantID(ctx, merchantID)
}

// CreateMenuItem verifies store ownership and inserts a new menu item.
func (s *StoreService) CreateMenuItem(ctx context.Context, merchantID uuid.UUID, storeID uuid.UUID, req *CreateMenuItemRequest) (*MenuItem, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Verify store ownership
	st, err := s.repo.GetStoreById(ctx, storeID)
	if err != nil {
		return nil, err
	}

	if st.MerchantID != merchantID {
		return nil, ErrUnauthorizedStoreAccess
	}

	item := &MenuItem{
		StoreID:     storeID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
	}

	if err := s.repo.CreateMenuItem(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// ListMenuItems returns all items for a store.
func (s *StoreService) ListMenuItems(ctx context.Context, storeID uuid.UUID) ([]MenuItem, error) {
	// Check if store exists
	if _, err := s.repo.GetStoreById(ctx, storeID); err != nil {
		return nil, err
	}

	items, err := s.repo.GetMenuItemsByStoreID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []MenuItem{}
	}
	return items, nil
}

// GetMenuItem returns a menu item by ID.
func (s *StoreService) GetMenuItem(ctx context.Context, id uuid.UUID) (*MenuItem, error) {
	return s.repo.GetMenuItemById(ctx, id)
}
