// Package extproc models the gateway's protocol-neutral external-processing decisions.
package extproc

type Phase string

const (
	RequestHeaders  Phase = "request_headers"
	RequestBody     Phase = "request_body"
	ResponseHeaders Phase = "response_headers"
	ResponseBody    Phase = "response_body"
)

type Decision struct {
	Phase    Phase
	Continue bool
	Status   int
	Headers  map[string]string
	Body     []byte
}

func Allow(p Phase) Decision { return Decision{Phase: p, Continue: true} }
func Deny(p Phase, status int, body []byte) Decision {
	return Decision{Phase: p, Status: status, Body: append([]byte(nil), body...)}
}
