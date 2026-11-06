// Package routing selects healthy weighted targets and bounds retries.
package routing

import (
	"math/rand/v2"
	"sync"
	"time"
)

type Target struct {
	Name    string
	Weight  int
	Healthy bool
}
type Selector struct {
	mu      sync.Mutex
	targets []Target
}

func New(targets []Target) *Selector {
	cp := append([]Target(nil), targets...)
	return &Selector{targets: cp}
}
func (s *Selector) Next() *Target {
	s.mu.Lock()
	defer s.mu.Unlock()
	total := 0
	for _, t := range s.targets {
		if t.Healthy && t.Weight > 0 {
			total += t.Weight
		}
	}
	if total == 0 {
		return nil
	}
	n := rand.IntN(total) //nolint:gosec // routing jitter is not a security boundary.
	for i := range s.targets {
		if !s.targets[i].Healthy || s.targets[i].Weight <= 0 {
			continue
		}
		n -= s.targets[i].Weight
		if n < 0 {
			return &s.targets[i]
		}
	}
	return nil
}
func Backoff(attempt int, base, max time.Duration) time.Duration {
	if attempt < 1 {
		attempt = 1
	}
	d := base << min(attempt-1, 10)
	if d > max {
		d = max
	}
	return d/2 + time.Duration(rand.Int64N(int64(d/2)+1)) //nolint:gosec // retry jitter is not a security boundary.
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
