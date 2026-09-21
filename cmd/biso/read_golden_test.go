package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"biso/internal/board"
)

// These are the golden tests of the two commands that read, `biso ls` and
// `biso get`: the compiled program runs on a real board and what it printed
// is compared, character for character, with the blocks transcribed from
// docs/spec/cmd/ls.md and docs/spec/cmd/get.md.
//
// The boards are built to be the boards of those examples, with the program
// itself wherever the program can build them, and the lines are not
// adjusted to whatever came out. The one thing built with a statement
// instead is the open question of MYP-60: the verb that asks one is
// `biso ask`, which belongs to the next step, so there is no call that
// could put it there yet.

// listingBoard is the board of the example of docs/spec/cmd/ls.md#salida.
// Its four visible tasks are the four of that block, and they come out in
// that order because of what they are and not because of how they were
// written: MYP-7 and MYP-11 both add up to an urgency of 19.0 and the
// identifier breaks the tie, MYP-19 adds up to 7.0 and MYP-23 to 4.0.
//
// MYP-40 is the task that depends on MYP-11, which is the `blocking` term
// of its urgency. It has to be unfinished for that term to count, so it is
// on the board and in the listing, with the lowest priority and nothing
// else, which puts it last.
func listingBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	for i := 1; i <= 40; i++ {
		m.run(t, "new", "Task "+strconv.Itoa(i)).assertCode(t, 0)
	}
	// Everything that is not one of the five is finished, so the default
	// listing leaves it out.
	var finished []string
	for i := 1; i <= 39; i++ {
		switch i {
		case 7, 11, 19, 23:
		default:
			finished = append(finished, "MYP-"+strconv.Itoa(i))
		}
	}
	m.run(t, append(append([]string{"set"}, finished...), "--status", "Done")...).assertCode(t, 0)

	m.run(t, "set", "MYP-7", "--title", "Crash on an empty repository",
		"--type", "bug", "--priority", "high", "--due", "2026-09-08",
		"--add-ac", "It does not crash", "--add-ac", "There is a test",
		"--add-ac", "The message names the repository",
		"--add-ac", "The exit code is 2").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--title", "Normalize CRLF in the diff",
		"--type", "bug", "--priority", "high", "--status", "In Progress",
		"--add-assignees", "@claude", "--add-labels", "parser",
		"--add-refs", "docs/bugs/BUG-02.md",
		"--add-ac", "The diff ignores CRLF",
		"--add-ac", "There is a test that covers it").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)
	m.run(t, "set", "MYP-19", "--title", "Retry the upload on 5xx",
		"--type", "task", "--priority", "high",
		"--add-ac", "It retries three times",
		"--add-ac", "It gives up with a message").assertCode(t, 0)
	m.run(t, "set", "MYP-23", "--title", "Rewrite the install section",
		"--type", "docs", "--priority", "medium",
		"--add-assignees", "@sara,@avilches",
		"--add-ac", "The section is rewritten").assertCode(t, 0)
	m.run(t, "set", "MYP-40", "--title", "Depends on the parser",
		"--priority", "low", "--add-deps", "MYP-11").assertCode(t, 0)
	return m
}

func TestListPrintsTheColumnsOfTheSpecification(t *testing.T) {
	m := listingBoard(t)

	// Four rows of the five that match, which is what the block of the
	// specification shows: the fifth is the task that makes MYP-11 block
	// something.
	got := m.run(t, "ls", "--limit", "4").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "ls-output.txt"), "the columns of biso ls")
}

