package main

import (
	"context"
	"errors"
	"fmt"
	"foodeli/platform/config"
	"foodeli/platform/database"
	"foodeli/platform/middleware"
	"foodeli/platform/outbox"
	"foodeli/platform/rabbitmq"
	platformRedis "foodeli/platform/redis"
	"foodeli/services/auth"
	"foodeli/services/cart"
	"foodeli/services/notification"
	"foodeli/services/order"
	"foodeli/services/payment"
	"foodeli/services/product"
	"foodeli/services/store"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type Resources struct {
	DB       *pgxpool.Pool
	Redis    *redis.Client
	RabbitMQ *rabbitmq.Client
}

func initInfrastructure(ctx context.Context, cfg *config.Config) (*Resources, func(), error) {
	//postgre
	log.Println("Connecting to PostgreSQL...")
	dbPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL connection pool: %v", err)
	}
	log.Println("Successfully connected to PostgreSQL")

	//redis
	log.Println("Connecting to Redis...")
	redisClient, err := platformRedis.NewRedisClient(ctx, cfg.RedisHost, cfg.RedisPort, cfg.RedisPass, cfg.RedisDB)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	log.Println("Successfully connected to Redis")

	//rabbitmq
	log.Println("Connecting to RabbitMQ...")
	rabbitClient, err := rabbitmq.NewRabbitMQClient(cfg.RabbitMQURL, cfg.RabbitMQExchange)
	if err != nil {
		log.Fatalf("Failed to initialize RabbitMQ client: %v", err)
	}
	log.Println("Successfully connected to RabbitMQ")

	//
	res := &Resources{
		DB:       dbPool,
		Redis:    redisClient,
		RabbitMQ: rabbitClient,
	}

	//cleanup
	cleanup := func() {
		log.Println("Cleaning up resources...")
		dbPool.Close()
		redisClient.Close()
		rabbitClient.Close()
		log.Println("All infrastructure resources closed cleanly")
	}

	return res, cleanup, nil
}

func setupRouter(cfg *config.Config, res *Resources, notifHub *notification.Hub) *gin.Engine {
	// 4. Initialize Auth Feature Package
	authRepo := auth.NewRepository(res.DB)
	authService := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpirationHours)
	authHandler := auth.NewHandler(authService)

	// 5. Initialize Store & Menu Feature Package
	storeRepo := store.NewRepository(res.DB)
	storeService := store.NewService(storeRepo)
	storeHandler := store.NewHandler(storeService)

	productRepo := product.NewRepository(res.DB)
	productService := product.NewService(productRepo, storeService)
	productHandler := product.NewHandler(productService)

	outboxRepo := outbox.NewPostgresRepository(res.DB)

	orderRepo := order.NewRepository(res.DB, outboxRepo)
	orderService := order.NewService(storeService, productService, orderRepo, res.RabbitMQ)
	orderHandler := order.NewHandler(orderService)

	// 6. Initialize Payment Feature Package
	paymentRepo := payment.NewPostgresRepository(res.DB)
	mockGateway := payment.NewMockGateway()
	tyrGateway := payment.NewTyrFintechGateway(cfg.TyrFintechURL, cfg.TyrFintechAPIKey, cfg.TyrFintechCallbackURL, cfg.TyrFintechWebhookURL)

	multiGateway := payment.NewMultiGateway(map[string]payment.PaymentGateway{
		payment.ProviderMock:       mockGateway,
		payment.ProviderTyrFintech: tyrGateway,
	}, mockGateway)
	
	paymentService := payment.NewService(paymentRepo, orderService, multiGateway, res.RabbitMQ, outboxRepo, cfg.TyrFintechCallbackURL, cfg.TyrFintechWebhookURL)
	paymentHandler := payment.NewHandler(paymentService)

	// 7. Initialize Cart Feature Package (Redis)
	cartRepo := cart.NewRedisRepository(res.Redis, 7*24*time.Hour)
	cartService := cart.NewService(cartRepo, storeService, productService, orderService)
	cartHandler := cart.NewHandler(cartService)

	// 8. Initialize Notification Feature Package (SSE)
	notificationHandler := notification.NewHandler(notifHub)

	authMiddleware := middleware.AuthMiddleware(cfg.JWTSecret)

	rateLimiters := RateLimiters{
		LoginLimiter: middleware.RateLimiter(res.Redis, middleware.RateLimiterConfig{
			Prefix: "auth:login",
			Limit:  5,
			Window: 1 * time.Minute,
		}),
		RegisterLimiter: middleware.RateLimiter(res.Redis, middleware.RateLimiterConfig{
			Prefix: "auth:register",
			Limit:  5,
			Window: 10 * time.Minute,
		}),
		WebhookLimiter: middleware.RateLimiter(res.Redis, middleware.RateLimiterConfig{
			Prefix: "webhook:tyrfintech",
			Limit:  30,
			Window: 1 * time.Minute,
		}),
	}

	return NewRouter(authHandler, storeHandler, productHandler, orderHandler, paymentHandler, cartHandler, notificationHandler, authMiddleware, rateLimiters)
}

func runServer(cfg *config.Config, router *gin.Engine) {
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      0, // Disabled for SSE long-lived streams
		IdleTimeout:       120 * time.Second,
	}

	//
	go func() {
		log.Printf("Foodeli Core API server listening on port %s (%s mode)", cfg.ServerPort, cfg.ServerEnv)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down Foodeli Core API server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Foodeli Core API server exited cleanly")
}

// startConsumers initializes and starts all asynchronous event workers and outbox relays.
func startConsumers(ctx context.Context, res *Resources, notifHub *notification.Hub) error {
	// 1. Asynchronous Notification Queue Consumer (RabbitMQ -> SMS/Push/SSE)
	notifConsumer, err := notification.NewConsumer(res.RabbitMQ, notifHub)
	if err != nil {
		return fmt.Errorf("failed to initialize notification consumer: %w", err)
	}
	if err := notifConsumer.Start(ctx); err != nil {
		return err
	}

	// 2. Transactional Outbox Relay Worker (PostgreSQL outbox_events -> RabbitMQ Exchange)
	outboxRepo := outbox.NewPostgresRepository(res.DB)
	outboxWorker := outbox.NewRelayWorker(outboxRepo, res.RabbitMQ, outbox.DefaultWorkerConfig())
	outboxWorker.Start(ctx)

	return nil
}
