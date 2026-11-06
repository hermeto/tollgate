// Package ratelimit provides a concurrency-safe local token bucket.
package ratelimit

import (
	"sync"
	"time"
)

type Limit struct {
	Capacity int64
	Per      time.Duration
}
type bucket struct {
	tokens float64
	at     time.Time
}
type Limiter struct {
	mu      sync.Mutex
	limits  map[string]Limit
	buckets map[string]bucket
}
type Decision struct {
	Allowed   bool
	Limit     int64
	Remaining int64
	Reset     time.Time
}

func New(limits map[string]Limit) *Limiter {
	cp := map[string]Limit{}
	for k, v := range limits {
		cp[k] = v
	}
	return &Limiter{limits: cp, buckets: map[string]bucket{}}
}
func (l *Limiter) Allow(key string, cost int64, now time.Time) Decision {
	l.mu.Lock()
	defer l.mu.Unlock()
	lim, ok := l.limits[key]
	if !ok || lim.Capacity <= 0 || lim.Per <= 0 {
		return Decision{Allowed: true}
	}
	b := l.buckets[key]
	if b.at.IsZero() {
		b.at = now
		b.tokens = float64(lim.Capacity)
	}
	elapsed := now.Sub(b.at)
	if elapsed > 0 {
		b.tokens += elapsed.Seconds() * float64(lim.Capacity) / lim.Per.Seconds()
		if b.tokens > float64(lim.Capacity) {
			b.tokens = float64(lim.Capacity)
		}
		b.at = now
	}
	allowed := b.tokens >= float64(cost)
	if allowed {
		b.tokens -= float64(cost)
	}
	l.buckets[key] = b
	remaining := int64(b.tokens)
	reset := now
	if !allowed {
		missing := float64(cost) - b.tokens
		reset = now.Add(time.Duration(missing / float64(lim.Capacity) * float64(lim.Per)))
	}
	return Decision{Allowed: allowed, Limit: lim.Capacity, Remaining: remaining, Reset: reset}
}
