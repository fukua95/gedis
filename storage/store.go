package storage

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/codecrafters-io/redis-starter-go/resp"
)

const (
	stringType = "string"
	streamType = "stream"
	nullType   = "none"
)

type Key string

type Value struct {
	v  string
	ex time.Duration // expiration
}

type Store struct {
	m           map[Key]Value
	streams     map[Key]*Stream
	streamsLock sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		m:       make(map[Key]Value),
		streams: make(map[Key]*Stream),
	}
}

// golang's func, map, slice 不支持 hash, 所以不能作为 map/sync.map 的 key.
// []byte 先转化为 string.
func (s *Store) Put(key string, value string, ex int64) {
	k := Key(key)
	v := Value{
		v:  value,
		ex: time.Duration(ex),
	}
	s.m[k] = v
}

func (s *Store) Get(key string) (string, bool) {
	k := Key(key)
	v, ok := s.m[k]
	if !ok {
		fmt.Printf("store: key=%v is not exist\n", k)
		return "", false
	}
	if s.HasExpired(v) {
		fmt.Printf("store: key=%v, ex=%v, now=%v is outdated", k, v.ex, time.Duration(time.Now().UnixMilli()))
		delete(s.m, k)
		return "", false
	}
	return v.v, true
}

func (s *Store) Scan() []Key {
	res := []Key{}
	del := []Key{}
	for k, v := range s.m {
		if s.HasExpired(v) {
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

func (s *Store) HasExpired(v Value) bool {
	return v.ex > 0 && v.ex < time.Duration(time.Now().UnixMilli())
}

func (s *Store) AddStream(key string, idStr string, pairs []string) (string, error) {
	if len(pairs)%2 != 0 {
		return "", errors.New("Store: AddStream para error")
	}
	id, err := s.generateID(key, idStr)
	if err != nil {
		return "", err
	}
	entry := &Entry{ID: id, KVs: make([]KV, len(pairs)/2)}
	for i := 0; i < len(pairs); i += 2 {
		entry.KVs[i/2].K = string(pairs[i])
		entry.KVs[i/2].V = string(pairs[i+1])
	}

	s.streamsLock.Lock()
	defer s.streamsLock.Unlock()

	if _, has := s.streams[Key(key)]; !has {
		s.streams[Key(key)] = &Stream{}
	}
	s.streams[Key(key)].Add(entry)
	return id.String(), nil
}

// [start, end]
func (s *Store) GetStream(key string, start string, end string) []*Entry {
	s.streamsLock.RLock()
	defer s.streamsLock.RUnlock()

	stream, has := s.streams[Key(key)]
	if !has {
		return nil
	}
	startID, _ := DecodeID(start)
	endID, _ := DecodeID(end)
	return stream.Get(startID, endID)
}

func (s *Store) StreamNewestID(key string) string {
	s.streamsLock.RLock()
	defer s.streamsLock.RUnlock()

	stream, has := s.streams[Key(key)]
	var id ID
	if !has {
		id = ID{timestamp: 0, seq: 0}
	} else {
		id = stream.LastEntry().ID
	}
	return id.String()
}

func (s *Store) ValueType(key string) string {
	if _, has := s.Get(key); has {
		return stringType
	}
	if _, has := s.streams[Key(key)]; has {
		return streamType
	}
	return nullType
}

func (s *Store) generateID(key string, idStr string) (ID, error) {
	id, err := DecodeID(idStr)
	if err != nil {
		return id, resp.ErrStreamIDIllegal
	}
	if id.timestamp == 0 && id.seq == 0 {
		return id, resp.ErrStreamIDIllegal
	}
	if id.timestamp == -1 {
		id.timestamp = time.Now().UnixMilli()
	}

	if stream, has := s.streams[Key(key)]; has {
		lastID := stream.LastEntry().ID
		if LessThan(id, lastID) || Equal(id, lastID) {
			return id, resp.ErrStreamIDInvalid
		}
		if id.seq == -1 {
			if id.timestamp == lastID.timestamp {
				id.seq = lastID.seq + 1
			} else {
				id.seq = 0
			}
		}
	} else {
		if id.seq == -1 {
			if id.timestamp == 0 {
				id.seq = 1
			} else {
				id.seq = 0
			}
		}
	}
	return id, nil
}
