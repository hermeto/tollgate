package pipeline

import (
	"context"
	"time"
)

// Timing stamps the request's arrival if the transport has not, so per-Stage records and the gateway
// overhead are measured from one origin. It is otherwise a no-op: the pipeline times every hook.
type Timing struct{}

// Name implements Stage.
func (Timing) Name() string { return "timing" }

// OnRequest implements RequestHook.
func (Timing) OnRequest(_ context.Context, req *Request) Decision {
	if req.Received.IsZero() {
		req.Received = time.Now()
	}
	return Pass()
}

// StageTime sums the time each Stage spent across all its hook calls (input to the stage duration metric).
func StageTime(req *Request) map[string]time.Duration {
	out := make(map[string]time.Duration, len(req.Records))
	for _, r := range req.Records {
		out[r.Stage] += r.Duration
	}
	return out
}
