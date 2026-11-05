package sim

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/tokens"
)

var oaiAuth = map[string]string{"Authorization": "Bearer tg-example-key", "Content-Type": "application/json"}

func newOpenAI(t *testing.T, seed uint64) *httptest.Server {
	t.Helper()
	cfg := DefaultConfig(FlavorOpenAI)
	cfg.Seed = seed
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

const chatBody = `{"model":"gpt-4o-mini","messages":[{"role":"system","content":"Be brief."},{"role":"user","content":"Say hello."}]}`

func TestOpenAIChatMatchesFixture(t *testing.T) {
	ts := newOpenAI(t, 42)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	assertSameShape(t, "chat_completion.json", jsonShape(t, fixture(t, "openai/chat_completion.json")), jsonShape(t, body))
}

func TestOpenAIStreamMatchesFixtures(t *testing.T) {
	ts := newOpenAI(t, 42)
	for _, tc := range []struct{ fixture, extra string }{
		{"chat_completion_stream.sse", `"stream":true`},
		{"chat_completion_stream_usage.sse", `"stream":true,"stream_options":{"include_usage":true}`},
	} {
		req := strings.Replace(chatBody, `{"model"`, `{`+tc.extra+`,"model"`, 1)
		status, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", req, oaiAuth)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, body)
		}
		want := streamShape(t, parseSSE(t, fixture(t, "openai/"+tc.fixture)))
		assertSameShape(t, tc.fixture, want, streamShape(t, parseSSE(t, body)))
	}
}

