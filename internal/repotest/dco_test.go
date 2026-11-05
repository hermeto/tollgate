package repotest

import (
	"os/exec"
	"strings"
	"testing"
)

// gitRepo creates a throwaway repository with deterministic identity settings.
func gitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	git(t, dir, "init", "-q", "-b", "main")
	git(t, dir, "config", "user.name", "Test Author")
	git(t, dir, "config", "user.email", "author@example.com")
	git(t, dir, "config", "commit.gpgsign", "false")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "base")
	git(t, dir, "tag", "base")
	return dir
}

func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func TestDCOAllSigned(t *testing.T) {
	dir := gitRepo(t)
	git(t, dir, "commit", "-q", "--allow-empty", "-s", "-m", "feat: one")
	git(t, dir, "commit", "-q", "--allow-empty", "-s", "-m", "fix: two")
	out, _, code := runScript(t, dir, "check-dco.sh", nil, "base", "HEAD")
	if code != 0 {
		t.Fatalf("exit = %d, stdout %q", code, out)
	}
}

func TestDCOMissingSignoff(t *testing.T) {
	dir := gitRepo(t)
	git(t, dir, "commit", "-q", "--allow-empty", "-s", "-m", "feat: signed")
	git(t, dir, "commit", "-q", "--allow-empty", "-m", "fix: unsigned")
	out, _, code := runScript(t, dir, "check-dco.sh", nil, "base", "HEAD")
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
	if !strings.Contains(out, "fix: unsigned: missing Signed-off-by") || strings.Contains(out, "feat: signed:") {
		t.Errorf("stdout = %q", out)
	}
}

func TestDCOSignoffOnlyInBodyIsRejected(t *testing.T) {
	dir := gitRepo(t)
	msg := "feat: tricky\n\nSigned-off-by: Test Author <author@example.com> is quoted here\n\nnot a trailer block line\n"
	git(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
	_, _, code := runScript(t, dir, "check-dco.sh", nil, "base", "HEAD")
	if code != 1 {
		t.Fatalf("exit = %d, want 1", code)
	}
}

func TestDCOIgnoresMergeCommits(t *testing.T) {
	dir := gitRepo(t)
	git(t, dir, "checkout", "-q", "-b", "side")
	git(t, dir, "commit", "-q", "--allow-empty", "-s", "-m", "feat: side")
	git(t, dir, "checkout", "-q", "main")
	git(t, dir, "commit", "-q", "--allow-empty", "-s", "-m", "feat: main")
	git(t, dir, "merge", "-q", "--no-ff", "--no-edit", "side") // merge commit without sign-off
	out, _, code := runScript(t, dir, "check-dco.sh", nil, "base", "HEAD")
	if code != 0 {
		t.Fatalf("exit = %d, stdout %q", code, out)
	}
}

func TestDCOUsage(t *testing.T) {
	_, errOut, code := runScript(t, ".", "check-dco.sh", nil, "only-one")
	if code != 2 || !strings.Contains(errOut, "usage:") {
		t.Fatalf("exit = %d, stderr = %q", code, errOut)
	}
}
