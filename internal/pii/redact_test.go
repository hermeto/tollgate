package pii

import "testing"

func TestRedactSyntheticPII(t *testing.T) {
	out, m := Redact("contact jane@example.test or +1 416-555-0100")
	if out == "" || len(m) != 2 || out == "contact jane@example.test or +1 416-555-0100" {
		t.Fatalf("out=%q matches=%+v", out, m)
	}
}
