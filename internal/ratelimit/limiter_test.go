package ratelimit

import (
	"testing"
	"time"
)

func TestTokenBucketEnforcesCapacityAndRefills(t *testing.T) {
	now := time.Unix(0, 0)
	l := New(map[string]Limit{"team": {Capacity: 2, Per: time.Minute}})
	if !l.Allow("team", 2, now).Allowed {
		t.Fatal("initial request denied")
	}
	if d := l.Allow("team", 1, now); d.Allowed || d.Remaining != 0 {
		t.Fatalf("decision=%+v", d)
	}
	if !l.Allow("team", 1, now.Add(30*time.Second)).Allowed {
		t.Fatal("refill denied")
	}
}
