package cache

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
)

const (
	defaultRedisHost       = "localhost"
	defaultCacheExpitation = time.Millisecond * 50
)

var (
	redisClient *redis.Client
	redisPort   string
)

func TestMain(m *testing.M) {
	cleanup, err := newDockerRedis()
	if err != nil {
		log.Fatalf("could not run Redis container: %s", err)
	}
	// Run the tests
	code := m.Run()

	// Clean up resources
	cleanup()

	os.Exit(code)
}

func newDockerRedis() (func(), error) {
	// Create a new Pool with a Redis container
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, fmt.Errorf("failed to  construct pool: %w", err)
	}

	if err = pool.Client.Ping(); err != nil {
		return nil, fmt.Errorf("failed to connect to Docker: %w", err)
	}

	container, err := pool.Run("redis", "3.2", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker container: %w", err)
	}

	// Set up a Redis client to connect to the Docker container
	redisPort = container.GetPort("6379/tcp")
	redisClient = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", defaultRedisHost, redisPort), // The container is accessible at this address
	})

	// Wait for the Redis container to be ready
	if err = pool.Retry(func() error {
		return redisClient.Ping(context.Background()).Err()
	}); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	// Clean up resources
	cleanup := func() {
		if err = pool.Purge(container); err != nil {
			log.Fatalf("failed to purge resource: %s", err)
		}
	}

	return cleanup, nil
}

func TestRedisCacheIntegration(t *testing.T) {
	testCases := []struct {
		name     string
		keyAdd   []byte
		keyGet   []byte
		expired  bool
		expected bool
	}{
		{
			name:     "Test Add and Get - Key Exists",
			keyAdd:   []byte("test_key_exists"),
			keyGet:   []byte("test_key_exists"),
			expected: true,
		},
		{
			name:     "Test Add and Get - Key Does Not Exist",
			keyGet:   []byte("test_key_not_exists"),
			expected: false,
		},
		{
			name:     "Test Add and Get - Expired Key",
			keyAdd:   []byte("test_key_expired"),
			keyGet:   []byte("test_key_expired"),
			expired:  true,
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()

			cache, err := New(ctx, defaultRedisHost, ":"+redisPort, defaultCacheExpitation)
			assert.NoError(t, err)

			// Test Add and Get methods
			if tc.keyAdd != nil {
				err = cache.Add(tc.keyAdd)
				assert.NoError(t, err)
			}

			if tc.expired {
				time.Sleep(defaultCacheExpitation * 2)
			}

			exists, err := cache.Get(tc.keyGet)
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, exists)

			// Clean up
			if tc.keyAdd != nil {
				result := redisClient.Get(ctx, string(tc.keyAdd))
				assert.Equal(t, redis.Nil, result.Err())
			}
		})
	}
}
