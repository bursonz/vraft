package bak

import "sync"

// Storage 定义了KV接口
type Storage interface {
	Set(key string, value []byte)
	Get(key string) ([]byte, bool)
	HasData() bool
}

// MapStorage 使用Map存储[]bytes
type MapStorage struct {
	m  map[string][]byte
	mu sync.Mutex
}

func NewMapStorage() *MapStorage {
	m := make(map[string][]byte)
	return &MapStorage{
		m: m,
	}
}

func (s *MapStorage) Get(key string) ([]byte, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, found := s.m[key]
	return v, found
}

func (s *MapStorage) Set(key string, value []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
}

func (s *MapStorage) HasData() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.m) > 0
}
