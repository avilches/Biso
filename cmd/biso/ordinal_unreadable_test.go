package main

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"biso/internal/board"
)

// This file is the one anchor of the manual order that the program cannot
// reach on its own: a key already stored that does not keep its form
// (docs/spec/modelo-de-datos/orden-manual.md#una-clave-guardada-que-no-cumple-la-regla).
// It is not an error of its own, it is the general rule of a datum that
// cannot be interpreted, so what it has to do is what any other unreadable
// task does: drop out of the listing with the warning that names it, fail a
// targeted read, and show up in `biso doctor`.
//
// The board it runs over is a board of the shape an older binary left
// behind, which is the one way this really happens: back when `ordinal` was
// an integer the column carried no CHECK, SQLite hands those integers back
// as the integers they are, and database/sql turns each one into a string
// on the way into the field, so a 0 or a 3000 arrives here as a string that
// is not a key (docs/spec/estado-de-implementacion.md#el-orden-manual).

// ordinalColumn matches the whole declaration of the ordinal column, from
// its type to the comma that closes it, however its CHECK is written.
var ordinalColumn = regexp.MustCompile(`(?s)ordinal(\s+)TEXT.*?,\n`)

// dropTheCheckOfTheOrdinalColumn rewrites the stored schema so that the
// column accepts anything, which is what the column looked like before the
// manual order became a text key. It is the only way to get a malformed key
// into a board: the CHECK is precisely what keeps this program from writing
// one, and turning it off per connection would not do, because `biso
// doctor` runs an integrity check that reads the constraint back.
func dropTheCheckOfTheOrdinalColumn(t *testing.T, dir string) {
	t.Helper()
	b, err := board.Open(&board.Location{ID: board.MarkerID(dir), Dir: dir}, board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()

	before := schemaOfTheTaskTable(t, b)
	if !strings.Contains(before, "CHECK") {
		t.Fatalf("the task table has no CHECK to drop, so this test no longer sets up what it says:\n%s",
			before)
	}
	after := ordinalColumn.ReplaceAllString(before, "ordinal${1}TEXT,\n")
	if after == before || strings.Contains(afterTheOrdinalLine(after), "CHECK") {
		t.Fatalf("the ordinal column was not stripped of its CHECK:\n%s", after)
	}

	for _, statement := range []string{
		`PRAGMA writable_schema = ON`,
		`UPDATE sqlite_schema SET sql = ? WHERE type = 'table' AND name = 'task'`,
		`PRAGMA writable_schema = RESET`,
	} {
		var args []any
		if strings.HasPrefix(statement, "UPDATE") {
			args = []any{after}
		}
		if _, err := b.Store.Exec(statement, args...); err != nil {
			t.Fatalf("%s: %v", statement, err)
		}
	}
}

func schemaOfTheTaskTable(t *testing.T, b *board.Board) string {
	t.Helper()
	rows, err := b.Store.Query(`SELECT sql FROM sqlite_schema WHERE type = 'table' AND name = 'task'`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	if !rows.Next() {
		t.Fatal("the board has no task table")
	}
	var text string
	if err := rows.Scan(&text); err != nil {
		t.Fatal(err)
	}
	return text
}

// afterTheOrdinalLine is what the schema says from the ordinal column on,
// which is where a CHECK that survived the rewrite would still be.
func afterTheOrdinalLine(schema string) string {
	at := strings.Index(schema, "ordinal")
	if at < 0 {
		return schema
	}
	rest := schema[at:]
	if end := strings.Index(rest, "description"); end > 0 {
		return rest[:end]
	}
	return rest
}

// writeOrdinalDirectly puts a value in the column without going through the
// program, which is only possible once the CHECK is gone.
func writeOrdinalDirectly(t *testing.T, dir string, keys map[string]string) {
	t.Helper()
	b, err := board.Open(&board.Location{ID: board.MarkerID(dir), Dir: dir}, board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	for id, key := range keys {
		if _, err := b.Store.Exec(`UPDATE task SET ordinal = ? WHERE id = ?`, key, id); err != nil {
			t.Fatalf("writing the key %q on %s: %v", key, id, err)
		}
	}
}

func TestAStoredOrdinalThatIsNotAKeyLeavesTheTaskOutAndIsReported(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")
	m.run(t, "new", "A task that stays readable", "--ordinal", "last").assertCode(t, 0)
	m.run(t, "new", "An integer where a key goes").assertCode(t, 0)
	m.run(t, "new", "A key that ends in zero").assertCode(t, 0)

	dropTheCheckOfTheOrdinalColumn(t, dir)
	writeOrdinalDirectly(t, dir, map[string]string{"MYP-2": "3000", "MYP-3": "m0"})

	listed := m.run(t, "ls", "--ids").assertCode(t, 0)

	assertEqual(t, listed.stdout, "MYP-1\n", "the rest of the listing")
	assertEqual(t, listed.stderr,
		"warning: 2 tasks could not be read and were skipped: MYP-2, MYP-3\n",
		"the warning that names the two tasks whose stored key is not one")

	// A targeted read of one of them fails, because there is nothing else
	// to answer with, and the message says which field it choked on.
	got := m.run(t, "get", "MYP-2").assertCode(t, 3)
	if !strings.Contains(got.stderr, "is not an ordinal key") {
		t.Errorf("the refusal does not name the key:\n%s", got.stderr)
	}

	report := m.run(t, "doctor").assertCode(t, 6)
	for _, id := range []string{"MYP-2", "MYP-3"} {
		if !strings.Contains(report.stdout, id+`  task "`+id+`" could not be parsed: ordinal:`) {
			t.Errorf("biso doctor does not report %s:\n%s", id, report.stdout)
		}
	}
	if strings.Contains(report.stdout, "MYP-1") {
		t.Errorf("biso doctor reported the readable task too:\n%s", report.stdout)
	}

	// The integers that do happen to keep the form of a key are read as
	// keys, so what the check refuses is the value and not the older
	// column. And they are compared by code point like any other key, which
	// is why the 12 comes out before the 7 and the manual order of an old
	// board is not the one it had: one more reason to recreate it rather
	// than to keep using it
	// (docs/spec/estado-de-implementacion.md#el-orden-manual).
	writeOrdinalDirectly(t, dir, map[string]string{"MYP-2": "7", "MYP-3": "12"})
	after := m.run(t, "ls", "--sort", "ordinal", "--ids").assertCode(t, 0)
	assertEqual(t, after.stdout, "MYP-3\nMYP-2\nMYP-1\n",
		"the order of two integers reinterpreted as keys, compared by code point")
}
