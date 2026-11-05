package repotest

import (
	"bytes"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"
)

// repoRoot is the repository root relative to this package directory.
var repoRoot = filepath.Join("..", "..")

// runScript runs a script from scripts/ and returns stdout, stderr and the exit code.
func runScript(t *testing.T, dir, script string, stdin []byte, args ...string) (string, string, int) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repoRoot, "scripts", script))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), path, args...)
	cmd.Dir = dir
	cmd.Stdin = bytes.NewReader(stdin)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	code := 0
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		code = exitErr.ExitCode()
	} else if err != nil {
		t.Fatalf("run %s: %v", script, err)
	}
	return stdout.String(), stderr.String(), code
}
