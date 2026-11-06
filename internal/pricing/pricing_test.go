package pricing

import (
	"math"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadAndEffectiveCost(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "openai.yaml"), []byte("provider: openai\nprices:\n- model: m\n  effective_at: 2026-01-01T00:00:00Z\n  currency: USD\n  input_usd_per_million: 1\n  output_usd_per_million: 2\n  cached_input_usd_per_million: 0.5\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	tab, err := Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	p, ok := tab.Lookup("m", time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC))
	if !ok || math.Abs(p.Cost(1000, 500, 200)-0.0019) > 1e-12 {
		t.Fatalf("price=%+v cost=%v", p, p.Cost(1000, 500, 200))
	}
}
