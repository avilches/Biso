package cli

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
)

// ImportOptions holds the parsed arguments of the import subcommand.
type ImportOptions struct {
	// BacklogDir is the Backlog.md data directory, the one holding tasks/.
	BacklogDir string
	// Project is the directory biso resolves the destination board from.
	Project string
	// Out is where the NDJSON is written, "-" meaning standard output.
	Out string
	// Biso is the biso binary to run.
	Biso string
	// Strict makes import write nothing and exit with code 5 on findings.
	Strict bool
}

// importValueFlags are the import flags that take a value, keyed by every
// spelling the flag package accepts ("-name" and "--name"). It drives the
// positional/flag split below, since the standard flag package stops
// parsing at the first non-flag argument and the import signature puts the
// positional <backlog-dir> before the flags.
var importValueFlags = map[string]bool{
	"-project": true, "--project": true,
	"-out": true, "--out": true,
	"-biso": true, "--biso": true,
}

const importHelp = `Usage:
  backlog.md-migrate import <backlog-dir> --project <dir> [--out <file|->] [--biso <path>] [--strict]

Imports a Backlog.md board into a biso board, writing NDJSON for
"biso new --from" on the given output.

Arguments:
  <backlog-dir>     Backlog.md data directory, the one containing tasks/

Flags:
  --project <dir>   Directory biso resolves the destination board from (required)
  --out <file|->    Where to write the NDJSON (default: - for standard output)
  --biso <path>     The biso binary to run (default: biso from PATH)
  --strict          If there are findings, write nothing and exit with code 5
`

// ParseImportArgs parses the arguments of the import subcommand (the
// program arguments after "import").
//
// When --help or -h is given, it writes the help text to stdout and
// returns ok=false with exitCode=0. On a usage error (a missing required
// argument, or an unknown flag) it writes a usage message to stderr and
// returns ok=false with exitCode=2. Otherwise it returns the parsed
// options with ok=true.
func ParseImportArgs(args []string, stdout, stderr io.Writer) (opts ImportOptions, exitCode int, ok bool) {
	positional, flagArgs := splitImportArgs(args)

	fs := flag.NewFlagSet("import", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	out := fs.String("out", "-", "where to write the NDJSON")
	biso := fs.String("biso", "biso", "the biso binary to run")
	project := fs.String("project", "", "directory biso resolves the destination board from")
	strict := fs.Bool("strict", false, "write nothing and exit with code 5 on findings")

	if err := fs.Parse(flagArgs); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, importHelp)
			return ImportOptions{}, 0, false
		}
		fmt.Fprintf(stderr, "backlog.md-migrate: %s\n", err)
		fmt.Fprint(stderr, topLevelUsage)
		return ImportOptions{}, 2, false
	}

	// Anything the flag package still treated as a non-flag argument is an
	// unrecognized flag that splitImportArgs could not attribute a value
	// to, or a genuine extra positional argument. Either way it is folded
	// into the positional list so the checks below report it consistently.
	positional = append(positional, fs.Args()...)

	if len(positional) == 0 {
		fmt.Fprintln(stderr, "backlog.md-migrate: missing <backlog-dir>")
		fmt.Fprint(stderr, topLevelUsage)
		return ImportOptions{}, 2, false
	}
	if len(positional) > 1 {
		fmt.Fprintf(stderr, "backlog.md-migrate: unexpected argument %q\n", positional[1])
		fmt.Fprint(stderr, topLevelUsage)
		return ImportOptions{}, 2, false
	}
	if *project == "" {
		fmt.Fprintln(stderr, "backlog.md-migrate: missing required flag --project")
		fmt.Fprint(stderr, topLevelUsage)
		return ImportOptions{}, 2, false
	}

	return ImportOptions{
		BacklogDir: positional[0],
		Project:    *project,
		Out:        *out,
		Biso:       *biso,
		Strict:     *strict,
	}, 0, true
}

// splitImportArgs separates the positional arguments from the flag
// arguments in args. It exists because the standard flag package stops
// parsing flags at the first non-flag argument, while the import signature
// places the positional <backlog-dir> before the flags
// (backlog.md-migrate import <backlog-dir> --project <dir> ...).
//
// It only needs to know which flags take a value, so it can also pull
// along that value instead of misreading it as positional; it does not
// need to know every valid flag name, since an unrecognized flag is still
// reported correctly by flag.FlagSet.Parse afterwards.
func splitImportArgs(args []string) (positional, flagArgs []string) {
	for i := 0; i < len(args); i++ {
		a := args[i]

		if a == "--" {
			positional = append(positional, args[i+1:]...)
			break
		}

		if !strings.HasPrefix(a, "-") || a == "-" {
			positional = append(positional, a)
			continue
		}

		flagArgs = append(flagArgs, a)

		name := a
		if idx := strings.Index(a, "="); idx >= 0 {
			name = a[:idx]
		}
		if importValueFlags[name] && !strings.Contains(a, "=") && i+1 < len(args) {
			i++
			flagArgs = append(flagArgs, args[i])
		}
	}
	return positional, flagArgs
}
