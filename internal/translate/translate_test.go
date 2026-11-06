package translate

import (
	"encoding/json"
	"testing"

	"github.com/hermeto/tollgate/internal/canonical"
)

func TestOpenAIToAnthropic(t *testing.T) {
	max := 32
	body, err := OpenAIToAnthropic(canonical.Request{Model: "claude", Messages: []canonical.Message{{Role: "system", Content: json.RawMessage(`"be concise"`)}, {Role: "user", Content: json.RawMessage(`"hi"`)}}, MaxTokens: &max})
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got["max_tokens"] != float64(32) {
		t.Fatalf("body = %s", body)
	}
}
func TestTranslationRejectsUnsupportedRole(t *testing.T) {
	_, err := OpenAIToAnthropic(canonical.Request{Model: "claude", Messages: []canonical.Message{{Role: "developer"}}, MaxTokens: new(int)})
	if err == nil {
		t.Fatal("unsupported role accepted")
	}
}
