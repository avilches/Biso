package main

import (
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// This file walks the two tables of behaviour, the one of
// docs/spec/cmd/ls.md#comportamiento-caso-a-caso and the one of
// docs/spec/cmd/get.md#comportamiento-caso-a-caso, through the compiled
// program.

// smallBoard is a board with a handful of tasks and nothing special about
// it, for the cases that only need a board to exist.
func smallBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	for i := 1; i <= 4; i++ {
		m.run(t, "new", "Task "+strconv.Itoa(i)).assertCode(t, 0)
	}
	m.run(t, "set", "MYP-1", "--add-labels", "frontend,backend",
		"--add-assignees", "@claude").assertCode(t, 0)
	m.run(t, "set", "MYP-2", "--add-labels", "frontend").assertCode(t, 0)
	return m
}

// TestListRejectsAValueTheVocabularyDoesNotHave is the first row of that
// table: a filter can never answer an empty list because the value was
// misspelled, because whoever reads the empty list takes it for a fact
// about the board.
func TestListRejectsAValueTheVocabularyDoesNotHave(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "ls", "--type", "epic").assertCode(t, 3)

	assertEqual(t, got.stderr,
		"error: unknown type: \"epic\"\n"+
			"       valid types on this board: task, bug, docs\n",
		"the message of an unknown type")
	assertEqual(t, got.stdout, "", "the standard output of a rejected filter")
}

// TestListSuggestsTheClosestLabelAndAssignee is the second row: a label or
// a person the board does not have is exit code 3 with up to five of the
// closest ones (docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué).
func TestListSuggestsTheClosestLabelAndAssignee(t *testing.T) {
	m := smallBoard(t)

	label := m.run(t, "ls", "-l", "fronted").assertCode(t, 3)
	assertEqual(t, label.stderr,
		"error: unknown label: \"fronted\"\nhint: did you mean: frontend?\n",
		"the message of an unknown label")

	assignee := m.run(t, "ls", "-a", "@clude").assertCode(t, 3)
	assertEqual(t, assignee.stderr,
		"error: unknown assignee: \"@clude\"\nhint: did you mean: @claude?\n",
		"the message of an unknown assignee")
}

// TestListUncheckedTurnsOffThoseTwoChecksAndNothingElse is the third row,
// with the limit that page puts on it: the configured vocabularies keep
// validating.
func TestListUncheckedTurnsOffThoseTwoChecksAndNothingElse(t *testing.T) {
	m := smallBoard(t)

	accepted := m.run(t, "ls", "-l", "fronted", "--unchecked").assertCode(t, 0)
	assertEqual(t, accepted.stdout, "", "a label nobody has, with --unchecked")

	// And the vocabulary of a configured field still answers exit code 3.
	m.run(t, "ls", "--type", "epic", "--unchecked").assertCode(t, 3)
}

// TestListMineWithoutAnIdentityIsExitCodeSix is the row of its own that the
// table of exit codes of docs/spec/cmd/ls.md gives it, and the message of
// docs/spec/invocacion.md#variables-de-entorno.
func TestListMineWithoutAnIdentityIsExitCodeSix(t *testing.T) {
	m := smallBoard(t)
	delete(m.env, "BISO_ME")

	got := m.run(t, "ls", "--mine").assertCode(t, 6)

	assertEqual(t, got.stderr,
		"error: --mine needs an identity; set BISO_ME, "+
			"or add \"me\" to ~/.biso/config.json\n",
		"the message of --mine without an identity")
}

// TestListHidesTheTerminalStatusUnlessItIsNamed is the rule of the default
// value of -s: the terminal status is out unless -s names it or
// --any-status asks for it, and --not-status does not bring it back.
func TestListHidesTheTerminalStatusUnlessItIsNamed(t *testing.T) {
	m := smallBoard(t)
	m.run(t, "set", "MYP-4", "--status", "Done").assertCode(t, 0)

	for _, c := range []struct {
		argv []string
		want []string
		what string
	}{
		{[]string{"ls", "--ids"}, []string{"MYP-1", "MYP-2", "MYP-3"}, "the default"},
		{[]string{"ls", "--ids", "-s", "Done"}, []string{"MYP-4"}, "-s Done"},
		{[]string{"ls", "--ids", "--any-status"},
			[]string{"MYP-1", "MYP-2", "MYP-3", "MYP-4"}, "--any-status"},
		{[]string{"ls", "--ids", "--not-status", "To Do"}, nil, "--not-status on its own"},
	} {
		got := m.run(t, c.argv...).assertCode(t, 0)
		assertEqual(t, got.stdout, joinLines(c.want), "the listing with "+c.what)
	}
}

