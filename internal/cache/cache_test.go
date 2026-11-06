package cache

import (
	"testing"
	"time"
)

func TestExactCacheTTL(t *testing.T) {
	now := time.Unix(0, 0)
	s := New(2, time.Minute)
	k := Key("POST", "/v1/chat", []byte("x"))
	s.Put(k, Entry{Status: 200, Body: []byte("ok")}, now)
	if e, ok := s.Get(k, now.Add(time.Second)); !ok || string(e.Body) != "ok" {
		t.Fatal(e, ok)
	}
	if _, ok := s.Get(k, now.Add(2*time.Minute)); ok {
		t.Fatal("expired entry returned")
	}
}
