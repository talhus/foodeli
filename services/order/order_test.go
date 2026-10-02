package order_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"foodeli/platform/middleware"
	"foodeli/services/order"
	"foodeli/services/product"
	"foodeli/services/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type MockStoreChecker struct {
	stores map[uuid.UUID]*store.Store
}

type MockEventPublisher struct{}

func (m *MockEventPublisher) PublishJSON(ctx context.Context, routingKey string, payload any) error {
	return nil
}

// get store
func (m *MockStoreChecker) GetStore(ctx context.Context, id uuid.UUID) (*store.Store, error) {
	st, exists := m.stores[id]
	if !exists {
		return nil, order.ErrStoreNotFound
	}
	return st, nil
}

// get merchant store
func (m *MockStoreChecker) GetMerchantStore(ctx context.Context, merchantID uuid.UUID) (*store.Store, error) {
	for _, st := range m.stores {
		if st.MerchantID == merchantID {
			return st, nil
		}
	}
	return nil, order.ErrStoreNotFound
}

//mock product checker

type MockProductChecker struct {
	products map[uuid.UUID]*product.Product
}

// get product by id
func (m *MockProductChecker) GetProductById(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	product, exists := m.products[id]
	if !exists {
		return nil, order.ErrProductNotFound
	}
	return product, nil
}

// mock order repo
type MockOrderRepository struct {
	orders map[uuid.UUID]*order.Order
}

func NewMockOrderRepository() *MockOrderRepository {
	return &MockOrderRepository{
		orders: make(map[uuid.UUID]*order.Order),
	}
}

// REPOS
func (m *MockOrderRepository) CreateOrder(ctx context.Context, req *order.Order) (*order.Order, error) {
	if req.IdempotencyKey != "" {
		for _, existing := range m.orders {
			if existing.UserID == req.UserID && existing.IdempotencyKey == req.IdempotencyKey {
				return existing, nil
			}
		}
	}
	m.orders[req.ID] = req
	return req, nil
}

func (m *MockOrderRepository) GetById(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	orders, exists := m.orders[id]
	if !exists {
		return nil, order.ErrOrderNotFound
	}
	return orders, nil
}

func (m *MockOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]order.Order, error) {
	var orders []order.Order
	for _, o := range m.orders {
		if o.UserID == userID {
			orders = append(orders, *o)
		}
	}
	return orders, nil
}

func (m *MockOrderRepository) ListByStoreID(ctx context.Context, storeID uuid.UUID) ([]order.Order, error) {
	var orders []order.Order
	for _, o := range m.orders {
		if o.StoreID == storeID {
			orders = append(orders, *o)
		}
	}
	return orders, nil
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	orders, exists := m.orders[id]
	if !exists {
		return order.ErrOrderNotFound
	}
	orders.Status = status
	return nil
}

