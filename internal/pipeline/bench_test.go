package pipeline

import (
	"context"
	"net/http"
	"testing"
)

// BenchmarkEmptyPipeline measures the request-ID and timing Stages through the request and response
// phases (spec 003 SC-002). Baseline numbers are recorded in docs/local-stack.md.
func BenchmarkEmptyPipeline(b *testing.B) {
	p, err := New(RequestID{}, Timing{})
	if err != nil {
		b.Fatal(err)
	}
	ctx := context.Background()
	header := http.Header{"Content-Type": {"application/json"}}
	body := []byte(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`)
	b.ReportAllocs()
	for b.Loop() {
		req := NewRequest(ProtocolOpenAI, http.MethodPost, "/v1/chat/completions", header, body)
		passed, _ := p.Request(ctx, req)
		p.Response(ctx, req, &Response{Status: 200, Header: http.Header{}}, passed)
	}
}

// BenchmarkChunkNoHooks is the per-event cost when no Stage inspects streams.
func BenchmarkChunkNoHooks(b *testing.B) {
	p, _ := New(RequestID{}, Timing{})
	req := NewRequest(ProtocolOpenAI, http.MethodPost, "/v1/chat/completions", http.Header{}, nil)
	c := &Chunk{Data: []byte(`data: {"choices":[{"delta":{"content":"hi"}}]}`)}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		p.Chunk(ctx, req, c, 2)
	}
}
