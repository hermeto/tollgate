package server

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hermeto/tollgate/internal/config"
	"github.com/hermeto/tollgate/sim"
)

func BenchmarkPassthrough(b *testing.B) {
	provider, err := sim.New(sim.DefaultConfig(sim.FlavorOpenAI), sim.Options{Logger: slog.Default()})
	if err != nil {
		b.Fatal(err)
	}
	providerServer := httptest.NewServer(provider.Handler())
	defer providerServer.Close()
	cfg := config.Default()
	cfg.Upstream.BaseURL = providerServer.URL
	cfg.Upstream.CredentialEnv = "TOLLGATE_BENCH_KEY"
	b.Setenv("TOLLGATE_BENCH_KEY", "tg-example-bench")
	cfg.Listen.API, cfg.Listen.Ops = ":0", ":1"
	s, err := New(cfg)
	if err != nil {
		b.Fatal(err)
	}
	request := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/chat/completions", strings.NewReader(`{"model":"gpt-4o-mini","messages":[{"role":"user","content":"hi"}]}`))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		response := httptest.NewRecorder()
		s.handle(response, request.Clone(context.Background()))
		if response.Code != 200 {
			b.Fatalf("status %d: %s", response.Code, response.Body.String())
		}
	}
}