func generateTestToken(userID uuid.UUID, role string, secret string) string {
	claims := middleware.JWTClaims{
		UserID: userID,
		Email:  "test@example.com",
		Role:   role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func setupOrderRouter() (*gin.Engine, *MockStoreChecker, *MockProductChecker, *MockOrderRepository, string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockStore := &MockStoreChecker{stores: make(map[uuid.UUID]*store.Store)}
	mockProduct := &MockProductChecker{products: make(map[uuid.UUID]*product.Product)}
	mockOrderRepo := NewMockOrderRepository()
	mockEvt := &MockEventPublisher{}

	service := order.NewService(mockStore, mockProduct, mockOrderRepo, mockEvt)
	handler := order.NewHandler(service)

	jwtSecret := "order_test_secret"
	v1 := r.Group("/api/v1")
	handler.RegisterRoutes(v1, middleware.AuthMiddleware(jwtSecret))

	return r, mockStore, mockProduct, mockOrderRepo, jwtSecret
}

// ========================
// TESTS
// ========================
func TestOrderCreationAndIdempotency(t *testing.T) {
	r, mockStore, mockProduct, mockOrderRepo, secret := setupOrderRouter()

	merchantID := uuid.New()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	// Seed store and product into mocks
	mockStore.stores[storeID] = &store.Store{
		ID:         storeID,
		MerchantID: merchantID,
		Name:       "Pizza Hub",
		Address:    "Moda Cad. No:42 Istanbul",
	}

	mockProduct.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "Margherita Pizza",
		Price:   1500, // 15.00 TL in cents
		Stock:   10,
	}

	customerToken := generateTestToken(customerID, "customer", secret)

	orderPayload := order.CreateOrderRequest{
		StoreID:         storeID,
		DeliveryAddress: "Kadikoy Moda Cad. No:10 Istanbul",
		Items: []order.OrderItemRequest{
			{
				ProductID: productID,
				Quantity:  2,
			},
		},
	}
	body, _ := json.Marshal(orderPayload)

	// 1. First Checkout Request with Idempotency-Key
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(body))
	req1.Header.Set("Authorization", "Bearer "+customerToken)
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Idempotency-Key", "idemp-key-001")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created on first checkout, got %d. Body: %s", w1.Code, w1.Body.String())
	}

	var resp1 struct {
		Data order.Order `json:"data"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &resp1); err != nil {
		t.Fatalf("failed to parse order response: %v", err)
	}

	firstOrder := resp1.Data
	if firstOrder.TotalAmount != 3000 { // 2 * 1500
		t.Errorf("expected total_amount 3000, got %d", firstOrder.TotalAmount)
	}
	if firstOrder.Status != order.StatusPending {
		t.Errorf("expected status %s, got %s", order.StatusPending, firstOrder.Status)
	}
	if len(firstOrder.Items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(firstOrder.Items))
	}
	if firstOrder.Items[0].ProductName != "Margherita Pizza" {
		t.Errorf("expected product_name 'Margherita Pizza', got %s", firstOrder.Items[0].ProductName)
	}
	if firstOrder.Items[0].UnitPrice != 1500 {
		t.Errorf("expected unit_price 1500, got %d", firstOrder.Items[0].UnitPrice)
	}
	if firstOrder.Items[0].Subtotal != 3000 {
		t.Errorf("expected subtotal 3000, got %d", firstOrder.Items[0].Subtotal)
	}

	// 2. Second Checkout Request with the EXACT SAME Idempotency-Key (Duplicate submission)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(body))
	req2.Header.Set("Authorization", "Bearer "+customerToken)
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Idempotency-Key", "idemp-key-001")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected status 201 on idempotent checkout, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var resp2 struct {
		Data order.Order `json:"data"`
	}
	if err := json.Unmarshal(w2.Body.Bytes(), &resp2); err != nil {
		t.Fatalf("failed to parse second order response: %v", err)
	}

	secondOrder := resp2.Data
	if secondOrder.ID != firstOrder.ID {
		t.Errorf("expected idempotent response to return identical order ID %s, got %s", firstOrder.ID, secondOrder.ID)
	}

	// 3. Verify that only 1 order exists in the repository
	if len(mockOrderRepo.orders) != 1 {
		t.Errorf("expected exactly 1 order in storage due to idempotency, got %d", len(mockOrderRepo.orders))
	}
}

func TestOrderValidationAndStockErrors(t *testing.T) {
	r, mockStore, mockProduct, _, secret := setupOrderRouter()
	merchantID := uuid.New()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	// Mock Store
	mockStore.stores[storeID] = &store.Store{
		ID:         storeID,
		MerchantID: merchantID,
		Name:       "Burger Hub",
		Address:    "Tampere",
	}

	// Mock Product with only 3 in stock
	mockProduct.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "CheeseBurger",
		Price:   2500,
		Stock:   3,
	}

	customerToken := generateTestToken(customerID, "customer", secret)

	tests := []struct {
		name           string
		storeID        uuid.UUID
		items          []order.OrderItemRequest
		expectedStatus int
	}{
		{
			name:           "EmptyItems",
			storeID:        storeID,
			items:          []order.OrderItemRequest{},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "NonExistentStore",
			storeID:        uuid.New(),
			items:          []order.OrderItemRequest{{ProductID: productID, Quantity: 1}},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "NonExistentProduct",
			storeID:        storeID,
			items:          []order.OrderItemRequest{{ProductID: uuid.New(), Quantity: 1}},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "InsufficientStock",
			storeID:        storeID,
			items:          []order.OrderItemRequest{{ProductID: productID, Quantity: 10}}, // Stock is only 3!
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			payload := order.CreateOrderRequest{
				StoreID:         tc.storeID,
				DeliveryAddress: "Lielahti 40, Tampere",
				Items:           tc.items,
			}
			body, _ := json.Marshal(payload)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/orders", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+customerToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("expected status %d, got %d. Body: %s", tc.expectedStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestOrderStatusTransitionsAndPermissions(t *testing.T) {
	r, mockStore, _, mockOrderRepo, secret := setupOrderRouter()

	merchantAID := uuid.New()
	merchantBID := uuid.New()
	customerAID := uuid.New()
	customerBID := uuid.New()
	storeAID := uuid.New()
	storeBID := uuid.New()
	orderID := uuid.New()

	// Seed Stores
	mockStore.stores[storeAID] = &store.Store{ID: storeAID, MerchantID: merchantAID, Name: "Store A"}
	mockStore.stores[storeBID] = &store.Store{ID: storeBID, MerchantID: merchantBID, Name: "Store B"}

	// Seed Initial Order in Store A with status PENDING
	mockOrderRepo.orders[orderID] = &order.Order{
		ID:              orderID,
		UserID:          customerAID,
		StoreID:         storeAID,
		DeliveryAddress: "Customer A Address",
		TotalAmount:     5000,
		Status:          order.StatusPending,
	}

	merchantAToken := generateTestToken(merchantAID, "merchant", secret)
	merchantBToken := generateTestToken(merchantBID, "merchant", secret)
	customerBToken := generateTestToken(customerBID, "customer", secret)

	// 1. Customer B tries to view Customer A's order -> Should be 403 Forbidden
	t.Run("CustomerCannotViewOthersOrder", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+customerBToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", w.Code)
		}
	})

	// 2. Merchant B (owner of Store B) tries to update Store A's order -> Should be 403 Forbidden
	t.Run("MerchantCannotUpdateOthersStoreOrder", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"status": order.StatusConfirmed})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/"+orderID.String()+"/status", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+merchantBToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403 Forbidden, got %d", w.Code)
		}
	})

	// 3. Valid Lifecycle Transitions by Merchant A: PENDING -> CONFIRMED -> PREPARING -> DELIVERED
	transitions := []string{order.StatusConfirmed, order.StatusPreparing, order.StatusDelivered}
	for _, nextStatus := range transitions {
		t.Run("TransitionTo_"+nextStatus, func(t *testing.T) {
			body, _ := json.Marshal(map[string]string{"status": nextStatus})
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/"+orderID.String()+"/status", bytes.NewBuffer(body))
			req.Header.Set("Authorization", "Bearer "+merchantAToken)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("expected 200 OK for transition to %s, got %d. Body: %s", nextStatus, w.Code, w.Body.String())
			}
		})
	}

	// 4. Invalid Transition: Order is already DELIVERED, attempting to transition back to PREPARING -> 400 Bad Request
	t.Run("InvalidTransitionFromDelivered", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"status": order.StatusPreparing})
		req := httptest.NewRequest(http.MethodPatch, "/api/v1/orders/"+orderID.String()+"/status", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+merchantAToken)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 Bad Request for invalid transition from DELIVERED, got %d", w.Code)
		}
	})
}

func TestListOrders(t *testing.T) {
	r, mockStore, _, mockOrderRepo, secret := setupOrderRouter()

	merchantID := uuid.New()
	customerID := uuid.New()
	storeID := uuid.New()

	mockStore.stores[storeID] = &store.Store{ID: storeID, MerchantID: merchantID, Name: "Store X"}

	order1 := &order.Order{
		ID:          uuid.New(),
		UserID:      customerID,
		StoreID:     storeID,
		TotalAmount: 1000,
		Status:      order.StatusPending,
	}
	mockOrderRepo.orders[order1.ID] = order1

	customerToken := generateTestToken(customerID, "customer", secret)
	merchantToken := generateTestToken(merchantID, "merchant", secret)

	// 1. List Customer Orders
	t.Run("ListCustomerOrders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/customer/mine", nil)
		req.Header.Set("Authorization", "Bearer "+customerToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var resp struct {
			Data []order.Order `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Errorf("expected 1 order, got %d", len(resp.Data))
		}
	})

	// 2. List Merchant Orders
	t.Run("ListMerchantOrders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/orders/merchant/mine", nil)
		req.Header.Set("Authorization", "Bearer "+merchantToken)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", w.Code)
		}

		var resp struct {
			Data []order.Order `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("failed to parse response: %v", err)
		}
		if len(resp.Data) != 1 {
			t.Errorf("expected 1 order, got %d", len(resp.Data))
		}
	})
}
