package cart

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type Repository interface {
	GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error)
	SaveCart(ctx context.Context, cart *Cart) error
	DeleteCart(ctx context.Context, userID uuid.UUID) error
}

type RedisRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisRepository(client *redis.Client, ttl time.Duration) Repository {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	return &RedisRepository{
		client: client,
		ttl:    ttl,
	}
}

// ==================================================
// METHODS
// ==================================================
func (r *RedisRepository) cartKey(userID uuid.UUID) string {
	return fmt.Sprintf("cart:%s", userID.String())
}

// get cart
func (r *RedisRepository) GetCart(ctx context.Context, userID uuid.UUID) (*Cart, error) {
	key := r.cartKey(userID)
	data, err := r.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("failed to get cart from redis %w", err)
	}
	var cart Cart
	if err := json.Unmarshal(data, &cart); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cart data: %w", err)
	}
	return &cart, nil
}

// save cart
func (r *RedisRepository) SaveCart(ctx context.Context, cart *Cart) error {
	key := r.cartKey(cart.UserID)
	data, err := json.Marshal(cart)
	if err != nil {
		return fmt.Errorf("failed to marshal cart data: %w", err)
	}

	if err := r.client.Set(ctx, key, data, r.ttl).Err(); err != nil {
		return fmt.Errorf("failed to set cart in redis: %w", err)
	}

	return nil
}

// delete cart
func (r *RedisRepository) DeleteCart(ctx context.Context, userID uuid.UUID) error {
	key := r.cartKey(userID)
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete cart from redis %w", err)
	}
	return nil
}
