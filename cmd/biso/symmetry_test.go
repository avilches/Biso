package main

import (
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/board"
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
// What is compared is the export of the two boards, byte for byte. It is
// the strongest comparison available from outside the program and the one
// that means the most: every non-derived field of every task is in it, and
// two boards whose exports are identical cannot differ in any of them.

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

	assertEqual(t, destination.run(t, "export").assertCode(t, 0).stdout, dump,
		"the export of the board the import produced")

	// The counter of the criterion keys is deduced and not exported, so it
	// is checked where it shows: the next criterion of a task whose highest
	// imported key was #3 is #4 on both boards.
	source.run(t, "set", "MYP-1", "--add-ac", "One more").assertCode(t, 0)
	destination.run(t, "set", "MYP-1", "--add-ac", "One more").assertCode(t, 0)
	assertEqual(t,
		destination.run(t, "get", "MYP-1", "--section", "ac").assertCode(t, 0).stdout,
		source.run(t, "get", "MYP-1", "--section", "ac").assertCode(t, 0).stdout,
		"the criteria of the two boards after adding one to each")
}

func TestSnapshotAndInitFromLeaveTwoIdenticalBoards(t *testing.T) {
	source := sourceBoard(t)
	source.run(t, "snapshot", "--vcs", "none").assertCode(t, 0)
	sourceDir := source.boardDir(t)

	// The restore is run from outside the source project, so that its
	// pointer does not collide with the one the source already has
	// (docs/spec/cmd/export.md).
	elsewhere := filepath.Join(source.home, "elsewhere")
	source.mkdir(t, elsewhere)
	restored := filepath.Join(source.home, "restored-board")
	source.at(elsewhere).run(t, "init", "--at", restored, "--from", sourceDir).assertCode(t, 0)

	// Nothing of the vocabulary was declared by hand, and the restored
	// board writes the same two files as the one it came from.
	source.at(restored).run(t, "snapshot", "--vcs", "none").assertCode(t, 0)
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
