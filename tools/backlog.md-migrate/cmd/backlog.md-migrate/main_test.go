package main

import (
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestBinaryIsWiredUp builds the binary once and runs it, to check that
// main.go actually calls into internal/cli and turns its result into a
// process exit code. The exhaustive argument matrix lives in
// internal/cli, which is faster to run and easier to debug.
func TestBinaryIsWiredUp(t *testing.T) {
	binPath := filepath.Join(t.TempDir(), "backlog.md-migrate")

	build := exec.Command("go", "build", "-o", binPath, ".")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build failed: %s\n%s", err, out)
	}

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
	if code != 1 {
		t.Fatalf("got exit code %d, want 1, output: %s", code, out)
	}
	if !strings.Contains(string(out), "not implemented yet") {
		t.Fatalf("expected a not-implemented message, got: %s", out)
	}
}
