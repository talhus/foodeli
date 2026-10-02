package order

import (
	"context"
	"foodeli/platform/events"
	"foodeli/services/product"
	"foodeli/services/store"

	"github.com/google/uuid"
)

type Service interface {
	GetOrderById(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*Order, error)
	CreateOrder(ctx context.Context, userID uuid.UUID, req *CreateOrderRequest, idempotencyKey string) (*Order, error)
	ListCustomerOrders(ctx context.Context, userID uuid.UUID) ([]Order, error)
	ListMerchantOrders(ctx context.Context, merchantID uuid.UUID) ([]Order, error)
	UpdateOrderStatus(ctx context.Context, merchantID, orderID uuid.UUID, newStatus string) (*Order, error)
	MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error
}

type StoreChecker interface {
	GetMerchantStore(ctx context.Context, merchantID uuid.UUID) (*store.Store, error)
	GetStore(ctx context.Context, id uuid.UUID) (*store.Store, error)
}

type ProductChecker interface {
	GetProductById(ctx context.Context, id uuid.UUID) (*product.Product, error)
}

type EventPublisher interface {
	PublishJSON(ctx context.Context, routingKey string, payload any) error
}
type OrderService struct {
	storeChecker   StoreChecker
	productChecker ProductChecker
	repository     Repository
	eventPublisher EventPublisher
}

func NewService(s StoreChecker, p ProductChecker, r Repository, e EventPublisher) *OrderService {
	return &OrderService{
		productChecker: p,
		storeChecker:   s,
		repository:     r,
		eventPublisher: e,
	}
}

//================================================
// methods
//================================================

// get order
func (s *OrderService) GetOrderById(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*Order, error) {
	o, err := s.repository.GetById(ctx, orderID)
	if err != nil {
		return nil, err
	}

	switch role {
	case "customer":
		if o.UserID != requesterID {
			return nil, ErrUnauthorizedOrderAccess
		}
	case "merchant":
		st, err := s.storeChecker.GetMerchantStore(ctx, requesterID)
		if err != nil || o.StoreID != st.ID {
			return nil, ErrUnauthorizedOrderAccess
		}
	case "admin":
		//admin can see everything
	default:
		return nil, ErrUnauthorizedOrderAccess

	}

	return o, nil
}

// create order
func (s *OrderService) CreateOrder(ctx context.Context, userID uuid.UUID, req *CreateOrderRequest, idempotencyKey string) (*Order, error) {
	//check cart
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	//check shop
	if _, err := s.storeChecker.GetStore(ctx, req.StoreID); err != nil {
		return nil, ErrStoreNotFound
	}

	orderID := uuid.New()
	var totalAmount int64
	orderItems := make([]OrderItem, 0, len(req.Items))

	//validate porudcts, and snapshot it
	for _, itemReq := range req.Items {
		prod, err := s.productChecker.GetProductById(ctx, itemReq.ProductID)
		if err != nil {
			return nil, ErrProductNotFound
		}

		//validate storeid
		if prod.StoreID != req.StoreID {
			return nil, ErrProductNotFound
		}

		//pre check stock
		if prod.Stock < itemReq.Quantity {
			return nil, ErrInsufficientStock
		}

		subtotal := prod.Price * int64(itemReq.Quantity)
		totalAmount += subtotal

		//add items snapshot
		orderItems = append(orderItems, OrderItem{
			ID:          uuid.New(),
			OrderID:     orderID,
			ProductID:   itemReq.ProductID,
			ProductName: prod.Name,
			UnitPrice:   prod.Price,
			Quantity:    itemReq.Quantity,
			Subtotal:    subtotal,
		})
	}

	//order entity
	order := &Order{
		ID:              orderID,
		UserID:          userID,
		StoreID:         req.StoreID,
		DeliveryAddress: req.DeliveryAddress,
		TotalAmount:     totalAmount,
		Status:          StatusPending,
		IdempotencyKey:  idempotencyKey,
		Items:           orderItems,
	}

	//save db
	createOrder, err := s.repository.CreateOrder(ctx, order)
	if err != nil {
		return nil, err
	}

	//publish event
	if s.eventPublisher != nil {
		_ = s.eventPublisher.PublishJSON(ctx, events.OrderCreatedRoutingKey, &events.OrderCreatedPayload{
			OrderID:         createOrder.ID,
			UserID:          createOrder.UserID,
			StoreID:         createOrder.StoreID,
			TotalAmount:     createOrder.TotalAmount,
			DeliveryAddress: createOrder.DeliveryAddress,
			ItemCount:       len(createOrder.Items),
			OccurredAt:      createOrder.CreatedAt,
		})
	}

	return createOrder, nil
}

// list customer oders
func (s *OrderService) ListCustomerOrders(ctx context.Context, userID uuid.UUID) ([]Order, error) {
	return s.repository.ListByUserID(ctx, userID)
}

// list merchant orders
func (s *OrderService) ListMerchantOrders(ctx context.Context, merchantID uuid.UUID) ([]Order, error) {
	st, err := s.storeChecker.GetMerchantStore(ctx, merchantID)
	if err != nil {
		return nil, err
	}
	return s.repository.ListByStoreID(ctx, st.ID)
}

// update order
func (s *OrderService) UpdateOrderStatus(ctx context.Context, merchantID, orderID uuid.UUID, newStatus string) (*Order, error) {
	st, err := s.storeChecker.GetMerchantStore(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	o, err := s.repository.GetById(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.StoreID != st.ID {
		return nil, ErrUnauthorizedOrderAccess
	}

	//validate status transition
	if !isValidTransition(o.Status, newStatus) {
		return nil, ErrInvalidOrderStatus
	}

	if err := s.repository.UpdateStatus(ctx, orderID, newStatus); err != nil {
		return nil, err
	}

	return s.repository.GetById(ctx, orderID)
}

// MarkOrderPaid transitions an order from PENDING to CONFIRMED upon successful payment.
func (s *OrderService) MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error {
	o, err := s.repository.GetById(ctx, orderID)
	if err != nil {
		return err
	}
	if o.Status != StatusPending {
		return ErrInvalidOrderStatus
	}
	return s.repository.UpdateStatus(ctx, orderID, StatusConfirmed)
}

// helper
func isValidTransition(current, next string) bool {
	switch current {
	case StatusPending:
		return next == StatusConfirmed || next == StatusCancelled
	case StatusConfirmed:
		return next == StatusPreparing || next == StatusCancelled
	case StatusPreparing:
		return next == StatusDelivered || next == StatusCancelled
	default:
		return false
	}
}
