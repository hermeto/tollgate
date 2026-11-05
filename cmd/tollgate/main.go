// Command tollgate is the Tollgate LLM gateway binary.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/hermeto/tollgate/internal/version"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

// run executes the CLI and returns the process exit code.
func run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	switch args[0] {
	case "version", "--version", "-v":
		_, _ = fmt.Fprintln(stdout, version.String())
		return 0
	case "help", "--help", "-h":
		usage(stdout)
		return 0
	default:
		_, _ = fmt.Fprintf(stderr, "tollgate: unknown command %q\n\n", args[0])
		usage(stderr)
		return 2
	}
}

func usage(w io.Writer) {
	_, _ = fmt.Fprint(w, "Usage: tollgate <command>\n\n"+
		"Commands:\n"+
		"  version   Print version information\n"+
		"  help      Show this help\n")
}
