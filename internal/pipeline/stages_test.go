package pipeline

import (
	"strings"
	"testing"
	"time"
)

func TestRequestIDPropagatesValidID(t *testing.T) {
	req := newReq()
	req.Header.Set("X-Request-Id", "client-abc-123")
	RequestID{}.OnRequest(t.Context(), req)
	if req.ID != "client-abc-123" || req.UpstreamHeader.Get("X-Request-Id") != "client-abc-123" || req.ResponseHeader.Get("X-Request-Id") != "client-abc-123" {
		t.Fatalf("id=%q upstream=%q response=%q", req.ID, req.UpstreamHeader.Get("X-Request-Id"), req.ResponseHeader.Get("X-Request-Id"))
	}
}

func TestRequestIDGeneratesWhenMissingOrInvalid(t *testing.T) {
	for _, in := range []string{"", "has space", strings.Repeat("x", 129), "bad\x00byte"} {
		req := newReq()
		if in != "" {
			req.Header.Set("X-Request-Id", in)
		}
		RequestID{}.OnRequest(t.Context(), req)
		if !strings.HasPrefix(req.ID, "req_") || len(req.ID) != 30 || req.ID == in {
			t.Errorf("input %q → id %q", in, req.ID)
		}
	}
	a, b := newReq(), newReq()
	RequestID{}.OnRequest(t.Context(), a)
	RequestID{}.OnRequest(t.Context(), b)
	if a.ID == b.ID {
		t.Fatal("generated ids collide")
	}
}

func TestTimingStampsArrivalOnce(t *testing.T) {
	req := newReq()
	Timing{}.OnRequest(t.Context(), req)
	first := req.Received
	if first.IsZero() {
		t.Fatal("arrival not stamped")
	}
	time.Sleep(time.Millisecond)
	Timing{}.OnRequest(t.Context(), req)
	if !req.Received.Equal(first) {
		t.Fatal("arrival overwritten")
	}
}