// TestListCutsTheTitleAtAHundredCellsBeforeAligning is the two steps of
// docs/spec/cmd/ls.md#salida in the order that page fixes: a title longer
// than a hundred cells is cut, counting the three dots, and only then does
// the width of column 5 get computed, so one long title does not widen the
// table beyond a hundred.
func TestListCutsTheTitleAtAHundredCellsBeforeAligning(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", strings.Repeat("a", 150)).assertCode(t, 0)
	m.run(t, "new", "Short").assertCode(t, 0)

	got := m.run(t, "ls").assertCode(t, 0)

	lines := strings.Split(strings.TrimSuffix(got.stdout, "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("the listing has %d lines: %q", len(lines), got.stdout)
	}
	title := strings.Repeat("a", 97) + "..."
	if !strings.Contains(lines[0], title) {
		t.Errorf("the long title was not cut to a hundred cells: %q", lines[0])
	}
	if strings.Contains(lines[0], title+"a") {
		t.Errorf("the long title kept more than a hundred cells: %q", lines[0])
	}
	// Both rows pad column 5 to the same width, which is the width of the
	// cut title and not of the title that was stored.
	if len(lines[0]) != len(lines[1]) {
		t.Errorf("the two rows are not aligned:\n%q\n%q", lines[0], lines[1])
	}
}

// TestListMeasuresTheColumnsInTerminalCells is the unit of that algorithm:
// an ideograph measures two cells, so a title of three of them is as wide
// as six letters and the column is padded to the wider of the two.
func TestListMeasuresTheColumnsInTerminalCells(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "日本語").assertCode(t, 0)
	m.run(t, "new", "abcdef").assertCode(t, 0)

	got := m.run(t, "ls").assertCode(t, 0)

	// The two titles are as wide as each other, six cells, so neither of
	// them is padded and the two spaces after each one are the separator
	// and nothing else. Counting the ideographs as three characters would
	// have padded that row with three more spaces.
	rows := strings.Split(strings.TrimSuffix(got.stdout, "\n"), "\n")
	if !strings.Contains(rows[0], "日本語  -  -  -") {
		t.Errorf("the ideographs were not measured in cells: %q", rows[0])
	}
	if !strings.Contains(rows[1], "abcdef  -  -  -") {
		t.Errorf("the second row is padded differently: %q", rows[1])
	}
}

// boardOfFiftyEight is the board the other three blocks of
// docs/spec/cmd/ls.md#salida speak of: fifty eight tasks that match, which
// is what makes the default limit hide twenty eight of them.
func boardOfFiftyEight(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	dir := filepath.Join(m.home, ".biso", "boards", "my-project-3f9a2b1c")
	m.buildBoard(t, dir, "3f9a2b1c", "My project", "MYP", 58, 0, 58)
	m.write(t, filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	return m
}

func TestListSaysOnStderrWhatItLeftOut(t *testing.T) {
	m := boardOfFiftyEight(t)

	got := m.run(t, "ls").assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "ls-truncated.txt"), "the truncation warning")
	if lines := strings.Count(got.stdout, "\n"); lines != 30 {
		t.Errorf("the listing printed %d rows and not the default thirty", lines)
	}
}

func TestListCountPrintsOneNumberAndNothingElse(t *testing.T) {
	m := boardOfFiftyEight(t)

	got := m.run(t, "ls", "--count").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "ls-count.txt"), "the output of biso ls --count")
	if got.stderr != "" {
		t.Errorf("biso ls --count wrote to stderr: %q", got.stderr)
	}
}

func TestListIdsPrintsIdentifiersOnePerLine(t *testing.T) {
	m := listingBoard(t)

	got := m.run(t, "ls", "--ids", "--limit", "2").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "ls-ids.txt"), "the output of biso ls --ids")
}

// TestListLimitZeroPrintsNoRowAndTheTotal is the row of the table of
// docs/spec/cmd/ls.md#comportamiento-caso-a-caso: it is the way of counting
// without --count.
func TestListLimitZeroPrintsNoRowAndTheTotal(t *testing.T) {
	m := boardOfFiftyEight(t)

	got := m.run(t, "ls", "--limit", "0").assertCode(t, 0)

	assertEqual(t, got.stdout, "", "the standard output of biso ls --limit 0")
	if !strings.Contains(got.stderr, "58 more tasks match; showing 0 of 58") {
		t.Errorf("the total is not in the warning: %q", got.stderr)
	}
}

// TestListWithNoResultIsAFactAndNotAFailure is the other row of that table:
// a filter that is valid and matches nothing exits 0 with a note.
func TestListWithNoResultIsAFactAndNotAFailure(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	got := m.run(t, "ls", "--type", "bug").assertCode(t, 0)

	assertEqual(t, got.stdout, "", "the standard output of a listing with no result")
	assertEqual(t, got.stderr, "note: no tasks match\n", "the note of a listing with no result")
}

func TestGetPrintsTheCardOfTheSpecification(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11").assertCode(t, 0)

	assertEqual(t, got.stdout, m.withTheDatesOfTheSpecification(got.stdout, fixture(t, "get-output.txt")),
		"the card of biso get")
}

func TestGetSectionPrintsOnlyThatSection(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--section", "ac").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "get-section-ac.txt"), "biso get --section ac")
}

