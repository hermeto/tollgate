package sim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/tokens"
)

var antAuth = map[string]string{"X-Api-Key": "tg-example-key", "Anthropic-Version": "2023-06-01", "Content-Type": "application/json"}

func newAnthropic(t *testing.T, seed uint64) *httptest.Server {
	t.Helper()
	cfg := DefaultConfig(FlavorAnthropic)
	cfg.Seed = seed
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

const msgBody = `{"model":"claude-sonnet-4-5","max_tokens":256,"system":"Be brief.","messages":[{"role":"user","content":"Say hello."}]}`

// eventShapes maps "event/subtype" to the data shape of the first such event in a stream.
func eventShapes(t *testing.T, frames []sseFrame) map[string]any {
	t.Helper()
	out := map[string]any{}
	for _, f := range frames {
		var probe struct {
			Delta        struct{ Type string } `json:"delta"`
			ContentBlock struct{ Type string } `json:"content_block"`
		}
		_ = json.Unmarshal([]byte(f.data), &probe)
		key := f.event + "/" + probe.Delta.Type + probe.ContentBlock.Type
		if _, seen := out[key]; !seen {
			out[key] = jsonShape(t, []byte(f.data))
		}
	}
	return out
}

func eventNames(frames []sseFrame) []string {
	var out []string
	for _, f := range frames {
		if len(out) == 0 || out[len(out)-1] != f.event {
			out = append(out, f.event)
		}
	}
	return out
}

func TestAnthropicMessageMatchesFixture(t *testing.T) {
	ts := newAnthropic(t, 42)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, antAuth)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	assertShapeSubset(t, "message.json", jsonShape(t, fixture(t, "anthropic/message.json")), jsonShape(t, body))
}

func TestAnthropicStreamMatchesFixtures(t *testing.T) {
	ts := newAnthropic(t, 42)
	req := strings.Replace(msgBody, `{"model"`, `{"stream":true,"model"`, 1)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", req, antAuth)
	if status != http.StatusOK {
		t.Fatalf("status %d: %s", status, body)
	}
	got := parseSSE(t, body)
	want := parseSSE(t, fixture(t, "anthropic/message_stream.sse"))
	if g, w := strings.Join(eventNames(got), ","), strings.Join(eventNames(want), ","); g != w {
		t.Fatalf("event order\n got: %s\nwant: %s", g, w)
	}
	gotShapes := eventShapes(t, got)
	for _, fx := range []string{"anthropic/message_stream.sse", "anthropic/message_tool_use_stream.sse"} {
		for key, shape := range eventShapes(t, parseSSE(t, fixture(t, fx))) {
			if g, ok := gotShapes[key]; ok {
				assertShapeSubset(t, fx+" "+key, shape, g)
			}
		}
	}
}

func TestAnthropicToolUseMatchesFixtures(t *testing.T) {
	ts := newAnthropic(t, 42)
	tools := `"tools":[{"name":"get_weather","input_schema":{"type":"object"}}]`
	for _, choice := range []string{`{"type":"any"}`, `{"type":"tool","name":"get_weather"}`} {
		req := `{"model":"claude-sonnet-4-5","max_tokens":256,"messages":[{"role":"user","content":"Weather in Paris?"}],` + tools + `,"tool_choice":` + choice + `}`
		status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", req, antAuth)
		if status != http.StatusOK {
			t.Fatalf("status %d: %s", status, body)
		}
		assertShapeSubset(t, "message_tool_use.json", jsonShape(t, fixture(t, "anthropic/message_tool_use.json")), jsonShape(t, body))
		if !strings.Contains(string(body), `"stop_reason":"tool_use"`) {
			t.Fatalf("stop reason: %s", body)
		}
		_, sbody := doReq(t, http.MethodPost, ts.URL+"/v1/messages", strings.Replace(req, `{"model"`, `{"stream":true,"model"`, 1), antAuth)
		got := eventShapes(t, parseSSE(t, sbody))
		want := eventShapes(t, parseSSE(t, fixture(t, "anthropic/message_tool_use_stream.sse")))
		for _, key := range []string{"content_block_start/tool_use", "content_block_delta/input_json_delta", "message_delta/"} {
			g, ok := got[key]
			if !ok {
				t.Fatalf("stream lacks %s", key)
			}
			assertShapeSubset(t, key, want[key], g)
		}
	}
	req := `{"model":"claude-sonnet-4-5","max_tokens":64,"messages":[{"role":"user","content":"hi"}],` + tools + `,"tool_choice":{"type":"auto"}}`
	if _, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", req, antAuth); strings.Contains(string(body), "tool_use") {
		t.Fatalf("auto tool_choice forced a tool: %s", body)
	}
}

