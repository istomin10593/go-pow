package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// Default value.
	value = "value"

	// GetAndDeleteScript is a Lua script that gets a key from cache and deletes it.
	getAndDeleteScript = `
	local value = redis.call('GET', KEYS[1])
	if value then
		redis.call('DEL', KEYS[1])
		return value
	else
		return nil
	end
	`
)

// RedisCache - cache for random values.
type RedisCache struct {
	ctx        context.Context
	client     *redis.Client
	expiration time.Duration
}

// New - create new instance of RedisCache.
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
	_, err := c.client.Eval(c.ctx, getAndDeleteScript, []string{string(key)}).Result()

	if errors.Is(err, redis.Nil) {
		// Key does not exist in cache.
		return false, nil
	}

	if err != nil {
		// Some other error occurred.
		return false, err
	}

	return true, nil
}
