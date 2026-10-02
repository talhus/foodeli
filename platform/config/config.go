package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration parameters for the application.
type Config struct {
	ServerPort string
	ServerEnv  string

	DatabaseURL string

	RedisHost string
	RedisPort string
	RedisPass string
	RedisDB   int

	RabbitMQURL      string
	RabbitMQExchange string

	JWTSecret          string
	JWTExpirationHours int

	TyrFintechURL         string
	TyrFintechAPIKey      string
	TyrFintechCallbackURL string
	TyrFintechWebhookURL  string
}

// Load loads configuration from .env and environment variables.
func Load() (*Config, error) {
	// Try loading from root .env or current dir .env, ignoring if file not found
	_ = godotenv.Load("../../.env")
	_ = godotenv.Load(".env")

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),
		ServerEnv:  getEnv("SERVER_ENV", "development"),

		DatabaseURL: getEnv("DATABASE_URL", "postgres://foodeli:foodeli_secret@localhost:5432/foodeli_db?sslmode=disable"),

		RedisHost: getEnv("REDIS_HOST", "localhost"),
		RedisPort: getEnv("REDIS_PORT", "6379"),
		RedisPass: getEnv("REDIS_PASSWORD", ""),
		RedisDB:   getEnvAsInt("REDIS_DB", 0),

		RabbitMQURL:      getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		RabbitMQExchange: getEnv("RABBITMQ_EXCHANGE", "foodeli.events"),

		JWTSecret:          getEnv("JWT_SECRET", "foodeli_super_secret_jwt_key_2026_wolt_level"),
		JWTExpirationHours: getEnvAsInt("JWT_EXPIRATION_HOURS", 72),

		TyrFintechURL:         getEnv("TYRFINTECH_URL", "http://localhost:8081"),
		TyrFintechAPIKey:      getEnv("TYRFINTECH_API_KEY", "tyr_live_foodeli_secret_key_12345"),
		TyrFintechCallbackURL: getEnv("TYRFINTECH_CALLBACK_URL", "http://localhost:3000/orders/callback"),
		TyrFintechWebhookURL:  getEnv("TYRFINTECH_WEBHOOK_URL", "http://localhost:8080/api/v1/webhooks/tyrfintech"),
	}

	if cfg.JWTSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET must not be empty")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr, ok := os.LookupEnv(key); ok {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}
