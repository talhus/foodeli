package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response envelope.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error structure.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Success writes a successful JSON response.
func Success(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response{
		Success: true,
		Data:    data,
	})
}

// Error writes an error JSON response.
func Error(c *gin.Context, statusCode int, code, message string) {
	c.JSON(statusCode, Response{
		Success: false,
		Error: &APIError{
			Code:    code,
			Message: message,
		},
	})
}

// BadRequest writes a 400 Bad Request response.
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized writes a 401 Unauthorized response.
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden writes a 403 Forbidden response.
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound writes a 404 Not Found response.
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict writes a 409 Conflict response.
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message)
}

// InternalServerError writes a 500 Internal Server Error response.
func InternalServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", message)
}

// TooManyRequests writes a 429 Too Many Requests response.
func TooManyRequests(c *gin.Context, message string) {
	Error(c, http.StatusTooManyRequests, "TOO_MANY_REQUESTS", message)
}

