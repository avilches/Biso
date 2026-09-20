package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// This is the integration test the strategy of section 7 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md
// asks for: it compiles biso and runs the real program, which is the only
// way to exercise what a process does and not what a function returns, the
// exit code of the process included.
//
// Every call gets a home directory of its own, so the machine this suite
// runs on is never read and never written.

var (
	buildOnce  sync.Once
	binaryPath string
	buildErr   error
)

// binary compiles the program once for the whole suite.
func binary(t *testing.T) string {
	t.Helper()
	buildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "biso-binary")
		if err != nil {
			buildErr = err
			return
		}
		binaryPath = filepath.Join(dir, "biso")
		out, err := exec.Command("go", "build", "-o", binaryPath, ".").CombinedOutput()
		if err != nil {
			buildErr = err
			binaryPath = string(out)
		}
	})
	if buildErr != nil {
		t.Fatalf("biso did not build: %v\n%s", buildErr, binaryPath)
	}
	return binaryPath
}

type call struct {
	code   int
	stdout string
	stderr string
}

func run(t *testing.T, home, dir string, argv ...string) call {
	t.Helper()
	return runIn(t, dir, []string{"HOME=" + home}, argv...)
}

// runIn executes the compiled program with a working directory and the
// environment variables a test needs on top of the ones it inherits.
func runIn(t *testing.T, dir string, env []string, argv ...string) call {
	t.Helper()
	cmd := exec.Command(binary(t), argv...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("biso could not be run: %v", err)
	}
	return call{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

// assertCode fails the test when the call did not end the way it should
// have, printing what it wrote, which is what a failure of this suite needs
// to be readable.
func (c call) assertCode(t *testing.T, want int) call {
	t.Helper()
	if c.code != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			c.code, want, c.stdout, c.stderr)
	}
	return c
}

func TestTheRealProgramCreatesABoardAndFindsItAgain(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Before there is a board, `biso where` says so and the process exits
	// with 20.
	got := run(t, home, dir, "where")
	if got.code != 20 {
		t.Fatalf("exit code = %d, want 20\n%s", got.code, got.stderr)
	}
	if !strings.Contains(got.stderr, "no board here, and none configured for this project") {
		t.Errorf("stderr = %q", got.stderr)
	}

	got = run(t, home, dir, "init", "My project", "--prefix", "MYP")
	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", got.code, got.stderr)
	}
	if !strings.HasPrefix(got.stdout, "Created board \"My project\"\n") {
		t.Errorf("stdout = %q", got.stdout)
	}

	// The board is a directory with a database, a marker and an exclusion
	// file, and the project has a pointer.
	entries, err := os.ReadDir(filepath.Join(home, ".biso", "boards"))
	if err != nil || len(entries) != 1 {
		t.Fatalf("the boards root holds %v (%v)", entries, err)
	}
	board := filepath.Join(home, ".biso", "boards", entries[0].Name())
	for _, name := range []string{"board.db", ".gitignore"} {
		if _, err := os.Stat(filepath.Join(board, name)); err != nil {
			t.Errorf("the board has no %s", name)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".biso.json")); err != nil {
		t.Error("the project has no pointer")
	}

	// And from a subdirectory of the project, `biso where` finds it.
	deep := filepath.Join(dir, "src", "cli")
	if err := os.MkdirAll(deep, 0o755); err != nil {
		t.Fatal(err)
	}
	got = run(t, home, deep, "where")
	if got.code != 0 {
		t.Fatalf("exit code = %d, want 0\n%s", got.code, got.stderr)
	}
	if !strings.Contains(got.stdout, "path     "+board) {
		t.Errorf("stdout = %q, want the board at %s", got.stdout, board)
	}
	if !strings.Contains(got.stdout, "tasks    0 not archived, 0 archived, no id assigned yet") {
		t.Errorf("stdout = %q", got.stdout)
	}

	// A second init over the same project is the error of a board that
	// already exists, with exit code 2.
	got = run(t, home, dir, "init")
	if got.code != 2 {
		t.Fatalf("exit code = %d, want 2\n%s", got.code, got.stderr)
	}
}

func TestTheRealProgramAnswersTheHelpAndTheVersion(t *testing.T) {
	home := t.TempDir()
	got := run(t, home, home, "--version")
	if got.code != 0 || got.stdout != "biso 1.0.0\n" {
		t.Errorf("biso --version = %d, %q", got.code, got.stdout)
	}
	got = run(t, home, home, "--help")
	if got.code != 0 || !strings.HasPrefix(got.stdout, "biso 1.0.0 - the task board of this project.") {
		t.Errorf("biso --help = %d, %q", got.code, got.stdout)
	}
}

func TestTheRealProgramWritesDataOnStdoutAndEverythingElseOnStderr(t *testing.T) {
	home := t.TempDir()
	dir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}

	got := run(t, home, dir, "init", "My project", "--at", "board")
	if got.code != 0 {
		t.Fatalf("exit code = %d\n%s", got.code, got.stderr)
	}
	// Redirecting stdout to a file leaves a clean file of data, and
	// redirecting it to /dev/null loses no note at all
	// (docs/spec/salida-y-terminal.md#stdout-stderr-y-qué-va-en-cada-uno).
	if strings.Contains(got.stdout, "note:") {
		t.Errorf("a note travelled on stdout: %q", got.stdout)
	}
	if !strings.Contains(got.stderr, "note: the board lives inside this project.") {
		t.Errorf("stderr = %q", got.stderr)
	}
}
