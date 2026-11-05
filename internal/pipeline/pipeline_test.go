package pipeline

import (
	"context"
	"net/http"
	"slices"
	"strings"
	"testing"
)

// recorder is a test Stage that logs every hook call and returns scripted decisions.
type recorder struct {
	name    string
	log     *[]string
	onReq   Decision
	onResp  Decision
	onChunk Decision
}

func (r *recorder) Name() string { return r.name }

func (r *recorder) OnRequest(_ context.Context, req *Request) Decision {
	*r.log = append(*r.log, "req:"+r.name)
	return r.onReq
}

func (r *recorder) OnResponse(_ context.Context, _ *Request, resp *Response) Decision {
	*r.log = append(*r.log, "resp:"+r.name)
	resp.Header.Add("X-Seen-By", r.name)
	return r.onResp
}

func (r *recorder) OnChunk(_ context.Context, _ *Request, c *Chunk) Decision {
	*r.log = append(*r.log, "chunk:"+r.name)
	c.Data = append(c.Data, r.name...)
	return r.onChunk
}

func stages(log *[]string, names ...string) []*recorder {
	out := make([]*recorder, len(names))
	for i, n := range names {
		out[i] = &recorder{name: n, log: log}
	}
	return out
}

func build(t *testing.T, rs []*recorder) *Pipeline {
	t.Helper()
	ss := make([]Stage, len(rs))
	for i, r := range rs {
		ss[i] = r
	}
	p, err := New(ss...)
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func newReq() *Request {
	return NewRequest(ProtocolOpenAI, http.MethodPost, "/v1/chat/completions", http.Header{}, []byte(`{}`))
}

func TestHookOrder(t *testing.T) {
	var log []string
	p := build(t, stages(&log, "A", "B", "C"))
	req := newReq()
	passed, d := p.Request(t.Context(), req)
	if passed != 3 || d.Kind != Continue {
		t.Fatalf("passed=%d decision=%v", passed, d.Kind)
	}
	resp := &Response{Status: 200, Header: http.Header{}}
	p.Response(t.Context(), req, resp, passed)
	c := &Chunk{Data: []byte("x")}
	p.Chunk(t.Context(), req, c, passed)
	want := []string{"req:A", "req:B", "req:C", "resp:C", "resp:B", "resp:A", "chunk:C", "chunk:B", "chunk:A"}
	if !slices.Equal(log, want) {
		t.Fatalf("calls = %v\nwant  %v", log, want)
	}
	if string(c.Data) != "xCBA" || strings.Join(resp.Header.Values("X-Seen-By"), "") != "CBA" {
		t.Fatalf("mutations: chunk %q, header %v", c.Data, resp.Header.Values("X-Seen-By"))
	}
	if len(req.Records) != 6 { // chunk hooks are not recorded per event
		t.Fatalf("records = %d, want 6", len(req.Records))
	}
}

func TestDenyStopsAndOnlyOuterStagesSeeTheResponse(t *testing.T) {
	var log []string
	rs := stages(&log, "A", "B", "C")
	rs[1].onReq = Denied(http.StatusForbidden, "permission_error", "model not allowed")
	p := build(t, rs)
	req := newReq()
	passed, d := p.Request(t.Context(), req)
	if passed != 1 || d.Kind != Deny || d.Status != http.StatusForbidden {
		t.Fatalf("passed=%d decision=%+v", passed, d)
	}
	p.Response(t.Context(), req, &Response{Status: d.Status, Header: http.Header{}}, passed)
	want := []string{"req:A", "req:B", "resp:A"}
	if !slices.Equal(log, want) {
		t.Fatalf("calls = %v, want %v (C and B's response hook must not run)", log, want)
	}
}

func TestRespondShortCircuits(t *testing.T) {
	var log []string
	rs := stages(&log, "A", "B", "C")
	cached := &Response{Status: 200, Header: http.Header{}, Body: []byte(`{"cached":true}`)}
	rs[1].onReq = Responded(cached)
	p := build(t, rs)
	req := newReq()
	passed, d := p.Request(t.Context(), req)
	if passed != 1 || d.Kind != Respond || d.Response != cached {
		t.Fatalf("passed=%d decision=%+v", passed, d)
	}
	p.Response(t.Context(), req, d.Response, passed)
	if !slices.Equal(log, []string{"req:A", "req:B", "resp:A"}) || cached.Header.Get("X-Seen-By") != "A" {
		t.Fatalf("calls = %v, header %v", log, cached.Header)
	}
}

func TestRespondWithoutResponseIsAnInternalError(t *testing.T) {
	var log []string
	rs := stages(&log, "A")
	rs[0].onReq = Decision{Kind: Respond}
	p := build(t, rs)
	_, d := p.Request(t.Context(), newReq())
	if d.Kind != Deny || d.Status != http.StatusInternalServerError {
		t.Fatalf("decision = %+v", d)
	}
}

func TestResponseAndChunkDeny(t *testing.T) {
	var log []string
	rs := stages(&log, "A", "B")
	rs[1].onResp = Denied(http.StatusBadGateway, "server_error", "blocked")
	rs[1].onChunk = Denied(http.StatusBadGateway, "server_error", "blocked chunk")
	p := build(t, rs)
	req := newReq()
	passed, _ := p.Request(t.Context(), req)
	if d := p.Response(t.Context(), req, &Response{Header: http.Header{}}, passed); d.Kind != Deny {
		t.Fatalf("response decision = %v", d.Kind)
	}
	if d := p.Chunk(t.Context(), req, &Chunk{}, passed); d.Kind != Deny || d.Message != "blocked chunk" {
		t.Fatalf("chunk decision = %+v", d)
	}
	// B denied first in both phases, so A was not called after it.
	if slices.Contains(log[2:], "resp:A") || slices.Contains(log[2:], "chunk:A") {
		t.Fatalf("calls after deny = %v", log)
	}
}

func TestMutateIsRecorded(t *testing.T) {
	var log []string
	rs := stages(&log, "A")
	rs[0].onReq = Mutated()
	p := build(t, rs)
	req := newReq()
	p.Request(t.Context(), req)
	if req.Records[0].Kind != Mutate || req.Records[0].Phase != PhaseRequest || req.Records[0].Stage != "A" {
		t.Fatalf("record = %+v", req.Records[0])
	}
	if StageTime(req)["A"] < 0 {
		t.Fatal("negative stage time")
	}
}

type onlyName struct{ n string }

func (o onlyName) Name() string { return o.n }

func TestNewValidatesStages(t *testing.T) {
	var log []string
	for name, ss := range map[string][]Stage{
		"nil":       {nil},
		"empty":     {&recorder{name: "", log: &log}},
		"duplicate": {&recorder{name: "A", log: &log}, &recorder{name: "A", log: &log}},
		"no hooks":  {onlyName{"x"}},
	} {
		if _, err := New(ss...); err == nil {
			t.Errorf("%s: New accepted invalid stages", name)
		}
	}
	p, err := New(RequestID{}, Timing{})
	if err != nil || !slices.Equal(p.Names(), []string{"request_id", "timing"}) || p.HasChunkHooks() {
		t.Fatalf("names = %v, err %v, chunks %v", p.Names(), err, p.HasChunkHooks())
	}
}

func TestKindAndPhaseNames(t *testing.T) {
	if Deny.String() != "deny" || PhaseChunk.String() != "chunk" || Kind(9).String() != "kind(9)" || Phase(9).String() != "phase(9)" {
		t.Fatal("unexpected names")
	}
}