// TestListArchivedTasksAreOutUnlessAskedFor is the other default of the
// same kind.
func TestListArchivedTasksAreOutUnlessAskedFor(t *testing.T) {
	m := smallBoard(t)
	m.archive(t, "MYP-2")

	for _, c := range []struct {
		argv []string
		want []string
	}{
		{[]string{"ls", "--ids"}, []string{"MYP-1", "MYP-3", "MYP-4"}},
		{[]string{"ls", "--ids", "--archived"},
			[]string{"MYP-1", "MYP-2", "MYP-3", "MYP-4"}},
		{[]string{"ls", "--ids", "--only-archived"}, []string{"MYP-2"}},
	} {
		got := m.run(t, c.argv...).assertCode(t, 0)
		assertEqual(t, got.stdout, joinLines(c.want),
			"the listing with "+strings.Join(c.argv[2:], " "))
	}
}

// TestListCombinesLabelsWithAndAndLabelOrWithOr is the one exception of the
// combination rules of docs/spec/cmd/ls.md#parámetros.
func TestListCombinesLabelsWithAndAndLabelOrWithOr(t *testing.T) {
	m := smallBoard(t)

	both := m.run(t, "ls", "--ids", "-l", "frontend", "-l", "backend").assertCode(t, 0)
	assertEqual(t, both.stdout, "MYP-1\n", "two labels ANDed")

	either := m.run(t, "ls", "--ids", "--label-or", "frontend",
		"--label-or", "backend").assertCode(t, 0)
	assertEqual(t, either.stdout, "MYP-1\nMYP-2\n", "two labels ORed")
}

// TestListFiltersByTheShapeOfTheBoardAndNotByStatusNames covers the four
// filters that ask about the board and not about a name: --blocked,
// --waiting, --active and their opposites.
func TestListFiltersByTheShapeOfTheBoardAndNotByStatusNames(t *testing.T) {
	m := smallBoard(t)
	m.run(t, "set", "MYP-1", "--add-deps", "MYP-2").assertCode(t, 0)
	m.run(t, "set", "MYP-3", "--status", "In Progress").assertCode(t, 0)
	// An open question is what --waiting asks about, and the verb that
	// asks one belongs to the next step, so the fixture writes it.
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET question_author = ?,
			question_asked_at = ?, question_body = ? WHERE id = 'MYP-4'`,
		"@sara", "2026-09-06T09:30:00Z", "Which endpoint?")

	for _, c := range []struct {
		flag string
		want []string
	}{
		{"--blocked", []string{"MYP-1"}},
		{"--not-blocked", []string{"MYP-2", "MYP-3", "MYP-4"}},
		{"--waiting", []string{"MYP-4"}},
		// In the default order and not in identifier order: MYP-2 blocks
		// MYP-1 and scores for it, and MYP-1 is blocked and scores
		// against (docs/spec/modelo-de-datos/urgencia.md).
		{"--not-waiting", []string{"MYP-2", "MYP-3", "MYP-1"}},
		{"--active", []string{"MYP-3"}},
		{"--not-active", []string{"MYP-2", "MYP-4", "MYP-1"}},
	} {
		got := m.run(t, "ls", "--ids", c.flag).assertCode(t, 0)
		assertEqual(t, got.stdout, joinLines(c.want), "the listing with "+c.flag)
	}
}

// TestListSortRejectsAFieldThatDoesNotExist is the usage error of
// docs/spec/cmd/ls.md#códigos-de-salida, with the valid fields named.
func TestListSortRejectsAFieldThatDoesNotExist(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "ls", "--sort", "importance").assertCode(t, 2)

	assertEqual(t, got.stderr,
		"error: --sort: unknown value: \"importance\"\n"+
			"       valid sort fields: urgency, id, ordinal, due, updated, created, title\n",
		"the message of an unknown sort field")
}

func TestListRejectsANegativeLimit(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "ls", "--limit", "-1").assertCode(t, 2)

	if !strings.Contains(got.stderr, "--limit: not a whole number of rows") {
		t.Errorf("the message of a negative limit is %q", got.stderr)
	}
}

// TestListSkipsATaskItCannotReadAndSaysSo is the row of the table that
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// owns: a set read never aborts because of one bad task and never hides it.
func TestListSkipsATaskItCannotReadAndSaysSo(t *testing.T) {
	m := smallBoard(t)
	// A priority the board does not configure any more is one of the ways
	// a task becomes unreadable.
	m.execOnBoard(t, m.boardDir(t),
		`UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "ls", "--ids").assertCode(t, 0)

	assertEqual(t, got.stdout, "MYP-1\nMYP-3\nMYP-4\n", "the rest of the listing")
	if !strings.Contains(got.stderr,
		"warning: 1 task could not be read and was skipped: MYP-2") {
		t.Errorf("the skipped task is not named: %q", got.stderr)
	}
}

