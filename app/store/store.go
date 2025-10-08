package store

import (
	"sync"
	"time"
)

type Entry struct {
	Value string
	TTL   time.Time
}

type Store struct {
	mu    sync.Mutex
	store map[string]Entry
}

func NewStore() *Store {
	return &Store{store: make(map[string]Entry)}
}

func (s *Store) Set(key string, value string, ttl *time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl != nil {
		s.store[key] = Entry{Value: value, TTL: time.Now().Add(*ttl)}
	} else {
		s.store[key] = Entry{Value: value, TTL: time.Time{}}
	}
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.store[key]
	if !ok {
		return "", false
	}

	if !value.TTL.IsZero() && value.TTL.Before(time.Now()) {
		delete(s.store, key)
		return "", false
	}

	return value.Value, true
}
