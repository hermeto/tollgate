// Package version holds build metadata injected at link time with -ldflags.
package version

import "fmt"

// Build metadata. GoReleaser overrides these with -X flags.
var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// String returns the human-readable version line printed by `tollgate version`.
func String() string {
	return fmt.Sprintf("tollgate %s (commit %s, built %s)", Version, Commit, Date)
}
