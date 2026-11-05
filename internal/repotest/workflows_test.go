package repotest

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"go.yaml.in/yaml/v4"
)

type workflow struct {
	Permissions any            `yaml:"permissions"`
	Concurrency any            `yaml:"concurrency"`
	Jobs        map[string]job `yaml:"jobs"`
}

type job struct {
	Uses           string `yaml:"uses"`
	TimeoutMinutes int    `yaml:"timeout-minutes"`
	Permissions    any    `yaml:"permissions"`
}

func workflowFiles(t *testing.T) []string {
	t.Helper()
	files, err := filepath.Glob(filepath.Join(repoRoot, ".github", "workflows", "*.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("no workflow files found")
	}
	return files
}

// usesLine matches a `uses:` reference to a remote action.
var usesLine = regexp.MustCompile(`^\s*-?\s*uses:\s*([^\s#]+)\s*(#.*)?$`)

// pinned matches owner/repo[/path]@<40-hex sha>.
var pinned = regexp.MustCompile(`^[\w.-]+/[\w.-]+(/[\w./-]+)?@[0-9a-f]{40}$`)

var versionComment = regexp.MustCompile(`^#\s*v\d+(\.\d+)*`)

func TestActionsArePinnedBySHA(t *testing.T) {
	for _, f := range workflowFiles(t) {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(data), "\n") {
			m := usesLine.FindStringSubmatch(line)
			if m == nil || strings.HasPrefix(m[1], "./") {
				continue // local reusable workflow or not a uses line
			}
			if !pinned.MatchString(m[1]) {
				t.Errorf("%s:%d: %q is not pinned to a full commit SHA", filepath.Base(f), i+1, m[1])
			}
			if !versionComment.MatchString(strings.TrimSpace(m[2])) {
				t.Errorf("%s:%d: %q lacks a version comment (# vX.Y.Z)", filepath.Base(f), i+1, m[1])
			}
		}
	}
}

func TestWorkflowsDeclarePermissionsConcurrencyAndTimeouts(t *testing.T) {
	for _, f := range workflowFiles(t) {
		data, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		var wf workflow
		if err := yaml.Unmarshal(data, &wf); err != nil {
			t.Fatalf("%s: %v", f, err)
		}
		name := filepath.Base(f)
		if wf.Permissions == nil {
			t.Errorf("%s: missing top-level permissions", name)
		} else if m, ok := wf.Permissions.(map[string]any); !ok || len(m) != 0 {
			t.Errorf("%s: top-level permissions must be {} (grant per job)", name)
		}
		if wf.Concurrency == nil {
			t.Errorf("%s: missing concurrency group", name)
		}
		for id, j := range wf.Jobs {
			if j.Uses != "" {
				continue // reusable workflow call: timeout lives in the called workflow
			}
			if j.TimeoutMinutes <= 0 {
				t.Errorf("%s: job %q has no timeout-minutes", name, id)
			}
			if j.Permissions == nil {
				t.Errorf("%s: job %q has no explicit permissions", name, id)
			}
		}
	}
}

func TestCIGateNeedsEveryRequiredJob(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatal(err)
	}
	var wf struct {
		Jobs map[string]struct {
			If    string   `yaml:"if"`
			Needs []string `yaml:"needs"`
		} `yaml:"jobs"`
	}
	if err := yaml.Unmarshal(data, &wf); err != nil {
		t.Fatal(err)
	}
	gate, ok := wf.Jobs["ci-gate"]
	if !ok {
		t.Fatal("ci.yml has no ci-gate job")
	}
	if gate.If != "always()" {
		t.Errorf("ci-gate must run with if: always(), got %q", gate.If)
	}
	needed := map[string]bool{}
	for _, n := range gate.Needs {
		needed[n] = true
	}
	for id := range wf.Jobs {
		if id != "ci-gate" && !needed[id] {
			t.Errorf("job %q is not aggregated by ci-gate", id)
		}
	}
}
