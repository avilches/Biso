package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These tests walk, through the compiled program, the rule of
// docs/decisiones/detalles.md#una-referencia-con-un-salto-de-línea-se-rechaza-al-escribirla:
// a reference that carries a `\r` or a `\n` is refused with the same
// malformed_string_value that already rejects one in the title, the author
// or the text of a criterion (docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string),
// through the three ways a reference reaches a task: the --add-refs flag,
// `new --from` and `init --from`. All three converge on Task.Validate, so
// these tests fail if that single check is ever removed or narrowed.

// TestAddRefsRejectsAReferenceWithANewline is the flag: --add-refs splits
// its value on a comma and not on a line break, so a value that carries one
// stays a single reference and reaches Task.Validate whole.
func TestAddRefsRejectsAReferenceWithANewline(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	got := m.run(t, "new", "A task", "--add-refs", "first line\nsecond line")

	got.assertCode(t, 2)
	assertEqual(t, got.stderr,
		"error: malformed reference: \"first line\\nsecond line\"\n"+
			"hint: a string field cannot contain a newline or a carriage return\n",
		"the failure of --add-refs with a newline in a reference")
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a refused --add-refs left behind")

	asJSON := m.run(t, "new", "A task", "--add-refs", "first line\nsecond line", "--json").assertCode(t, 2)
	var envelope struct {
		Error struct {
			ExitCode int    `json:"exitCode"`
			Code     string `json:"code"`
			Field    string `json:"field"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(asJSON.stderr), &envelope); err != nil {
		t.Fatalf("the error envelope is not JSON: %v\nstderr:\n%s", err, asJSON.stderr)
	}
	if envelope.Error.ExitCode != 2 || envelope.Error.Code != "malformed_string_value" ||
		envelope.Error.Field != "reference" {
		t.Errorf("the error object = %+v", envelope.Error)
	}
}

// TestNewFromRejectsAReferenceWithANewline is `new --from`: a batch line
// whose references carry a `\n` fails the whole file with exit code 7, and
// the failure underneath is the same malformed_string_value, stamped with
// the line it came from.
func TestNewFromRejectsAReferenceWithANewline(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"Bad","references":["docs/a.md","first line\nsecond line"]}`)

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	if !strings.Contains(got.stderr, `line 1: malformed reference: "first line\nsecond line"`) {
		t.Errorf("stderr does not name the reference:\n%s", got.stderr)
	}
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks an invalid batch left behind")

	asJSON := m.run(t, "new", "--from", path, "--json").assertCode(t, 7)
	var envelope struct {
		Error struct {
			Details []struct {
				ExitCode int    `json:"exitCode"`
				Code     string `json:"code"`
				Field    string `json:"field"`
				Message  string `json:"message"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(asJSON.stderr), &envelope); err != nil {
		t.Fatalf("the error envelope is not JSON: %v\nstderr:\n%s", err, asJSON.stderr)
	}
	if d := envelope.Error.Details; len(d) != 1 || d[0].Code != "malformed_string_value" ||
		d[0].ExitCode != 2 || d[0].Field != "reference" ||
		d[0].Message != `line 1: malformed reference: "first line\nsecond line"` {
		t.Errorf("the details = %+v", d)
	}
}

// TestInitFromRejectsAReferenceWithANewline is `init --from`: the same
// batch engine judges a snapshot's tasks.ndjson before the new board is
// created, so a reference with a newline there refuses the whole restore
// and leaves nothing behind.
func TestInitFromRejectsAReferenceWithANewline(t *testing.T) {
	m, edited := restoredSnapshot(t,
		`{"title":"Bad","references":["docs/a.md","first line\nsecond line"]}`+"\n")
	into := filepath.Join(m.home, "not-restored-newline-reference")

	got := m.restoreInto(t, edited, into).assertCode(t, 7)

	if !strings.Contains(got.stderr, `line 1: malformed reference: "first line\nsecond line"`) {
		t.Errorf("the failure does not name the reference:\n%s", got.stderr)
	}
	if _, err := os.Stat(into); err == nil {
		t.Errorf("a failed restore left %s behind", into)
	}
}

// TestSetAddRefsAcceptsAnOrdinaryReferenceWithACommaAndABackslash is the
// negative case: the same flag still takes what it always took, so the new
// check does not reject a value that never carried a line break.
func TestSetAddRefsAcceptsAnOrdinaryReferenceWithACommaAndABackslash(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "A task").assertCode(t, 0)

	got := m.run(t, "set", "MYP-1", "--add-refs", `notes/a\,b.md,C:\\dir\\`)

	got.assertCode(t, 0)
}
