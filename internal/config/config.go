// Package config loads and validates the gateway's YAML configuration.
package config

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/hermeto/tollgate/internal/yamlconf"
)

type Duration time.Duration

func (d *Duration) UnmarshalJSON(data []byte) error {
	var value string
	if err := unmarshalJSON(data, &value); err != nil {
		return fmt.Errorf("duration must be a string: %w", err)
	}
	v, err := time.ParseDuration(value)
	if err != nil {
		return fmt.Errorf("invalid duration %q: %w", value, err)
	}
	*d = Duration(v)
	return nil
}

func (d Duration) String() string { return time.Duration(d).String() }

type Config struct {
	Listen       ListenConfig              `json:"listen"`
	Upstream     UpstreamConfig            `json:"upstream"`
	Limits       LimitsConfig              `json:"limits"`
	Shutdown     ShutdownConfig            `json:"shutdown"`
	Pipeline     []string                  `json:"pipeline"`
	Dependencies []DependencyConfig        `json:"dependencies"`
	Telemetry    TelemetryConfig           `json:"telemetry"`
	Models       map[string][]TargetConfig `json:"models"`
	Identity     IdentityConfig            `json:"identity"`
}
type ListenConfig struct {
	API string `json:"api"`
	Ops string `json:"ops"`
}
type UpstreamConfig struct {
	BaseURL       string        `json:"base_url"`
	CredentialEnv string        `json:"credential_env"`
	Timeouts      TimeoutConfig `json:"timeouts"`
}
type TimeoutConfig struct {
	Connect        Duration `json:"connect"`
	ResponseHeader Duration `json:"response_header"`
	Idle           Duration `json:"idle"`
}
type LimitsConfig struct {
	MaxBodyBytes int64 `json:"max_body_bytes"`
}
type ShutdownConfig struct {
	DrainTimeout Duration `json:"drain_timeout"`
}
type DependencyConfig struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Address  string `json:"address"`
	FailMode string `json:"fail_mode"`
}
type TelemetryConfig struct {
	OTLPEndpoint      string  `json:"otlp_endpoint"`
	ServiceName       string  `json:"service_name"`
	TracesSampleRatio float64 `json:"traces_sample_ratio"`
}
type TargetConfig struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
	BaseURL  string `json:"base_url"`
}
type IdentityConfig struct {
	KeyStore  string `json:"key_store"`
	PolicyDir string `json:"policy_dir"`
}

// Load reads path, applies documented defaults, and validates semantic constraints.
func Load(path string) (Config, error) {
	doc, err := yamlconf.ReadFile(path, "gateway.schema.json")
	if err != nil {
		return Config{}, err
	}
	c := Default()
	if err := unmarshalJSON(doc, &c); err != nil {
		return Config{}, fmt.Errorf("decode gateway config: %w", err)
	}
	if err := c.Validate(); err != nil {
		return Config{}, err
	}
	return c, nil
}

func Default() Config {
	return Config{
		Listen:    ListenConfig{API: ":8080", Ops: ":8081"},
		Limits:    LimitsConfig{MaxBodyBytes: 8 * 1024 * 1024},
		Shutdown:  ShutdownConfig{DrainTimeout: Duration(30 * time.Second)},
		Pipeline:  []string{"request_id", "timing"},
		Telemetry: TelemetryConfig{ServiceName: "tollgate", TracesSampleRatio: 1},
		Models:    map[string][]TargetConfig{},
		Upstream:  UpstreamConfig{Timeouts: TimeoutConfig{Connect: Duration(2 * time.Second), ResponseHeader: Duration(60 * time.Second), Idle: Duration(90 * time.Second)}},
	}
}

func (c Config) Validate() error {
	if c.Listen.API == "" || c.Listen.Ops == "" {
		return fmt.Errorf("listen.api and listen.ops are required")
	}
	if c.Listen.API == c.Listen.Ops {
		return fmt.Errorf("listen.api and listen.ops must differ")
	}
	u, err := url.Parse(c.Upstream.BaseURL)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("upstream.base_url must be an http or https URL")
	}
	if c.Limits.MaxBodyBytes < 1024 || c.Limits.MaxBodyBytes > 64*1024*1024 {
		return fmt.Errorf("limits.max_body_bytes must be between 1024 and 67108864")
	}
	if c.Shutdown.DrainTimeout <= 0 || c.Upstream.Timeouts.Connect <= 0 || c.Upstream.Timeouts.ResponseHeader <= 0 || c.Upstream.Timeouts.Idle <= 0 {
		return fmt.Errorf("timeouts and shutdown.drain_timeout must be positive")
	}
	seen := map[string]bool{}
	for _, stage := range c.Pipeline {
		if stage != "request_id" && stage != "timing" {
			return fmt.Errorf("pipeline: unknown stage %q", stage)
		}
		if seen[stage] {
			return fmt.Errorf("pipeline: duplicate stage %q", stage)
		}
		seen[stage] = true
	}
	for i, dep := range c.Dependencies {
		if dep.Name == "" || dep.Address == "" || (dep.Kind != "redis" && dep.Kind != "postgres") || (dep.FailMode != "open" && dep.FailMode != "closed") {
			return fmt.Errorf("dependencies[%d] is invalid", i)
		}
	}
	for logical, targets := range c.Models {
		if logical == "" || len(targets) == 0 {
			return fmt.Errorf("models[%q] must have at least one target", logical)
		}
		for i, target := range targets {
			if (target.Provider != "openai" && target.Provider != "anthropic") || target.Model == "" || target.BaseURL == "" {
				return fmt.Errorf("models[%q][%d] is invalid", logical, i)
			}
		}
	}
	return nil
}

func unmarshalJSON(data []byte, dst any) error { return json.Unmarshal(data, dst) }
