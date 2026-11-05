# Contract: `internal/pipeline`

```go
type Stage interface{ Name() string }
type RequestHook interface{ OnRequest(ctx context.Context, req *Request) Decision }
type ResponseHook interface{ OnResponse(ctx context.Context, req *Request, resp *Response) Decision }
type ChunkHook interface{ OnChunk(ctx context.Context, req *Request, c *Chunk) Decision }

func New(stages ...Stage) (*Pipeline, error)          // rejects nil, unnamed, duplicate and hookless Stages
func (p *Pipeline) Request(ctx, req) (passed int, d Decision)
func (p *Pipeline) Response(ctx, req, resp, passed int) Decision
func (p *Pipeline) Chunk(ctx, req, c, passed int) Decision
```

Rules: request hooks run in order and stop at the first Deny or Respond; `passed` is the number of
Stages that let the request through (the stopping Stage is not counted). Response and chunk hooks run in
reverse over those passed Stages, so a Stage never sees the response it caused (a short-circuit's own
response still flows through the outer Stages). Respond without a response becomes a 500 Deny. A Deny from a response hook replaces the upstream response with an error (only before
anything was sent); a Deny from a chunk hook ends the stream with an error event. Every hook call is
timed and recorded in `Request.Decisions`.
