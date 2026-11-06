// Package pipeline defines the contract every policy implements (a Stage) and runs Stages around the
// upstream call: request hooks in order, response and chunk hooks in reverse (contracts/pipeline.md in
// specs/003-local-stack). It is transport-independent so the ext_proc adapter (feature 014) reuses it.
package pipeline

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"
)

// Protocol is the wire protocol the client speaks.
type Protocol string

// Supported client protocols.
const (
	ProtocolOpenAI    Protocol = "openai"
	ProtocolAnthropic Protocol = "anthropic"
)

// Request is one client request as seen by the Stages. Stages may change Header and Body (and must
// then return Mutated), and add headers for the upstream or the client response.
type Request struct {
	ID       string
	Protocol Protocol
	Method   string
	Path     string
	Header   http.Header // client headers
	Body     []byte      // raw request body
	Stream   bool        // client asked for a streamed response
	Received time.Time   // arrival at the gateway

	UpstreamHeader  http.Header // extra headers sent upstream
	UpstreamBaseURL string      // optional per-logical-model target override
	KeyPrefix       string
	Team            string
	Project         string
	ResponseHeader  http.Header // extra headers sent to the client on every outcome

	Records []Record // one entry per hook call, in call order
}

// NewRequest returns a Request with initialised header maps.
func NewRequest(protocol Protocol, method, path string, header http.Header, body []byte) *Request {
	return &Request{
		Protocol: protocol, Method: method, Path: path, Header: header, Body: body,
		UpstreamHeader: http.Header{}, ResponseHeader: http.Header{},
		Records: make([]Record, 0, 8),
	}
}

// Response is the upstream (or Stage-generated) response before it reaches the client.
type Response struct {
	Status int
	Header http.Header
	Body   []byte // non-streaming responses only
	Stream bool
}

// Chunk is one server-sent event of a streamed response (raw bytes, without the trailing blank line).
type Chunk struct {
	Data []byte
}

// Kind is a Stage's decision.
type Kind uint8

// Decision kinds.
const (
	Continue Kind = iota // let the request or response through unchanged
	Mutate               // let it through; the Stage changed it in place
	Deny                 // stop with a protocol-correct error
	Respond              // stop and answer with Decision.Response (short-circuit)
)

var kindNames = [...]string{"continue", "mutate", "deny", "respond"}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return fmt.Sprintf("kind(%d)", k)
}

// Decision is the outcome of one hook call.
type Decision struct {
	Kind     Kind
	Status   int       // Deny: HTTP status
	Type     string    // Deny: provider error type, e.g. "invalid_request_error"
	Message  string    // Deny: human-readable reason (never request content)
	Response *Response // Respond: the response to send instead of calling the upstream
}

// Pass returns a Continue decision.
func Pass() Decision { return Decision{} }

// Mutated returns a Mutate decision.
func Mutated() Decision { return Decision{Kind: Mutate} }

// Denied returns a Deny decision.
func Denied(status int, errType, msg string) Decision {
	return Decision{Kind: Deny, Status: status, Type: errType, Message: msg}
}

// Responded returns a Respond decision.
func Responded(r *Response) Decision { return Decision{Kind: Respond, Response: r} }

// Stops reports whether the decision ends the current phase.
func (d Decision) Stops() bool { return d.Kind == Deny || d.Kind == Respond }

// Phase is where in the exchange a hook ran.
type Phase uint8

// Hook phases.
const (
	PhaseRequest Phase = iota
	PhaseResponse
	PhaseChunk
)

var phaseNames = [...]string{"request", "response", "chunk"}

func (p Phase) String() string {
	if int(p) < len(phaseNames) {
		return phaseNames[p]
	}
	return fmt.Sprintf("phase(%d)", p)
}

// Record is the timing and outcome of one hook call (input to spans, metrics and audit).
type Record struct {
	Stage    string
	Phase    Phase
	Kind     Kind
	Duration time.Duration
}

// Stage is a named policy step. It implements any of RequestHook, ResponseHook and ChunkHook.
type Stage interface {
	Name() string
}

// RequestHook runs before the upstream call.
type RequestHook interface {
	OnRequest(ctx context.Context, req *Request) Decision
}

// ResponseHook runs on the response before anything is sent to the client.
type ResponseHook interface {
	OnResponse(ctx context.Context, req *Request, resp *Response) Decision
}

