// Command tollgate is the Tollgate LLM gateway binary.
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

// usageError marks errors caused by invalid command-line usage (exit code 2).
type usageError struct{ err error }

func (e usageError) Error() string { return e.err.Error() }
func (e usageError) Unwrap() error { return e.err }

// run executes the CLI and returns the process exit code: 0 success, 1 runtime error, 2 usage error.
func run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	root := newRootCmd()
	root.SetArgs(args)
	root.SetOut(stdout)
	root.SetErr(stderr)
	err := root.ExecuteContext(ctx)
	if err == nil {
		return 0
	}
	_, _ = fmt.Fprintf(stderr, "tollgate: %v\n", err)
	var ue usageError
	if errors.As(err, &ue) {
		_, _ = fmt.Fprintf(stderr, "Run 'tollgate --help' for usage.\n")
		return 2
	}
	return 1
}
