package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/sim"
)

func TestAnthropicPassthrough(t *testing.T) {
	provider, err := sim.New(sim.DefaultConfig(sim.FlavorAnthropic), sim.Options{Logger: slog.Default()})
	if err != nil {
		t.Fatal(err)
	}
	providerServer := httptest.NewServer(provider.Handler())
	defer providerServer.Close()
	cfg := config.Default()
	cfg.Upstream.BaseURL = providerServer.URL
	cfg.Upstream.CredentialEnv = "TOLLGATE_TEST_KEY"
	t.Setenv("TOLLGATE_TEST_KEY", "tg-example-test")
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/messages", strings.NewReader(`{"model":"claude-sonnet-5-5","max_tokens":8,"messages":[{"role":"user","content":"hi"}]}`))
	response := httptest.NewRecorder()
	s.handle(response, req)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if !strings.Contains(response.Body.String(), `"type":"message"`) {
		t.Fatalf("not Anthropic response: %s", response.Body.String())
	}
}

func TestLogicalModelsEndpoint(t *testing.T) {
	cfg := config.Default()
	cfg.Models = map[string][]config.TargetConfig{"chat-default": {{Provider: "openai", Model: "gpt-4o-mini", BaseURL: "http://127.0.0.1:9001"}}}
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	response := httptest.NewRecorder()
	s.handle(response, httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/models", nil))
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "chat-default") {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}
