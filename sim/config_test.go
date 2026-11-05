package sim

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func writeConfig(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "sim.yaml")
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadConfigDefaults(t *testing.T) {
	cfg, err := LoadConfig("", Overrides{Flavor: FlavorAnthropic})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Port != 9002 || cfg.Host != "127.0.0.1" || cfg.Seed != 42 || cfg.RequestLog.Size != 1000 || cfg.Embeddings.Dimensions != 1536 {
		t.Fatalf("unexpected defaults: %+v", cfg)
	}
	if len(cfg.Models) == 0 || !strings.HasPrefix(cfg.Models[0], "claude-") {
		t.Fatalf("anthropic default models = %v", cfg.Models)
	}
}

func TestLoadConfigFileAndOverrides(t *testing.T) {
	p := filepath.Join("testdata", "config-full.yaml")
	seed := uint64(99)
	cfg, err := LoadConfig(p, Overrides{Port: 9200, Seed: &seed})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Flavor != FlavorOpenAI || cfg.Port != 9200 || cfg.Seed != 99 {
		t.Fatalf("overrides not applied: %+v", cfg)
	}
	if cfg.Latency.TTFT.Value != 200 || cfg.OutputTokens.Kind != "normal" {
		t.Fatalf("file values or defaults lost: %+v", cfg)
	}
	if cfg.Faults.ErrorRates[429] != 0.25 || cfg.Faults.RetryAfterSeconds != 3 {
		t.Fatalf("faults = %+v", cfg.Faults)
	}
	if time.Duration(cfg.Faults.Outage.Duration) != time.Minute || time.Duration(cfg.Faults.Outage.StartAfter) != 10*time.Second {
		t.Fatalf("outage = %+v", cfg.Faults.Outage)
	}
}

func TestLoadConfigErrors(t *testing.T) {
	cases := []struct {
		name, body, wantErr string
		o                   Overrides
	}{
		{"unknown field", "flavor: openai\nbogus: 1\n", "additional properties", Overrides{}},
		{"rates above one", "flavor: openai\nfaults: {error_rates: {500: 0.7}, reset_rate: 0.5}\n", "must be <= 1", Overrides{}},
		{"min above max", "flavor: openai\noutput_tokens: {kind: fixed, value: 5, min: 10, max: 2}\n", "min 10 > max 2", Overrides{}},
		{"no flavor", "port: 9001\n", "flavor must be", Overrides{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadConfig(writeConfig(t, tc.body), tc.o)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("error = %v, want substring %q", err, tc.wantErr)
			}
		})
	}
}

func TestDurationJSONRoundTrip(t *testing.T) {
	d := Duration(1500 * time.Millisecond)
	b, err := d.MarshalJSON()
	if err != nil || string(b) != `"1.5s"` {
		t.Fatalf("marshal = %s, %v", b, err)
	}
	var back Duration
	if err := back.UnmarshalJSON(b); err != nil || back != d {
		t.Fatalf("unmarshal = %v, %v", back, err)
	}
	if err := back.UnmarshalJSON([]byte(`5`)); err == nil {
		t.Fatal("numeric duration accepted")
	}
}
