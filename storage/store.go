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
	if s.IsExpired(v) {
		fmt.Printf("store: key=%v, ex=%v, now=%v is outdated", k, v.ex, time.Duration(time.Now().UnixMilli()))
		delete(s.m, k)
		return nil, false
	}
	return v.v, true
}

func (s *Store) Scan() []Key {
	res := []Key{}
	del := []Key{}
	for k, v := range s.m {
		if s.IsExpired(v) {
			del = append(del, k)
		} else {
			res = append(res, k)
		}
	}
	for _, k := range del {
		delete(s.m, k)
	}
	return res
}

func (s *Store) IsExpired(v Value) bool {
	return v.ex > 0 && v.ex < time.Duration(time.Now().UnixMilli())
}
