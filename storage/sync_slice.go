package storage

import "sync"

type SyncSlice[T interface{}] struct {
	mu sync.Mutex
	s  []T
}

func (s *SyncSlice[T]) Append(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.s = append(s.s, v)
}

func (s *SyncSlice[T]) Data() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	data := make([]T, len(s.s))
	copy(data, s.s)
	return data
}
