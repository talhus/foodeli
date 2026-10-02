package auth

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Role constants defining available access levels in Foodeli.
const (
	RoleCustomer = "customer"
	RoleMerchant = "merchant"
	RoleAdmin    = "admin"
)

// User represents a registered user entity in the database.
type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// RegisterRequest defines the payload for creating a new user or merchant account.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role" binding:"required"`
}

// Validate checks business rules for the registration payload.
func (r *RegisterRequest) Validate() error {
	r.Email = strings.TrimSpace(strings.ToLower(r.Email))
	r.Role = strings.TrimSpace(strings.ToLower(r.Role))

	// Admins cannot be registered publicly; only customer or merchant
	if r.Role != RoleCustomer && r.Role != RoleMerchant {
		return errors.New("role must be either 'customer' or 'merchant'")
	}

	if len(r.Password) < 6 {
		return errors.New("password must be at least 6 characters long")
	}

	return nil
}

// LoginRequest defines the credentials payload for authentication.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse represents the sanitized user data returned to clients.
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

// AuthResponse defines the authentication response with JWT token.
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// ToResponse maps a User entity to a sanitized UserResponse.
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
}
