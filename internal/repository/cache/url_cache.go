package cache

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/cachehandler"
	"github.com/Gurveer1510/urlshortner/internal/domain"
	"github.com/redis/go-redis/v9"
)

const urlKeyPrefix = "url:"

type CachedURLRepository struct {
	store domain.UrlStore
	redis *cachehandler.RedisClient
}

func NewCachedURLRepository(store domain.UrlStore, r *cachehandler.RedisClient) *CachedURLRepository {
	return &CachedURLRepository{
		store: store,
		redis: r,
	}
}

func (c *CachedURLRepository) Get(ctx context.Context, code string) (*domain.Link, error) {
	key := urlKeyPrefix + code
	// key := code
	log.Println("IN THE RDIS GET")
	val, err := c.redis.Get(ctx, key)
	if err == nil {
		var u domain.Link
		if jsonErr := json.Unmarshal([]byte(val), &u); jsonErr == nil {
			return &u, nil
		}
	}

	if !errors.Is(err, redis.Nil) {
		log.Printf("redis Get error for %s: %v", key, err)
	}

	link, err := c.store.Get(ctx, code)
	if err != nil {
		return nil, err
	}

	ttl := 25 * time.Hour
	if link.ExpiresAt != nil {
		remaining := time.Until(*link.ExpiresAt)
		if remaining <= 0 {
			return link, nil
		}
		ttl = remaining
	}

	if data, err := json.Marshal(link); err == nil {
		_ = c.redis.Set(ctx, key, string(data), ttl)
	}
	return link, nil
}

func (c *CachedURLRepository) Save(ctx context.Context, userId string, req domain.UrlReq) error {
    return c.store.Save(ctx, userId, req)
}

func (c *CachedURLRepository) SaveClick(ctx context.Context, ip, code string) error {
    return c.store.SaveClick(ctx, ip, code)
}

func (c *CachedURLRepository) GetStats(ctx context.Context, code string) (*domain.StatsResp, error) {
    return c.store.GetStats(ctx, code)
}