func TestOpenAIToolCallsMatchFixtures(t *testing.T) {
	ts := newOpenAI(t, 42)
	tools := `"tools":[{"type":"function","function":{"name":"get_current_weather","parameters":{"type":"object"}}}]`
	for _, choice := range []string{`"required"`, `{"type":"function","function":{"name":"get_current_weather"}}`} {
		req := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Weather in Boston?"}],` + tools + `,"tool_choice":` + choice + `}`
		status, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", req, oaiAuth)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, body)
		}
		assertSameShape(t, "chat_completion_tool_call.json", jsonShape(t, fixture(t, "openai/chat_completion_tool_call.json")), jsonShape(t, body))
		if !strings.Contains(string(body), `"name":"get_current_weather","arguments":"{}"`) {
			t.Fatalf("tool call body = %s", body)
		}
		streamReq := strings.Replace(req, `{"model"`, `{"stream":true,"model"`, 1)
		_, sbody := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", streamReq, oaiAuth)
		want := streamShape(t, parseSSE(t, fixture(t, "openai/chat_completion_tool_call_stream.sse")))
		assertSameShape(t, "chat_completion_tool_call_stream.sse", want, streamShape(t, parseSSE(t, sbody)))
	}
	// tool_choice "auto" does not force a call.
	req := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}],` + tools + `,"tool_choice":"auto"}`
	_, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", req, oaiAuth)
	if strings.Contains(string(body), "tool_calls") {
		t.Fatalf("auto tool_choice produced a tool call: %s", body)
	}
}

func TestOpenAIUsageFollowsTheCountingMethod(t *testing.T) {
	ts := newOpenAI(t, 42)
	_, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth)
	var resp oaiCompletion
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	wantPrompt := tokens.CountChat([]tokens.Message{{Role: "system", Content: "Be brief."}, {Role: "user", Content: "Say hello."}})
	u := resp.Usage
	if u.PromptTokens != wantPrompt {
		t.Errorf("prompt_tokens = %d, want %d", u.PromptTokens, wantPrompt)
	}
	if got := tokens.Count(*resp.Choices[0].Message.Content); u.CompletionTokens != got {
		t.Errorf("completion_tokens = %d, but the content counts %d tokens", u.CompletionTokens, got)
	}
	if u.TotalTokens != u.PromptTokens+u.CompletionTokens {
		t.Errorf("total_tokens = %d, want %d", u.TotalTokens, u.PromptTokens+u.CompletionTokens)
	}
}

func TestOpenAIMaxTokensCutsWithLength(t *testing.T) {
	ts := newOpenAI(t, 42)
	for _, limit := range []string{`"max_tokens":5`, `"max_completion_tokens":5,"max_tokens":50`} {
		req := strings.Replace(chatBody, `{"model"`, `{`+limit+`,"model"`, 1)
		_, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", req, oaiAuth)
		var resp oaiCompletion
		if err := json.Unmarshal(body, &resp); err != nil {
			t.Fatal(err)
		}
		if resp.Choices[0].FinishReason != "length" || resp.Usage.CompletionTokens != 5 {
			t.Fatalf("%s: finish=%q completion=%d", limit, resp.Choices[0].FinishReason, resp.Usage.CompletionTokens)
		}
	}
}

func TestOpenAIStreamUsageMatchesContent(t *testing.T) {
	ts := newOpenAI(t, 7)
	req := strings.Replace(chatBody, `{"model"`, `{"stream":true,"stream_options":{"include_usage":true},"model"`, 1)
	_, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", req, oaiAuth)
	var text strings.Builder
	var usage oaiUsage
	for _, f := range parseSSE(t, body) {
		if f.data == "[DONE]" {
			continue
		}
		var c struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
			Usage *oaiUsage `json:"usage"`
		}
		if err := json.Unmarshal([]byte(f.data), &c); err != nil {
			t.Fatal(err)
		}
		for _, ch := range c.Choices {
			text.WriteString(ch.Delta.Content)
		}
		if c.Usage != nil {
			usage = *c.Usage
		}
	}
	if usage.CompletionTokens == 0 || usage.CompletionTokens != tokens.Count(text.String()) {
		t.Fatalf("streamed usage %d vs content tokens %d", usage.CompletionTokens, tokens.Count(text.String()))
	}
}

func TestOpenAIErrors(t *testing.T) {
	ts := newOpenAI(t, 42)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, map[string]string{"Content-Type": "application/json"})
	if status != http.StatusUnauthorized {
		t.Fatalf("missing key: %d", status)
	}
	assertSameShape(t, "error_401.json", jsonShape(t, fixture(t, "openai/error_401.json")), jsonShape(t, body))

	unknown := strings.Replace(chatBody, "gpt-4o-mini", "gpt-unknown", 1)
	status, body = doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", unknown, oaiAuth)
	if status != http.StatusNotFound || !strings.Contains(string(body), "model_not_found") {
		t.Fatalf("unknown model: %d %s", status, body)
	}
	for _, bad := range []string{`{not json`, `{"model":"gpt-4o-mini","messages":[]}`, `{"messages":[{"role":"user","content":"x"}]}`,
		`{"model":"gpt-4o-mini","n":2,"messages":[{"role":"user","content":"x"}]}`} {
		if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", bad, oaiAuth); status != http.StatusBadRequest {
			t.Errorf("%s: status %d, want 400", bad, status)
		}
	}
	big := `{"model":"gpt-4o-mini","messages":[{"role":"user","content":"` + strings.Repeat("a", maxBodyBytes) + `"}]}`
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", big, oaiAuth); status != http.StatusRequestEntityTooLarge {
		t.Errorf("oversized body: status %d, want 413", status)
	}
}

var volatile = regexp.MustCompile(`"(id|created)":("[^"]*"|\d+)`)

// TestOpenAIReplayIsDeterministic covers spec SC-004: the same seed and request sequence give
// byte-identical bodies once ids and timestamps are removed.
func TestOpenAIReplayIsDeterministic(t *testing.T) {
	requests := []string{
		chatBody,
		strings.Replace(chatBody, `{"model"`, `{"stream":true,"stream_options":{"include_usage":true},"model"`, 1),
		`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"Another question"}],"max_tokens":12}`,
	}
	run := func(seed uint64) []string {
		ts := newOpenAI(t, seed)
		var out []string
		for _, r := range requests {
			_, body := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", r, oaiAuth)
			out = append(out, volatile.ReplaceAllString(string(body), `"$1":"*"`))
		}
		_, emb := doReq(t, http.MethodPost, ts.URL+"/v1/embeddings", `{"model":"text-embedding-3-small","input":"x"}`, oaiAuth)
		return append(out, string(emb))
	}
	a, b, c := run(42), run(42), run(43)
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("request %d differs between replays with the same seed", i)
		}
	}
	if a[0] == c[0] {
		t.Fatal("different seeds produced the same answer")
	}
}

func TestOpenAIStreamStopsWhenClientLeaves(t *testing.T) {
	cfg := DefaultConfig(FlavorOpenAI)
	cfg.OutputTokens = Distribution{Kind: "fixed", Value: 1000}
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // the client is already gone
	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/v1/chat/completions",
		strings.NewReader(strings.Replace(chatBody, `{"model"`, `{"stream":true,"model"`, 1)))
	req.Header.Set("Authorization", "Bearer tg-example-key")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if n := strings.Count(rec.Body.String(), "data: "); n > 1 {
		t.Fatalf("server kept streaming after the client left: %d frames", n)
	}
}
