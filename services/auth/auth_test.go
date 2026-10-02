package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"foodeli/platform/middleware"
	"foodeli/services/auth"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// MockRepository implements auth.Repository for isolated unit tests.
type MockRepository struct {
	users map[string]*auth.User
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users: make(map[string]*auth.User),
	}
}

func (m *MockRepository) CreateUser(ctx context.Context, user *auth.User) error {
	if _, exists := m.users[user.Email]; exists {
		return auth.ErrUserAlreadyExists
	}
	user.ID = uuid.New()
	m.users[user.Email] = user
	return nil
}

func (m *MockRepository) FindByEmail(ctx context.Context, email string) (*auth.User, error) {
	u, exists := m.users[email]
	if !exists {
		return nil, auth.ErrUserNotFound
	}
	return u, nil
}

func (m *MockRepository) FindById(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, auth.ErrUserNotFound
}

func setupTestRouter() (*gin.Engine, *MockRepository) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockRepo := NewMockRepository()
	jwtSecret := "test_secret_key"
	authService := auth.NewService(mockRepo, jwtSecret, 24)
	authHandler := auth.NewHandler(authService)

	v1 := r.Group("/api/v1")
	authHandler.RegisterRoutes(v1, middleware.AuthMiddleware(jwtSecret))

	return r, mockRepo
}

func TestRegisterAndLogin(t *testing.T) {
	r, _ := setupTestRouter()

	// 1. Register Customer
	regPayload := auth.RegisterRequest{
		Email:    "test@foodeli.fi",
		Password: "password123",
		Role:     "customer",
	}
	body, _ := json.Marshal(regPayload)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
	}

	// 2. Duplicate Register
	wDup := httptest.NewRecorder()
	reqDup := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	reqDup.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wDup, reqDup)

	if wDup.Code != http.StatusConflict {
		t.Fatalf("expected duplicate to return 409, got %d", wDup.Code)
	}

	// 3. Login
	loginPayload := auth.LoginRequest{
		Email:    "test@foodeli.fi",
		Password: "password123",
	}
	loginBody, _ := json.Marshal(loginPayload)
	reqLogin := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(loginBody))
	reqLogin.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	r.ServeHTTP(wLogin, reqLogin)

	if wLogin.Code != http.StatusOK {
		t.Fatalf("expected login status 200, got %d", wLogin.Code)
	}

	var loginResp struct {
		Success bool `json:"success"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wLogin.Body.Bytes(), &loginResp)

	if loginResp.Data.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	// 4. Authenticated /me
	reqMe := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	reqMe.Header.Set("Authorization", "Bearer "+loginResp.Data.Token)
	wMe := httptest.NewRecorder()
	r.ServeHTTP(wMe, reqMe)

	if wMe.Code != http.StatusOK {
		t.Fatalf("expected /me status 200, got %d", wMe.Code)
	}
}
