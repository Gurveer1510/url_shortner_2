package store

import (
	"errors"
	"sync"
)

var ErrNotFound = errors.New("short code not found")
var ErrConflict = errors.New("Duplicate code found")

type Store interface {
	Save(code, url string) error
	Get(code string) (string, error)
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