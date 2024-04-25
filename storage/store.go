package storage

import (
	"sync"
	"time"
)

type Key string

type Value struct {
	v  []byte
	ex time.Duration // expiration
}

type Store struct {
	m sync.Map
}

func NewStore() *Store {
	return &Store{}
}

// golang's func, map, slice 不支持 hash, 所以不能作为 map/sync.map 的 key.
// []byte 先转化为 string.
func (s *Store) Put(key []byte, value []byte) {
	k := Key(key)
	v := Value{v: value}
	s.m.Store(k, v)
}

func (s *Store) PutEx(key []byte, value []byte, ex int) {
	k := Key(key)
	v := Value{
		v:  value,
		ex: time.Duration(time.Now().UnixMilli() + int64(ex)),
	}
	s.m.Store(k, v)
}

func (s *Store) Get(key []byte) ([]byte, bool) {
	k := Key(key)
	val, ok := s.m.Load(k)
	if !ok {
		return nil, false
	}
	v := val.(Value)
	if v.ex > 0 && v.ex < time.Duration(time.Now().UnixMilli()) {
		s.m.Delete(k)
		return nil, false
	}
	return v.v, true
}
