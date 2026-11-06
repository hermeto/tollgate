package anthropic

import "testing"

func TestDecodeMessages(t *testing.T) {
	req, err := DecodeMessages([]byte(`{"model":"m","max_tokens":32,"messages":[{"role":"user","content":"hi"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	if req.Protocol != "anthropic" {
		t.Fatal(req.Protocol)
	}
}

func FuzzDecodeMessages(f *testing.F) {
	f.Add([]byte(`{"model":"m","max_tokens":1,"messages":[]}`))
	f.Fuzz(func(t *testing.T, body []byte) { _, _ = DecodeMessages(body) })
}
