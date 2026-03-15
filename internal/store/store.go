package store

import (
	"errors"
	"sync"

	urltype "github.com/Gurveer1510/urlshortner/internal/urlType"
)

var ErrNotFound = errors.New("short code not found")
var ErrConflict = errors.New("Duplicate code found")
var ErrExpiredCode = errors.New("Code is expired")

type Store interface {
	Save(urlReq urltype.UrlReq) error
	Get(code string) (*urltype.UrlReq, error)
	SaveClick(ipAddress, code string) error
}

type InMemory struct {
	mu sync.RWMutex
	data map[string]string
}

func NewInMemory() *InMemory {
	return &InMemory{
		data: make(map[string]string),
	}
}

func (s *InMemory) Save(code, url string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[code] = url
	return nil
}

func (s *InMemory) Get(code string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if u, ok := s.data[code]; ok {
		return u, nil
	}
	return "", ErrNotFound
}