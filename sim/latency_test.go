package sim

import (
	"bufio"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// streamTimes returns the arrival time of every "data:" line of a streaming request, relative to
// when the request was sent.
func streamTimes(t *testing.T, url, body string, headers map[string]string) []time.Duration {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	start := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out []time.Duration
	sc := bufio.NewScanner(resp.Body)
	for sc.Scan() {
		if strings.HasPrefix(sc.Text(), "data: ") {
			out = append(out, time.Since(start))
		}
	}
	return out
}

func timedServer(t *testing.T, f Flavor, mutate func(*Config)) *httptest.Server {
	t.Helper()
	cfg := DefaultConfig(f)
	mutate(&cfg)
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

// TestStreamTimingTolerances covers spec US3 AS1: TTFT 200 ms, inter-token 20 ms, 10 tokens →
// first content after 200–220 ms, whole stream 380–420 ms. Timers are retried up to 3 times so a
// noisy CI runner does not flake; the lower bounds are never relaxed.
func TestStreamTimingTolerances(t *testing.T) {
	ts := timedServer(t, FlavorOpenAI, func(c *Config) {
		c.Latency.TTFT = Distribution{Kind: "fixed", Value: 200}
		c.Latency.InterToken = Distribution{Kind: "fixed", Value: 20}
		c.OutputTokens = Distribution{Kind: "fixed", Value: 10}
	})
	body := strings.Replace(chatBody, `{"model"`, `{"stream":true,"model"`, 1)
	var first, last time.Duration
	for attempt := range 3 {
		times := streamTimes(t, ts.URL+"/v1/chat/completions", body, oaiAuth)
		// frames: role+first content arrive together, then 9 content chunks, finish, [DONE]
		if len(times) < 12 {
			t.Fatalf("frames = %d", len(times))
		}
		first, last = times[1], times[10]
		if first < 200*time.Millisecond || last < 380*time.Millisecond {
			t.Fatalf("faster than configured: first %v, last content %v", first, last)
		}
		if first <= 220*time.Millisecond && last <= 420*time.Millisecond {
			return
		}
		t.Logf("attempt %d outside tolerance: first %v, last content %v", attempt+1, first, last)
	}
	t.Fatalf("first content %v (want 200–220 ms), last content %v (want 380–420 ms)", first, last)
}

func TestNonStreamingWaitsForTheWholeResponse(t *testing.T) {
	ts := timedServer(t, FlavorAnthropic, func(c *Config) {
		c.Latency.TTFT = Distribution{Kind: "fixed", Value: 50}
		c.Latency.InterToken = Distribution{Kind: "fixed", Value: 5}
		c.OutputTokens = Distribution{Kind: "fixed", Value: 11}
	})
	start := time.Now()
	if status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/messages", msgBody, antAuth); status != http.StatusOK {
		t.Fatalf("status %d", status)
	}
	if d := time.Since(start); d < 100*time.Millisecond {
		t.Fatalf("non-streaming response took %v, want ≥ 100 ms (50 + 10×5)", d)
	}
}

func TestPaceSamplesAreDeterministic(t *testing.T) {
	h := requestHash("/v1/messages", []byte(msgBody))
	sample := func() []time.Duration {
		rng := newSeeds(11).perRequest(h)
		d := Distribution{Kind: "lognormal", Mu: 3, Sigma: 0.5}
		var out []time.Duration
		for range 20 {
			out = append(out, millis(d.sample(rng)))
		}
		return out
	}
	a, b := sample(), sample()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("sample %d differs: %v vs %v", i, a[i], b[i])
		}
	}
}

func TestWaitStopsWhenClientLeaves(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	start := time.Now()
	if wait(ctx, 10*time.Second) {
		t.Fatal("wait reported success on a cancelled context")
	}
	if time.Since(start) > time.Second {
		t.Fatal("wait ignored cancellation")
	}
	if !wait(t.Context(), 0) {
		t.Fatal("zero wait should succeed")
	}
}
