// Package cache provides bounded exact-request response caching.
package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"
)

type Entry struct {
	Status    int
	Header    map[string]string
	Body      []byte
	ExpiresAt time.Time
}
type Store struct {
	mu    sync.Mutex
	max   int
	ttl   time.Duration
	items map[string]Entry
}

func New(max int, ttl time.Duration) *Store {
	return &Store{max: max, ttl: ttl, items: map[string]Entry{}}
}
func Key(method, path string, body []byte) string {
	h := sha256.New()
	h.Write([]byte(method))
	h.Write([]byte{0})
	h.Write([]byte(path))
	h.Write([]byte{0})
	h.Write(body)
	return hex.EncodeToString(h.Sum(nil))
}
func (s *Store) Get(key string, now time.Time) (Entry, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	e, ok := s.items[key]
	if !ok || !now.Before(e.ExpiresAt) {
		delete(s.items, key)
		return Entry{}, false
	}
	e.Body = append([]byte(nil), e.Body...)
	return e, true
}
func (s *Store) Put(key string, e Entry, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.max > 0 && len(s.items) >= s.max {
		for k := range s.items {
			delete(s.items, k)
			break
		}
	}
	e.ExpiresAt = now.Add(s.ttl)
	e.Body = append([]byte(nil), e.Body...)
	s.items[key] = e
}
