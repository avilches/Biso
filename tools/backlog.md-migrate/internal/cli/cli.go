// Package cli holds the command line parsing, validation, and orchestration
// for backlog.md-migrate, so that it can be exercised with unit tests
// without invoking the compiled binary. cmd/backlog.md-migrate/main.go stays
// a thin wrapper that calls Run and translates its result into a process
// exit code.
package cli

import (
	"encoding/json"
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

	// docs/especificacion.md, "Qué lee del origen", and the task-70 phase 5
	// brief: source.Read treats a missing tasks/ directory as an empty
	// board, which is correct for its own read-anything-that-is-there
	// purpose but not for import's exit code 3 ("no tiene tasks/"). This
	// check runs before source.Read for that reason, and it also covers
	// <backlog-dir> itself not existing or not being a directory, since a
	// missing or non-directory <backlog-dir> can never contain a tasks/
	// directory either.
	tasksDir := filepath.Join(opts.BacklogDir, "tasks")
	if info, err := os.Stat(tasksDir); err != nil || !info.IsDir() {
		fmt.Fprintf(stderr, "backlog.md-migrate: %s has no tasks/ directory\n", opts.BacklogDir)
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
	// findings (phase 2, reading) come first, since they happened first in
	// the pipeline, followed by every finding the conversion engine raised
	// (phases 4a and 4b, in the deterministic order Assemble already
	// produced them).
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
		// Not one of docs/especificacion.md's own exit codes: writing the
		// output failing (a full disk, a permission error) is outside the
		// three input/output conditions the specification's exit code
		// table names (source unreadable, destination unresponsive,
		// findings present). This is a deliberate implementer decision,
		// reported alongside the rest of phase 5's work: a generic
		// unexpected-failure code, distinct from every code the
		// specification does define.
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
		return encodeNDJSON(stdout, lines)
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

	if err := encodeNDJSON(tmp, lines); err != nil {
		tmp.Close()
		return fmt.Errorf("writing %s: %w", out, err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("closing temporary file for %s: %w", out, err)
	}

	if err := os.Rename(tmpName, out); err != nil {
		return fmt.Errorf("renaming temporary file to %s: %w", out, err)
	}
	return nil
}

// encodeNDJSON writes one compact JSON object per line to w, in convert.Line
// order, disabling HTML escaping so a title, description, or comment body
// containing '<', '>', or '&' is not silently rewritten into a Unicode
// escape sequence: this output is NDJSON for "biso new --from", never HTML.
func encodeNDJSON(w io.Writer, lines []convert.Line) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, line := range lines {
		if err := enc.Encode(line); err != nil {
			return err
		}
	}
	return nil
}
