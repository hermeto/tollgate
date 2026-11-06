package yamlconf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestToJSONNormalizesAndValidates(t *testing.T) {
	out, err := ToJSON([]byte("flavor: openai\nfaults:\n  error_rates: {429: 0.5}\n"), "sim.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(out), `"429":0.5`) {
		t.Fatalf("json = %s", out)
	}
	if _, err := ToJSON([]byte("bogus: 1\n"), "sim.schema.json"); err == nil {
		t.Fatal("schema violation accepted")
	}
	if _, err := ToJSON([]byte(":\n- ["), "sim.schema.json"); err == nil {
		t.Fatal("invalid YAML accepted")
	}
	if out, err := ToJSON(nil, "sim.schema.json"); err != nil || string(out) != "{}" {
		t.Fatalf("empty = %s, %v", out, err)
	}
}

func TestReadFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "c.yaml")
	if err := os.WriteFile(p, []byte("flavor: anthropic\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if out, err := ReadFile(p, "sim.schema.json"); err != nil || !strings.Contains(string(out), "anthropic") {
		t.Fatalf("ReadFile = %s, %v", out, err)
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing.yaml"), "sim.schema.json"); err == nil {
		t.Fatal("missing file accepted")
	}
	if out, err := ReadFile("", "sim.schema.json"); err != nil || string(out) != "{}" {
		t.Fatalf("empty path = %s, %v", out, err)
	}
}
