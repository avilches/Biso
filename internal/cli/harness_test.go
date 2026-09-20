package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// This file is the harness every test of this package runs through: a whole
// machine in a temporary directory, with its own home, its own boards root
// and its own clock, so that nothing a test does can reach the machine it
// runs on and nothing the machine has can change what a test sees.
//
// These tests drive Run in this process, which is what lets them fix the
// clock and the identifiers a call mints. The golden tests, where what is
// checked is the literal text of a command, live in cmd/biso instead and
// run the compiled program.

// fixedNow is the instant every envelope of these tests is stamped with. It
// is the one the examples of docs/spec/ carry.
var fixedNow = time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC)

type machine struct {
	t    *testing.T
	home string
	dir  string
	env  map[string]string
	ids  []string
}

// newMachine builds one: a home directory and a project directory inside it,
// which is the ordinary shape of docs/spec/resolucion-del-tablero.md.
func newMachine(t *testing.T) *machine {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	return &machine{t: t, home: home, dir: dir, env: map[string]string{}}
}

// at moves the working directory of the next call, creating it if it is not
// there.
func (m *machine) at(dir string) *machine {
	m.t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		m.t.Fatal(err)
	}
	clone := *m
	clone.dir = dir
	return &clone
}

// withIDs queues the identifiers the next calls will mint, so that a test
// knows what the board will be called.
func (m *machine) withIDs(ids ...string) *machine {
	m.ids = append(m.ids, ids...)
	return m
}

type outcome struct {
	code   int
	stdout string
	stderr string
}

// run calls the program the way cmd/biso/main.go does, with neither of its
// streams on a terminal, which is how a test harness always sees them.
func (m *machine) run(argv ...string) outcome {
	m.t.Helper()
	return m.runWithTerminal(false, argv...)
}

// runWithTerminal is run with the answer the two streams give when asked
// whether they are a terminal, which is the one thing biso ever asks about
// it (docs/spec/salida-y-terminal.md#interactividad-terminal-y-color).
func (m *machine) runWithTerminal(isTerminal bool, argv ...string) outcome {
	m.t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(argv, Streams{
		StdoutIsTerminal: isTerminal,
		StderrIsTerminal: isTerminal,
		Stdout:           &stdout,
		Stderr:           &stderr,
		Stdin:            strings.NewReader(""),
		Getenv: func(name string) (string, bool) {
			v, ok := m.env[name]
			return v, ok
		},
		Dir:   m.dir,
		Home:  m.home,
		Now:   func() time.Time { return fixedNow },
		NewID: m.nextID,
	})
	return outcome{code: code, stdout: stdout.String(), stderr: stderr.String()}
}

func (m *machine) nextID() (string, error) {
	if len(m.ids) == 0 {
		return "00000000", nil
	}
	id := m.ids[0]
	m.ids = m.ids[1:]
	return id, nil
}

// boardsRoot is the machine's default root, where a board created without
// --at goes.
func (m *machine) boardsRoot() string {
	return filepath.Join(m.home, ".biso", "boards")
}

func (m *machine) writeFile(path, content string) {
	m.t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		m.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		m.t.Fatal(err)
	}
}

func (m *machine) read(path string) string {
	m.t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		m.t.Fatal(err)
	}
	return string(b)
}

func (m *machine) exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// assertCode fails the test when the call did not end the way it should
// have, printing what it wrote, which is what a failure of this suite needs
// to be readable.
func (o outcome) assertCode(t *testing.T, want int) outcome {
	t.Helper()
	if o.code != want {
		t.Fatalf("exit code = %d, want %d\nstdout:\n%s\nstderr:\n%s",
			o.code, want, o.stdout, o.stderr)
	}
	return o
}

// assertEqual compares two texts character for character, and shows them
// both when they differ.
func assertEqual(t *testing.T, got, want, what string) {
	t.Helper()
	if got == want {
		return
	}
	t.Errorf("%s does not match, character for character.\n--- got ---\n%s\n--- want ---\n%s",
		what, got, want)
}
