package extproc

import "testing"

func TestDecisionsCopyBody(t *testing.T) {
	b := []byte("x")
	d := Deny(RequestBody, 403, b)
	b[0] = 'y'
	if d.Continue || string(d.Body) != "x" {
		t.Fatal(d)
	}
}
