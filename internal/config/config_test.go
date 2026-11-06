package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeConfig(t *testing.T, text string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "gateway.yaml")
	if err := os.WriteFile(p, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadAppliesDefaults(t *testing.T) {
	c, err := Load(writeConfig(t, "upstream:\n  base_url: http://127.0.0.1:9001\n"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Listen.API != ":8080" || c.Listen.Ops != ":8081" || c.Limits.MaxBodyBytes != 8*1024*1024 {
		t.Fatalf("defaults = %+v", c)
	}
	if time.Duration(c.Upstream.Timeouts.ResponseHeader) != 60*time.Second || c.Telemetry.ServiceName != "tollgate" {
		t.Fatalf("nested defaults = %+v", c)
	}
}

func TestLoadParsesDurationsAndRejectsUnknownFields(t *testing.T) {
	c, err := Load(writeConfig(t, "listen: {api: ':18080', ops: ':18081'}\nupstream:\n  base_url: https://example.test\n  timeouts: {connect: 250ms, response_header: 3s, idle: 4s}\nshutdown: {drain_timeout: 5s}\n"))
	if err != nil {
		t.Fatal(err)
	}
	if time.Duration(c.Upstream.Timeouts.Connect) != 250*time.Millisecond || time.Duration(c.Shutdown.DrainTimeout) != 5*time.Second {
		t.Fatalf("durations = %+v", c)
	}
	if _, err := Load(writeConfig(t, "upstream: {base_url: http://localhost:1}\nunknown: true\n")); err == nil {
		t.Fatal("unknown field accepted")
	}
}

func TestLoadRejectsInvalidSemanticConfig(t *testing.T) {
	cases := []string{
		"listen: {api: ':8080', ops: ':8080'}\nupstream: {base_url: http://localhost:1}\n",
		"upstream: {base_url: ftp://localhost:1}\n",
		"upstream: {base_url: http://localhost:1}\npipeline: [timing, timing]\n",
	}
	for _, text := range cases {
		if _, err := Load(writeConfig(t, text)); err == nil {
			t.Fatalf("accepted invalid config: %s", text)
		}
	}
}
