// Package cli holds the command line parsing and validation for
// backlog.md-migrate, so that it can be exercised with unit tests without
// invoking the compiled binary. cmd/backlog.md-migrate/main.go stays a thin
// wrapper that calls Run and translates its result into a process exit
// code.
package cli

import (
	"fmt"
	"io"
)

// topLevelUsage is printed on stderr whenever the top level command line is
// wrong: no subcommand at all, or a subcommand other than "import".
const topLevelUsage = "usage: backlog.md-migrate import <backlog-dir> --project <dir> [--out <file|->] [--biso <path>] [--strict]\n"

// Run parses args (the program arguments without the program name itself)
// and returns the process exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, topLevelUsage)
		return 2
	}

	switch args[0] {
	case "import":
		return runImport(args[1:], stdout, stderr)
	default:
		fmt.Fprintf(stderr, "backlog.md-migrate: unknown command %q\n", args[0])
		fmt.Fprint(stderr, topLevelUsage)
		return 2
	}
}

// runImport parses the import subcommand arguments and, once they are
// valid, runs the (not yet implemented) conversion.
func runImport(args []string, stdout, stderr io.Writer) int {
	_, code, ok := ParseImportArgs(args, stdout, stderr)
	if !ok {
		return code
	}

	fmt.Fprintln(stderr, "backlog.md-migrate: import not implemented yet")
	// Exit code 1 is a phase 1 scaffolding placeholder, not part of the
	// specification (docs/especificacion.md only defines 0, 2, 3, 4 and 5
	// for import). It disappears once phase 4/5 implement the real
	// conversion, which will exit with one of those codes instead.
	return 1
}