// TestGetSectionIsPrintedInTheOrderOfTheCard is the rule of
// docs/spec/cmd/get.md: several sections come out in the fixed order of the
// card and never in the order they were asked for.
func TestGetSectionIsPrintedInTheOrderOfTheCard(t *testing.T) {
	m := cardBoard(t)

	one := m.run(t, "get", "MYP-11", "--section", "plan,ac").assertCode(t, 0)
	other := m.run(t, "get", "MYP-11", "--section", "ac,plan").assertCode(t, 0)

	assertEqual(t, other.stdout, one.stdout, "the two orders of --section")
	if strings.Index(one.stdout, "## Acceptance Criteria") >
		strings.Index(one.stdout, "## Implementation Plan") {
		t.Errorf("the sections came out in the order they were asked for:\n%s", one.stdout)
	}
}

// TestGetSectionOfAnEmptySectionPrintsNothing is the last row of the table
// of docs/spec/cmd/get.md: an empty section that was asked for explicitly
// is left out, and when nothing is left the output is empty and the code is
// still 0.
func TestGetSectionOfAnEmptySectionPrintsNothing(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--section", "summary").assertCode(t, 0)

	assertEqual(t, got.stdout, "", "biso get --section summary of a task with no summary")
}

// TestGetWithoutSectionPrintsTheEightSectionsMarkingTheEmptyOnes is the
// other half of that rule, and the reason the card has exactly eight
// sections and not nine.
func TestGetWithoutSectionPrintsTheEightSectionsMarkingTheEmptyOnes(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-9").assertCode(t, 0)

	for _, heading := range []string{
		"## Description", "## Acceptance Criteria", "## Implementation Plan",
		"## Implementation Notes", "## Final Summary", "## Comments",
		"## Open Question",
	} {
		if !strings.Contains(got.stdout, heading) {
			t.Errorf("the whole card does not print %q:\n%s", heading, got.stdout)
		}
	}
	// A task with nothing but a title has seven of the eight sections
	// empty, and every one of them is printed and marked.
	if n := strings.Count(got.stdout, "(empty)"); n != 7 {
		t.Errorf("the whole card marked %d empty sections and not seven:\n%s", n, got.stdout)
	}
}

func TestGetExplainsTheUrgencyTermByTerm(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--explain-urgency", "--section", "ac").assertCode(t, 0)

	explanation := got.stdout[strings.Index(got.stdout, "urgency 19.0"):]
	assertEqual(t, explanation, fixture(t, "get-explain.txt"), "the explanation of the urgency")
}

func TestGetExplainsATerminalUrgencyWithOneLine(t *testing.T) {
	m := cardBoard(t)
	m.run(t, "set", "MYP-19", "--status", "Done").assertCode(t, 0)

	got := m.run(t, "get", "MYP-19", "--explain-urgency", "--section", "meta").assertCode(t, 0)

	explanation := got.stdout[strings.Index(got.stdout, "urgency 0.0"):]
	assertEqual(t, explanation, fixture(t, "get-terminal-urgency.txt"),
		"the explanation of a terminal task's urgency")
}

// TestGetPrintsTheOpenQuestionWithItsAuthorAndItsInstant is the third block
// of docs/spec/cmd/get.md#salida. The question is written with a statement
// because `biso ask` is the next step's command.
func TestGetPrintsTheOpenQuestionWithItsAuthorAndItsInstant(t *testing.T) {
	m := newMachine(t)
	dir := filepath.Join(m.home, ".biso", "boards", "my-project-3f9a2b1c")
	m.buildBoard(t, dir, "3f9a2b1c", "My project", "MYP", 60, 0, 60)
	m.write(t, filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	m.execOnBoard(t, dir, `UPDATE task SET title = ?,
			question_author = ?, question_asked_at = ?, question_body = ?
		WHERE id = 'MYP-60'`,
		"Confirm the retry budget for the upload endpoint",
		"@claude", "2026-09-06T09:30:00Z",
		"Should the retry budget be shared with the download endpoint or kept separate?")

	got := m.run(t, "get", "MYP-60", "--section", "question").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "get-question.txt"), "biso get --section question")
}