// TestGetOfAnUnreadableTaskIsExitCodeThree is the other half of that rule:
// a targeted read of the same task fails, because there is nothing else to
// answer with.
func TestGetOfAnUnreadableTaskIsExitCodeThree(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t),
		`UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "get", "MYP-2").assertCode(t, 3)

	if !strings.Contains(got.stderr, "MYP-2 cannot be read") {
		t.Errorf("the reason is not in the message: %q", got.stderr)
	}
}

// TestGetByTextSaysWhichTaskItMatched is the third row of the table of
// docs/spec/cmd/get.md.
func TestGetByTextSaysWhichTaskItMatched(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "CRLF", "--section", "meta").assertCode(t, 0)

	if !strings.Contains(got.stderr, "note: \"CRLF\" matched MYP-11\n") {
		t.Errorf("the note of a text that matched one task is %q", got.stderr)
	}
}

// TestGetByTextWithSeveralMatchesPrintsTheCandidates is the second row, and
// the half of docs/spec/referencias.md that was missing: the candidates
// come out on stdout in the format of `biso ls`, and the error line on
// stderr.
func TestGetByTextWithSeveralMatchesPrintsTheCandidates(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "get", "Task").assertCode(t, 5)

	assertEqual(t, got.stderr, "error: \"Task\" matches 4 tasks\n",
		"the error of an ambiguous reference")
	// The same four rows `biso ls` would print, in the same order.
	listing := m.run(t, "ls").assertCode(t, 0)
	assertEqual(t, got.stdout, listing.stdout, "the candidates of an ambiguous reference")
}

// TestGetOfAnArchivedTaskSaysSo is the fourth row.
func TestGetOfAnArchivedTaskSaysSo(t *testing.T) {
	m := smallBoard(t)
	m.archive(t, "MYP-2")

	got := m.run(t, "get", "MYP-2", "--section", "meta").assertCode(t, 0)

	if !strings.Contains(got.stderr, "note: MYP-2 is archived\n") {
		t.Errorf("the note of an archived task is %q", got.stderr)
	}
}

// TestGetRejectsASectionThatDoesNotExist is the sixth row: exit code 2 with
// the eight valid names, which is what makes it eight sections and not
// nine.
func TestGetRejectsASectionThatDoesNotExist(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "get", "MYP-1", "--section", "dod").assertCode(t, 2)

	assertEqual(t, got.stderr,
		"error: --section: unknown value: \"dod\"\n"+
			"       valid sections: meta, desc, ac, plan, notes, summary, "+
			"comments, question\n",
		"the message of an unknown section")
}

// TestGetCountsARepeatedSectionOnceAndWarns is the rule any repeatable flag
// follows (docs/spec/valores-de-entrada.md#repetición-y-listas-separadas-por-comas).
func TestGetCountsARepeatedSectionOnceAndWarns(t *testing.T) {
	m := cardBoard(t)

	got := m.run(t, "get", "MYP-11", "--section", "ac,ac").assertCode(t, 0)

	if n := strings.Count(got.stdout, "## Acceptance Criteria"); n != 1 {
		t.Errorf("the repeated section was printed %d times", n)
	}
	if !strings.Contains(got.stderr, "--section: \"ac\" given twice, kept once") {
		t.Errorf("the repetition was not warned about: %q", got.stderr)
	}
}

func TestGetRefusesToBeToldBothHowToReadTheReference(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "get", "MYP-1", "--id", "--match").assertCode(t, 2)

	assertEqual(t, got.stderr, "error: --id and --match cannot be used together\n",
		"the message of --id with --match")
}

// TestGetTakesTheThreeShapesOfAnIdentifier is the grammar of
// docs/spec/referencias.md#la-gramática.
func TestGetTakesTheThreeShapesOfAnIdentifier(t *testing.T) {
	m := smallBoard(t)

	for _, ref := range []string{"MYP-1", "myp-1", "1", "#1"} {
		got := m.run(t, "get", ref, "--section", "meta").assertCode(t, 0)
		if !strings.HasPrefix(got.stdout, "MYP-1  Task 1\n") {
			t.Errorf("%q did not resolve to MYP-1: %q", ref, got.stdout)
		}
	}
}

// TestTheSameTextIsWorthTheSameWhenFilteringAsWhenWriting is the acceptance
// table of docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos,
// run in both directions in the same test.
//
// The writing half of it lives in write_golden_test.go; this one is what
// closes the gap that TASK-11 could not close from inside internal/match,
// where there is one implementation and no command: here the two commands
// are really run, and a command that validated a vocabulary on its own
// instead of calling match.Match would show up as a different resolved
// value or a different message.
func TestTheSameTextIsWorthTheSameWhenFilteringAsWhenWriting(t *testing.T) {
	for _, c := range []struct {
		typed string
		code  int
	}{
		{typed: "To Do"},
		{typed: "todo"},
		{typed: "TO_DO"},
		{typed: "In-Progress"},
		{typed: "Pending", code: 3},
		{typed: "", code: 3},
	} {
		t.Run("status "+c.typed, func(t *testing.T) {
			m := newMachine(t)
			m.env["BISO_ME"] = "@avilches"
			m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
			m.run(t, "new", "A task").assertCode(t, 0)

			written := m.run(t, "set", "MYP-1", "--status", c.typed)
			filtered := m.run(t, "ls", "--ids", "--status", c.typed)

			written.assertCode(t, orZero(c.code))
			filtered.assertCode(t, orZero(c.code))
			if c.code != 0 {
				// The same code, the same message and the same list of
				// valid values in both directions.
				assertEqual(t, filtered.stderr, written.stderr,
					"the message of "+strconv.Quote(c.typed)+" in the two directions")
				return
			}
			// And when it resolves, the filter finds the task the write
			// left in that status.
			assertEqual(t, filtered.stdout, "MYP-1\n",
				"the task written with "+strconv.Quote(c.typed)+" and filtered by it")
		})
	}
}

func orZero(code int) int { return code }

// joinLines writes the identifiers one per line, which is what --ids
// prints, and the empty string for a listing with nothing in it.
func joinLines(ids []string) string {
	var b strings.Builder
	for _, id := range ids {
		b.WriteString(id)
		b.WriteString("\n")
	}
	return b.String()
}

// boardDir is where this machine's only board lives, which the tests that
// write a fixture straight into the database need.
func (m *machine) boardDir(t *testing.T) string {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(m.home, ".biso", "boards", "*"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("this machine has %d boards and not one: %v", len(matches), err)
	}
	return matches[0]
}

// archive marks a task archived. `biso archive` is the command that does
// it and it belongs to a later step, so the fixture writes the column.
func (m *machine) archive(t *testing.T, id string) {
	t.Helper()
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET archived = 1 WHERE id = ?`, id)
}

