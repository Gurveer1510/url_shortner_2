package cachehandler

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(addr string) (*RedisClient, error) {
	opt, err := redis.ParseURL(addr) // handles rediss:// with TLS
    if err != nil {
        // fallback for plain addr format
        opt = &redis.Options{Addr: addr}
    }
    rdb := redis.NewClient(opt)

	return &RedisClient{
		client: rdb,
	}, nil
}

func (r *RedisClient) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
} 

func (r *RedisClient) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

func (r *RedisClient) Del(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisClient) Client() *redis.Client {
	return r.client
}