// cardBoard is the board of the examples of docs/spec/cmd/get.md: MYP-11
// with the body, the lease and the assignee those blocks show, and MYP-40
// depending on it, which is the `blocking` term of the urgency 19.0 they
// print.
//
// MYP-11 is created by @claude with --start, which is what leaves it in the
// active status, assigned to @claude and with the lease in @claude's hands,
// and its author is @avilches because --author says so. Every later call on
// it is made by @claude too, so the lease is renewed instead of being
// warned about.
func cardBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	for i := 1; i <= 10; i++ {
		m.run(t, "new", "Task "+strconv.Itoa(i)).assertCode(t, 0)
	}

	m.env["BISO_ME"] = "@claude"
	m.run(t, "new", "Normalize CRLF in the diff", "--start", "--author", "@avilches",
		"--type", "bug", "--priority", "high",
		"--add-labels", "parser", "--add-refs", "docs/bugs/BUG-02.md",
		"--append-desc", "The diff compares byte by byte and marks as different "+
			"two lines that only\ndiffer in the line ending.",
		"--append-plan", "1. Read the parser.\n2. Add the CRLF case.",
		"--append-note", "The parser already normalized LF, CRLF was missing.",
		"--add-ac", "The diff ignores CRLF",
		"--add-ac", "A criterion that will be removed",
		"--add-ac", "There is a test that covers it").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--rm-ac", "2", "--check-ac", "1").assertCode(t, 0)
	m.run(t, "set", "MYP-11",
		"--comment", "A user with a Windows clone reported this.",
		"--comment-author", "@avilches").assertCode(t, 0)

	m.env["BISO_ME"] = "@avilches"
	for i := 12; i <= 39; i++ {
		m.run(t, "new", "Task "+strconv.Itoa(i)).assertCode(t, 0)
	}
	m.run(t, "new", "Depends on the parser", "--add-deps", "MYP-11").assertCode(t, 0)
	return m
}

// withTheDatesOfTheSpecification puts this run's instants into the block of
// the specification. A card prints when the task was created, updated and
// commented, and no flag of the program fixes those: what the test checks
// is every other character of the block.
func (m *machine) withTheDatesOfTheSpecification(got, want string) string {
	gotLines := strings.Split(got, "\n")
	wantLines := strings.Split(want, "\n")
	for i := range wantLines {
		if i >= len(gotLines) {
			break
		}
		switch {
		case strings.HasPrefix(wantLines[i], "created "),
			strings.HasPrefix(wantLines[i], "lease "),
			strings.HasPrefix(wantLines[i], "#1  @avilches, "):
			wantLines[i] = gotLines[i]
		}
	}
	return strings.Join(wantLines, "\n")
}

// execOnBoard runs one statement against a board's database, which is how a
// fixture writes what no command of this step can write yet.
func (m *machine) execOnBoard(t *testing.T, dir, statement string, args ...any) {
	t.Helper()
	b, err := board.Open(&board.Location{ID: board.MarkerID(dir), Dir: dir}, board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	defer b.Close()
	if _, err := b.Store.Exec(statement, args...); err != nil {
		t.Fatal(fmt.Errorf("%s: %w", statement, err))
	}
}

// The two envelopes of these commands are checked key by key and not value
// by value: the schema of the specification is written about one board of
// one example, with its own identifiers and its own instants, and a real
// run cannot be made to repeat either. What has to match exactly is the
// shape, which is what a program consuming the output depends on
// (docs/spec/contrato-json.md#números-fechas-y-ausencias: the key that is
// documented for a kind is always there, whatever it is worth).

func TestListEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := listingBoard(t)

	got := m.run(t, "ls", "--limit", "4", "--json").assertCode(t, 0)

	assertSameShape(t, got.stdout, fixture(t, "ls-json.txt"))
}

func TestGetEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := cardBoard(t)

	// The schema of that page is the one of a call that asked for the
	// explanation, which is the only flag that adds a key.
	got := m.run(t, "get", "MYP-11", "--explain-urgency", "--json").assertCode(t, 0)

	assertSameShape(t, got.stdout, fixture(t, "get-json.txt"))
}

// TestTheFiltersOfTheEnvelopeAreTheEffectiveOnes is the promise of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls: data.filters is not an
// echo of the flags that were written, it is the filter that produced the
// listing, with the defaults resolved.
func TestTheFiltersOfTheEnvelopeAreTheEffectiveOnes(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "ls", "--mine", "--json").assertCode(t, 0)

	filters := envelopeOf(t, got.stdout)["data"].(map[string]any)["filters"].(map[string]any)
	// --mine resolved to the identity it used, and --status resolved to every
	// status but the terminal one.
	assertJSONEqual(t, filters["assignee"], []any{"@avilches"}, "the assignee filter")
	assertJSONEqual(t, filters["status"], []any{"To Do", "In Progress"}, "the status filter")
	assertJSONEqual(t, filters["blocked"], nil, "a filter nobody asked for")
	assertJSONEqual(t, filters["overdue"], false, "a unary filter nobody asked for")
}

