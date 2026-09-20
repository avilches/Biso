package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/board"
)

// This file is the harness the tests of this package run through: a whole
// machine in a temporary directory, with its own home and its own project,
// so that nothing a test does can reach the machine it runs on and nothing
// that machine has can change what a test sees. Every call goes through the
// compiled program.

type machine struct {
	home string
	dir  string
	env  map[string]string
}

// newMachine builds one: a home directory and a project directory inside it,
// which is the ordinary shape of docs/spec/resolucion-del-tablero.md.
func newMachine(t *testing.T) *machine {
	t.Helper()
	// The temporary directory is resolved first, because on some systems it
	// hangs off a symbolic link (/var is /private/var on macOS) and the
	// process would then see a home that is not a prefix of its own working
	// directory, which is not what any real machine looks like.
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	m := &machine{home: home, dir: filepath.Join(home, "my-project"), env: map[string]string{}}
	m.mkdir(t, m.dir)
	return m
}

// at moves the working directory of the next call, creating it if it is not
// there.
func (m *machine) at(dir string) *machine {
	clone := *m
	clone.dir = dir
	return &clone
}

func (m *machine) mkdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
}

func (m *machine) write(t *testing.T, path, content string) {
	t.Helper()
	m.mkdir(t, filepath.Dir(path))
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (m *machine) read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// run executes the compiled program in this machine and answers what the
// process wrote and what it exited with.
func (m *machine) run(t *testing.T, argv ...string) call {
	t.Helper()
	env := []string{"HOME=" + m.home}
	for name, value := range m.env {
		env = append(env, name+"="+value)
	}
	return runIn(t, m.dir, env, argv...)
}

// substituted is a fixture with the two absolute paths of the examples
// turned into the ones of this machine. Everything else of the block, the
// alignment of the columns included, is compared as the specification wrote
// it: a temporary home is the one thing a test cannot help changing.
func (m *machine) substituted(fixture string) string {
	fixture = strings.ReplaceAll(fixture, "/Volumes/work", filepath.Join(m.home, "Volumes", "work"))
	return strings.ReplaceAll(fixture, "/Users/avilches", m.home)
}

// projectOfTheSpecification is the directory the examples of
// docs/spec/cmd/where.md are called from.
func (m *machine) projectOfTheSpecification() string {
	return filepath.Join(m.home, "Hub", "Projects", "My project")
}

// buildTheBoardOfTheSpecification makes the board those examples describe,
// with its two hundred and forty eight tasks, and points the project at it.
// It is built here and not with `biso init` because its id and its counts
// are part of the text being checked, and no flag of the program fixes an
// id: that is the whole difference between a fixture and a call.
func (m *machine) buildTheBoardOfTheSpecification(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(m.home, ".biso", "boards", "my-project-3f9a2b1c")
	m.buildBoard(t, dir, "3f9a2b1c", "My project", "MYP", 248, 31, 290)

	project := m.projectOfTheSpecification()
	m.mkdir(t, project)
	m.write(t, filepath.Join(project, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	return dir
}

// buildBoard writes a board with the given identity, name and counts: as
// many tasks as the two counts add up to, with the counter left at the
// highest identifier the board is meant to have handed out.
func (m *machine) buildBoard(t *testing.T, dir, id, name, prefix string, notArchived, archived, highest int) {
	t.Helper()
	b, err := board.Create(dir, id, board.DefaultConfig(name, prefix), board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if err := board.WriteMarker(dir, id); err != nil {
		t.Fatal(err)
	}
	if notArchived+archived == 0 {
		return
	}
	// One statement for the whole board: this is a fixture and not a use
	// of the program, and creating them one by one would be two hundred
	// and seventy nine transactions to check one line of text.
	if _, err := b.Store.Exec(`
		WITH RECURSIVE n(i) AS (
			SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < ?
		)
		INSERT INTO task (
			id, num, title, status, type, priority, parent, author, due,
			ordinal, description, plan, notes, summary, created_at, updated_at,
			archived, lease_expires_at, lease_holder, question_author,
			question_asked_at, question_body, next_criterion_key, next_comment_key
		)
		SELECT ? || '-' || i, i, 'Task ' || i, ?, 'task', 'medium', '', '', '',
			i, '', '', '', '', '2026-09-06T09:12:04Z', '2026-09-06T09:12:04Z',
			CASE WHEN i > ? THEN 1 ELSE 0 END, '', '', '', '', '', 1, 1
		FROM n`,
		notArchived+archived, prefix, board.DefaultConfig(name, prefix).InitialStatus,
		notArchived); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Store.Exec(`UPDATE board_counter SET last_task_num = ?`, highest); err != nil {
		t.Fatal(err)
	}
}
