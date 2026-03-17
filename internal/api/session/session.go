package session

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	UserEmail string
	CreatedAt time.Time
	ExpiresAt time.Time
}

type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
}

func NewSessionStore() *SessionStore {
	s := &SessionStore{sessions: make(map[string]*Session)}
	go s.cleanUpLoop()
	return s
}

func (s *SessionStore) Create(email string) (string, error) {
	id, err := generateSessionID()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = &Session{
		UserEmail: email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(7*24*time.Hour),
	}

	return id, nil
}

func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *SessionStore) Get(id string) ( *Session, bool ){
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok || time.Now().After(session.ExpiresAt) {
		return nil, false
	}
	return session, true
}

func (s *SessionStore) cleanUpLoop() {
	ticker := time.NewTicker(1 * time.Hour)
	for range ticker.C {
		s.mu.Lock()
		for id, session := range s.sessions {
			if time.Now().After(session.ExpiresAt) {
				s.Delete(id)
			}
		}
		s.mu.Unlock()
	}
}

func generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
