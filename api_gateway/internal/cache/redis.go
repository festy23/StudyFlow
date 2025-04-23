package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCache struct {
	rdb *redis.Client
}

func NewRedisCache(rdb *redis.Client) *RedisCache {
	return &RedisCache{rdb: rdb}
}

func (r *RedisCache) Get(ctx context.Context, key string) ([]byte, bool) {
	val, err := r.rdb.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) || err != nil {
		return nil, false
	}
	return val, true
}

func (r *RedisCache) Set(ctx context.Context, key string, data []byte, ttl time.Duration) {
	r.rdb.Set(ctx, key, data, ttl)
}

func (r *RedisCache) Delete(ctx context.Context, key string) {
	r.rdb.Del(ctx, key)
}
