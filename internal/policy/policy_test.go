package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInheritanceAndReloadRetention(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "teams"), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "org.yaml"), []byte("enabled: true\nmodels: [gpt-4o-mini]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "teams", "payments.yaml"), []byte("models: [gpt-4o-mini, claude-sonnet-5-5]\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	l, err := NewLoader(dir)
	if err != nil {
		t.Fatal(err)
	}
	p := l.Snapshot().Effective("payments", "checkout")
	if len(p.Models) != 2 || p.Provenance["models"] != "teams/payments.yaml" {
		t.Fatalf("policy = %+v", p)
	}
	if err := os.WriteFile(filepath.Join(dir, "teams", "payments.yaml"), []byte("mode: invalid\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := l.Reload(); err == nil {
		t.Fatal("invalid reload accepted")
	}
	if allowed, _ := l.Snapshot().Allows("payments", "checkout", "gpt-4o-mini"); !allowed {
		t.Fatal("previous snapshot was lost")
	}
}
