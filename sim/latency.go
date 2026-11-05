package sim

import (
	"context"
	"math/rand/v2"
	"time"
)

// pace drives the timing of one response: time to first token, inter-token delays and the
// timeout and stall faults (spec FR-005, FR-007). Delays are absolute deadlines counted from the
// request's arrival, so parsing and writing time is absorbed instead of added (spec SC-007), and
// every wait ends early when the client leaves.
type pace struct {
	ctx      context.Context
	rng      *rand.Rand
	ttft     Distribution
	itl      Distribution
	hold     time.Duration // timeout-before-first-token fault
	stallAt  int           // content tokens sent before the stall; -1 means no stall
	stallFor time.Duration
	at       time.Time // deadline of the previous event; starts at the request's arrival
}

func millis(v float64) time.Duration { return time.Duration(v * float64(time.Millisecond)) }

// waitUntil sleeps until the deadline or until ctx is done, and reports whether it was reached.
func waitUntil(ctx context.Context, deadline time.Time) bool {
	return wait(ctx, time.Until(deadline))
}

// wait sleeps for d or until ctx is done, and reports whether the full delay elapsed.
func wait(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return true
	case <-ctx.Done():
		return false
	}
}

// first waits until the first token is ready. Nothing, not even headers, is sent before it returns,
// so time to first token equals time to first byte.
func (p *pace) first() bool {
	d := p.hold + millis(p.ttft.sample(p.rng))
	if p.stallAt == 0 {
		d += p.stallFor
	}
	p.at = p.at.Add(d)
	return waitUntil(p.ctx, p.at)
}

// next waits before content token number sent+1 (sent ≥ 1 tokens already delivered).
func (p *pace) next(sent int) bool {
	d := millis(p.itl.sample(p.rng))
	if p.stallAt > 0 && sent == p.stallAt {
		d += p.stallFor
	}
	p.at = p.at.Add(d)
	return waitUntil(p.ctx, p.at)
}

// whole waits for a complete non-streaming response of n tokens: first token plus n−1 gaps.
// Mid-stream stalls apply to streams only.
func (p *pace) whole(n int) bool {
	d := p.hold + millis(p.ttft.sample(p.rng))
	for i := 1; i < n; i++ {
		d += millis(p.itl.sample(p.rng))
	}
	p.at = p.at.Add(d)
	return waitUntil(p.ctx, p.at)
}
