// Package policy loads deterministic org/team/project policy snapshots.
package policy

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sync/atomic"

	"github.com/hermeto/tollgate/schemas"

	"go.yaml.in/yaml/v4"
)

type Policy struct {
	Enabled    bool              `yaml:"enabled" json:"enabled"`
	Mode       string            `yaml:"mode" json:"mode"`
	Models     []string          `yaml:"models" json:"models"`
	Provenance map[string]string `yaml:"-" json:"provenance,omitempty"`
}
type Snapshot struct {
	Org      Policy
	Teams    map[string]Policy
	Projects map[string]map[string]Policy
}
type Loader struct {
	dir     string
	current atomic.Pointer[Snapshot]
}

func Load(dir string) (*Snapshot, error) {
	org, err := readPolicy(filepath.Join(dir, "org.yaml"))
	if err != nil {
		return nil, err
	}
	snap := &Snapshot{Org: org, Teams: map[string]Policy{}, Projects: map[string]map[string]Policy{}}
	teamFiles, err := filepath.Glob(filepath.Join(dir, "teams", "*.yaml"))
	if err != nil {
		return nil, err
	}
	for _, path := range teamFiles {
		p, err := readPolicy(path)
		if err != nil {
			return nil, err
		}
		name := filepath.Base(path)
		snap.Teams[name[:len(name)-5]] = p
	}
	projectFiles, err := filepath.Glob(filepath.Join(dir, "projects", "*", "*.yaml"))
	if err != nil {
		return nil, err
	}
	for _, path := range projectFiles {
		team := filepath.Base(filepath.Dir(path))
		name := filepath.Base(path)
		p, err := readPolicy(path)
		if err != nil {
			return nil, err
		}
		if snap.Projects[team] == nil {
			snap.Projects[team] = map[string]Policy{}
		}
		snap.Projects[team][name[:len(name)-5]] = p
	}
	return snap, nil
}
func (s *Snapshot) Effective(team, project string) Policy {
	out := clone(s.Org)
	merge(&out, s.Org, "org.yaml")
	if p, ok := s.Teams[team]; ok {
		merge(&out, p, "teams/"+team+".yaml")
	}
	if p, ok := s.Projects[team][project]; ok {
		merge(&out, p, "projects/"+team+"/"+project+".yaml")
	}
	return out
}
func (s *Snapshot) Allows(team, project, model string) (bool, bool) {
	p := s.Effective(team, project)
	if len(p.Models) == 0 {
		return true, p.Mode == "dry-run"
	}
	return slices.Contains(p.Models, model), p.Mode == "dry-run"
}
func NewLoader(dir string) (*Loader, error) {
	snap, err := Load(dir)
	if err != nil {
		return nil, err
	}
	l := &Loader{dir: dir}
	l.current.Store(snap)
	return l, nil
}
func (l *Loader) Snapshot() *Snapshot { return l.current.Load() }
func (l *Loader) Reload() error {
	snap, err := Load(l.dir)
	if err != nil {
		return err
	}
	l.current.Store(snap)
	return nil
}
func readPolicy(path string) (Policy, error) {
	body, err := os.ReadFile(path) //nolint:gosec // path is constrained to the configured policy directory.
	if err != nil {
		return Policy{}, fmt.Errorf("read policy %s: %w", path, err)
	}
	var p Policy
	if err := yaml.Unmarshal(body, &p); err != nil {
		return Policy{}, fmt.Errorf("parse policy %s: %w", path, err)
	}
	if p.Mode == "" {
		p.Mode = "enforce"
	}
	if err := schemas.Validate("policy.schema.json", p); err != nil {
		return Policy{}, fmt.Errorf("validate policy %s: %w", path, err)
	}
	if p.Mode != "enforce" && p.Mode != "dry-run" {
		return Policy{}, fmt.Errorf("policy %s: mode must be enforce or dry-run", path)
	}
	return p, nil
}
func clone(p Policy) Policy {
	p.Models = append([]string(nil), p.Models...)
	p.Provenance = map[string]string{}
	return p
}
func merge(dst *Policy, src Policy, source string) {
	if src.Enabled {
		dst.Enabled = true
		dst.Provenance["enabled"] = source
	}
	if src.Mode != "" {
		dst.Mode = src.Mode
		dst.Provenance["mode"] = source
	}
	if src.Models != nil {
		dst.Models = append([]string(nil), src.Models...)
		dst.Provenance["models"] = source
	}
}
