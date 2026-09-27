// Package cli holds the command line parsing, validation, and orchestration
// for backlog.md-migrate, so that it can be exercised with unit tests
// without invoking the compiled binary. cmd/backlog.md-migrate/main.go stays
// a thin wrapper that calls Run and translates its result into a process
// exit code.
package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"backlog.md-migrate/internal/convert"
	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
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
// valid, runs the conversion end to end: reading the source board, reading
// the destination board, assembling the batch (internal/convert.Assemble),
// writing its NDJSON, and printing every accumulated Finding, following the
// exit code table of docs/especificacion.md, "Códigos de salida".
func runImport(args []string, stdout, stderr io.Writer) int {
	opts, code, ok := ParseImportArgs(args, stdout, stderr)
	if !ok {
		return code
	}

	// docs/especificacion.md, "Qué lee del origen": source.Read treats a
	// missing tasks/ directory as an empty board, which is correct for its
	// own read-anything-that-is-there purpose but not for import's exit
	// code 3. This check runs before source.Read for that reason, and
	// distinguishes three different conditions rather than collapsing them
	// into one message: <backlog-dir> itself missing or not a directory,
	// <backlog-dir> present but with no tasks/ subdirectory, and any other
	// os.Stat failure on tasks/ (a permission error, for example), which is
	// reported with its own underlying error rather than described as if
	// tasks/ were simply absent.
	if info, err := os.Stat(opts.BacklogDir); err != nil {
		fmt.Fprintf(stderr, "backlog.md-migrate: %s does not exist: %s\n", opts.BacklogDir, err)
		return 3
	} else if !info.IsDir() {
		fmt.Fprintf(stderr, "backlog.md-migrate: %s is not a directory\n", opts.BacklogDir)
		return 3
	}

	tasksDir := filepath.Join(opts.BacklogDir, "tasks")
	if info, err := os.Stat(tasksDir); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(stderr, "backlog.md-migrate: %s has no tasks/ directory\n", opts.BacklogDir)
		} else {
			fmt.Fprintf(stderr, "backlog.md-migrate: could not check %s: %s\n", tasksDir, err)
		}
		return 3
	} else if !info.IsDir() {
		fmt.Fprintf(stderr, "backlog.md-migrate: %s exists but is not a directory\n", tasksDir)
		return 3
	}

	sourceBoard, err := source.Read(opts.BacklogDir)
	if err != nil {
		fmt.Fprintf(stderr, "backlog.md-migrate: could not read the source board: %s\n", err)
		return 3
	}

	destBoard, err := destination.Read(opts.Biso, opts.Project)
	if err != nil {
		fmt.Fprintf(stderr, "backlog.md-migrate: could not read the destination board: %s\n", err)
		return 4
	}

	lines, convertFindings, err := convert.Assemble(sourceBoard, destBoard)
	if err != nil {
		fmt.Fprintf(stderr, "backlog.md-migrate: could not resolve source identifiers: %s\n", err)
		return 3
	}

	// docs/especificacion.md, "Los hallazgos": the source board's own
	// findings (raised while reading it) come first, since they happened
	// first in the pipeline, followed by every finding the conversion
	// engine raised (converting each task, then resolving identifiers, in
	// the deterministic order Assemble already produced them).
	findings := append(append([]source.Finding(nil), sourceBoard.Findings...), convertFindings...)

	if opts.Strict && len(findings) > 0 {
		// docs/especificacion.md, "Códigos de salida": with --strict and
		// findings present, nothing is written at all, not even a
		// temporary file, only the findings on stderr and exit code 5.
		printFindings(stderr, findings)
		return 5
	}

	if err := writeNDJSON(lines, opts.Out, stdout); err != nil {
		fmt.Fprintf(stderr, "backlog.md-migrate: could not write the NDJSON output: %s\n", err)
		// docs/especificacion.md, "Códigos de salida", code 1: the residual
		// code for a write failure unrelated to the source, the
		// destination, or findings. The findings already accumulated (from
		// reading the source and converting it) are still worth printing
		// here, even though the write itself failed, since they are useful
		// for diagnosing the run.
		printFindings(stderr, findings)
		return 1
	}

	printFindings(stderr, findings)
	if len(findings) > 0 {
		return 5
	}
	return 0
}

// printFindings writes every finding to w, one per line, in the format
// source.Finding.String already implements
// (docs/especificacion.md, "Los hallazgos").
func printFindings(w io.Writer, findings []source.Finding) {
	for _, f := range findings {
		fmt.Fprintln(w, f.String())
	}
}

// writeNDJSON writes lines as NDJSON (docs/especificacion.md, "La salida")
// to out, or to stdout when out is "-". Writing to a real file never leaves
// a partial file behind: it is written to a temporary file in out's own
// directory first (so the final rename stays on the same filesystem) and
// renamed into place only once every line has been written successfully
// (docs/especificacion.md, "Códigos de salida", closing note). Any error
// leaves out untouched and removes the temporary file.
func writeNDJSON(lines []convert.Line, out string, stdout io.Writer) error {
	if out == "" || out == "-" {
		return convert.EncodeNDJSON(stdout, lines)
	}

	dir := filepath.Dir(out)
	tmp, err := os.CreateTemp(dir, ".backlog.md-migrate-*.ndjson.tmp")
	if err != nil {
		return fmt.Errorf("creating a temporary file next to %s: %w", out, err)
	}
	tmpName := tmp.Name()
	// If anything below fails before the rename, or the rename itself
	// fails, this removes the leftover temporary file; once the rename
	// succeeds, tmpName no longer exists and Remove is a silent no-op.
	defer os.Remove(tmpName)

	if err := convert.EncodeNDJSON(tmp, lines); err != nil {
		tmp.Close()
		return fmt.Errorf("writing %s: %w", out, err)
	}
	// os.CreateTemp creates the file with mode 0600 regardless of the
	// process umask, since it is meant for private scratch files; this one
	// is about to become the final output file, so it gets the ordinary
	// 0644 permissions a written file would otherwise have, set explicitly
	// rather than left to whatever the process umask happens to be.
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		return fmt.Errorf("setting permissions on the temporary file for %s: %w", out, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temporary file for %s: %w", out, err)
	}

	if err := os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("renaming temporary file to %s: %w", out, err)
	}
	return nil
}
