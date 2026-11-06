// Package pricing loads effective-dated model prices and computes token costs.
package pricing

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"go.yaml.in/yaml/v4"
)

type Price struct {
	Model                    string    `yaml:"model"`
	EffectiveAt              time.Time `yaml:"effective_at"`
	Currency                 string    `yaml:"currency"`
	InputUSDPerMillion       float64   `yaml:"input_usd_per_million"`
	OutputUSDPerMillion      float64   `yaml:"output_usd_per_million"`
	CachedInputUSDPerMillion float64   `yaml:"cached_input_usd_per_million"`
	Source                   string    `yaml:"source"`
}
type table struct {
	Provider string  `yaml:"provider"`
	Prices   []Price `yaml:"prices"`
}
type Table struct {
	Provider string
	prices   map[string][]Price
}

func Load(dir string) (*Table, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "*.yaml"))
	if err != nil {
		return nil, err
	}
	t := &Table{prices: map[string][]Price{}}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var in table
		if err := yaml.Unmarshal(body, &in); err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}
		if in.Provider == "" || len(in.Prices) == 0 {
			return nil, fmt.Errorf("%s: provider and prices are required", path)
		}
		for _, p := range in.Prices {
			if p.Model == "" || p.Currency != "USD" || p.EffectiveAt.IsZero() || p.InputUSDPerMillion < 0 || p.OutputUSDPerMillion < 0 || p.CachedInputUSDPerMillion < 0 {
				return nil, fmt.Errorf("%s: invalid price for %s", path, p.Model)
			}
			t.prices[p.Model] = append(t.prices[p.Model], p)
		}
	}
	for model := range t.prices {
		sort.Slice(t.prices[model], func(i, j int) bool { return t.prices[model][i].EffectiveAt.Before(t.prices[model][j].EffectiveAt) })
	}
	return t, nil
}

func (t *Table) Lookup(model string, at time.Time) (Price, bool) {
	var out Price
	for _, p := range t.prices[model] {
		if p.EffectiveAt.After(at) {
			break
		}
		out = p
	}
	return out, out.Model != ""
}
func (p Price) Cost(input, output, cached int) float64 {
	return float64(input-cached)*p.InputUSDPerMillion/1e6 + float64(cached)*p.CachedInputUSDPerMillion/1e6 + float64(output)*p.OutputUSDPerMillion/1e6
}
