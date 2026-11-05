package sim

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func putFaults(t *testing.T, url, method, body string) (int, []byte) {
	t.Helper()
	return doReq(t, method, url+"/_sim/faults", body, map[string]string{"Content-Type": "application/json"})
}

func TestFaultStatusWithRetryAfter(t *testing.T) {
	ts := newOpenAI(t, 42)
	if status, body := putFaults(t, ts.URL, http.MethodPut, `{"error_rates":{"429":1},"retry_after_seconds":2}`); status != http.StatusOK {
		t.Fatalf("PUT: %d %s", status, body)
	}
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/v1/chat/completions", strings.NewReader(chatBody))
	req.Header.Set("Authorization", "Bearer k")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
	if resp.StatusCode != http.StatusTooManyRequests || resp.Header.Get("Retry-After") != "2" {
		t.Fatalf("status %d retry-after %q", resp.StatusCode, resp.Header.Get("Retry-After"))
	}
	raw, _ := json.Marshal(body)
	assertSameShape(t, "error_429.json", jsonShape(t, fixture(t, "openai/error_429.json")), jsonShape(t, raw))

	// Cleared live: the next request succeeds (SC-005).
	putFaults(t, ts.URL, http.MethodDelete, "")
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth); status != http.StatusOK {
		t.Fatalf("after DELETE: %d", status)
	}
}

func TestAnthropicOverloadedFault(t *testing.T) {
	ts := newAnthropic(t, 42)
	putFaults(t, ts.URL, http.MethodPut, `{"error_rates":{"529":1}}`)
	status, body := doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, antAuth)
	if status != 529 {
		t.Fatalf("status %d", status)
	}
	assertShapeSubset(t, "error_529.json", jsonShape(t, fixture(t, "anthropic/error_529.json")), jsonShape(t, body))
	putFaults(t, ts.URL, http.MethodPut, `{"error_rates":{"429":1}}`)
	_, body = doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, antAuth)
	assertShapeSubset(t, "error_429.json", jsonShape(t, fixture(t, "anthropic/error_429.json")), jsonShape(t, body))
}

func TestOutageWindow(t *testing.T) {
	ts := newOpenAI(t, 42)
	putFaults(t, ts.URL, http.MethodPut, `{"outage":{"start_after":"150ms","duration":"200ms"}}`)
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth); status != http.StatusOK {
		t.Fatalf("before window: %d", status)
	}
	time.Sleep(200 * time.Millisecond)
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth); status != http.StatusServiceUnavailable {
		t.Fatalf("inside window: %d", status)
	}
	time.Sleep(250 * time.Millisecond)
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", chatBody, oaiAuth); status != http.StatusOK {
		t.Fatalf("after window: %d", status)
	}
}

func TestConnectionReset(t *testing.T) {
	ts := newOpenAI(t, 42)
	putFaults(t, ts.URL, http.MethodPut, `{"reset_rate":1}`)
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/v1/chat/completions", strings.NewReader(chatBody))
	req.Header.Set("Authorization", "Bearer k")
	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		resp.Body.Close()
		t.Fatalf("expected a connection error, got status %d", resp.StatusCode)
	}
	recs := readLog(t, ts.URL)
	if len(recs) != 1 || recs[0].Status != 0 {
		t.Fatalf("request log after reset: %+v", recs)
	}
}

func TestTimeoutBeforeFirstToken(t *testing.T) {
	ts := newOpenAI(t, 42)
	putFaults(t, ts.URL, http.MethodPut, `{"timeout_before_first_token":{"rate":1,"duration":"300ms"}}`)
	body := strings.Replace(chatBody, `{"model"`, `{"stream":true,"model"`, 1)
	times := streamTimes(t, ts.URL+"/v1/chat/completions", body, oaiAuth)
	if len(times) == 0 || times[0] < 300*time.Millisecond {
		t.Fatalf("first byte after %v, want ≥ 300 ms", times)
	}
	client := &http.Client{Timeout: 100 * time.Millisecond}
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, ts.URL+"/v1/chat/completions", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer k")
	if resp, err := client.Do(req); err == nil {
		resp.Body.Close()
		t.Fatal("client with a 100 ms timeout got a response")
	}
}

