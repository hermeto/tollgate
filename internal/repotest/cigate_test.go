package repotest

import (
	"strings"
	"testing"
)

func TestCIGate(t *testing.T) {
	allSuccess := `{"lint":{"result":"success"},"test":{"result":"success"},"security":{"result":"success"},"sdk-compat":{"result":"success"},"e2e-local":{"result":"success"},` +
		`"pr-title":{"result":"success"},"dco":{"result":"success"},"dependency-review":{"result":"success"}}`
	cases := []struct {
		name     string
		event    string
		needs    string
		wantCode int
		wantOut  string
	}{
		{"all success on PR", "pull_request", allSuccess, 0, "all required jobs passed"},
		{"one failure", "pull_request", `{"lint":{"result":"success"},"test":{"result":"failure"}}`, 1, "test: failure"},
		{"cancelled fails", "pull_request", `{"lint":{"result":"cancelled"}}`, 1, "lint: cancelled"},
		{"skipped fails on PR", "pull_request", `{"lint":{"result":"success"},"dco":{"result":"skipped"}}`, 1, "dco: skipped"},
		{"pr-only skipped on push passes", "push",
			`{"lint":{"result":"success"},"test":{"result":"success"},"e2e-local":{"result":"success"},"pr-title":{"result":"skipped"},"dco":{"result":"skipped"},"dependency-review":{"result":"skipped"}}`,
			0, "all required jobs passed"},
		{"non pr-only skipped on push fails", "push", `{"lint":{"result":"skipped"},"dco":{"result":"skipped"}}`, 1, "lint: skipped"},
		{"pull_request_target treated as PR", "pull_request_target", `{"dco":{"result":"skipped"}}`, 1, "dco: skipped"},
		{"missing result fails", "pull_request", `{"lint":{}}`, 1, "lint: missing"},
		{"empty object fails", "pull_request", `{}`, 1, "no jobs to check"},
		{"invalid json", "pull_request", `not json`, 2, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out, _, code := runScript(t, ".", "ci-gate.sh", []byte(tc.needs), tc.event)
			if code != tc.wantCode {
				t.Fatalf("exit = %d, want %d (stdout %q)", code, tc.wantCode, out)
			}
			if !strings.Contains(out, tc.wantOut) {
				t.Errorf("stdout = %q, want substring %q", out, tc.wantOut)
			}
		})
	}
}

func TestCIGateUsage(t *testing.T) {
	_, errOut, code := runScript(t, ".", "ci-gate.sh", nil)
	if code != 2 || !strings.Contains(errOut, "usage:") {
		t.Fatalf("exit = %d, stderr = %q", code, errOut)
	}
}

func TestCIGateReportsEveryFailure(t *testing.T) {
	out, _, code := runScript(t, ".", "ci-gate.sh",
		[]byte(`{"lint":{"result":"failure"},"test":{"result":"cancelled"}}`), "pull_request")
	if code != 1 {
		t.Fatalf("exit = %d", code)
	}
	for _, want := range []string{"lint: failure", "test: cancelled"} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout %q missing %q", out, want)
		}
	}
}
