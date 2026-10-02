package product

import (
	"context"
	"errors"
	"foodeli/services/store"

	"github.com/google/uuid"
)

var (
	ErrUnauthorizedProductAccess = errors.New("unauthorized product access")
)

type StoreChecker interface {
	GetMerchantStore(ctx context.Context, merchantID uuid.UUID) (*store.Store, error)
}

type Service interface {
	CreateProduct(ctx context.Context, merchantID uuid.UUID, req *CreateProductRequest) (*Product, error)
	GetProductById(ctx context.Context, id uuid.UUID) (*Product, error)
	ListProductsByStoreId(ctx context.Context, storeId uuid.UUID) ([]Product, error)
	UpdateProduct(ctx context.Context, merchantID, productID uuid.UUID, req *UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, merchantID, productID uuid.UUID) error
}

type ProductService struct {
	productRepo  Repository
	storeChecker StoreChecker
}

func NewService(productRepo Repository, storeChecker StoreChecker) Service {
	return &ProductService{
		productRepo:  productRepo,
		storeChecker: storeChecker,
	}
}

// create
func (s *ProductService) CreateProduct(ctx context.Context, merchantID uuid.UUID, req *CreateProductRequest) (*Product, error) {
	store, err := s.storeChecker.GetMerchantStore(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	product := &Product{
		ID:          uuid.New(),
		StoreID:     store.ID,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// GET PRODUCT
func (s *ProductService) GetProductById(ctx context.Context, id uuid.UUID) (*Product, error) {
	return s.productRepo.GetById(ctx, id)
}

// LIST PRODUCT BY STORE ID
func (s *ProductService) ListProductsByStoreId(ctx context.Context, id uuid.UUID) ([]Product, error) {
	return s.productRepo.ListByStoreId(ctx, id)
}

// UPDATE PRODUCT
func (s *ProductService) UpdateProduct(ctx context.Context, merchantID, productID uuid.UUID, req *UpdateProductRequest) (*Product, error) {
	product, err := s.productRepo.GetById(ctx, productID)
	if err != nil {
		return nil, err
	}

	store, err := s.storeChecker.GetMerchantStore(ctx, merchantID)
	if err != nil || product.StoreID != store.ID {
		return nil, err
	}
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.Stock = req.Stock

	if err := s.productRepo.Update(ctx, product); err != nil {
		return nil, err
	}
	return product, nil
}

// DELETE PRODUCT
func (s *ProductService) DeleteProduct(ctx context.Context, merchantID, productID uuid.UUID) error {
	product, err := s.productRepo.GetById(ctx, productID)
	if err != nil {
		return err
	}
	store, err := s.storeChecker.GetMerchantStore(ctx, merchantID)
	if err != nil || product.StoreID != store.ID {
		return ErrUnauthorizedProductAccess
	}
	return s.productRepo.Delete(ctx, product.ID)
}
