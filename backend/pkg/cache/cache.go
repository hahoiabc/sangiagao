package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache defines a generic cache interface.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPrefix(ctx context.Context, prefix string) error
	Exists(ctx context.Context, key string) (bool, error)
	CountByPrefix(ctx context.Context, prefix string) (int, error)
	KeysByPrefix(ctx context.Context, prefix string) ([]string, error)
	// Incr atomically increments a key by 1 and returns the new value.
	// If the key does not exist, it is created with value 1 and the given TTL.
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

// RedisCache implements Cache using Redis.
type RedisCache struct {
	client *redis.Client
}

// NewRedisCache creates a new Redis-backed cache.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{client: client}
}

func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	return val, err
}

func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *RedisCache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	n, err := c.client.Exists(ctx, key).Result()
	return n > 0, err
}

func (c *RedisCache) CountByPrefix(ctx context.Context, prefix string) (int, error) {
	count := 0
	iter := c.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		count++
	}
	if err := iter.Err(); err != nil {
		return 0, err
	}
	return count, nil
}

func (c *RedisCache) KeysByPrefix(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	iter := c.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}

func (c *RedisCache) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	val, err := c.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	// Set TTL only on first increment (val == 1)
	if val == 1 && ttl > 0 {
		c.client.Expire(ctx, key, ttl)
	}
	return val, nil
}

func (c *RedisCache) DeleteByPrefix(ctx context.Context, prefix string) error {
	iter := c.client.Scan(ctx, 0, prefix+"*", 100).Iterator()
	pipe := c.client.Pipeline()
	count := 0
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
		count++
		if count%100 == 0 {
			if _, err := pipe.Exec(ctx); err != nil {
				return err
			}
		}
	}
	if err := iter.Err(); err != nil {
		return err
	}
	if count%100 != 0 {
		if _, err := pipe.Exec(ctx); err != nil {
			return err
		}
	}
	return nil
}
