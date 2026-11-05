package main

import (
	"bytes"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()
	var out, errOut bytes.Buffer
	code = run(t.Context(), args, &out, &errOut)
	return out.String(), errOut.String(), code
}

func TestRunVersion(t *testing.T) {
	for _, arg := range []string{"version", "--version", "-v"} {
		out, _, code := runCLI(t, arg)
		if code != 0 {
			t.Fatalf("run(%q) exit = %d, want 0", arg, code)
		}
		if want := "tollgate dev (commit none, built unknown)\n"; out != want {
			t.Errorf("run(%q) stdout = %q, want %q", arg, out, want)
		}
	}
}

func TestRunUnknownCommand(t *testing.T) {
	_, errOut, code := runCLI(t, "bogus")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, `unknown command "bogus"`) {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	_, errOut, code := runCLI(t)
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
	if !strings.Contains(errOut, "Usage:") || !strings.Contains(errOut, "tollgate [command]") {
		t.Errorf("stderr = %q", errOut)
	}
}

func TestRunBadFlagIsUsageError(t *testing.T) {
	_, _, code := runCLI(t, "version", "--nope")
	if code != 2 {
		t.Fatalf("exit = %d, want 2", code)
	}
}
