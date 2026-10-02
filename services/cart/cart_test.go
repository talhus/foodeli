package cart_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"foodeli/platform/middleware"
	"foodeli/services/cart"
	"foodeli/services/order"
	"foodeli/services/product"
	"foodeli/services/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ======================================================================
// MOCKS
// ======================================================================

type MockCartRepository struct {
	mu    sync.Mutex
	carts map[uuid.UUID]*cart.Cart
}

func NewMockCartRepository() *MockCartRepository {
	return &MockCartRepository{
		carts: make(map[uuid.UUID]*cart.Cart),
	}
}

func (m *MockCartRepository) GetCart(ctx context.Context, userID uuid.UUID) (*cart.Cart, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	c, exists := m.carts[userID]
	if !exists {
		return nil, cart.ErrCartNotFound
	}

	// Deep copy to prevent pointer mutation issues in tests
	data, _ := json.Marshal(c)
	var copyCart cart.Cart
	_ = json.Unmarshal(data, &copyCart)
	return &copyCart, nil
}

func (m *MockCartRepository) SaveCart(ctx context.Context, c *cart.Cart) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, _ := json.Marshal(c)
	var copyCart cart.Cart
	_ = json.Unmarshal(data, &copyCart)
	m.carts[c.UserID] = &copyCart
	return nil
}

func (m *MockCartRepository) DeleteCart(ctx context.Context, userID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.carts, userID)
	return nil
}

type MockStoreChecker struct {
	mu     sync.Mutex
	stores map[uuid.UUID]*store.Store
}

func (m *MockStoreChecker) GetStore(ctx context.Context, id uuid.UUID) (*store.Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s, exists := m.stores[id]
	if !exists {
		return nil, cart.ErrProductNotFound
	}
	return s, nil
}

type MockProductChecker struct {
	mu       sync.Mutex
	products map[uuid.UUID]*product.Product
}

func (m *MockProductChecker) GetProductById(ctx context.Context, id uuid.UUID) (*product.Product, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p, exists := m.products[id]
	if !exists {
		return nil, cart.ErrProductNotFound
	}
	return p, nil
}

type MockOrderCreator struct {
	mu           sync.Mutex
	createdOrder *order.Order
}

