package pipeline

import (
	"context"
	"crypto/rand"
)

// RequestID propagates a valid incoming x-request-id or generates one, and sends it upstream and back
// to the client (spec 003 FR-008).
type RequestID struct{}

// Name implements Stage.
func (RequestID) Name() string { return "request_id" }

// OnRequest implements RequestHook.
func (RequestID) OnRequest(_ context.Context, req *Request) Decision {
	id := req.Header.Get("X-Request-Id")
	if !validRequestID(id) {
		id = "req_" + rand.Text()
	}
	req.ID = id
	req.UpstreamHeader.Set("X-Request-Id", id)
	req.ResponseHeader.Set("X-Request-Id", id)
	return Pass()
}

// validRequestID accepts 1–128 printable ASCII characters (no spaces), so ids are safe in logs and headers.
func validRequestID(id string) bool {
	if id == "" || len(id) > 128 {
		return false
	}
	for i := range len(id) {
		if id[i] < 0x21 || id[i] > 0x7e {
			return false
		}
	}
	return true
}
