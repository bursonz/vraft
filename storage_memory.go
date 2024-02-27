package raft

import "sync"

type MemoryStorage struct {
	m  map[string][]byte
	mu sync.Mutex
}

func NewMemoryStorage() *MemoryStorage {
	m := make(map[string][]byte)
	return &MemoryStorage{m: m}
}

func (s *MemoryStorage) Set(key string, value []byte) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.m[key] = value
	return true, nil
}

func (s *MemoryStorage) Get(key string) ([]byte, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	value, found := s.m[key]
	return value, found, nil
}

func (s *MemoryStorage) HasData() (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.m) > 0, nil
}

func (s *MemoryStorage) Entries(lo, hi, maxSize int) (Entry, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MemoryStorage) Term(i int) (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MemoryStorage) LastIndex() (int, error) {
	//TODO implement me
	panic("implement me")
}

func (s *MemoryStorage) FirstIndex() (int, error) {
	//TODO implement me
	panic("implement me")
}