func (m *MockOrderCreator) CreateOrder(ctx context.Context, userID uuid.UUID, req *order.CreateOrderRequest, idempotencyKey string) (*order.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var total int64
	items := make([]order.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = order.OrderItem{
			ID:        uuid.New(),
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	ord := &order.Order{
		ID:              uuid.New(),
		UserID:          userID,
		StoreID:         req.StoreID,
		DeliveryAddress: req.DeliveryAddress,
		TotalAmount:     total,
		Status:          order.StatusPending,
		Items:           items,
		IdempotencyKey:  idempotencyKey,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	m.createdOrder = ord
	return ord, nil
}

// ======================================================================
// TEST SETUP HELPERS
// ======================================================================

func generateTestToken(userID uuid.UUID, role string, secret string) string {
	claims := middleware.JWTClaims{
		UserID: userID,
		Email:  "customer@example.com",
		Role:   role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

type testFixture struct {
	router       *gin.Engine
	cartRepo     *MockCartRepository
	storeChecker *MockStoreChecker
	prodChecker  *MockProductChecker
	orderCreator *MockOrderCreator
	secret       string
}

func setupCartRouter() *testFixture {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	cartRepo := NewMockCartRepository()
	storeChecker := &MockStoreChecker{stores: make(map[uuid.UUID]*store.Store)}
	prodChecker := &MockProductChecker{products: make(map[uuid.UUID]*product.Product)}
	orderCreator := &MockOrderCreator{}

	service := cart.NewService(cartRepo, storeChecker, prodChecker, orderCreator)
	handler := cart.NewHandler(service)

	secret := "cart_test_secret_key"
	v1 := r.Group("/api/v1")
	handler.RegisterRoutes(v1, middleware.AuthMiddleware(secret))

	return &testFixture{
		router:       r,
		cartRepo:     cartRepo,
		storeChecker: storeChecker,
		prodChecker:  prodChecker,
		orderCreator: orderCreator,
		secret:       secret,
	}
}

// ======================================================================
// UNIT TESTS
// ======================================================================

func TestCart_AddItem_And_RecalculateTotals(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	// Seed store and product
	f.storeChecker.stores[storeID] = &store.Store{
		ID:   storeID,
		Name: "Burger House",
	}
	f.prodChecker.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "Cheeseburger",
		Price:   12000, // 120.00 TRY
		Stock:   10,
	}

	token := generateTestToken(customerID, "customer", f.secret)

	// 1. Add item first time (qty: 2)
	addReq := cart.AddToCartRequest{
		StoreID:   storeID,
		ProductID: productID,
		Quantity:  2,
	}
	body, _ := json.Marshal(addReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data cart.Cart `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res.Data.TotalAmount != 24000 {
		t.Errorf("expected total amount 24000, got %d", res.Data.TotalAmount)
	}
	if res.Data.ItemCount != 2 {
		t.Errorf("expected item count 2, got %d", res.Data.ItemCount)
	}
	if *res.Data.StoreID != storeID {
		t.Errorf("expected store ID %s, got %s", storeID, *res.Data.StoreID)
	}

	// 2. Add same item again (qty: 1) -> should increment quantity to 3
	addReq2 := cart.AddToCartRequest{
		StoreID:   storeID,
		ProductID: productID,
		Quantity:  1,
	}
	body2, _ := json.Marshal(addReq2)

	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Authorization", "Bearer "+token)

	w2 := httptest.NewRecorder()
	f.router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200, got %d: %s", w2.Code, w2.Body.String())
	}

	_ = json.Unmarshal(w2.Body.Bytes(), &res)
	if res.Data.ItemCount != 3 {
		t.Errorf("expected item count 3, got %d", res.Data.ItemCount)
	}
	if res.Data.TotalAmount != 36000 {
		t.Errorf("expected total amount 36000, got %d", res.Data.TotalAmount)
	}
}

func TestCart_SingleStoreConstraint_And_Force(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	storeA := uuid.New()
	storeB := uuid.New()
	prodA := uuid.New()
	prodB := uuid.New()

	// Seed 2 distinct stores and products
	f.storeChecker.stores[storeA] = &store.Store{ID: storeA, Name: "Store A"}
	f.storeChecker.stores[storeB] = &store.Store{ID: storeB, Name: "Store B"}

	f.prodChecker.products[prodA] = &product.Product{ID: prodA, StoreID: storeA, Name: "Pizza", Price: 15000, Stock: 10}
	f.prodChecker.products[prodB] = &product.Product{ID: prodB, StoreID: storeB, Name: "Sushi", Price: 20000, Stock: 10}

	token := generateTestToken(customerID, "customer", f.secret)

	// Step 1: Add item from Store A
	addA := cart.AddToCartRequest{StoreID: storeA, ProductID: prodA, Quantity: 1}
	bodyA, _ := json.Marshal(addA)
	reqA := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bodyA))
	reqA.Header.Set("Content-Type", "application/json")
	reqA.Header.Set("Authorization", "Bearer "+token)
	wA := httptest.NewRecorder()
	f.router.ServeHTTP(wA, reqA)
	if wA.Code != http.StatusOK {
		t.Fatalf("failed to add item from Store A: %s", wA.Body.String())
	}

	// Step 2: Try to add item from Store B without force -> 409 Conflict
	addB := cart.AddToCartRequest{StoreID: storeB, ProductID: prodB, Quantity: 1, Force: false}
	bodyB, _ := json.Marshal(addB)
	reqB := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bodyB))
	reqB.Header.Set("Content-Type", "application/json")
	reqB.Header.Set("Authorization", "Bearer "+token)
	wB := httptest.NewRecorder()
	f.router.ServeHTTP(wB, reqB)

	if wB.Code != http.StatusConflict {
		t.Fatalf("expected HTTP 409 Conflict on cross-store addition, got %d: %s", wB.Code, wB.Body.String())
	}

	// Step 3: Add item from Store B WITH Force: true -> should replace cart
	addBForce := cart.AddToCartRequest{StoreID: storeB, ProductID: prodB, Quantity: 2, Force: true}
	bodyBForce, _ := json.Marshal(addBForce)
	reqBForce := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(bodyBForce))
	reqBForce.Header.Set("Content-Type", "application/json")
	reqBForce.Header.Set("Authorization", "Bearer "+token)
	wBForce := httptest.NewRecorder()
	f.router.ServeHTTP(wBForce, reqBForce)

	if wBForce.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 with Force: true, got %d: %s", wBForce.Code, wBForce.Body.String())
	}

	var res struct {
		Data cart.Cart `json:"data"`
	}
	_ = json.Unmarshal(wBForce.Body.Bytes(), &res)

	if *res.Data.StoreID != storeB {
		t.Errorf("expected cart store to be Store B (%s), got %s", storeB, *res.Data.StoreID)
	}
	if len(res.Data.Items) != 1 || res.Data.Items[0].ProductID != prodB {
		t.Errorf("expected cart to contain only Store B's product")
	}
	if res.Data.ItemCount != 2 {
		t.Errorf("expected item count 2, got %d", res.Data.ItemCount)
	}
}

func TestCart_StockValidation(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	f.storeChecker.stores[storeID] = &store.Store{ID: storeID, Name: "Bakery"}
	f.prodChecker.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "Croissant",
		Price:   3500,
		Stock:   2, // Only 2 in stock!
	}

	token := generateTestToken(customerID, "customer", f.secret)

	// Attempt to add 5 items (stock is only 2)
	addReq := cart.AddToCartRequest{
		StoreID:   storeID,
		ProductID: productID,
		Quantity:  5,
	}
	body, _ := json.Marshal(addReq)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 Bad Request on stock overflow, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCart_UpdateQuantity_And_RemoveItem(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	f.storeChecker.stores[storeID] = &store.Store{ID: storeID, Name: "Taco Place"}
	f.prodChecker.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "Taco",
		Price:   5000,
		Stock:   10,
	}

	token := generateTestToken(customerID, "customer", f.secret)

	// 1. Add item with qty 2
	addReq := cart.AddToCartRequest{StoreID: storeID, ProductID: productID, Quantity: 2}
	body, _ := json.Marshal(addReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to add item: %s", w.Body.String())
	}

	// 2. Update item quantity to 4
	updateReq := cart.UpdateCartItemRequest{Quantity: 4}
	updateBody, _ := json.Marshal(updateReq)
	putReq := httptest.NewRequest(http.MethodPut, "/api/v1/cart/items/"+productID.String(), bytes.NewReader(updateBody))
	putReq.Header.Set("Content-Type", "application/json")
	putReq.Header.Set("Authorization", "Bearer "+token)
	putW := httptest.NewRecorder()
	f.router.ServeHTTP(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on quantity update, got %d: %s", putW.Code, putW.Body.String())
	}

	var res struct {
		Data cart.Cart `json:"data"`
	}
	_ = json.Unmarshal(putW.Body.Bytes(), &res)
	if res.Data.ItemCount != 4 || res.Data.TotalAmount != 20000 {
		t.Errorf("expected count 4 and total 20000, got count %d, total %d", res.Data.ItemCount, res.Data.TotalAmount)
	}

	// 3. Remove single item
	delReq := httptest.NewRequest(http.MethodDelete, "/api/v1/cart/items/"+productID.String(), nil)
	delReq.Header.Set("Authorization", "Bearer "+token)
	delW := httptest.NewRecorder()
	f.router.ServeHTTP(delW, delReq)

	if delW.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on item delete, got %d: %s", delW.Code, delW.Body.String())
	}

	_ = json.Unmarshal(delW.Body.Bytes(), &res)
	if len(res.Data.Items) != 0 || res.Data.TotalAmount != 0 {
		t.Errorf("expected empty cart after item removal")
	}

	// 4. Clear cart endpoint
	clearReq := httptest.NewRequest(http.MethodDelete, "/api/v1/cart", nil)
	clearReq.Header.Set("Authorization", "Bearer "+token)
	clearW := httptest.NewRecorder()
	f.router.ServeHTTP(clearW, clearReq)

	if clearW.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 on clear cart, got %d", clearW.Code)
	}
}

func TestCart_Checkout_Success_And_PurgeRedis(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	storeID := uuid.New()
	productID := uuid.New()

	f.storeChecker.stores[storeID] = &store.Store{ID: storeID, Name: "Noodle Bar"}
	f.prodChecker.products[productID] = &product.Product{
		ID:      productID,
		StoreID: storeID,
		Name:    "Ramen",
		Price:   18000,
		Stock:   10,
	}

	token := generateTestToken(customerID, "customer", f.secret)

	// 1. Add item to cart
	addReq := cart.AddToCartRequest{StoreID: storeID, ProductID: productID, Quantity: 2}
	body, _ := json.Marshal(addReq)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/cart/items", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	f.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("failed to add item: %s", w.Body.String())
	}

	// 2. Checkout
	checkoutReq := cart.CheckoutRequest{
		DeliveryAddress: "Ataturk Caddesi No: 42, Kadikoy, Istanbul",
	}
	checkoutBody, _ := json.Marshal(checkoutReq)
	cReq := httptest.NewRequest(http.MethodPost, "/api/v1/cart/checkout", bytes.NewReader(checkoutBody))
	cReq.Header.Set("Content-Type", "application/json")
	cReq.Header.Set("Authorization", "Bearer "+token)
	cReq.Header.Set("Idempotency-Key", "cart-checkout-test-key-1")

	cW := httptest.NewRecorder()
	f.router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusCreated {
		t.Fatalf("expected HTTP 201 Created on checkout, got %d: %s", cW.Code, cW.Body.String())
	}

	var res struct {
		Data order.Order `json:"data"`
	}
	_ = json.Unmarshal(cW.Body.Bytes(), &res)

	if res.Data.StoreID != storeID {
		t.Errorf("expected order store ID %s, got %s", storeID, res.Data.StoreID)
	}
	if res.Data.DeliveryAddress != "Ataturk Caddesi No: 42, Kadikoy, Istanbul" {
		t.Errorf("expected delivery address to match checkout request")
	}

	// 3. Verify Redis Cart was purged after checkout
	_, err := f.cartRepo.GetCart(context.Background(), customerID)
	if err != cart.ErrCartNotFound {
		t.Errorf("expected cart to be purged from Redis after checkout, got %v", err)
	}
}

func TestCart_Checkout_EmptyCart_ReturnsBadRequest(t *testing.T) {
	f := setupCartRouter()
	customerID := uuid.New()
	token := generateTestToken(customerID, "customer", f.secret)

	checkoutReq := cart.CheckoutRequest{
		DeliveryAddress: "Test Address",
	}
	checkoutBody, _ := json.Marshal(checkoutReq)
	cReq := httptest.NewRequest(http.MethodPost, "/api/v1/cart/checkout", bytes.NewReader(checkoutBody))
	cReq.Header.Set("Content-Type", "application/json")
	cReq.Header.Set("Authorization", "Bearer "+token)

	cW := httptest.NewRecorder()
	f.router.ServeHTTP(cW, cReq)

	if cW.Code != http.StatusBadRequest {
		t.Fatalf("expected HTTP 400 Bad Request for empty cart checkout, got %d: %s", cW.Code, cW.Body.String())
	}
}
