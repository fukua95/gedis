package storage

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
)

type ID struct {
	timestamp int64
	seq       int64
}

var (
	MinID = ID{timestamp: 0, seq: 1}
	MaxID = ID{timestamp: math.MaxInt64, seq: math.MaxInt64}
	anyID = ID{timestamp: -1, seq: -1}
)

func DecodeID(idStr string) (ID, error) {
	if idStr == "*" {
		return anyID, nil
	}
	v := strings.Split(idStr, "-")
	if len(v) != 2 {
		return ID{}, errors.New("decode stream id error")
	}

	id := ID{}
	f := func(i int) int64 {
		if v[i] == "*" {
			return -1
		}
		res, _ := strconv.ParseInt(v[i], 10, 64)
		return res
	}
	id.timestamp, id.seq = f(0), f(1)
	return id, nil
}

func LessThan(l ID, r ID) bool {
	if l.timestamp < r.timestamp {
		return true
	}

	if l.timestamp == r.timestamp && l.seq != -1 && l.seq < r.seq {
		return true
	}
	return false
}

func Equal(l ID, r ID) bool {
	return l.timestamp == r.timestamp && l.seq == r.seq
}

func GreaterThan(l ID, r ID) bool {
	return l.timestamp > r.timestamp || (l.timestamp == r.timestamp && l.seq > r.seq)
}

func (id *ID) String() string {
	return fmt.Sprintf("%s-%s", strconv.FormatInt(id.timestamp, 10), strconv.FormatInt(id.seq, 10))
}

type KV struct {
	K string
	V string
}

type Entry struct {
	ID  ID
	KVs []KV
}

type Stream struct {
	Entries []*Entry
}

func (s *Stream) Add(e *Entry) {
	s.Entries = append(s.Entries, e)
}

func (s *Stream) Get(start ID, end ID) []*Entry {
	res := []*Entry{}
	for _, e := range s.Entries {
		if LessThan(e.ID, start) {
			continue
		}
		if GreaterThan(e.ID, end) {
			break
		}
		res = append(res, e)
	}
	return res
}

func (s *Stream) LastEntry() *Entry {
	return s.Entries[len(s.Entries)-1]
}
