package main

import (
	"strings"
	"testing"
)

// These are the golden tests of the two commands that write: the compiled
// program is run on a real board and what it printed is compared, character
// for character, with the block transcribed from docs/spec/cmd/set.md.
//
// The examples of that page all speak of MYP-11, a task of the board with
// two acceptance criteria, one of them checked, in the active status, with
// high priority and with another unfinished task depending on it, which is
// what adds up to the urgency 19.0 the lines show
// (docs/spec/modelo-de-datos/urgencia.md). So the board is built to be that
// board, with the program itself, and the line is not adjusted to whatever
// came out.

// exampleBoard creates the board of the examples and leaves MYP-11 in the
// state those lines describe, except for its acceptance criteria, which each
// test sets up as its own case needs.
func exampleBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	// Eleven tasks, so that the one the examples name really is MYP-11.
	for i := 1; i <= 11; i++ {
		m.run(t, "new", "Task "+string(rune('0'+i%10))).assertCode(t, 0)
	}
	m.run(t, "set", "MYP-11",
		"--priority", "high",
		"--status", "In Progress",
		"--add-assignees", "@claude").assertCode(t, 0)
	// Another task that depends on MYP-11 and has not finished, which is
	// the term "bloquea" of the formula.
	m.run(t, "new", "Depends on the parser", "--add-deps", "MYP-11").assertCode(t, 0)
	return m
}

func TestSetPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "Dates keep their time zone").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)

	got := m.run(t, "set", "MYP-11", "--add-labels", "parser").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "set-status-line.txt"), "the status line of biso set")
}

func TestSetAnnouncesTheKeysOfTheCriteriaItCreates(t *testing.T) {
	m := exampleBoard(t)
	// Three criteria created and one removed leaves the counter at 4, which
	// is why the two the example adds get the keys #4 and #5.
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "A criterion that will be removed",
		"--add-ac", "Dates keep their time zone").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--rm-ac", "2").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)

	got := m.run(t, "set", "MYP-11",
		"--add-ac", "There is a test",
		"--add-ac", "Docs updated").assertCode(t, 0)

	assertEqual(t,
		"biso set MYP-11 --add-ac \"There is a test\" --add-ac \"Docs updated\"\n"+got.stdout,
		fixture(t, "set-added-ac.txt"),
		"the status line with the keys it created")
}

func TestSetDryRunPrintsTheSameLineUnderItsHeader(t *testing.T) {
	m := exampleBoard(t)
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "Dates keep their time zone").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)

	got := m.run(t, "set", "MYP-11", "--add-labels", "parser", "--dry-run").assertCode(t, 0)

	assertEqual(t,
		"biso set MYP-11 --add-labels parser --dry-run\n"+got.stdout,
		fixture(t, "set-dry-run.txt"),
		"the preview of biso set")

	// And it wrote nothing, which is the whole point of the flag: taking
	// the label back out finds nothing to take.
	back := m.run(t, "set", "MYP-11", "--rm-labels", "parser").assertCode(t, 0)
	if !strings.Contains(back.stderr, "not present, nothing removed") {
		t.Errorf("the preview had written the label after all: %s", back.stderr)
	}
}

func TestSetWarnsOnStderrWhenItReplacesANonEmptyList(t *testing.T) {
	m := exampleBoard(t)
	m.run(t, "set", "MYP-11", "--add-labels", "cli,parser").assertCode(t, 0)

	got := m.run(t, "set", "MYP-11", "--replace-labels", "urgent").assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "set-overwrite.txt"), "the overwrite warning")
}

// TestNewPrintsTheIdentifierAndNothingElse is the output block of
// docs/spec/cmd/new.md: one line per task created, with the identifier and
// nothing more, and never the status line of the other writing commands.
func TestNewPrintsTheIdentifierAndNothingElse(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	got := m.run(t, "new", "Normalize CRLF in the diff",
		"--type", "bug", "--priority", "high",
		"--add-ac", "The diff ignores CRLF").assertCode(t, 0)

	assertEqual(t, got.stdout, "MYP-1\n", "the output of biso new")
	if got.stderr != "" {
		t.Errorf("biso new wrote to stderr, and it had nothing to say: %s", got.stderr)
	}
}

// TestNewDryRunSaysWhatItWouldHaveDoneAndWritesNothing is the ending this
// branch decided for a preview over one single task
// (docs/spec/cmd/new.md#--dry-run-sobre-una-sola-tarea).
func TestNewDryRunSaysWhatItWouldHaveDoneAndWritesNothing(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	got := m.run(t, "new", "Normalize CRLF in the diff", "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stdout, "", "the standard output of a preview")
	assertEqual(t, got.stderr,
		"1 task would be created, nothing was written (--dry-run)\n",
		"the line of a preview")

	// The identifier was not spent either: the first real task takes it.
	real := m.run(t, "new", "The first real one").assertCode(t, 0)
	assertEqual(t, real.stdout, "MYP-1\n", "the identifier of the first task really created")
}

// TestTheSameTextIsWorthTheSameInBothDirections is the acceptance table of
// docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos,
// run through the compiled program. This is the writing half; the reading
// half is the same table through `biso ls`, and it checks the same thing:
// that neither command validates a vocabulary on its own instead of calling
// the one matching algorithm there is.
func TestTheSameTextIsWorthTheSameInBothDirections(t *testing.T) {
	for _, c := range []struct {
		typed    string
		resolves string
		code     int
	}{
		{typed: "To Do", resolves: "To Do"},
		{typed: "todo", resolves: "To Do"},
		{typed: "TO_DO", resolves: "To Do"},
		{typed: "In-Progress", resolves: "In Progress"},
		{typed: "Pending", code: 3},
		{typed: "", code: 3},
	} {
		t.Run("status "+c.typed, func(t *testing.T) {
			m := newMachine(t)
			m.env["BISO_ME"] = "@claude"
			m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
			m.run(t, "new", "A task", "--add-assignees", "@claude").assertCode(t, 0)

			got := m.run(t, "set", "MYP-1", "--status", c.typed)

			if c.code != 0 {
				got.assertCode(t, c.code)
				assertEqual(t, got.stderr,
					"error: unknown status: \""+c.typed+"\"\n"+
						"       valid statuses on this board: To Do, In Progress, Done\n",
					"the message of an unknown status")
				return
			}
			got.assertCode(t, 0)
			if !strings.Contains(got.stdout, "  "+c.resolves+"  ") {
				t.Errorf("%q was written as something other than %q: %s",
					c.typed, c.resolves, got.stdout)
			}
		})
	}
}
