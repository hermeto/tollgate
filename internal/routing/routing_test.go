package routing

import (
	"testing"
	"time"
)

func TestSelectorAndBackoff(t *testing.T) {
	s := New([]Target{{Name: "a", Weight: 1, Healthy: true}, {Name: "b", Weight: 1, Healthy: false}})
	if s.Next() == nil || s.Next().Name != "a" {
		t.Fatal("unhealthy target selected")
	}
	if d := Backoff(3, time.Millisecond, time.Second); d < time.Millisecond || d > 8*time.Millisecond {
		t.Fatal(d)
	}
}