func TestStallMidStream(t *testing.T) {
	cfg := DefaultConfig(FlavorAnthropic)
	cfg.OutputTokens = Distribution{Kind: "fixed", Value: 6}
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	putFaults(t, ts.URL, http.MethodPut, `{"stall_mid_stream":{"rate":1,"after_tokens":3,"duration":"300ms"}}`)
	times := streamTimes(t, ts.URL+"/v1/messages", strings.Replace(msgBody, `{"model"`, `{"stream":true,"model"`, 1), antAuth)
	// frames: message_start, content_block_start, ping, then deltas at indexes 3..8
	if len(times) < 9 {
		t.Fatalf("frames = %d", len(times))
	}
	if gap := times[6] - times[5]; gap < 300*time.Millisecond {
		t.Fatalf("gap after 3 tokens = %v, want ≥ 300 ms", gap)
	}
	if gap := times[5] - times[4]; gap > 100*time.Millisecond {
		t.Fatalf("stall happened too early: %v", gap)
	}
}

func TestFaultsLiveSetMergeReset(t *testing.T) {
	cfg := DefaultConfig(FlavorOpenAI)
	cfg.Faults = FaultConfig{RetryAfterSeconds: 9}
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	putFaults(t, ts.URL, http.MethodPut, `{"error_rates":{"500":0.1}}`)
	_, body := putFaults(t, ts.URL, http.MethodPatch, `{"reset_rate":0.2,"outage":{"duration":"1m"}}`)
	var got FaultConfig
	if err := json.Unmarshal(body, &got); err != nil {
		t.Fatal(err)
	}
	if got.ErrorRates[500] != 0.1 || got.ResetRate != 0.2 || got.Outage == nil || got.Outage.Start == nil || got.Outage.Status != 503 {
		t.Fatalf("merged config = %+v", got)
	}
	_, body = doReq(t, http.MethodGet, ts.URL+"/_sim/faults", "", nil)
	if !strings.Contains(string(body), `"reset_rate":0.2`) {
		t.Fatalf("GET = %s", body)
	}
	_, body = putFaults(t, ts.URL, http.MethodDelete, "")
	if !strings.Contains(string(body), `"retry_after_seconds":9`) || strings.Contains(string(body), "reset_rate") {
		t.Fatalf("DELETE did not restore the start-up config: %s", body)
	}
}

func TestFaultsRejectInvalidConfig(t *testing.T) {
	ts := newOpenAI(t, 42)
	for _, bad := range []string{`{"error_rates":{"500":0.8},"reset_rate":0.5}`, `{"bogus":1}`, `{"error_rates":{"529":0.1}}`, `not json`} {
		if status, body := putFaults(t, ts.URL, http.MethodPut, bad); status != http.StatusBadRequest {
			t.Errorf("%s: %d %s", bad, status, body)
		}
	}
}

func TestAdminEndpointsIgnoreFaults(t *testing.T) {
	ts := newOpenAI(t, 42)
	putFaults(t, ts.URL, http.MethodPut, `{"error_rates":{"500":1}}`)
	for _, path := range []string{"/_sim/healthz", "/_sim/faults", "/_sim/requests"} {
		if status, _ := doReq(t, http.MethodGet, ts.URL+path, "", nil); status != http.StatusOK {
			t.Errorf("%s: %d", path, status)
		}
	}
	if status, _ := doReq(t, http.MethodGet, ts.URL+"/v1/models", "", oaiAuth); status != http.StatusInternalServerError {
		t.Errorf("/v1/models should be affected by faults: %d", status)
	}
}

func TestFaultDecisionsAreDeterministic(t *testing.T) {
	run := func() string {
		cfg := DefaultConfig(FlavorOpenAI)
		cfg.Faults.ErrorRates = map[int]float64{500: 0.5}
		s, err := New(cfg, Options{})
		if err != nil {
			t.Fatal(err)
		}
		ts := httptest.NewServer(s.Handler())
		defer ts.Close()
		var b strings.Builder
		for i := range 20 {
			body := strings.Replace(chatBody, "Say hello.", "Question "+string(rune('a'+i%5)), 1)
			status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/chat/completions", body, oaiAuth)
			b.WriteString(http.StatusText(status)[:1])
		}
		return b.String()
	}
	a, b := run(), run()
	if a != b {
		t.Fatalf("fault sequence differs: %s vs %s", a, b)
	}
	if !strings.Contains(a, "I") || !strings.Contains(a, "O") {
		t.Fatalf("50%% error rate produced %s", a)
	}
}
