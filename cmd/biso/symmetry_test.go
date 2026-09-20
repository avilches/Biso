package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"biso/internal/board"

	_ "modernc.org/sqlite"
)

// This file is the guarantee of docs/spec/cmd/export.md: a board exported
// and imported into another one is the same board, field for field, with
// its identifiers, its dates and the keys of its criteria and of its
// comments included.
//
// There are two ways of doing it and they exercise different code, so there
// is one test for each: `biso export` with `biso new --from`, which needs
// the destination to declare the same vocabulary by hand, and `biso
// snapshot` with `biso init --from`, which brings the vocabulary with it.
//
// **What is compared is a dump of the two databases, and not the export of
// the two boards.** Comparing the exports would be comparing the output of
// one function with the output of that same function: a field the encoder
// forgot to write would be missing identically on both sides and the
// comparison would still pass, which is precisely how this test used to be
// able to pass with a board that had lost every createdAt. The dump below
// is built from the tables themselves, by SQL this file writes, and shares
// no code with anything `biso export` runs through. The other half of the
// same protection is
// TestExportWritesEveryFieldOfTheFormat in export_test.go, which compares
// one exported line against a fixture written by hand: between the two, a
// field that disappears from the format shows up either as two boards that
// differ or as a line that no longer matches the fixture.

// richBoard is a board carrying one of everything the interchange format
// can hold: the three kinds of date, an explicit identifier, criteria with
// gaps in their keys and with one checked, comments with their own keys and
// instants, an open question, a lease, an archived task, external fields,
// a parent, a dependency and every list.
//
// Its tasks come in through the batch because that is the only way to write
// some of them: an archived task, a created-at of last year and a lease
// that belongs to somebody who is not running the test have no flag of
// their own anywhere else.
const richBoard = `# every shape the format can carry
{"id":"MYP-1","title":"Write the parser","type":"bug","priority":"high","status":"Done",` +
	`"description":"A long description\nover two lines","labels":["parser","urgent"],` +
	`"references":["docs/bugs/BUG-02.md"],"documentation":["docs/parser.md"],` +
	`"modifiedFiles":["parser.go"],"ext":{"trello.card":"5f2a8c1e"},"author":"@sara",` +
	`"due":"2026-01-31","ordinal":7,"plan":"1. Read it","notes":"It was the CRLF",` +
	`"summary":"Done and tested",` +
	`"acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true},` +
	`{"key":3,"text":"There is a test","checked":false}],` +
	`"comments":[{"key":1,"author":"@avilches","createdAt":"2026-08-14T10:22:00Z","body":"Reported from Windows"},` +
	`{"key":4,"author":"@sara","createdAt":"2026-08-15T09:00:00Z","body":"Fixed"}],` +
	`"question":{"author":"@avilches","askedAt":"2026-08-16T09:00:00Z","body":"Is it a CRLF, or also a lone CR?"},` +
	`"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
{"id":"MYP-2","title":"Depends on the parser","status":"In Progress","assignees":["@sara"],` +
	`"dependencies":["MYP-1"],"leaseHolder":"@sara","leaseExpiresAt":"2126-09-08T14:00:00Z",` +
	`"createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
{"id":"MYP-9","title":"An archived one","archived":true,"type":"docs","priority":"low",` +
	`"parent":"MYP-1","createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-20T18:05:00Z"}
{"title":"No identifier of its own, and no dates either"}
`

// vocabulary is what a board has to declare to accept the tasks above.
var vocabulary = []string{"--prefix", "MYP", "--extensions", "trello.card"}

func sourceBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, append([]string{"init", "My project"}, vocabulary...)...).assertCode(t, 0)
	path := filepath.Join(m.dir, "rich.ndjson")
	m.write(t, path, richBoard)
	m.run(t, "new", "--from", path).assertCode(t, 0)
	// One more task written the ordinary way, so that the export carries a
	// task the program itself stamped and not only imported ones.
	m.run(t, "new", "Taken by whoever is calling", "--start").assertCode(t, 0)
	return m
}

func TestExportAndNewFromLeaveTwoIdenticalBoards(t *testing.T) {
	source := sourceBoard(t)
	dump := source.run(t, "export").assertCode(t, 0).stdout
	if strings.Count(dump, "\n") != 5 {
		t.Fatalf("the export has %d lines and the board has five tasks:\n%s",
			strings.Count(dump, "\n"), dump)
	}

	// The destination declares the same vocabulary by hand, which is what
	// this way of restoring asks of it (docs/spec/cmd/export.md).
	destination := newMachine(t)
	destination.env["BISO_ME"] = "@claude"
	destination.run(t, append([]string{"init", "Another project"}, vocabulary...)...).assertCode(t, 0)
	path := filepath.Join(destination.dir, "dump.ndjson")
	destination.write(t, path, dump)
	destination.run(t, "new", "--from", path).assertCode(t, 0)

	// The two boards, table by table and column by column. The two tables
	// of the board itself are left out of this comparison and only of this
	// one: `export` carries the tasks and never the vocabulary, so the
	// destination was created by hand with a name of its own and an
	// identity of its own (docs/spec/cmd/export.md#la-garantía-de-simetría).
	// The other way round, below, compares those two as well.
	assertEqual(t,
		dumpDatabase(t, destination.boardDir(t), taskTables...),
		dumpDatabase(t, source.boardDir(t), taskTables...),
		"the tasks of the board the import produced")

	// And the export of the two, which is the weaker comparison of the two
	// and is kept because it is the one that reads the board back out
	// through the format.
	assertEqual(t, destination.run(t, "export").assertCode(t, 0).stdout, dump,
		"the export of the board the import produced")
}

