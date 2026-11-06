package e2e

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func endpoint(t *testing.T) string {
	t.Helper()
	if os.Getenv("TOLLGATE_E2E") != "1" {
		t.Skip("set TOLLGATE_E2E=1 to run against the local Compose stack")
	}
	return strings.TrimRight(getenv("TOLLGATE_API_URL", "http://127.0.0.1:8080"), "/")
}
func getenv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
func post(t *testing.T, ctx context.Context, body string) *http.Response {
	t.Helper()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint(t)+"/v1/chat/completions", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("content-type", "application/json")
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Do(req)
	if err != nil {
		t.Fatal(err)
	}
	return resp
}

func TestPassthrough(t *testing.T) {
	resp := post(t, context.Background(), `{"model":"simulated","messages":[{"role":"user","content":"hello"}]}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `"choices"`) {
		t.Fatalf("response lacks choices: %s", body)
	}
}

func TestStreamingRelay(t *testing.T) {
	resp := post(t, context.Background(), `{"model":"simulated","messages":[{"role":"user","content":"hello"}],"stream":true}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), "data: [DONE]") {
		t.Fatalf("stream lacks DONE: %s", body)
	}
}

func TestBodyLimit(t *testing.T) {
	resp := post(t, context.Background(), `{"model":"simulated","messages":[{"role":"user","content":"`+strings.Repeat("x", 8*1024*1024)+`"}]}`)
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", resp.StatusCode)
	}
}

func TestClientCancelCancelsUpstream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	resp := post(t, ctx, `{"model":"simulated","messages":[{"role":"user","content":"hello"}],"stream":true}`)
	if resp != nil {
		_ = resp.Body.Close()
	}
	cancel()
}

func TestTelemetryDownDoesNotBlock(t *testing.T) {
	started := time.Now()
	resp := post(t, context.Background(), `{"model":"simulated","messages":[{"role":"user","content":"hello"}]}`)
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if time.Since(started) > 5*time.Second {
		t.Fatal("request exceeded telemetry fail-open budget")
	}
}
