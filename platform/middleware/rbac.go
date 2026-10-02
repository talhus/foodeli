package middleware

import (
	"foodeli/platform/response"

	"github.com/gin-gonic/gin"
)

// RequireRole checks if the authenticated user has one of the allowed roles.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := GetUserRole(c)
		if !ok {
			response.Unauthorized(c, "Authentication required")
			c.Abort()
			return
		}

		for _, allowed := range allowedRoles {
			if role == allowed {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "You do not have permission to access this resource")
		c.Abort()
	}
}
