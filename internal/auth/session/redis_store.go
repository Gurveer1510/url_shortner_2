package session

import (
	"context"
	"encoding/json"
	"time"

	"github.com/Gurveer1510/urlshortner/internal/cachehandler"
)

const sessionTTL = 7 * 24 * time.Hour
const sessionPrefix = "session:"

type RedisSessionStore struct {
	redis *cachehandler.RedisClient
}

func NewRedisSessionStore(redis *cachehandler.RedisClient) *RedisSessionStore {
	return &RedisSessionStore{redis: redis}
}

func (s *RedisSessionStore) Create(userId, email string) (*Session, error) {
	id, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	sess := &Session{
		UserId: userId,
		Id: id,
		UserEmail: email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sessionTTL),
	}

	data, err := json.Marshal(sess)
	if err != nil {
		return nil, err
	}

	if err := s.redis.Set(context.Background(), sessionPrefix+id, string(data), sessionTTL); err != nil {
		return nil, err
	}
	return sess, nil
}

func (s *RedisSessionStore) Get(id string) (*Session, bool) {
	val, err := s.redis.Get(context.Background(), sessionPrefix+id)
	if err != nil {
		return nil, false
	}

	var sess Session
	if err := json.Unmarshal([]byte(val), &sess); err !=nil{
		return nil, false
	}

	if time.Now().After(sess.ExpiresAt) {
		return nil, false
	}

	return &sess, true
}

func (s *RedisSessionStore) Delete(id string) {
	_ = s.redis.Del(context.Background(), sessionPrefix+id)
}