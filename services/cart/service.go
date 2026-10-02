package cart

import (
	"context"
	"errors"
	"foodeli/services/order"
	"foodeli/services/product"
	"foodeli/services/store"
	"time"

	"github.com/google/uuid"
)

type StoreChecker interface {
	GetStore(ctx context.Context, id uuid.UUID) (*store.Store, error)
}
type ProductChecker interface {
	GetProductById(ctx context.Context, id uuid.UUID) (*product.Product, error)
}
type OrderCreator interface {
	CreateOrder(ctx context.Context, userID uuid.UUID, req *order.CreateOrderRequest, idempotencyKey string) (*order.Order, error)
}

type Service interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error)
	AddItem(ctx context.Context, userID uuid.UUID, req *AddToCartRequest) (*Cart, error)
	UpdateItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*Cart, error)
	RemoveItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*Cart, error)
	ClearCart(ctx context.Context, userID uuid.UUID) error
	Checkout(ctx context.Context, userID uuid.UUID, req *CheckoutRequest, idempotencyKey string) (*order.Order, error)
}

type CartService struct {
	repo           Repository
	storeChecker   StoreChecker
	productChecker ProductChecker
	orderCreator   OrderCreator
}

func NewService(repo Repository, s StoreChecker, p ProductChecker, o OrderCreator) *CartService {
	return &CartService{
		repo:           repo,
		storeChecker:   s,
		productChecker: p,
		orderCreator:   o,
	}
}

// ======================================================
// METHODS
// ======================================================

// get cart
func (s *CartService) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrCartNotFound) {
			return &Cart{
				UserID:    userID,
				Items:     []CartItem{},
				UpdatedAt: time.Now().UTC(),
			}, nil
		}
		return nil, err
	}
	return cart, nil
}

// add item
func (s *CartService) AddItem(ctx context.Context, userID uuid.UUID, req *AddToCartRequest) (*Cart, error) {
	if req.Quantity <= 0 {
		return nil, ErrInvalidQuantity
	}

	//check store
	st, err := s.storeChecker.GetStore(ctx, req.StoreID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	prod, err := s.productChecker.GetProductById(ctx, req.ProductID)
	if err != nil || prod.StoreID != req.StoreID {
		return nil, ErrProductNotFound
	}

	cart, err := s.GetCart(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrCartNotFound) {
			cart = &Cart{
				UserID:    userID,
				StoreID:   &req.StoreID,
				StoreName: st.Name,
				Items:     []CartItem{},
			}
		} else {
			return nil, err
		}
	}

	//single store!
	if cart.StoreID != nil && *cart.StoreID != req.StoreID && len(cart.Items) > 0 {
		if !req.Force {
			return nil, ErrCrossStoreConflict
		}
		cart.StoreID = &req.StoreID
		cart.StoreName = st.Name
		cart.Items = []CartItem{}
	}

	itemIndex := -1
	for i, item := range cart.Items {
		if item.ProductID == req.ProductID {
			itemIndex = i
			break
		}
	}
	targetQuantity := req.Quantity
	if itemIndex >= 0 {
		targetQuantity += cart.Items[itemIndex].Quantity
	}

	if prod.Stock < targetQuantity {
		return nil, ErrInsufficientStock
	}

	if itemIndex >= 0 {
		cart.Items[itemIndex].Quantity = targetQuantity
		cart.Items[itemIndex].UnitPrice = prod.Price
	} else {
		cart.Items = append(cart.Items, CartItem{
			ProductID:   prod.ID,
			ProductName: prod.Name,
			UnitPrice:   prod.Price,
			Quantity:    req.Quantity,
		})
	}

	cart.StoreID = &req.StoreID
	cart.StoreName = st.Name
	cart.RecalculateTotals()

	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil

}

// update item quantity
func (s *CartService) UpdateItemQuantity(ctx context.Context, userID uuid.UUID, productID uuid.UUID, quantity int) (*Cart, error) {
	if quantity <= 0 {
		return s.RemoveItem(ctx, userID, productID)
	}

	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	itemIndex := -1
	for i, item := range cart.Items {
		if item.ProductID == productID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		return nil, ErrItemNotFoundInCart
	}

	prod, err := s.productChecker.GetProductById(ctx, productID)
	if err != nil {
		return nil, ErrProductNotFound
	}

	if prod.Stock < quantity {
		return nil, ErrInsufficientStock
	}

	cart.Items[itemIndex].Quantity = quantity
	cart.Items[itemIndex].UnitPrice = prod.Price

	cart.RecalculateTotals()

	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil
}

// remove item
func (s *CartService) RemoveItem(ctx context.Context, userID uuid.UUID, productID uuid.UUID) (*Cart, error) {
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	itemIndex := -1
	for i, item := range cart.Items {
		if item.ProductID == productID {
			itemIndex = i
			break
		}
	}

	if itemIndex == -1 {
		return nil, ErrItemNotFoundInCart
	}

	cart.Items = append(cart.Items[:itemIndex], cart.Items[itemIndex+1:]...)

	cart.RecalculateTotals()

	if err := s.repo.SaveCart(ctx, cart); err != nil {
		return nil, err
	}
	return cart, nil

}

// clear cart
func (s *CartService) ClearCart(ctx context.Context, userID uuid.UUID) error {
	return s.repo.DeleteCart(ctx, userID)
}

// checkout
func (s *CartService) Checkout(ctx context.Context, userID uuid.UUID, req *CheckoutRequest, idempotencyKey string) (*order.Order, error) {
	cart, err := s.repo.GetCart(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(cart.Items) == 0 || cart.StoreID == nil {
		return nil, ErrCartEmpty
	}

	orderItems := make([]order.OrderItemRequest, len(cart.Items))
	for i, item := range cart.Items {
		orderItems[i] = order.OrderItemRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	createOrderReq := &order.CreateOrderRequest{
		StoreID:         *cart.StoreID,
		DeliveryAddress: req.DeliveryAddress,
		Items:           orderItems,
	}

	createOrder, err := s.orderCreator.CreateOrder(ctx, userID, createOrderReq, idempotencyKey)
	if err != nil {
		return nil, err
	}

	_ = s.repo.DeleteCart(ctx, userID)

	return createOrder, nil
}
