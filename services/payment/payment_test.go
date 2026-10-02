package payment_test

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
	"foodeli/platform/outbox"
	"foodeli/services/order"
	"foodeli/services/payment"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ======================================================================
// MOCKS
// ======================================================================

type MockOrderManager struct {
	mu     sync.Mutex
	orders map[uuid.UUID]*order.Order
}

func NewMockOrderManager() *MockOrderManager {
	return &MockOrderManager{orders: make(map[uuid.UUID]*order.Order)}
}

func (m *MockOrderManager) GetOrderById(ctx context.Context, requesterID uuid.UUID, role string, orderID uuid.UUID) (*order.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ord, exists := m.orders[orderID]
	if !exists {
		return nil, order.ErrOrderNotFound
	}
	return ord, nil
}

func (m *MockOrderManager) MarkOrderPaid(ctx context.Context, orderID uuid.UUID) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ord, exists := m.orders[orderID]
	if !exists {
		return order.ErrOrderNotFound
	}
	if ord.Status != order.StatusPending {
		return order.ErrInvalidOrderStatus
	}
	ord.Status = order.StatusConfirmed
	return nil
}

type MockPaymentRepository struct {
	mu       sync.Mutex
	payments map[uuid.UUID]*payment.Payment
	byOrder  map[uuid.UUID]*payment.Payment
}

func NewMockPaymentRepository() *MockPaymentRepository {
	return &MockPaymentRepository{
		payments: make(map[uuid.UUID]*payment.Payment),
		byOrder:  make(map[uuid.UUID]*payment.Payment),
	}
}

func (m *MockPaymentRepository) CreatePayment(ctx context.Context, p *payment.Payment) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p.CreatedAt = time.Now()
	p.UpdatedAt = time.Now()
	m.payments[p.ID] = p
	m.byOrder[p.OrderID] = p
	return nil
}

func (m *MockPaymentRepository) GetById(ctx context.Context, paymentID uuid.UUID) (*payment.Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.payments[paymentID]
	if !exists {
		return nil, payment.ErrPaymentNotFound
	}
	return p, nil
}

func (m *MockPaymentRepository) GetByOrderId(ctx context.Context, orderID uuid.UUID) (*payment.Payment, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.byOrder[orderID]
	if !exists {
		return nil, payment.ErrPaymentNotFound
	}
	return p, nil
}

func (m *MockPaymentRepository) UpdateStatus(ctx context.Context, paymentID uuid.UUID, status string, failureReason *string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.payments[paymentID]
	if !exists {
		return payment.ErrPaymentNotFound
	}
	p.Status = status
	p.FailureReason = failureReason
	p.UpdatedAt = time.Now()
	return nil
}