// TestGetSectionCutsTheEnvelopeToWhatWasAsked is the exception the JSON
// contract declares: with --section, data.task carries id and the keys of
// those sections and no other.
func TestGetSectionCutsTheEnvelopeToWhatWasAsked(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--section", "ac", "--json").assertCode(t, 0)

	task := envelopeOf(t, got.stdout)["data"].(map[string]any)["task"].(map[string]any)
	if len(task) != 2 || task["id"] == nil || task["acceptanceCriteria"] == nil {
		t.Errorf("--section ac answered %d keys and not id plus the section: %v",
			len(task), keysOf(task))
	}
}

// TestEveryNumberOfTheUrgencyWritesItsDecimal is the rule of
// docs/spec/contrato-json.md#números-fechas-y-ausencias: the urgency and
// every term of its breakdown are decimals with one digit after the point.
// It reads the raw text and not the decoded document on purpose, because
// decoding is exactly what loses the difference between 6 and 6.0, and a
// consumer that reads the output as text sees it.
func TestEveryNumberOfTheUrgencyWritesItsDecimal(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--explain-urgency", "--json").assertCode(t, 0)

	for _, fragment := range []string{
		`"urgency": 19.0`,
		`"priority": 6.0`,
		`"value": 4.0`,
		`"blocking": 8.0`,
		`"blocked": 0.0`,
		`"due": 0.0`,
		`"criteria": 1.0`,
		`"age": 0.0`,
	} {
		if !strings.Contains(got.stdout, fragment) {
			t.Errorf("the envelope does not carry %s:\n%s", fragment, got.stdout)
		}
	}
}

// TestTheBreakdownOfATerminalTaskIsSevenZerosWithTheirDecimal is the
// paragraph of docs/spec/cmd/get.md that fixes the shape of the explanation
// of a task whose urgency is zero by definition: seven terms at 0.0 and the
// reason saying why.
func TestTheBreakdownOfATerminalTaskIsSevenZerosWithTheirDecimal(t *testing.T) {
	m := cardBoard(t)
	m.run(t, "set", "MYP-11", "--status", "Done").assertCode(t, 0)

	got := m.run(t, "get", "MYP-11", "--explain-urgency", "--json").assertCode(t, 0)

	for _, fragment := range []string{
		`"urgency": 0.0`,
		`"priority": 0.0`,
		`"value": 0.0`,
		`"reason": "not_active"`,
		`"blocking": 0.0`,
		`"blocked": 0.0`,
		`"due": 0.0`,
		`"criteria": 0.0`,
		`"age": 0.0`,
	} {
		if !strings.Contains(got.stdout, fragment) {
			t.Errorf("the envelope does not carry %s:\n%s", fragment, got.stdout)
		}
	}
}

func envelopeOf(t *testing.T, text string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(text), &envelope); err != nil {
		t.Fatalf("the envelope is not JSON: %v\n%s", err, text)
	}
	return envelope
}

func keysOf(object map[string]any) []string {
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func assertJSONEqual(t *testing.T, got, want any, what string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s is %#v and not %#v", what, got, want)
	}
}

// assertSameShape compares two JSON documents by their keys and by the type
// of what hangs off each one, and not by the values.
func assertSameShape(t *testing.T, got, want string) {
	t.Helper()
	var a, b any
	if err := json.Unmarshal([]byte(got), &a); err != nil {
		t.Fatalf("the program's envelope is not JSON: %v\n%s", err, got)
	}
	if err := json.Unmarshal([]byte(want), &b); err != nil {
		t.Fatalf("the specification's schema is not JSON: %v\n%s", err, want)
	}
	compareShape(t, "", a, b)
}

func compareShape(t *testing.T, path string, got, want any) {
	t.Helper()
	switch expected := want.(type) {
	case map[string]any:
		object, ok := got.(map[string]any)
		if !ok {
			t.Errorf("%s is %T and the schema has an object there", path, got)
			return
		}
		if a, b := keysOf(object), keysOf(expected); !reflect.DeepEqual(a, b) {
			t.Errorf("%s has the keys %v and the schema has %v", path, a, b)
			return
		}
		for key := range expected {
			compareShape(t, path+"."+key, object[key], expected[key])
		}
	case []any:
		list, ok := got.([]any)
		if !ok {
			t.Errorf("%s is %T and the schema has a list there", path, got)
			return
		}
		if len(expected) == 0 || len(list) == 0 {
			return
		}
		// Every element of a list has the same shape, so the first one of
		// each is enough and the lengths need not match: the schema shows
		// one task and a real board has as many as it has.
		compareShape(t, path+"[]", list[0], expected[0])
	case nil:
		// A key the schema shows as null can hold a value on a real board.
	default:
		if got == nil {
			// And the other way round: a key the schema shows with a value
			// can be null here, such as a task with no due date.
			return
		}
	}
}
