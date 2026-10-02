package store_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"foodeli/platform/middleware"
	"foodeli/services/store"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// MockRepository implements store.Repository
type MockRepository struct {
	stores    map[uuid.UUID]*store.Store
	menuItems map[uuid.UUID]*store.MenuItem
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		stores:    make(map[uuid.UUID]*store.Store),
		menuItems: make(map[uuid.UUID]*store.MenuItem),
	}
}

func (m *MockRepository) CreateStore(ctx context.Context, s *store.Store) error {
	for _, existing := range m.stores {
		if existing.MerchantID == s.MerchantID {
			return store.ErrMerchantStoreAlreadyExists
		}
	}
	s.ID = uuid.New()
	m.stores[s.ID] = s
	return nil
}

func (m *MockRepository) GetStoreById(ctx context.Context, id uuid.UUID) (*store.Store, error) {
	s, exists := m.stores[id]
	if !exists {
		return nil, store.ErrStoreNotFound
	}
	return s, nil
}

func (m *MockRepository) ListStores(ctx context.Context) ([]store.Store, error) {
	var list []store.Store
	for _, s := range m.stores {
		list = append(list, *s)
	}
	return list, nil
}

func (m *MockRepository) GetStoreByMerchantID(ctx context.Context, merchantID uuid.UUID) (*store.Store, error) {
	for _, s := range m.stores {
		if s.MerchantID == merchantID {
			return s, nil
		}
	}
	return nil, store.ErrStoreNotFound
}

func (m *MockRepository) CreateMenuItem(ctx context.Context, item *store.MenuItem) error {
	item.ID = uuid.New()
	m.menuItems[item.ID] = item
	return nil
}

func (m *MockRepository) GetMenuItemsByStoreID(ctx context.Context, storeID uuid.UUID) ([]store.MenuItem, error) {
	var list []store.MenuItem
	for _, item := range m.menuItems {
		if item.StoreID == storeID {
			list = append(list, *item)
		}
	}
	return list, nil
}

func (m *MockRepository) GetMenuItemById(ctx context.Context, id uuid.UUID) (*store.MenuItem, error) {
	item, exists := m.menuItems[id]
	if !exists {
		return nil, store.ErrMenuItemNotFound
	}
	return item, nil
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

func setupStoreRouter() (*gin.Engine, *MockRepository, string) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	mockRepo := NewMockRepository()
	service := store.NewService(mockRepo)
	handler := store.NewHandler(service)

	jwtSecret := "store_test_secret"
	v1 := r.Group("/api/v1")
	handler.RegisterRoutes(v1, middleware.AuthMiddleware(jwtSecret))

	return r, mockRepo, jwtSecret
}

func TestStoreAndMenuFlow(t *testing.T) {
	r, _, secret := setupStoreRouter()
	merchantID := uuid.New()
	customerID := uuid.New()

	merchantToken := generateTestToken(merchantID, "merchant", secret)
	customerToken := generateTestToken(customerID, "customer", secret)

	// 1. Customer tries to create a store -> Should be Forbidden (403)
	storePayload := store.CreateStoreRequest{
		Name:    "Wolt Pizza Hub",
		Address: "Mannerheimintie 10, Helsinki",
	}
	body, _ := json.Marshal(storePayload)
	req1 := httptest.NewRequest(http.MethodPost, "/api/v1/stores", bytes.NewBuffer(body))
	req1.Header.Set("Authorization", "Bearer "+customerToken)
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	if w1.Code != http.StatusForbidden {
		t.Fatalf("expected status 403 Forbidden for customer creating store, got %d", w1.Code)
	}

	// 2. Merchant creates a store -> Should succeed (201)
	req2 := httptest.NewRequest(http.MethodPost, "/api/v1/stores", bytes.NewBuffer(body))
	req2.Header.Set("Authorization", "Bearer "+merchantToken)
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created for merchant creating store, got %d. Body: %s", w2.Code, w2.Body.String())
	}

	var storeResp struct {
		Data store.Store `json:"data"`
	}
	_ = json.Unmarshal(w2.Body.Bytes(), &storeResp)
	storeID := storeResp.Data.ID

	// 2b. Merchant tries to create a 2nd store -> Should fail with 409 Conflict
	secondStorePayload := store.CreateStoreRequest{
		Name:    "Second Store Attempt",
		Address: "Another Street 2, Helsinki",
	}
	secondBody, _ := json.Marshal(secondStorePayload)
	req2b := httptest.NewRequest(http.MethodPost, "/api/v1/stores", bytes.NewBuffer(secondBody))
	req2b.Header.Set("Authorization", "Bearer "+merchantToken)
	req2b.Header.Set("Content-Type", "application/json")
	w2b := httptest.NewRecorder()
	r.ServeHTTP(w2b, req2b)

	if w2b.Code != http.StatusConflict {
		t.Fatalf("expected status 409 Conflict for creating 2nd store, got %d", w2b.Code)
	}

	// 3. Merchant adds a menu item -> Should succeed (201)
	itemPayload := store.CreateMenuItemRequest{
		Name:        "Margherita Pizza",
		Description: "San Marzano tomatoes, fresh mozzarella, basil",
		Price:       12.50,
	}
	itemBody, _ := json.Marshal(itemPayload)
	req3 := httptest.NewRequest(http.MethodPost, "/api/v1/stores/"+storeID.String()+"/menu", bytes.NewBuffer(itemBody))
	req3.Header.Set("Authorization", "Bearer "+merchantToken)
	req3.Header.Set("Content-Type", "application/json")
	w3 := httptest.NewRecorder()
	r.ServeHTTP(w3, req3)

	if w3.Code != http.StatusCreated {
		t.Fatalf("expected status 201 Created for adding menu item, got %d", w3.Code)
	}

	// 4. Public lists menu items -> Should return item (200)
	req4 := httptest.NewRequest(http.MethodGet, "/api/v1/stores/"+storeID.String()+"/menu", nil)
	w4 := httptest.NewRecorder()
	r.ServeHTTP(w4, req4)

	if w4.Code != http.StatusOK {
		t.Fatalf("expected status 200 OK for public menu listing, got %d", w4.Code)
	}
}
