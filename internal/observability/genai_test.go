package observability

import "testing"

func TestAttributes(t *testing.T) {
	if len(Attributes("openai", "m", "chat", 1, 2)) != 5 {
		t.Fatal("attribute set incomplete")
	}
}
