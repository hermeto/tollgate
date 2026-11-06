package server

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/internal/identity"
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

func TestIdentityAuthenticationAndPolicy(t *testing.T) {
	provider, err := sim.New(sim.DefaultConfig(sim.FlavorOpenAI), sim.Options{Logger: slog.Default()})
	if err != nil {
		t.Fatal(err)
	}
	providerServer := httptest.NewServer(provider.Handler())
	defer providerServer.Close()
	dir := t.TempDir()
	keyPath := filepath.Join(dir, "keys.json")
	store, err := identity.Open(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	expires := time.Now().Add(time.Hour)
	key, _, err := store.Create("payments", "checkout", &expires)
	if err != nil {
		t.Fatal(err)
	}
	policyDir := filepath.Join(dir, "policies")
	if err := os.MkdirAll(filepath.Join(policyDir, "teams"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(policyDir, "org.yaml"), []byte("enabled: true\nmodels: [gpt-4o-mini]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.Upstream.BaseURL = providerServer.URL
	cfg.Upstream.CredentialEnv = "TOLLGATE_TEST_KEY"
	t.Setenv("TOLLGATE_TEST_KEY", "tg-test-upstream")
	cfg.Identity = config.IdentityConfig{KeyStore: keyPath, PolicyDir: policyDir}
	s, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	request := func(auth, model string) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/v1/chat/completions", strings.NewReader(`{"model":"`+model+`","messages":[{"role":"user","content":"hi"}]}`))
		if auth != "" {
			req.Header.Set("Authorization", "Bearer "+auth)
		}
		response := httptest.NewRecorder()
		s.handle(response, req)
		return response
	}
	if got := request("", "gpt-4o-mini"); got.Code != http.StatusUnauthorized {
		t.Fatalf("missing key status = %d, body = %s", got.Code, got.Body.String())
	}
	if got := request(key, "blocked-model"); got.Code != http.StatusForbidden {
		t.Fatalf("policy denial status = %d, body = %s", got.Code, got.Body.String())
	}
	if got := request(key, "gpt-4o-mini"); got.Code != http.StatusOK || got.Header().Get("X-Tollgate-Team") != "payments" {
		t.Fatalf("authorized request status = %d, team = %q, body = %s", got.Code, got.Header().Get("X-Tollgate-Team"), got.Body.String())
	}
}
