# Contract: `internal/pipeline`

```go
type Stage interface{ Name() string }
type RequestHook interface{ OnRequest(ctx context.Context, req *Request) Decision }
type ResponseHook interface{ OnResponse(ctx context.Context, req *Request, resp *Response) Decision }
type ChunkHook interface{ OnChunk(ctx context.Context, req *Request, c *Chunk) Decision }

func New(stages ...Stage) (*Pipeline, error)          // rejects duplicate names
func (p *Pipeline) Request(ctx, req) (ran int, d Decision)
func (p *Pipeline) Response(ctx, req, resp, ran int) Decision
func (p *Pipeline) Chunk(ctx, req, c, ran int) Decision
```

Rules: request hooks run in order and stop at the first Deny or Respond; `ran` is the number of Stages
whose request hook ran (including the one that stopped). Response and chunk hooks run in reverse over
those Stages. A Deny from a response hook replaces the upstream response with an error (only before
anything was sent); a Deny from a chunk hook ends the stream with an error event. Every hook call is
timed and recorded in `Request.Decisions`.