func TestSnapshotAndInitFromLeaveTwoIdenticalBoards(t *testing.T) {
	source := sourceBoard(t)
	source.run(t, "snapshot", "--vcs", "none").assertCode(t, 0)
	sourceDir := source.boardDir(t)

	// The restore runs on a machine with a set of roots of its own, which
	// is how a snapshot is really restored: the identity it carries is free
	// there. Restoring it beside the board it came from is the duplicate
	// identity error of docs/spec/cmd/init.md.
	elsewhere := newMachine(t)
	elsewhere.env["BISO_ME"] = "@claude"
	restored := filepath.Join(elsewhere.home, "restored-board")
	elsewhere.run(t, "init", "--at", restored, "--from", sourceDir).assertCode(t, 0)

	// Nothing of the vocabulary was declared by hand, so this comparison
	// covers every table there is, the board's own identity and its twenty
	// configuration keys included.
	assertEqual(t,
		dumpDatabase(t, restored, allTables(t, restored)...),
		dumpDatabase(t, sourceDir, allTables(t, sourceDir)...),
		"the database of the restored board")

	// And the two files the restored board writes for itself, which is the
	// same comparison seen through the format.
	elsewhere.at(restored).run(t, "snapshot", "--vcs", "none").assertCode(t, 0)
	for _, name := range []string{board.SnapshotTasksFile, board.SnapshotConfigFile} {
		assertEqual(t,
			source.read(t, filepath.Join(restored, name)),
			source.read(t, filepath.Join(sourceDir, name)),
			name+" of the restored board")
	}

	// The identity travels with the snapshot, so the pointer a project had
	// committed keeps naming the same board.
	if got := board.MarkerID(restored); got != board.MarkerID(sourceDir) {
		t.Errorf("the restored board has id %q and the original one %q", got, board.MarkerID(sourceDir))
	}
}

// taskTables are the tables that hold the tasks, which is what `biso
// export` and `biso new --from` carry between two boards.
var taskTables = []string{
	"board_counter", "task", "task_comment", "task_criterion",
	"task_ext", "task_list_item",
}

// allTables are every table of a board's database, in alphabetical order,
// asked of the database itself so that a table added later is compared
// without anybody remembering to add it here.
func allTables(t *testing.T, dir string) []string {
	t.Helper()
	db := openDatabase(t, dir)
	defer db.Close()
	rows, err := db.Query(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		tables = append(tables, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	return tables
}

// dumpDatabase writes the whole content of the tables given: every column
// of every row, each row on its own line, sorted by everything it holds so
// that two boards written in a different order still dump the same.
//
// It reads the database directly and goes through no part of the program,
// which is the whole point: it can show a field that the export format has
// stopped carrying.
func dumpDatabase(t *testing.T, dir string, tables ...string) string {
	t.Helper()
	db := openDatabase(t, dir)
	defer db.Close()

	var out strings.Builder
	for _, table := range tables {
		columns := columnsOf(t, db, table)
		quoted := make([]string, 0, len(columns))
		for _, c := range columns {
			quoted = append(quoted, `"`+c+`"`)
		}
		rows, err := db.Query(fmt.Sprintf(`SELECT %s FROM "%s"`,
			strings.Join(quoted, ", "), table))
		if err != nil {
			t.Fatalf("reading %s of %s: %v", table, dir, err)
		}
		var lines []string
		for rows.Next() {
			values := make([]any, len(columns))
			into := make([]any, len(columns))
			for i := range values {
				into[i] = &values[i]
			}
			if err := rows.Scan(into...); err != nil {
				rows.Close()
				t.Fatal(err)
			}
			cells := make([]string, 0, len(columns))
			for i, c := range columns {
				cells = append(cells, fmt.Sprintf("%s=%s", c, cellText(values[i])))
			}
			lines = append(lines, table+": "+strings.Join(cells, " | "))
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		rows.Close()
		sort.Strings(lines)
		for _, line := range lines {
			out.WriteString(line + "\n")
		}
	}
	return out.String()
}

// cellText writes one value the way this comparison reads it, with NULL
// told apart from the empty string: the difference between a field with no
// value and a field whose value is nothing is a difference between two
// boards (docs/spec/contrato-json.md#números-fechas-y-ausencias).
func cellText(value any) string {
	switch v := value.(type) {
	case nil:
		return "<null>"
	case []byte:
		return fmt.Sprintf("%q", string(v))
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func columnsOf(t *testing.T, db *sql.DB, table string) []string {
	t.Helper()
	rows, err := db.Query(fmt.Sprintf(`PRAGMA table_info("%s")`, table))
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var columns []string
	for rows.Next() {
		var (
			cid        int
			name, kind string
			notNull    int
			dflt       any
			primary    int
		)
		if err := rows.Scan(&cid, &name, &kind, &notNull, &dflt, &primary); err != nil {
			t.Fatal(err)
		}
		columns = append(columns, name)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if len(columns) == 0 {
		t.Fatalf("%s has no columns", table)
	}
	return columns
}

func openDatabase(t *testing.T, dir string) *sql.DB {
	t.Helper()
	path := filepath.Join(dir, board.DatabaseFile)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("%s holds no database: %v", dir, err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	return db
}
