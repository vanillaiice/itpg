package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ctx    context.Context
}

const ErrRedisNil = redis.Nil

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

func (c *Cache) Close() error {
	return c.client.Close()
}

func (c *Cache) Set(key string, value any, ttl time.Duration) error {
	return c.client.Set(c.ctx, key, value, ttl).Err()
}

func (c *Cache) Get(key string) (string, error) {
	return c.client.Get(c.ctx, key).Result()
}
