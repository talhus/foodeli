package main

import (
	_ "foodeli/docs"
	"foodeli/platform/response"
	"foodeli/services/auth"
	"foodeli/services/cart"
	"foodeli/services/notification"
	"foodeli/services/order"
	"foodeli/services/payment"
	"foodeli/services/product"
	"foodeli/services/store"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RateLimiters holds rate limiter middlewares for sensitive endpoints.
type RateLimiters struct {
	LoginLimiter    gin.HandlerFunc
	RegisterLimiter gin.HandlerFunc
	WebhookLimiter  gin.HandlerFunc
}

func NewRouter(
	authHandler *auth.Handler,
	storeHandler *store.Handler,
	productHandler *product.Handler,
	orderHandler *order.Handler,
	paymentHandler *payment.Handler,
	cartHandler *cart.Handler,
	notificationHandler *notification.Handler,
	authMiddleware gin.HandlerFunc,
	limiters RateLimiters,
) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(corsMiddleware)

	// Swagger documentation UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health Check Route
	r.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "foodeli-core-api",
		})
	})

	v1 := r.Group("/api/v1")

	authHandler.RegisterRoutes(v1, authMiddleware, limiters.LoginLimiter, limiters.RegisterLimiter)
	storeHandler.RegisterRoutes(v1, authMiddleware)
	productHandler.RegisterRoutes(v1, authMiddleware)
	orderHandler.RegisterRoutes(v1, authMiddleware)
	paymentHandler.RegisterRoutes(v1, authMiddleware, limiters.WebhookLimiter)
	cartHandler.RegisterRoutes(v1, authMiddleware)
	notificationHandler.RegisterRoutes(v1, authMiddleware)

	return r
}

// CORS middleware
func corsMiddleware(c *gin.Context) {
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, Idempotency-Key")
	c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, PATCH, DELETE")

	if c.Request.Method == "OPTIONS" {
		c.AbortWithStatus(http.StatusNoContent)
		return
	}
	c.Next()
}
