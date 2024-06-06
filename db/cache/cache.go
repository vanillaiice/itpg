package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is a cache implementation.
type Cache struct {
	client *redis.Client
	ctx    context.Context
}

// ErrRedisNil is returned when a key is not found in redis.
const ErrRedisNil = redis.Nil

// New initializes a new cache.
func New(url string, ctx context.Context) (*Cache, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opts)

	if err = client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Cache{
		client: client,
		ctx:    ctx,
	}, nil
}

// Close closes the cache.
func (c *Cache) Close() error {
	return c.client.Close()
}

// Set sets a value in the cache.
func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

// Get gets a value from the cache.
func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}
