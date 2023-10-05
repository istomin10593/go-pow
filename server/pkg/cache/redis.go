package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
)

// Default value.
const (
	value = "value"
)

// RedisCache - cache for random values
type RedisCache struct {
	ctx        context.Context
	client     *redis.Client
	expiration time.Duration
}

// New - create new instance of RedisCache
func New(ctx context.Context, host, port string, expiration time.Duration) (*RedisCache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     host + port,
		Password: "",
		DB:       0,
	})

	// Test connection.
	err := rdb.Set(ctx, "key", "value", 0).Err()
	if err != nil {
		return nil, err
	}

	return &RedisCache{
		ctx:        ctx,
		client:     rdb,
		expiration: expiration,
	}, nil
}

// Add adds rand key to cache with expiration.
func (c *RedisCache) Add(key []byte) error {
	return c.client.Set(c.ctx, string(key), value, c.expiration).Err()
}

// Get gets key from cache.
func (c *RedisCache) Get(key []byte) (bool, error) {
	_, err := c.client.Get(c.ctx, string(key)).Result()
	if err == redis.Nil {
		return false, nil // Key does not exist in cache
	}

	if err != nil {
		return false, err // Some other error occurred
	}

	return true, nil // Key exists in cache
}

// Delete deletes key from cache.
func (c *RedisCache) Delete(key []byte) error {
	return c.client.Del(c.ctx, string(key)).Err()
}