func TestAnthropicUsageAndMaxTokens(t *testing.T) {
	ts := newAnthropic(t, 42)
	_, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, antAuth)
	var resp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
		StopReason string   `json:"stop_reason"`
		Usage      antUsage `json:"usage"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	wantIn := tokens.CountChat([]tokens.Message{{Role: "system", Content: "Be brief."}, {Role: "user", Content: "Say hello."}})
	if resp.Usage.InputTokens != wantIn || resp.Usage.OutputTokens != tokens.Count(resp.Content[0].Text) {
		t.Fatalf("usage %+v, want input %d output %d", resp.Usage, wantIn, tokens.Count(resp.Content[0].Text))
	}
	cut := strings.Replace(msgBody, `"max_tokens":256`, `"max_tokens":3`, 1)
	_, body = doReq(t, http.MethodPost, ts.URL+"/v1/messages", cut, antAuth)
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatal(err)
	}
	if resp.StopReason != "max_tokens" || resp.Usage.OutputTokens != 3 {
		t.Fatalf("max_tokens cut: stop=%q output=%d", resp.StopReason, resp.Usage.OutputTokens)
	}
}

func TestAnthropicErrors(t *testing.T) {
	ts := newAnthropic(t, 42)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, map[string]string{"Content-Type": "application/json"})
	if status != http.StatusUnauthorized {
		t.Fatalf("missing key: %d", status)
	}
	assertShapeSubset(t, "error_401.json", jsonShape(t, fixture(t, "anthropic/error_401.json")), jsonShape(t, body))
	status, body = doReq(t, http.MethodPost, ts.URL+"/v1/messages", strings.Replace(msgBody, "claude-sonnet-4-5", "claude-unknown", 1), antAuth)
	if status != http.StatusNotFound {
		t.Fatalf("unknown model: %d", status)
	}
	assertShapeSubset(t, "error_404.json", jsonShape(t, fixture(t, "anthropic/error_404.json")), jsonShape(t, body))
	for _, bad := range []string{`{not json`, `{"model":"claude-sonnet-4-5","messages":[{"role":"user","content":"x"}]}`,
		`{"model":"claude-sonnet-4-5","max_tokens":0,"messages":[{"role":"user","content":"x"}]}`,
		`{"model":"claude-sonnet-4-5","max_tokens":5,"messages":[]}`} {
		if status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", bad, antAuth); status != http.StatusBadRequest || !strings.Contains(string(body), "invalid_request_error") {
			t.Errorf("%s: %d %s", bad, status, body)
		}
	}
}

func TestAnthropicContentBlocksCount(t *testing.T) {
	ts := newAnthropic(t, 42)
	req := `{"model":"claude-sonnet-4-5","max_tokens":16,"system":[{"type":"text","text":"Be brief."}],` +
		`"messages":[{"role":"user","content":[{"type":"text","text":"Say hello."}]}]}`
	_, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", req, antAuth)
	_, plain := doReq(t, http.MethodPost, ts.URL+"/v1/messages", strings.Replace(msgBody, `"max_tokens":256`, `"max_tokens":16`, 1), antAuth)
	var a, b struct{ Usage antUsage }
	_ = json.Unmarshal(body, &a)
	_ = json.Unmarshal(plain, &b)
	if a.Usage.InputTokens != b.Usage.InputTokens {
		t.Fatalf("block content counted %d tokens, plain string %d", a.Usage.InputTokens, b.Usage.InputTokens)
	}
}

func TestAnthropicReplayIsDeterministic(t *testing.T) {
	run := func(seed uint64) []string {
		ts := newAnthropic(t, seed)
		var out []string
		for _, r := range []string{msgBody, strings.Replace(msgBody, `{"model"`, `{"stream":true,"model"`, 1)} {
			_, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", r, antAuth)
			out = append(out, volatile.ReplaceAllString(string(body), `"$1":"*"`))
		}
		return out
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