func (m *MockPaymentRepository) UpdateTransactionAndStatus(ctx context.Context, paymentID uuid.UUID, transactionID string, status string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, exists := m.payments[paymentID]
	if !exists {
		return payment.ErrPaymentNotFound
	}
	p.TransactionID = transactionID
	p.Status = status
	p.UpdatedAt = time.Now()
	return nil
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

func setupPaymentRouter() (*gin.Engine, *MockOrderManager, *MockPaymentRepository, *payment.MockGateway, string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockOrderManager := NewMockOrderManager()
	mockPaymentRepo := NewMockPaymentRepository()
	mockGateway := payment.NewMockGateway()

	service := payment.NewService(mockPaymentRepo, mockOrderManager, mockGateway, nil, nil, "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")
	handler := payment.NewHandler(service)

	jwtSecret := "payment_test_secret_key"
	v1 := r.Group("/api/v1")
	handler.RegisterRoutes(v1, middleware.AuthMiddleware(jwtSecret))

	return r, mockOrderManager, mockPaymentRepo, mockGateway, jwtSecret
}

// ======================================================================
// UNIT TESTS: PAYMENT PROCESSING & LIFECYCLE
// ======================================================================

func TestPaymentProcessing_Success(t *testing.T) {
	r, mockOrders, mockRepo, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	orderID := uuid.New()

	// Seed pending order
	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 25000, // 250.00 TRY/EUR in cents
		Status:      order.StatusPending,
	}

	reqBody := payment.ProcessPaymentRequest{
		OrderID:        orderID,
		PaymentMethod:  payment.MethodCreditCard,
		CardNumber:     "5555123456789010",
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "2028",
		CVV:            "123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	token := generateTestToken(customerID, "customer", secret)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var res struct {
		Data payment.Payment `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if res.Data.Status != payment.StatusSuccess {
		t.Errorf("expected payment status SUCCESS, got %s", res.Data.Status)
	}
	if res.Data.Amount != 25000 {
		t.Errorf("expected amount 25000, got %d", res.Data.Amount)
	}
	if res.Data.TransactionID == "" {
		t.Errorf("expected non-empty transaction ID")
	}

	// Verify order status transitioned to CONFIRMED
	if mockOrders.orders[orderID].Status != order.StatusConfirmed {
		t.Errorf("expected order status to be CONFIRMED, got %s", mockOrders.orders[orderID].Status)
	}

	// Verify payment persisted in repository
	persisted, err := mockRepo.GetByOrderId(context.Background(), orderID)
	if err != nil || persisted.Status != payment.StatusSuccess {
		t.Errorf("expected payment to be persisted with SUCCESS status")
	}
}

func TestPaymentProcessing_Declined_402(t *testing.T) {
	r, mockOrders, _, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 15000,
		Status:      order.StatusPending,
	}

	// Test 1: Insufficient funds (Card starting with 4000)
	t.Run("CardDeclined_InsufficientFunds", func(t *testing.T) {
		reqBody := payment.ProcessPaymentRequest{
			OrderID:        orderID,
			PaymentMethod:  payment.MethodCreditCard,
			CardNumber:     "4000123456789010", // starts with 4000
			CardHolderName: "Jane Doe",
			ExpireMonth:    "10",
			ExpireYear:     "2027",
			CVV:            "456",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("expected HTTP 402 Payment Required, got %d: %s", w.Code, w.Body.String())
		}

		// Verify order is still PENDING
		if mockOrders.orders[orderID].Status != order.StatusPending {
			t.Errorf("declined payment should leave order in PENDING status, got %s", mockOrders.orders[orderID].Status)
		}
	})

	// Test 2: Invalid CVV code (000)
	t.Run("CardDeclined_InvalidCVV", func(t *testing.T) {
		reqBody := payment.ProcessPaymentRequest{
			OrderID:        orderID,
			PaymentMethod:  payment.MethodCreditCard,
			CardNumber:     "5555123456789010",
			CardHolderName: "Jane Doe",
			ExpireMonth:    "10",
			ExpireYear:     "2027",
			CVV:            "000", // invalid CVV
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusPaymentRequired {
			t.Fatalf("expected HTTP 402 Payment Required, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestPaymentProcessing_ValidationErrors(t *testing.T) {
	r, mockOrders, _, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	otherCustomerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 10000,
		Status:      order.StatusPending,
	}

	t.Run("NonExistentOrder", func(t *testing.T) {
		reqBody := payment.ProcessPaymentRequest{
			OrderID:        uuid.New(), // doesn't exist
			PaymentMethod:  payment.MethodCreditCard,
			CardNumber:     "5555123456789010",
			CardHolderName: "John Doe",
			ExpireMonth:    "12",
			ExpireYear:     "2028",
			CVV:            "123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected HTTP 404 Not Found, got %d", w.Code)
		}
	})

	t.Run("UnauthorizedUserCannotPayOthersOrder", func(t *testing.T) {
		reqBody := payment.ProcessPaymentRequest{
			OrderID:        orderID, // belongs to customerID
			PaymentMethod:  payment.MethodCreditCard,
			CardNumber:     "5555123456789010",
			CardHolderName: "Imposter",
			ExpireMonth:    "12",
			ExpireYear:     "2028",
			CVV:            "123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		// Authenticated as otherCustomerID
		token := generateTestToken(otherCustomerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected HTTP 403 Forbidden, got %d", w.Code)
		}
	})

	t.Run("OrderNotPayable_WhenAlreadyConfirmedOrCancelled", func(t *testing.T) {
		cancelledOrderID := uuid.New()
		mockOrders.orders[cancelledOrderID] = &order.Order{
			ID:          cancelledOrderID,
			UserID:      customerID,
			TotalAmount: 5000,
			Status:      order.StatusCancelled,
		}

		reqBody := payment.ProcessPaymentRequest{
			OrderID:        cancelledOrderID,
			PaymentMethod:  payment.MethodCreditCard,
			CardNumber:     "5555123456789010",
			CardHolderName: "John Doe",
			ExpireMonth:    "12",
			ExpireYear:     "2028",
			CVV:            "123",
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected HTTP 400 Bad Request, got %d", w.Code)
		}
	})
}

func TestPaymentProcessing_Idempotency(t *testing.T) {
	r, mockOrders, mockRepo, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 30000,
		Status:      order.StatusPending,
	}

	// Pre-seed an existing successful payment for this order
	existingPayment := &payment.Payment{
		ID:            uuid.New(),
		OrderID:       orderID,
		UserID:        customerID,
		Amount:        30000,
		Currency:      "EUR",
		Provider:      payment.ProviderMock,
		PaymentMethod: payment.MethodCreditCard,
		TransactionID: "txn_mock_existing_12345",
		Status:        payment.StatusSuccess,
	}
	_ = mockRepo.CreatePayment(context.Background(), existingPayment)

	reqBody := payment.ProcessPaymentRequest{
		OrderID:        orderID,
		PaymentMethod:  payment.MethodCreditCard,
		CardNumber:     "5555123456789010",
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "2028",
		CVV:            "123",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	token := generateTestToken(customerID, "customer", secret)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/process", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected HTTP 200 OK, got %d", w.Code)
	}

	var res struct {
		Data payment.Payment `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	// Idempotency: should return the exact existing transaction ID without double charging
	if res.Data.TransactionID != "txn_mock_existing_12345" {
		t.Errorf("expected existing transaction ID 'txn_mock_existing_12345', got %s", res.Data.TransactionID)
	}
}

func TestGetPaymentByOrderID(t *testing.T) {
	r, _, mockRepo, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	otherCustomerID := uuid.New()
	adminID := uuid.New()
	orderID := uuid.New()

	existingPayment := &payment.Payment{
		ID:            uuid.New(),
		OrderID:       orderID,
		UserID:        customerID,
		Amount:        18500,
		Currency:      "EUR",
		Provider:      payment.ProviderMock,
		PaymentMethod: payment.MethodCreditCard,
		TransactionID: "txn_mock_receipt_999",
		Status:        payment.StatusSuccess,
	}
	_ = mockRepo.CreatePayment(context.Background(), existingPayment)

	t.Run("CustomerCanViewOwnPayment", func(t *testing.T) {
		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/order/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("OtherCustomerCannotViewPayment", func(t *testing.T) {
		token := generateTestToken(otherCustomerID, "customer", secret)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/order/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Fatalf("expected HTTP 403 Forbidden, got %d", w.Code)
		}
	})

	t.Run("AdminCanViewAnyPayment", func(t *testing.T) {
		token := generateTestToken(adminID, "admin", secret)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/order/"+orderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("NonExistentPaymentReturns404", func(t *testing.T) {
		randomOrderID := uuid.New()
		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodGet, "/api/v1/payments/order/"+randomOrderID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected HTTP 404 Not Found, got %d", w.Code)
		}
	})
}

// ======================================================================
// UNIT TESTS: TYRFINTECH GATEWAY CHARGE
// ======================================================================

func TestTyrFintechGateway_Charge(t *testing.T) {
	t.Run("Success_200", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost || r.URL.Path != "/api/v1/charges" {
				http.Error(w, "not found", http.StatusNotFound)
				return
			}
			if r.Header.Get("Authorization") != "Bearer test_key_123" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			if r.Header.Get("Idempotency-Key") == "" {
				http.Error(w, "missing idempotency key", http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":        true,
				"transaction_id": "trx_test_999",
				"status":         "SUCCEEDED",
				"failure_reason": "",
			})
		}))
		defer server.Close()

		gw := payment.NewTyrFintechGateway(server.URL, "test_key_123", "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")
		res, err := gw.Charge(context.Background(), &payment.ChargeRequest{
			Amount:         10000,
			Currency:       "EUR",
			CardNumber:     "4111111111110000",
			CardHolderName: "John Doe",
			ExpireMonth:    "12",
			ExpireYear:     "28",
			CVV:            "123",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !res.Success || res.TransactionID != "trx_test_999" {
			t.Fatalf("expected success with trx_test_999, got: %+v", res)
		}
	})

	t.Run("Declined_402", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusPaymentRequired)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"success":        false,
				"transaction_id": "",
				"status":         "FAILED",
				"failure_reason": "Insufficient funds",
			})
		}))
		defer server.Close()

		gw := payment.NewTyrFintechGateway(server.URL, "test_key_123", "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")
		res, err := gw.Charge(context.Background(), &payment.ChargeRequest{
			Amount:     10000,
			Currency:   "EUR",
			CardNumber: "4111111111115001",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Success || res.FailureReason != "Insufficient funds" {
			t.Fatalf("expected decline with 'Insufficient funds', got: %+v", res)
		}
	})

	t.Run("Unauthorized_401", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		gw := payment.NewTyrFintechGateway(server.URL, "invalid_key", "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")
		_, err := gw.Charge(context.Background(), &payment.ChargeRequest{
			Amount: 10000,
		})
		if err == nil {
			t.Fatalf("expected error on 401 unauthorized, got nil")
		}
	})
}

