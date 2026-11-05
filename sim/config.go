// Package sim implements deterministic OpenAI- and Anthropic-compatible provider simulators with a
// seeded latency model and live fault injection. It is a test double: it never contacts a real
// provider and must not be exposed outside a test or demo environment.
package sim

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.yaml.in/yaml/v4"

	"github.com/hermeto/tollgate/schemas"
)

// Flavor selects the wire protocol a simulator speaks.
type Flavor string

// Supported flavors.
const (
	FlavorOpenAI    Flavor = "openai"
	FlavorAnthropic Flavor = "anthropic"
)

// Config is the simulator configuration (schemas/sim.schema.json).
type Config struct {
	Flavor       Flavor           `json:"flavor"`
	Host         string           `json:"host,omitempty"`
	Port         int              `json:"port,omitempty"`
	Seed         uint64           `json:"seed"`
	Models       []string         `json:"models,omitempty"`
	Latency      LatencyConfig    `json:"latency"`
	OutputTokens Distribution     `json:"output_tokens"`
	Embeddings   EmbeddingsConfig `json:"embeddings"`
	RequestLog   RequestLogConfig `json:"request_log"`
	Faults       FaultConfig      `json:"faults"`
}

// LatencyConfig holds the time-to-first-token and inter-token delay distributions, in milliseconds.
type LatencyConfig struct {
	TTFT       Distribution `json:"ttft"`
	InterToken Distribution `json:"inter_token"`
}

// EmbeddingsConfig configures /v1/embeddings.
type EmbeddingsConfig struct {
	Dimensions int    `json:"dimensions,omitempty"`
	ForwardURL string `json:"forward_url,omitempty"`
}

// RequestLogConfig bounds the in-memory request log; Size 0 disables it.
type RequestLogConfig struct {
	Size int `json:"size"`
}

// Distribution describes a random variable: fixed, normal or lognormal, with optional clamps.
type Distribution struct {
	Kind   string   `json:"kind"`
	Value  float64  `json:"value,omitempty"`
	Mean   float64  `json:"mean,omitempty"`
	Stddev float64  `json:"stddev,omitempty"`
	Mu     float64  `json:"mu,omitempty"`
	Sigma  float64  `json:"sigma,omitempty"`
	Min    *float64 `json:"min,omitempty"`
	Max    *float64 `json:"max,omitempty"`
}

// FaultConfig describes injected failures. Rates are probabilities per request.
type FaultConfig struct {
	ErrorRates              map[int]float64 `json:"error_rates,omitempty"`
	RetryAfterSeconds       int             `json:"retry_after_seconds,omitempty"`
	ResetRate               float64         `json:"reset_rate,omitempty"`
	TimeoutBeforeFirstToken *TimeoutFault   `json:"timeout_before_first_token,omitempty"`
	StallMidStream          *StallFault     `json:"stall_mid_stream,omitempty"`
	Outage                  *OutageWindow   `json:"outage,omitempty"`
}

// TimeoutFault holds the response before the first byte for Duration, with probability Rate.
type TimeoutFault struct {
	Rate     float64  `json:"rate"`
	Duration Duration `json:"duration"`
}

// StallFault pauses a stream after AfterTokens content tokens for Duration, with probability Rate.
type StallFault struct {
	Rate        float64  `json:"rate"`
	AfterTokens int      `json:"after_tokens"`
	Duration    Duration `json:"duration"`
}

// OutageWindow fails every request with Status between Start and End. StartAfter and Duration are
// relative inputs; Start and End are resolved when the window is applied.
type OutageWindow struct {
	StartAfter Duration   `json:"start_after,omitempty"`
	Duration   Duration   `json:"duration"`
	Status     int        `json:"status,omitempty"`
	Start      *time.Time `json:"start,omitempty"`
	End        *time.Time `json:"end,omitempty"`
}

// Duration is a time.Duration encoded as a Go duration string ("5s") in JSON and YAML.
type Duration time.Duration

// MarshalJSON implements json.Marshaler.
func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(time.Duration(d).String())
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return fmt.Errorf("duration must be a string like \"5s\": %w", err)
	}
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(v)
	return nil
}

func ptr(f float64) *float64 { return &f }

// DefaultConfig returns the defaults for a flavor (data-model.md).
func DefaultConfig(f Flavor) Config {
	cfg := Config{
		Flavor: f,
		Host:   "127.0.0.1",
		Seed:   42,
		Latency: LatencyConfig{
			TTFT:       Distribution{Kind: "fixed", Value: 0},
			InterToken: Distribution{Kind: "fixed", Value: 0},
		},
		OutputTokens: Distribution{Kind: "normal", Mean: 120, Stddev: 40, Min: ptr(1), Max: ptr(1024)},
		Embeddings:   EmbeddingsConfig{Dimensions: 1536},
		RequestLog:   RequestLogConfig{Size: 1000},
	}
	switch f {
	case FlavorAnthropic:
		cfg.Port = 9002
		cfg.Models = []string{"claude-opus-4-1", "claude-sonnet-4-5", "claude-haiku-4-5"}
	default:
		cfg.Port = 9001
		cfg.Models = []string{"gpt-4o", "gpt-4o-mini", "gpt-4.1", "gpt-4.1-mini", "text-embedding-3-small", "text-embedding-3-large"}
	}
	return cfg
}

