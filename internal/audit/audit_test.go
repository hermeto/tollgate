package audit

import "testing"

func TestHashChain(t *testing.T) {
	l := New()
	l.Append("key.create", "admin", "tg_abc")
	l.Append("policy.reload", "admin", "payments")
	if err := l.Verify(); err != nil {
		t.Fatal(err)
	}
	e := l.Events()
	e[0].Action = "tampered"
	if err := l.Verify(); err != nil {
		t.Fatal(err)
	}
}
