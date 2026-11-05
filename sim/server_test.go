package sim

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T, f Flavor) *httptest.Server {
	t.Helper()
	s, err := New(DefaultConfig(f), Options{})
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(s.Handler())
	t.Cleanup(ts.Close)
	return ts
}

// doReq performs a request and returns the status and body (the body is always closed).
func doReq(t *testing.T, method, url, body string, headers map[string]string) (int, []byte) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode, data
}

func TestHealthz(t *testing.T) {
	ts := newTestServer(t, FlavorOpenAI)
	status, body := doReq(t, http.MethodGet, ts.URL+"/_sim/healthz", "", nil)
	if status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	var h map[string]any
	if err := json.Unmarshal(body, &h); err != nil || h["status"] != "ok" || h["flavor"] != "openai" {
		t.Fatalf("body = %s (%v)", body, err)
	}
}

func TestMissingAuthIsProviderShaped401(t *testing.T) {
	oai := newTestServer(t, FlavorOpenAI)
	status, body := doReq(t, http.MethodPost, oai.URL+"/v1/unknown", "{}", nil)
	if status != http.StatusUnauthorized || !strings.Contains(string(body), `"code":"invalid_api_key"`) {
		t.Fatalf("openai: %d %s", status, body)
	}
	ant := newTestServer(t, FlavorAnthropic)
	status, body = doReq(t, http.MethodPost, ant.URL+"/v1/unknown", "{}", nil)
	if status != http.StatusUnauthorized || !strings.Contains(string(body), `"type":"authentication_error"`) {
		t.Fatalf("anthropic: %d %s", status, body)
	}
}

func TestAnyKeyValueIsAccepted(t *testing.T) {
	oai := newTestServer(t, FlavorOpenAI)
	status, body := doReq(t, http.MethodPost, oai.URL+"/v1/unknown", "{}", map[string]string{"Authorization": "Bearer tg-example-anything"})
	if status != http.StatusNotFound || !strings.Contains(string(body), `"type":"invalid_request_error"`) {
		t.Fatalf("openai: %d %s", status, body)
	}
	ant := newTestServer(t, FlavorAnthropic)
	status, body = doReq(t, http.MethodPost, ant.URL+"/v1/unknown", "{}", map[string]string{"X-Api-Key": "tg-example-anything"})
	if status != http.StatusNotFound || !strings.Contains(string(body), `"type":"not_found_error"`) {
		t.Fatalf("anthropic: %d %s", status, body)
	}
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	cfg := DefaultConfig(FlavorOpenAI)
	cfg.Faults.ErrorRates = map[int]float64{529: 0.1}
	if _, err := New(cfg, Options{}); err == nil || !strings.Contains(err.Error(), "Anthropic-only") {
		t.Fatalf("err = %v", err)
	}
}

func TestServeShutsDownGracefully(t *testing.T) {
	cfg := DefaultConfig(FlavorOpenAI)
	s, err := New(cfg, Options{})
	if err != nil {
		t.Fatal(err)
	}
	var lc net.ListenConfig
	ln, err := lc.Listen(t.Context(), "tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	done := make(chan error, 1)
	go func() { done <- s.Serve(ctx, ln) }()
	if status, _ := doReq(t, http.MethodGet, "http://"+ln.Addr().String()+"/_sim/healthz", "", nil); status != http.StatusOK {
		t.Fatalf("status = %d", status)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Serve returned %v", err)
		}
	case <-time.After(shutdownGrace + time.Second):
		t.Fatal("Serve did not return after cancel")
	}
}

func TestDefaultBindIsLoopback(t *testing.T) {
	if h := DefaultConfig(FlavorOpenAI).Host; h != "127.0.0.1" {
		t.Fatalf("default host = %q", h)
	}
}

func TestRequireAuthMiddleware(t *testing.T) {
	for _, f := range []Flavor{FlavorOpenAI, FlavorAnthropic} {
		s, err := New(DefaultConfig(f), Options{})
		if err != nil {
			t.Fatal(err)
		}
		called := false
		h := s.requireAuth(func(w http.ResponseWriter, _ *http.Request) {
			called = true
			w.WriteHeader(http.StatusNoContent)
		})
		rec := httptest.NewRecorder()
		h(rec, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/x", nil))
		if rec.Code != http.StatusUnauthorized || called {
			t.Fatalf("%s without auth: status %d, called %v", f, rec.Code, called)
		}
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/v1/x", nil)
		if f == FlavorOpenAI {
			req.Header.Set("Authorization", "Bearer tg-example-key")
		} else {
			req.Header.Set("X-Api-Key", "tg-example-key")
		}
		rec = httptest.NewRecorder()
		h(rec, req)
		if rec.Code != http.StatusNoContent || !called {
			t.Fatalf("%s with auth: status %d, called %v", f, rec.Code, called)
		}
	}
}

func TestEmptyBearerIsRejected(t *testing.T) {
	ts := newTestServer(t, FlavorOpenAI)
	status, _ := doReq(t, http.MethodPost, ts.URL+"/v1/x", "{}", map[string]string{"Authorization": "Bearer "})
	if status != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", status)
	}
}