// Overrides are command-line values that take precedence over the configuration file.
type Overrides struct {
	Flavor Flavor
	Host   string
	Port   int
	Seed   *uint64
}

// LoadConfig reads an optional YAML file, validates it against the schema, applies flavor defaults
// and command-line overrides, and checks semantic rules the schema cannot express.
func LoadConfig(path string, o Overrides) (Config, error) {
	var raw any = map[string]any{}
	if path != "" {
		data, err := os.ReadFile(path) // #nosec G304 -- operator-supplied config path
		if err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
		if err := yaml.Unmarshal(data, &raw); err != nil {
			return Config{}, fmt.Errorf("parse config %s: %w", path, err)
		}
		if raw == nil {
			raw = map[string]any{}
		}
	}
	raw = normalizeKeys(raw)
	if err := schemas.Validate("sim.schema.json", raw); err != nil {
		return Config{}, err
	}
	doc, err := json.Marshal(raw)
	if err != nil {
		return Config{}, err
	}
	var head struct {
		Flavor Flavor `json:"flavor"`
	}
	if err := json.Unmarshal(doc, &head); err != nil {
		return Config{}, err
	}
	flavor := head.Flavor
	if o.Flavor != "" {
		flavor = o.Flavor
	}
	if flavor != FlavorOpenAI && flavor != FlavorAnthropic {
		return Config{}, fmt.Errorf("flavor must be %q or %q, got %q", FlavorOpenAI, FlavorAnthropic, flavor)
	}
	cfg := DefaultConfig(flavor)
	if err := json.Unmarshal(doc, &cfg); err != nil {
		return Config{}, fmt.Errorf("decode config: %w", err)
	}
	cfg.Flavor = flavor
	if o.Host != "" {
		cfg.Host = o.Host
	}
	if o.Port != 0 {
		cfg.Port = o.Port
	}
	if o.Seed != nil {
		cfg.Seed = *o.Seed
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// Validate checks rules the JSON Schema cannot express.
func (c Config) Validate() error {
	var errs []error
	for name, d := range map[string]Distribution{
		"latency.ttft": c.Latency.TTFT, "latency.inter_token": c.Latency.InterToken, "output_tokens": c.OutputTokens,
	} {
		if d.Min != nil && d.Max != nil && *d.Min > *d.Max {
			errs = append(errs, fmt.Errorf("%s: min %v > max %v", name, *d.Min, *d.Max))
		}
	}
	if err := c.Faults.Validate(); err != nil {
		errs = append(errs, err)
	}
	if _, ok := c.Faults.ErrorRates[529]; ok && c.Flavor != FlavorAnthropic {
		errs = append(errs, errors.New("faults.error_rates: status 529 (overloaded) is Anthropic-only"))
	}
	if c.Port < 1 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port %d out of range", c.Port))
	}
	return errors.Join(errs...)
}

// Validate checks that the fault probabilities that decide a request's outcome sum to at most 1.
func (f FaultConfig) Validate() error {
	sum := f.ResetRate
	for status, r := range f.ErrorRates {
		switch status {
		case 429, 500, 503, 529:
		default:
			return fmt.Errorf("faults.error_rates: unsupported status %d", status)
		}
		if r < 0 || r > 1 {
			return fmt.Errorf("faults.error_rates[%d]: rate %v outside 0..1", status, r)
		}
		sum += r
	}
	if sum > 1+1e-9 {
		return fmt.Errorf("faults: error_rates plus reset_rate sum to %.3f, must be <= 1", sum)
	}
	return nil
}

// normalizeKeys converts YAML maps with non-string keys (e.g. 429: 0.5) into JSON-compatible maps.
func normalizeKeys(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			t[k] = normalizeKeys(val)
		}
		return t
	case map[any]any:
		m := make(map[string]any, len(t))
		for k, val := range t {
			switch kk := k.(type) {
			case string:
				m[kk] = normalizeKeys(val)
			case int:
				m[strconv.Itoa(kk)] = normalizeKeys(val)
			default:
				m[fmt.Sprint(kk)] = normalizeKeys(val)
			}
		}
		return m
	case []any:
		for i := range t {
			t[i] = normalizeKeys(t[i])
		}
		return t
	default:
		return v
	}
}
