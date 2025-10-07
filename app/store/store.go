package store

import "sync"

type Store struct {
	mu    sync.Mutex
	store map[string]string
}

func NewStore() *Store {
	return &Store{store: make(map[string]string)}
}

func (s *Store) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[key] = value
}

func (s *Store) Get(key string) (string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, ok := s.store[key]
	return value, ok
}
