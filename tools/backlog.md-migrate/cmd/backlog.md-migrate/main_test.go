package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// buildBinary compiles this command into a temporary directory and returns
// its path, failing the test on a build error.
func buildBinary(t *testing.T) string {
	t.Helper()

	binPath := filepath.Join(t.TempDir(), "backlog.md-migrate")
	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %s\n%s", err, out)
	}
	return binPath
}

// TestBinaryIsWiredUp builds the binary once and runs it, to check that
// main.go actually calls into internal/cli and turns its result into a
// process exit code. The exhaustive argument matrix lives in
// internal/cli, which is faster to run and easier to debug.
func TestBinaryIsWiredUp(t *testing.T) {
	binPath := buildBinary(t)

	cmd := exec.Command(binPath, "import", "./backlog", "--project", ".")
	out, err := cmd.CombinedOutput()

	exitErr, isExitErr := err.(*exec.ExitError)
	if err != nil && !isExitErr {
		t.Fatalf("could not run the binary: %s", err)
	}
	code := 0
	if isExitErr {
		code = exitErr.ExitCode()
	}
	// "./backlog" does not exist under this package's own directory, so it
	// has no tasks/ subdirectory either: docs/especificacion.md, "Códigos de
	// salida", maps that to exit code 3. The exhaustive exit code matrix
	// lives in internal/cli, which is faster to run and easier to debug.
	if code != 3 {
		t.Fatalf("got exit code %d, want 3, output: %s", code, out)
	}
	if !strings.Contains(string(out), "tasks/") {
		t.Fatalf("expected a message about the missing tasks/ directory, got: %s", out)
	}
}

// findRepoRoot returns this repository's root directory, computed from this
// test file's own path via runtime.Caller, the same technique
// internal/destination/destination_test.go's findBisoBinary uses.
func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not report this file's path")
	}
	// thisFile is <repo>/tools/backlog.md-migrate/cmd/backlog.md-migrate/main_test.go.
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
}

// exitCode reports the process exit code a command's error represents: 0
// for a nil error (the command exited cleanly), the wrapped code for an
// *exec.ExitError, or a fatal test failure for any other kind of error
// (the command could not even be started).
func exitCode(t *testing.T, err error) int {
	t.Helper()
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	t.Fatalf("could not run the command: %s", err)
	return -1
}

// runCLI runs bin with args, optionally inside dir (dir == "" leaves the
// working directory as this test binary's own), and fails the test, with
// its combined output, if it exits with anything other than 0.
func runCLI(t *testing.T, bin, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %s: %v\n%s", bin, strings.Join(args, " "), err, out)
	}
}

// TestEndToEndSmokeAgainstRealBisoAndBacklogCLIs is phase 5's own smoke
// test, task-70 phase 5 brief, "Pruebas": it exercises the compiled
// backlog.md-migrate binary end to end against a real Backlog.md source
// board (generated with the actual backlog CLI when it is on PATH) and a
// real biso destination board (bin/biso, built by the repository's own
// Makefile), and confirms the NDJSON import produces is accepted by
// "biso new --from --dry-run" with no errors. It only confirms the wiring
// works; the rich test matrix for every acceptance criterion #7 case is
// phase 6's job, not this test's.
func TestEndToEndSmokeAgainstRealBisoAndBacklogCLIs(t *testing.T) {
	repoRoot := findRepoRoot(t)
	biso := filepath.Join(repoRoot, "bin", "biso")
	if _, err := os.Stat(biso); err != nil {
		t.Skip("bin/biso not found; run `make build` from the repository root first")
	}

	backlogCLI, err := exec.LookPath("backlog")
	if err != nil {
		t.Skip("backlog CLI not found on PATH, cannot generate a real source board for this smoke test")
	}

	// A minimal real Backlog.md source board, generated with the actual
	// Backlog.md CLI rather than a hand-written fixture, per the task-70
	// phase 5 brief's own preference for this end-to-end test.
	sourceRoot := t.TempDir()
	runCLI(t, backlogCLI, sourceRoot, "init", "Smoke", "--defaults", "--no-git", "--agent-instructions", "none")
	runCLI(t, backlogCLI, sourceRoot, "task", "create", "Smoke test task", "--plain")

	// A minimal real biso destination board.
	destRoot := t.TempDir()
	project := filepath.Join(destRoot, "project")
	board := filepath.Join(destRoot, "board")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", project, err)
	}
	runCLI(t, biso, "", "-C", project, "init", "Fixture", "--prefix", "BISO", "--at", board)

	binPath := buildBinary(t)
	outFile := filepath.Join(t.TempDir(), "tasks.ndjson")

	importCmd := exec.Command(
		binPath, "import", filepath.Join(sourceRoot, "backlog"),
		"--project", project, "--biso", biso, "--out", outFile,
	)
	out, err := importCmd.CombinedOutput()
	if code := exitCode(t, err); code != 0 {
		t.Fatalf("backlog.md-migrate import failed with exit code %d: %s", code, out)
	}

	dryRun := exec.Command(biso, "--cwd", project, "new", "--from", outFile, "--dry-run")
	dryRunOut, err := dryRun.CombinedOutput()
	if code := exitCode(t, err); code != 0 {
		t.Fatalf("biso new --from --dry-run rejected the NDJSON produced by import (exit code %d): %s", code, dryRunOut)
	}
}
