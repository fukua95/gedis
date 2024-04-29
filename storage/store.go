package storage

import (
	"fmt"
	"time"
)

type Key string

type Value struct {
	v  []byte
	ex time.Duration // expiration
}

type Store struct {
	m map[Key]Value
}

func NewStore() *Store {
	return &Store{
		m: make(map[Key]Value),
	}
}

// golang's func, map, slice 不支持 hash, 所以不能作为 map/sync.map 的 key.
// []byte 先转化为 string.
func (s *Store) Put(key []byte, value []byte, ex int64) {
	if ex > 0 {
		ex += time.Now().UnixMilli()
	}
	k := Key(key)
	v := Value{
		v:  value,
		ex: time.Duration(ex),
	}
	s.m[k] = v
}

func (s *Store) Get(key []byte) ([]byte, bool) {
	k := Key(key)
	v, ok := s.m[k]
	if !ok {
		fmt.Printf("store: key=%v is not exist\n", k)
		return nil, false
	}
	if v.ex > 0 && v.ex < time.Duration(time.Now().UnixMilli()) {
		fmt.Printf("store: key=%v, ex=%v, now=%v is outdated", k, v.ex, time.Duration(time.Now().UnixMilli()))
		delete(s.m, k)
		return nil, false
	}
	return v.v, true
}