// ChunkHook runs on each streamed event before it is sent to the client.
type ChunkHook interface {
	OnChunk(ctx context.Context, req *Request, c *Chunk) Decision
}

type entry struct {
	name  string
	req   RequestHook
	resp  ResponseHook
	chunk ChunkHook
}

// Pipeline runs an ordered list of Stages. It is immutable and safe for concurrent use.
type Pipeline struct {
	entries   []entry
	hasChunks bool
}

// New builds a pipeline. Stage names must be unique and non-empty.
func New(stages ...Stage) (*Pipeline, error) {
	p := &Pipeline{entries: make([]entry, 0, len(stages))}
	seen := map[string]bool{}
	for i, s := range stages {
		if s == nil {
			return nil, fmt.Errorf("pipeline: stage %d is nil", i)
		}
		name := s.Name()
		if name == "" {
			return nil, fmt.Errorf("pipeline: stage %d has an empty name", i)
		}
		if seen[name] {
			return nil, fmt.Errorf("pipeline: duplicate stage %q", name)
		}
		seen[name] = true
		e := entry{name: name}
		e.req, _ = s.(RequestHook)
		e.resp, _ = s.(ResponseHook)
		e.chunk, _ = s.(ChunkHook)
		if e.req == nil && e.resp == nil && e.chunk == nil {
			return nil, fmt.Errorf("pipeline: stage %q implements no hook", name)
		}
		p.hasChunks = p.hasChunks || e.chunk != nil
		p.entries = append(p.entries, e)
	}
	return p, nil
}

// Names returns the Stage names in order.
func (p *Pipeline) Names() []string {
	out := make([]string, len(p.entries))
	for i, e := range p.entries {
		out[i] = e.name
	}
	return out
}

// HasChunkHooks reports whether any Stage inspects streamed events.
func (p *Pipeline) HasChunkHooks() bool { return p.hasChunks }

// ErrInvalidDecision is returned in a Deny when a Stage returns Respond without a response.
var ErrInvalidDecision = errors.New("pipeline: respond decision without a response")

// Request runs request hooks in order. It returns how many Stages let the request through and the
// decision that stopped it (Continue when every Stage passed). Response and chunk hooks later run
// only over those passed Stages, in reverse: a Stage that denies or short-circuits does not see the
// response it caused.
func (p *Pipeline) Request(ctx context.Context, req *Request) (passed int, d Decision) {
	for i := range p.entries {
		e := &p.entries[i]
		if e.req == nil {
			continue
		}
		start := time.Now()
		d = e.req.OnRequest(ctx, req)
		if d.Kind == Respond && d.Response == nil {
			d = Denied(http.StatusInternalServerError, "server_error", ErrInvalidDecision.Error())
		}
		req.Records = append(req.Records, Record{Stage: e.name, Phase: PhaseRequest, Kind: d.Kind, Duration: time.Since(start)})
		if d.Stops() {
			return i, d
		}
	}
	return len(p.entries), Pass()
}

// Response runs response hooks over the first `passed` Stages in reverse and stops at the first Deny.
// Respond is not allowed in the response phase and is treated as Continue.
func (p *Pipeline) Response(ctx context.Context, req *Request, resp *Response, passed int) Decision {
	for i := min(passed, len(p.entries)) - 1; i >= 0; i-- {
		e := &p.entries[i]
		if e.resp == nil {
			continue
		}
		start := time.Now()
		d := e.resp.OnResponse(ctx, req, resp)
		if d.Kind == Respond {
			d = Pass()
		}
		req.Records = append(req.Records, Record{Stage: e.name, Phase: PhaseResponse, Kind: d.Kind, Duration: time.Since(start)})
		if d.Kind == Deny {
			return d
		}
	}
	return Pass()
}

// Chunk runs chunk hooks over the first `passed` Stages in reverse and stops at the first Deny, which
// ends the stream with an error event. Chunk records are not kept per event (only their total time
// matters); the caller measures the chunk phase.
func (p *Pipeline) Chunk(ctx context.Context, req *Request, c *Chunk, passed int) Decision {
	if !p.hasChunks {
		return Pass()
	}
	for i := min(passed, len(p.entries)) - 1; i >= 0; i-- {
		if h := p.entries[i].chunk; h != nil {
			if d := h.OnChunk(ctx, req, c); d.Kind == Deny {
				return d
			}
		}
	}
	return Pass()
}