// TestPrintShowsTheWholeCardOfEveryTaskAWriteAffected is the global flag of
// docs/spec/cmd/flags-globales.md, which had nothing to print until
// `biso get` existed: the card it names is that command's card.
func TestPrintShowsTheWholeCardOfEveryTaskAWriteAffected(t *testing.T) {
	m := smallBoard(t)

	written := m.run(t, "set", "MYP-1", "MYP-3", "--priority", "high",
		"--print").assertCode(t, 0)
	card := m.run(t, "get", "MYP-1").assertCode(t, 0)

	if !strings.HasPrefix(written.stdout, card.stdout) {
		t.Errorf("--print did not print the card of biso get:\n%s", written.stdout)
	}
	// One card per task, separated by a blank line, and never the status
	// line as well: every datum of that line is inside the card.
	if !strings.Contains(written.stdout, "\nMYP-3  Task 3\n") {
		t.Errorf("--print left out the second task:\n%s", written.stdout)
	}
	if strings.Contains(written.stdout, "urgency 5.8  ") {
		t.Errorf("--print printed the status line as well:\n%s", written.stdout)
	}
}

// TestNewPrintShowsTheCardOfTheTaskItCreated is the same flag on the other
// writing command of this step.
func TestNewPrintShowsTheCardOfTheTaskItCreated(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "new", "A brand new task", "--print").assertCode(t, 0)

	if !strings.HasPrefix(got.stdout, "MYP-5  A brand new task\n") {
		t.Errorf("biso new --print printed %q", got.stdout)
	}
	if !strings.Contains(got.stdout, "## Description") {
		t.Errorf("biso new --print printed something other than the card:\n%s", got.stdout)
	}
}

