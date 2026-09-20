package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the cases of docs/spec/cmd/export.md that show from outside the
// program: what the dump carries by default, what the two flags of this
// command change, and the exit codes of its table.

// exportBoard is a board with one task in each of the three states this
// command tells apart: the initial one, the terminal one and the archive.
func exportBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "Still to do").assertCode(t, 0)
	m.run(t, "new", "Finished", "--status", "Done").assertCode(t, 0)
	m.run(t, "new", "Archived and finished", "--status", "Done").assertCode(t, 0)
	m.archive(t, "MYP-3")
	return m
}

func TestExportTakesTheWholeBoardWithNoFilters(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "export").assertCode(t, 0)

	// `biso ls` would leave out the finished one and the archived one; this
	// command inherits neither default (docs/spec/cmd/export.md).
	for _, id := range []string{"MYP-1", "MYP-2", "MYP-3"} {
		if !strings.Contains(got.stdout, `"id":"`+id+`"`) {
			t.Errorf("the dump does not carry %s:\n%s", id, got.stdout)
		}
	}
	if lines := strings.Count(got.stdout, "\n"); lines != 3 {
		t.Errorf("the dump has %d lines and the board has three tasks", lines)
	}
}

func TestExportLeavesTheArchivedOutOnRequest(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "export", "--no-archived").assertCode(t, 0)

	if strings.Contains(got.stdout, `"id":"MYP-3"`) {
		t.Errorf("--no-archived kept the archived task:\n%s", got.stdout)
	}
	if lines := strings.Count(got.stdout, "\n"); lines != 2 {
		t.Errorf("the dump has %d lines and two tasks are not archived", lines)
	}
}

func TestExportAppliesTheFiltersOfAListing(t *testing.T) {
	m := exportBoard(t)

	// The example of the specification: an explicit -s filters by that value
	// as it is, the terminal status included.
	got := m.run(t, "export", "-s", "Done", "--no-archived").assertCode(t, 0)

	if !strings.Contains(got.stdout, `"id":"MYP-2"`) || strings.Count(got.stdout, "\n") != 1 {
		t.Errorf("-s Done --no-archived answered:\n%s", got.stdout)
	}
}

func TestExportWritesTheFileMinusOAsks(t *testing.T) {
	m := exportBoard(t)
	path := filepath.Join(m.dir, "backup.ndjson")

	got := m.run(t, "export", "-o", path).assertCode(t, 0)

	if got.stdout != "" {
		t.Errorf("with -o the dump still went to stdout:\n%s", got.stdout)
	}
	if lines := strings.Count(m.read(t, path), "\n"); lines != 3 {
		t.Errorf("the file has %d lines and the board has three tasks", lines)
	}
}

func TestExportRefusesTheFlagsItsTableDoesNotHave(t *testing.T) {
	m := exportBoard(t)
	for _, argv := range [][]string{
		{"export", "--json"},
		{"export", "--dry-run"},
		{"export", "--print"},
		{"export", "--sort", "urgency"},
		{"export", "--limit", "5"},
		{"export", "--all"},
		{"export", "--ids"},
		{"export", "--count"},
		{"export", "--archived"},
		{"export", "--only-archived"},
	} {
		got := m.run(t, argv...)
		if got.code != 2 {
			t.Errorf("biso %s exited %d and not 2", strings.Join(argv, " "), got.code)
		}
		if got.stdout != "" {
			t.Errorf("biso %s wrote to stdout: %s", strings.Join(argv, " "), got.stdout)
		}
	}
}

func TestExportRefusesJSONInPlainText(t *testing.T) {
	m := exportBoard(t)

	got := m.run(t, "export", "--json").assertCode(t, 2)

	// --json is itself the invalid part of the call, so there is no
	// envelope to wrap the refusal in
	// (docs/spec/contrato-json.md#los-errores-en-json).
	if strings.HasPrefix(strings.TrimSpace(got.stderr), "{") {
		t.Errorf("the refusal came as a JSON envelope:\n%s", got.stderr)
	}
	assertEqual(t, got.stderr,
		"error: --json does not apply to export, whose output is already NDJSON\n",
		"the refusal of biso export --json")
}

func TestExportRefusesAFilterValueTheBoardDoesNotHave(t *testing.T) {
	m := exportBoard(t)
	m.run(t, "export", "-s", "Pendiente").assertCode(t, 3)
}

func TestExportAnswersSixWhenItSkippedATask(t *testing.T) {
	m := exportBoard(t)
	// A date the program did not write is one of the ways a task becomes
	// undecodable
	// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
	m.execOnBoard(t, m.boardDir(t),
		`UPDATE task SET created_at = 'yesterday' WHERE id = 'MYP-1'`)

	got := m.run(t, "export")

	if got.code != 6 {
		t.Errorf("the export of a board with an unreadable task exited %d and not 6", got.code)
	}
	if !strings.Contains(got.stderr, "could not be read and was skipped") {
		t.Errorf("the skip was not named:\n%s", got.stderr)
	}
	if lines := strings.Count(got.stdout, "\n"); lines != 2 {
		t.Errorf("the dump has %d lines, and the two readable tasks were still written", lines)
	}
}

func TestExportWithoutABoard(t *testing.T) {
	m := newMachine(t)
	m.run(t, "export").assertCode(t, 20)
}

func TestExportCannotWriteWhereItWasTold(t *testing.T) {
	m := exportBoard(t)
	closed := filepath.Join(m.dir, "closed")
	m.mkdir(t, closed)
	if err := os.Chmod(closed, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(closed, 0o700) })

	m.run(t, "export", "-o", filepath.Join(closed, "backup.ndjson")).assertCode(t, 8)
}
