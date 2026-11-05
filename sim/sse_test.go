package sim

import (
	"net/http/httptest"
	"testing"
)

func TestSSEFraming(t *testing.T) {
	rec := httptest.NewRecorder()
	s := startSSE(rec)
	if err := s.event("message_stop", []byte(`{"type":"message_stop"}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.event("", []byte("[DONE]")); err != nil {
		t.Fatal(err)
	}
	want := "event: message_stop\ndata: {\"type\":\"message_stop\"}\n\ndata: [DONE]\n\n"
	if rec.Body.String() != want {
		t.Fatalf("body = %q, want %q", rec.Body.String(), want)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q", ct)
	}
	if !rec.Flushed {
		t.Fatal("events were not flushed")
	}
}