// ======================================================================
// UNIT TESTS: HOSTED CHECKOUT SESSION (FLOW 2A)
// ======================================================================

func TestPaymentService_CreateCheckoutSession(t *testing.T) {
	r, mockOrders, _, _, secret := setupPaymentRouter()

	customerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 15000, // 150.00 EUR
		Status:      order.StatusPending,
	}

	t.Run("Success_CreatesSession", func(t *testing.T) {
		reqBody := payment.CreateCheckoutSessionRequest{
			OrderID: orderID,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(customerID, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout-session", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var res struct {
			Data payment.CheckoutSessionResponse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if res.Data.SessionID == "" {
			t.Errorf("expected session_id to be populated")
		}
		if res.Data.CheckoutURL == "" {
			t.Errorf("expected checkout_url to be populated")
		}
		if res.Data.Status != "PENDING" {
			t.Errorf("expected status PENDING, got %s", res.Data.Status)
		}
	})

	t.Run("UnauthorizedOrderAccess", func(t *testing.T) {
		otherUser := uuid.New()
		reqBody := payment.CreateCheckoutSessionRequest{
			OrderID: orderID,
		}
		bodyBytes, _ := json.Marshal(reqBody)

		token := generateTestToken(otherUser, "customer", secret)
		req := httptest.NewRequest(http.MethodPost, "/api/v1/payments/checkout-session", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected HTTP 403 Forbidden, got %d", w.Code)
		}
	})
}

func TestTyrFintechGateway_CreateCheckoutSession(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/checkout/sessions" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer live_key_999" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var req map[string]any
		_ = json.NewDecoder(r.Body).Decode(&req)
		if req["amount"] == nil || req["order_id"] == nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"session_id":   "cs_live_12345",
			"checkout_url": "http://localhost:3000/checkout?session_id=cs_live_12345",
			"status":       "PENDING",
			"expires_at":   "2026-09-30T00:00:00Z",
		})
	}))
	defer server.Close()

	gw := payment.NewTyrFintechGateway(server.URL, "live_key_999", "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")
	resp, err := gw.CreateCheckoutSession(context.Background(), &payment.GatewayCheckoutSessionRequest{
		OrderID:  "order_uuid_123",
		Amount:   15000,
		Currency: "EUR",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.SessionID != "cs_live_12345" {
		t.Errorf("expected session_id 'cs_live_12345', got %s", resp.SessionID)
	}
	if resp.CheckoutURL != "http://localhost:3000/checkout?session_id=cs_live_12345" {
		t.Errorf("unexpected checkout_url: %s", resp.CheckoutURL)
	}
}

// ======================================================================
// UNIT TESTS: ASYNCHRONOUS WEBHOOK RECEIVER (FLOW 2C)
// ======================================================================

func TestPaymentHandler_TyrFintechWebhook(t *testing.T) {
	r, mockOrders, mockRepo, _, _ := setupPaymentRouter()

	customerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 15000,
		Status:      order.StatusPending,
	}

	t.Run("Success_FirstDeliverySettlesOrder", func(t *testing.T) {
		webhookPayload := payment.TyrFintechWebhookPayload{
			Event:         "checkout.session.paid",
			TransactionID: "trx_webhook_12345",
			SessionID:     "cs_live_12345",
			OrderID:       orderID.String(),
			Amount:        15000,
			Currency:      "EUR",
			Status:        "SUCCEEDED",
			Timestamp:     time.Now().Unix(),
		}
		bodyBytes, _ := json.Marshal(webhookPayload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/tyrfintech", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 OK, got %d: %s", w.Code, w.Body.String())
		}

		var res map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if res["received"] != true {
			t.Errorf("expected received: true in response body")
		}

		// Verify order is now CONFIRMED
		if mockOrders.orders[orderID].Status != order.StatusConfirmed {
			t.Errorf("expected order status to be CONFIRMED, got %s", mockOrders.orders[orderID].Status)
		}

		// Verify payment record persisted with SUCCESS
		persisted, err := mockRepo.GetByOrderId(context.Background(), orderID)
		if err != nil || persisted.Status != payment.StatusSuccess {
			t.Errorf("expected payment to be persisted with SUCCESS status")
		}
		if persisted.TransactionID != "trx_webhook_12345" {
			t.Errorf("expected transaction ID 'trx_webhook_12345', got %s", persisted.TransactionID)
		}
	})

	t.Run("Success_IdempotentDuplicateDelivery", func(t *testing.T) {
		// Send the duplicate webhook
		webhookPayload := payment.TyrFintechWebhookPayload{
			Event:         "checkout.session.paid",
			TransactionID: "trx_webhook_12345",
			SessionID:     "cs_live_12345",
			OrderID:       orderID.String(),
			Amount:        15000,
			Currency:      "EUR",
			Status:        "SUCCEEDED",
			Timestamp:     time.Now().Unix(),
		}
		bodyBytes, _ := json.Marshal(webhookPayload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/tyrfintech", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected HTTP 200 OK on duplicate, got %d: %s", w.Code, w.Body.String())
		}

		var res map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &res)
		if res["received"] != true {
			t.Errorf("expected received: true in response body")
		}
	})

	t.Run("NotFound_InvalidOrderID", func(t *testing.T) {
		webhookPayload := payment.TyrFintechWebhookPayload{
			Event:         "checkout.session.paid",
			TransactionID: "trx_unknown",
			OrderID:       uuid.New().String(),
			Amount:        10000,
			Status:        "SUCCEEDED",
		}
		bodyBytes, _ := json.Marshal(webhookPayload)

		req := httptest.NewRequest(http.MethodPost, "/api/v1/webhooks/tyrfintech", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected HTTP 404 for non-existent order, got %d", w.Code)
		}
	})
}

