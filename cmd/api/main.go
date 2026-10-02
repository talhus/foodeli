package main

import (
	"context"
	"log"

	"foodeli/platform/config"
	"foodeli/services/notification"

	"github.com/gin-gonic/gin"
)

// @title Foodeli Core API
// @version 1.0
// @description High-performance Modular Monolith backend for food & marketplace delivery platform.
// @termsOfService http://swagger.io/terms/

// @contact.name Foodeli Engineering
// @contact.url https://github.com/iamtbay/foodeli

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token. Example: "Bearer eyJhbGciOi..."

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration %v", err)
	}

	if cfg.ServerEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	resources, cleanup, err := initInfrastructure(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize infrastructure %v", err)
	}
	defer cleanup()

	notifHub := notification.NewHub()

	router := setupRouter(cfg, resources, notifHub)

	// Start consumers in background
	if err := startConsumers(ctx, resources, notifHub); err != nil {
		log.Fatalf("Failed to start consumers: %v", err)
	}
	log.Println("Event consumers started")

	runServer(cfg, router)

}