// TestSetByTextWithSeveralMatchesPrintsTheCandidates is the same ending as
// in `biso get`: the resolution of a reference has no variant per command
// (docs/spec/referencias.md).
func TestSetByTextWithSeveralMatchesPrintsTheCandidates(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "set", "Task", "--priority", "high").assertCode(t, 5)

	assertEqual(t, got.stderr, "error: \"Task\" matches 4 tasks\n",
		"the error of an ambiguous reference in biso set")
	listing := m.run(t, "ls").assertCode(t, 0)
	assertEqual(t, got.stdout, listing.stdout, "the candidates biso set printed")
	// And it wrote nothing: the priority of the first task is still empty.
	after := m.run(t, "ls", "--priority", "high", "--ids").assertCode(t, 0)
	assertEqual(t, after.stdout, "", "the tasks a failed write left behind")
}

// TestGetWithJsonAnswersTheCandidatesEnvelope is the one command that
// answers a data envelope with a failing exit code, which the table of
// docs/spec/contrato-json.md#el-sobre gives to `get` and to no other.
func TestGetWithJsonAnswersTheCandidatesEnvelope(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "get", "Task", "--json").assertCode(t, 5)

	envelope := envelopeOf(t, got.stdout)
	if envelope["kind"] != "task.candidates" {
		t.Errorf("the kind is %v and not task.candidates", envelope["kind"])
	}
	tasks := envelope["data"].(map[string]any)["tasks"].([]any)
	if len(tasks) != 4 {
		t.Errorf("the envelope carries %d candidates and not four", len(tasks))
	}
	if envelope["error"] != nil {
		t.Error("a data envelope carries an error as well")
	}
}

// TestListWithAnAmbiguousParentAnswersTheOrdinaryErrorEnvelope is the other
// side of that rule: every command but `get` answers the error envelope.
func TestListWithAnAmbiguousParentAnswersTheOrdinaryErrorEnvelope(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "ls", "-p", "Task", "--json").assertCode(t, 5)

	envelope := envelopeOf(t, got.stderr)
	if envelope["kind"] != "error" {
		t.Errorf("the kind is %v and not error", envelope["kind"])
	}
	if envelope["error"].(map[string]any)["code"] != "ambiguous_reference" {
		t.Errorf("the code is %v", envelope["error"])
	}
}

// TestListParentFiltersTheSubtasksOfTheResolvedTask, with the identifier
// resolved and carried in the effective filter.
func TestListParentFiltersTheSubtasksOfTheResolvedTask(t *testing.T) {
	m := smallBoard(t)
	m.run(t, "set", "MYP-2", "--parent", "MYP-1").assertCode(t, 0)
	m.run(t, "set", "MYP-3", "--parent", "MYP-1").assertCode(t, 0)

	got := m.run(t, "ls", "--ids", "-p", "1").assertCode(t, 0)

	assertEqual(t, got.stdout, "MYP-2\nMYP-3\n", "the subtasks of MYP-1")

	envelope := m.run(t, "ls", "-p", "1", "--json").assertCode(t, 0)
	filters := envelopeOf(t, envelope.stdout)["data"].(map[string]any)["filters"].(map[string]any)
	if filters["parent"] != "MYP-1" {
		t.Errorf("the effective parent filter is %v and not the resolved id", filters["parent"])
	}
}

// TestListParentThatDoesNotExistIsExitCodeFour is the row of the table of
// exit codes of docs/spec/cmd/ls.md.
func TestListParentThatDoesNotExistIsExitCodeFour(t *testing.T) {
	m := smallBoard(t)

	got := m.run(t, "ls", "-p", "MYP-999").assertCode(t, 4)

	if !strings.Contains(got.stderr, "MYP-999 has never existed on this board") {
		t.Errorf("the message is %q", got.stderr)
	}
}
