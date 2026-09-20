package main

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the two transversal endings of
// docs/spec/garantias.md that belong to no command in particular, so no
// command's own tests ever covered them: a store the environment will not
// let this program write to, which is exit code 8, and a database that opens
// and then turns out to be damaged, which is exit code 21. Both are reached
// through several commands on purpose, because what is being checked is that
// the classification lives in the layer that talks to SQLite and not in the
// one command that happened to notice first.

// skipIfRootCanWriteAnything leaves out the tests that take away a
// permission, because the superuser is not refused by any of them and the
// test would be checking nothing.
func skipIfRootCanWriteAnything(t *testing.T) {
	t.Helper()
	if os.Geteuid() == 0 {
		t.Skip("the superuser is never refused by file permissions")
	}
}

// boardOf is the directory of the board a test of this file built.
func boardOf(m *machine) string { return filepath.Join(m.dir, "board") }

// TestAWriteThatThePermissionsRefuseIsTheEnvironmentFailing is the row of
// docs/spec/garantias.md#qué-pasa-cuando-el-almacén-no-se-puede-escribir
// over a database file that cannot be written: the board is sound, the
// request is sound, and what failed is the machine, which is exit code 8 and
// not the unforeseen failure of exit code 1.
func TestAWriteThatThePermissionsRefuseIsTheEnvironmentFailing(t *testing.T) {
	skipIfRootCanWriteAnything(t)
	m := oneTaskBoard(t)
	db := filepath.Join(boardOf(m), "board.db")
	if err := os.Chmod(db, 0o444); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(db, 0o644) })

	got := m.run(t, "set", "MYP-1", "--title", "Another title").assertCode(t, 8)

	if !strings.Contains(got.stderr, db+" cannot be written: ") {
		t.Errorf("the error does not name the file that refused the write:\n%s", got.stderr)
	}
	if !strings.Contains(got.stderr, "nothing was written") {
		t.Errorf("the error does not say that nothing was written:\n%s", got.stderr)
	}
	// And the same failure carries its own `code` in the envelope, the one
	// docs/spec/contrato-json.md already lists for exit code 8.
	envelope := m.run(t, "set", "MYP-1", "--title", "Another title", "--json").assertCode(t, 8)
	if !strings.Contains(envelope.stderr, `"code": "io_error"`) {
		t.Errorf("the envelope does not carry io_error:\n%s", envelope.stderr)
	}
}

// TestABoardDirectoryThatCannotBeWrittenFailsTheSameWay is the other half of
// that same path: the failure happens while opening, because WAL mode has to
// write to open, so nothing has read the board's identifier yet. The answer
// is still exit code 8, and the message names the file and never leaves a
// hole where an identifier would go.
func TestABoardDirectoryThatCannotBeWrittenFailsTheSameWay(t *testing.T) {
	skipIfRootCanWriteAnything(t)
	m := oneTaskBoard(t)
	dir := boardOf(m)
	for _, name := range []string{"board.db-wal", "board.db-shm"} {
		os.Remove(filepath.Join(dir, name))
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })

	for _, argv := range [][]string{
		{"ls"},
		{"get", "MYP-1"},
		{"doctor"},
		{"doctor", "--fix"},
	} {
		got := m.run(t, argv...).assertCode(t, 8)
		if !strings.Contains(got.stderr, filepath.Join(dir, "board.db")+" cannot be written: ") {
			t.Errorf("biso %s: %s", strings.Join(argv, " "), got.stderr)
		}
		if strings.Contains(got.stderr, "board 's database") {
			t.Errorf("biso %s printed a message with a hole where the id goes:\n%s",
				strings.Join(argv, " "), got.stderr)
		}
	}
}

// TestADamagedPageIsTheDatabaseThatCannotBeRead is the second case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// reached the way it really happens: not with a file that is not a database
// at all, which fails as soon as it is opened, but with one whose header is
// intact and whose pages are not. The damage then surfaces halfway through a
// query, and it has to end up as the same exit code 21 and the same three
// lines, from whichever command ran into it.
func TestADamagedPageIsTheDatabaseThatCannotBeRead(t *testing.T) {
	m := oneTaskBoard(t)
	damageTheTaskPage(t, boardOf(m))

	for _, argv := range [][]string{
		{"ls"},
		{"get", "MYP-1"},
		{"prime"},
		{"set", "MYP-1", "--title", "Another title"},
		{"archive", "MYP-1"},
		{"doctor"},
	} {
		got := m.run(t, argv...).assertCode(t, 21)
		if !strings.Contains(got.stderr, "'s database could not be read") {
			t.Errorf("biso %s said %q", strings.Join(argv, " "), got.stderr)
		}
	}

	envelope := m.run(t, "ls", "--json").assertCode(t, 21)
	if !strings.Contains(envelope.stderr, `"code": "database_unreadable"`) {
		t.Errorf("the envelope does not carry database_unreadable:\n%s", envelope.stderr)
	}
}

// damageTheTaskPage overwrites the page that holds the `task` table, leaving
// the first page, the one that carries the file header and the schema,
// untouched. That is what makes the damage invisible until something reads a
// task, which is the whole point of this fixture.
func damageTheTaskPage(t *testing.T, dir string) {
	t.Helper()
	path := filepath.Join(dir, "board.db")

	db, err := sql.Open("sqlite", "file:"+path)
	if err != nil {
		t.Fatal(err)
	}
	var rootPage, pageSize int
	if err := db.QueryRow(
		`SELECT rootpage FROM sqlite_master WHERE name = 'task'`).Scan(&rootPage); err != nil {
		t.Fatal(err)
	}
	if err := db.QueryRow(`PRAGMA page_size`).Scan(&pageSize); err != nil {
		t.Fatal(err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"board.db-wal", "board.db-shm"} {
		os.Remove(filepath.Join(dir, name))
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	start := (rootPage - 1) * pageSize
	if start+pageSize > len(content) {
		t.Fatalf("the database is %d bytes and page %d does not fit in it", len(content), rootPage)
	}
	for i := start; i < start+pageSize; i++ {
		content[i] = 0xff
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatal(err)
	}
}
