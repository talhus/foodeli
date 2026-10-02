package middleware

import (
	"errors"
	"strings"

	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	CtxUserIDKey   = "current_user_id"
	CtxUserRoleKey = "current_user_role"
	CtxUserEmail   = "current_user_email"
)

// JWTClaims defines the standard claims stored inside the JWT token.
type JWTClaims struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT Bearer tokens and injects user details into the request context.
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			cookieToken, err := c.Cookie("accessToken")
			if err != nil || cookieToken == "" {
				queryToken := c.Query("token")
				if queryToken == "" {
					response.Unauthorized(c, "Authorization header is required")
					c.Abort()
					return
				}
				tokenString = queryToken
			} else {
				tokenString = cookieToken
			}
		} else {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				response.Unauthorized(c, "Authorization format must be 'Bearer <token>'")
				c.Abort()
				return
			}
			tokenString = parts[1]
		}

		claims := &JWTClaims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Inject into Gin Context
		c.Set(CtxUserIDKey, claims.UserID)
		c.Set(CtxUserEmail, claims.Email)
		c.Set(CtxUserRoleKey, claims.Role)

		c.Next()
	}
}

// GetUserID retrieves the authenticated user's ID from the Gin context.
func GetUserID(c *gin.Context) (uuid.UUID, bool) {
	val, exists := c.Get(CtxUserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := val.(uuid.UUID)
	return id, ok
}

// GetUserRole retrieves the authenticated user's role from the Gin context.
func GetUserRole(c *gin.Context) (string, bool) {
	val, exists := c.Get(CtxUserRoleKey)
	if !exists {
		return "", false
	}
	role, ok := val.(string)
	return role, ok
}
