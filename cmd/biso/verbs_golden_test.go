package main

import (
	"strings"
	"testing"
)

// These are the golden tests of the six verbs of the cycle: the compiled
// program is run on the board of the examples and what it printed is
// compared, character for character, with the blocks transcribed from
// docs/spec/cmd/verbos-del-ciclo.md.
//
// The board is the same one the golden tests of `biso set` build, because
// every one of those examples speaks of MYP-11 and of the urgency 19.0 that
// board produces. What each test sets up on top of it is the criteria its
// own line counts.

// withTwoCriteria leaves MYP-11 with the two acceptance criteria of the
// examples, neither of them checked.
func withTwoCriteria(t *testing.T, m *machine) {
	t.Helper()
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "Dates keep their time zone").assertCode(t, 0)
}

// withOneChecked is the same two with the first one checked, which is the
// `ac 1/2` of most of the examples.
func withOneChecked(t *testing.T, m *machine) {
	t.Helper()
	withTwoCriteria(t, m)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)
}

func TestStartPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)

	got := m.run(t, "start", "MYP-11",
		"--append-plan", "1. Read the parser. 2. Add the CRLF case.").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "start-status-line.txt"),
		"the status line of biso start")
}

func TestStartDryRunPrintsTheSameLineUnderItsHeader(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)

	got := m.run(t, "start", "MYP-11", "--dry-run").assertCode(t, 0)

	assertEqual(t, "$ biso start MYP-11 --dry-run\n"+got.stdout,
		fixture(t, "start-dry-run.txt"), "the preview of biso start")
}

func TestStartOverAnArchivedTaskPrintsTheErrorOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)
	archive(t, m, "MYP-11")

	got := m.run(t, "start", "MYP-11").assertCode(t, 6)

	assertEqual(t, got.stderr, fixture(t, "start-archived.txt"),
		"the refusal of biso start over an archived task")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso start")
}

// The refusal of `biso start` over a task that is already closed, with the
// pointer to the one flag that would have reopened it in the same call.
func TestStartOverAFinishedTaskPrintsTheErrorOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)
	m.run(t, "set", "MYP-11", "--status", "Done").assertCode(t, 0)

	got := m.run(t, "start", "MYP-11").assertCode(t, 6)

	assertEqual(t, got.stdout, "", "the standard output of a failed biso start")
	if !strings.HasSuffix(got.stderr, fixture(t, "start-already-finished.txt")) {
		t.Errorf("stderr = %q, want it to end with the refusal of the specification", got.stderr)
	}
}

func TestNoteWithNoTextPrintsTheErrorOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "note", "MYP-11").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "note-missing-text.txt"),
		"the refusal of a biso note with nothing to append")
}

// The same refusal in the two verbs whose text is a question and an answer,
// each one with the noun its own page wrote.
func TestAskAndAnswerWithNoTextPrintTheErrorsOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)

	ask := m.run(t, "ask", "MYP-11").assertCode(t, 2)
	assertEqual(t, ask.stderr, fixture(t, "ask-missing-text.txt"),
		"the refusal of a biso ask with no question")

	answer := m.run(t, "answer", "MYP-11").assertCode(t, 2)
	assertEqual(t, answer.stderr, fixture(t, "answer-missing-text.txt"),
		"the refusal of a biso answer with no answer")
}

// --comment-author signs a comment, it does not write one, so a call that
// carries nothing else is the refusal above and not a write that says
// nothing (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
func TestCommentWithItsAuthorAloneIsTheRefusalOfACallWithNoText(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "comment", "MYP-11", "--comment-author", "@trello:juan").assertCode(t, 2)

	if !strings.HasPrefix(got.stderr, "error: biso comment needs a text to append\n") {
		t.Errorf("stderr = %q, want the refusal of the specification", got.stderr)
	}
	if got.stdout != "" {
		t.Errorf("stdout = %q, and nothing was written", got.stdout)
	}
}

func TestCommentRefusesASecondIdentifierWithItsOwnMessage(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "comment", "MYP-1", "MYP-2").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "comment-id-like.txt"),
		"the refusal of a positional that looks like an id")
}

// --strict names everything that is missing at once, with the same text the
// warnings would have carried without it.
func TestFinishWithStrictPrintsTheErrorOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "There is a test that covers it",
		"--check-ac", "1",
		"--append-summary", "Normalizes CRLF").assertCode(t, 0)
	// exampleBoard leaves the board at MYP-12, so one task in between puts
	// the unfinished subtask of the example at MYP-14.
	m.run(t, "new", "Not a subtask").assertCode(t, 0)
	m.run(t, "new", "A live subtask", "--parent", "MYP-11").assertCode(t, 0)

	got := m.run(t, "finish", "MYP-11", "--strict").assertCode(t, 6)

	assertEqual(t, got.stderr, fixture(t, "finish-strict.txt"),
		"the refusal of biso finish --strict")
	if got.stdout != "" {
		t.Errorf("stdout = %q, and --strict writes nothing", got.stdout)
	}
}

func TestNotePrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)

	got := m.run(t, "note", "MYP-11",
		"The parser already normalized LF, CRLF was missing").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "note-status-line.txt"),
		"the status line of biso note")
}

// The trap `biso note` sets on purpose, because it takes one single task
// while `biso set` takes several.
func TestNoteRefusesASecondIdentifierWithTheMessageOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "note", "MYP-1", "MYP-2").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "note-id-like.txt"),
		"the refusal of a positional that looks like an id")
}

func TestCommentPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)

	got := m.run(t, "comment", "MYP-11",
		"A user with a Windows clone reported this").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "comment-status-line.txt"),
		"the status line of biso comment")
	if !strings.Contains(got.stderr, "note: comment #1 by @claude") {
		t.Errorf("stderr = %q, want the key the comment took", got.stderr)
	}
}

func TestFinishPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)

	got := m.run(t, "finish", "MYP-11", "--check-ac", "all",
		"--append-summary", "Normalizes CRLF").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "finish-status-line.txt"),
		"the status line of biso finish")
}

func TestFinishDryRunPrintsTheSameLineUnderItsHeader(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)
	m.run(t, "set", "MYP-11", "--append-summary", "Normalizes CRLF").assertCode(t, 0)

	got := m.run(t, "finish", "MYP-11", "--check-ac", "all", "--dry-run").assertCode(t, 0)

	assertEqual(t, "$ biso finish MYP-11 --check-ac all --dry-run\n"+got.stdout,
		fixture(t, "finish-dry-run.txt"), "the preview of biso finish")
}

// The warning lists the criteria that are missing, under the message and
// indented, which is the one warning of the program that carries display
// lines of its own.
func TestFinishListsTheCriteriaThatAreStillUnchecked(t *testing.T) {
	m := exampleBoard(t)
	// Three criteria created and one removed leaves the two of the
	// example holding the keys #2 and #3, which is what the block shows.
	m.run(t, "set", "MYP-11",
		"--add-ac", "A criterion that will be removed",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "There is a test that covers it").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--rm-ac", "1").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "2").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--append-summary", "Normalizes CRLF").assertCode(t, 0)

	got := m.run(t, "finish", "MYP-11").assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "finish-ac-warning.txt"),
		"the warning of the unchecked criteria")
}

func TestFinishListsTheUnfinishedSubtasksAndMarksTheArchivedOnes(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)
	m.run(t, "set", "MYP-11", "--check-ac", "all",
		"--append-summary", "Normalizes CRLF").assertCode(t, 0)
	// exampleBoard leaves the board at MYP-12, so one task in between
	// puts the two subtasks of the example at MYP-14 and MYP-15.
	m.run(t, "new", "Not a subtask").assertCode(t, 0)
	m.run(t, "new", "A live subtask", "--parent", "MYP-11").assertCode(t, 0)
	m.run(t, "new", "An archived subtask", "--parent", "MYP-11").assertCode(t, 0)
	archive(t, m, "MYP-15")

	got := m.run(t, "finish", "MYP-11").assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "finish-subtasks-warning.txt"),
		"the warning of the unfinished subtasks")
}

func TestAskPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)

	got := m.run(t, "ask", "MYP-11",
		"Do we normalize binary files too, or only text?").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "ask-status-line.txt"),
		"the status line of biso ask")
}

func TestAskPrintsTheTwoRefusalsOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)
	m.run(t, "ask", "MYP-11", "Do we normalize binary files too?").assertCode(t, 0)

	second := m.run(t, "ask", "MYP-11", "And what about symlinks?").assertCode(t, 6)
	assertEqual(t, second.stderr, fixture(t, "ask-open-question.txt"),
		"the refusal of a second question")

	m.run(t, "answer", "MYP-11", "Only text files.").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--status", "Done").assertCode(t, 0)

	finished := m.run(t, "ask", "MYP-11", "And what about symlinks?").assertCode(t, 6)
	assertEqual(t, finished.stderr, fixture(t, "ask-finished.txt"),
		"the refusal over a finished task")
}

func TestAskRefusesASecondIdentifierWithItsOwnMessage(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "ask", "MYP-1", "MYP-2").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "ask-id-like.txt"),
		"the refusal of a positional that looks like an id")
}

func TestAnswerPrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)
	m.run(t, "ask", "MYP-11", "Do we normalize binary files too?").assertCode(t, 0)

	got := m.run(t, "answer", "MYP-11",
		"Only text files. Binary ones are skipped entirely.").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "answer-status-line.txt"),
		"the status line of biso answer")
}

func TestAnswerWithoutAQuestionPrintsTheErrorOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withOneChecked(t, m)

	got := m.run(t, "answer", "MYP-11", "Only text files.").assertCode(t, 6)

	assertEqual(t, got.stderr, fixture(t, "answer-no-question.txt"),
		"the refusal of answering nothing")
}

func TestAnswerRefusesASecondIdentifierWithItsOwnMessage(t *testing.T) {
	m := exampleBoard(t)

	got := m.run(t, "answer", "MYP-1", "MYP-2").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "answer-id-like.txt"),
		"the refusal of a positional that looks like an id")
}

// The whole cycle of the page's opening block, run as three calls, which is
// what those three lines promise a task's life looks like.
func TestTheThreeCallsOfTheCycleOfTheSpecification(t *testing.T) {
	m := exampleBoard(t)
	withTwoCriteria(t, m)

	m.run(t, "start", "MYP-11",
		"--append-plan", "1. Read the parser. 2. Add the CRLF case.").assertCode(t, 0)
	m.run(t, "note", "MYP-11",
		"The parser already normalized LF, CRLF was missing").assertCode(t, 0)
	got := m.run(t, "finish", "MYP-11", "--check-ac", "all",
		"--append-summary",
		"Normalize CRLF in the diff, verified with the tests.").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "finish-status-line.txt"),
		"the status line the cycle ends with")
	if got.stderr != "" {
		t.Errorf("stderr = %q, and the cycle left nothing missing", got.stderr)
	}
}

// archive flips the one field no command of this step writes: `biso archive`
// belongs to the last step of the implementation, and three of these
// examples speak of an archived task. It is a fixture of the board and not a
// use of the program, so it is written straight into the database.
func archive(t *testing.T, m *machine, id string) {
	t.Helper()
	b := openTheBoard(t, m)
	defer b.Close()
	if _, err := b.Store.Exec("UPDATE task SET archived = 1 WHERE id = ?", id); err != nil {
		t.Fatal(err)
	}
}
