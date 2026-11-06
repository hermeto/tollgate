package identity

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCreateAuthenticateRevokeAndPersist(t *testing.T) {
	path := filepath.Join(t.TempDir(), "keys.json")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	raw, r, err := s.Create("payments", "checkout", nil)
	if err != nil {
		t.Fatal(err)
	}
	got, ok := s.Authenticate(raw, time.Now())
	if !ok || got.Team != "payments" || r.Prefix == "" {
		t.Fatal("key did not authenticate")
	}
	if err := s.Revoke(r.Prefix); err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Authenticate(raw, time.Now()); ok {
		t.Fatal("revoked key authenticated")
	}
	reloaded, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(reloaded.List()) != 1 {
		t.Fatal("persisted key missing")
	}
}
func TestExpiredKey(t *testing.T) {
	s, _ := Open("")
	expires := time.Now().Add(-time.Second)
	raw, _, err := s.Create("t", "p", &expires)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := s.Authenticate(raw, time.Now()); ok {
		t.Fatal("expired key authenticated")
	}
}