func TestPaymentService_OutboxEnqueue(t *testing.T) {
	mockOrders := NewMockOrderManager()
	mockRepo := NewMockPaymentRepository()
	mockGW := payment.NewMockGateway()
	mockOutbox := outbox.NewMockRepository()

	service := payment.NewService(mockRepo, mockOrders, mockGW, nil, mockOutbox, "http://localhost:3000/orders/callback", "http://localhost:8080/api/v1/webhooks/tyrfintech")

	customerID := uuid.New()
	orderID := uuid.New()

	mockOrders.orders[orderID] = &order.Order{
		ID:          orderID,
		UserID:      customerID,
		TotalAmount: 20000,
		Status:      order.StatusPending,
	}

	req := &payment.ProcessPaymentRequest{
		OrderID:        orderID,
		PaymentMethod:  payment.MethodCreditCard,
		CardNumber:     "5555123456789010",
		CardHolderName: "John Doe",
		ExpireMonth:    "12",
		ExpireYear:     "2028",
		CVV:            "123",
	}

	ctx := context.Background()
	p, err := service.ProcessPayment(ctx, customerID, req)
	if err != nil {
		t.Fatalf("unexpected payment processing error: %v", err)
	}

	if p.Status != payment.StatusSuccess {
		t.Fatalf("expected payment status SUCCESS, got %s", p.Status)
	}

	// Verify event was enqueued into outbox
	pendingEvents, err := mockOutbox.FetchPending(ctx, 10)
	if err != nil {
		t.Fatalf("failed to fetch outbox events: %v", err)
	}
	if len(pendingEvents) != 1 {
		t.Fatalf("expected 1 outbox event enqueued, got %d", len(pendingEvents))
	}

	ev := pendingEvents[0]
	if ev.AggregateType != outbox.AggregatePayment || ev.AggregateID != p.ID {
		t.Errorf("unexpected outbox event: %+v", ev)
	}
}




