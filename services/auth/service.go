package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"foodeli/platform/middleware"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
)

// Service defines business operations for authentication and user profiles.
type Service interface {
	Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error)
	Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error)
	GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error)
}

// AuthService implements Service.
type AuthService struct {
	repo               Repository
	jwtSecret          string
	jwtExpirationHours int
}

// NewService creates a new instance of AuthService.
func NewService(repo Repository, jwtSecret string, jwtExpirationHours int) *AuthService {
	return &AuthService{
		repo:               repo,
		jwtSecret:          jwtSecret,
		jwtExpirationHours: jwtExpirationHours,
	}
}

// Register
func (s *AuthService) Register(ctx context.Context, req *RegisterRequest) (*AuthResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Role:         req.Role,
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// Login
func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	return &AuthResponse{
		Token: token,
		User:  user.ToResponse(),
	}, nil
}

// GetProfile
func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.FindById(ctx, userID)
	if err != nil {
		return nil, err
	}

	resp := user.ToResponse()
	return &resp, nil
}

func (s *AuthService) generateToken(user *User) (string, error) {
	expiration := time.Duration(s.jwtExpirationHours) * time.Hour
	claims := middleware.JWTClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "foodeli-core-api",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}
