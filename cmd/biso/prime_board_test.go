package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

// These are the cases of the BOARD block of `biso prime`: the two last
// steps of the cascade of docs/spec/presupuestos.md#el-presupuesto-de-tamaño,
// which are the ones that reach a board with no task at all, the archived
// tasks that block leaves out, and the empty vocabulary.

// TestTheCascadeTrimsTheVocabulariesOfTheBoardBlock is step 6, and the case
// that makes the cap true for every board and not only for every board that
// has tasks: this one holds no task whatsoever, so the four blocks the
// first five steps trim are not even printed, and the whole summary is the
// BOARD block (docs/spec/cmd/prime.md#el-recorte-en-cascada).
func TestTheCascadeTrimsTheVocabulariesOfTheBoardBlock(t *testing.T) {
	m := wordyBoard(t)

	got := m.run(t, "prime").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	board := boardBlockOf(got.stdout)
	// The two vocabularies go first, and with forty long names each there
	// is room for none of them.
	for _, line := range []string{
		"  types       +40 more\n",
		"  priorities  +40 more\n",
	} {
		if !strings.Contains(board, line) {
			t.Errorf("the BOARD block does not carry %q:\n%s", line, board)
		}
	}
	// The counts line is the last of the three to lose anything, so it
	// still names statuses and only then says how many it left out.
	counts := strings.Split(board, "\n")[1]
	if !strings.Contains(counts, "A rather long status name number 01 0 |") {
		t.Errorf("the counts line was trimmed before the two vocabularies:\n%s", counts)
	}
	if !strings.Contains(counts, " | +") || !strings.HasSuffix(counts, " more") {
		t.Errorf("the counts line does not say how many statuses it left out:\n%s", counts)
	}
}

// TestTheCascadeLeavesTheBoardBlockWholeWhileTheBlocksOfTasksCanStillGive is
// the order of those steps: the BOARD block is the sixth and not the first,
// so a board whose summary overflows because of its tasks prints its
// vocabulary whole.
func TestTheCascadeLeavesTheBoardBlockWholeWhileTheBlocksOfTasksCanStillGive(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 60, waiting: 40, assigned: 60})

	got := m.run(t, "prime", "--limit", "30").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	board := boardBlockOf(got.stdout)
	if strings.Contains(board, "more") {
		t.Errorf("the BOARD block was trimmed while the tasks could still give:\n%s", board)
	}
	for _, line := range []string{
		"  To Do 200 | In Progress 100 | Done 0\n",
		"  types       task, bug, docs\n",
		"  priorities  high, medium, low\n",
	} {
		if !strings.Contains(board, line) {
			t.Errorf("the BOARD block does not carry %q:\n%s", line, board)
		}
	}
}

// TestTheBoardBlockIsCutInTheMiddleAsALastResort is step 7, the floor of
// the whole cascade: a board name is free text of any length, so dropping
// elements can never bound the block on its own and the cap would have an
// exception without this cut.
func TestTheBoardBlockIsCutInTheMiddleAsALastResort(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", strings.Repeat("Name ", 2000), "--prefix", "MYP",
		"--at", "board").assertCode(t, 0)

	got := m.run(t, "prime").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	board := boardBlockOf(got.stdout)
	if !strings.HasSuffix(board, "...\n") {
		t.Errorf("the BOARD block was not closed with the three dots:\n%s", board)
	}
	// What is left is the beginning of the block and nothing else: the cut
	// happened inside the first line, so no other line of the block made it.
	if strings.Count(board, "\n") != 1 || !strings.HasPrefix(board, "BOARD  Name Name") {
		t.Errorf("the cut did not keep the beginning of the block:\n%s", board)
	}
	if !strings.Contains(got.stdout, "\n\nCOMMANDS  (") {
		t.Errorf("the cut ate the blank line that follows the block:\n%s", got.stdout)
	}
}

// TestPrimeLeavesTheArchivedTasksOutOfEveryLine is the rule of
// docs/spec/cmd/prime.md#la-salida-literal: the four blocks exclude the
// archived tasks, and so does the counts line, which would otherwise
// describe tasks no block is showing.
func TestPrimeLeavesTheArchivedTasksOutOfEveryLine(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")
	m.run(t, "new", "A task that is still live").assertCode(t, 0)
	m.run(t, "new", "A task that was archived").assertCode(t, 0)
	m.execOnBoard(t, dir, `UPDATE task SET archived = 1 WHERE id = 'MYP-2'`)

	got := m.run(t, "prime").assertCode(t, 0)

	if strings.Contains(got.stdout, "MYP-2") {
		t.Errorf("the archived task was listed:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "  To Do 1 | In Progress 0 | Done 0\n") {
		t.Errorf("the counts line counts the archived task:\n%s", got.stdout)
	}
	if n := len(blockOf(got.stdout, "NEXT UP")); n != 1 {
		t.Errorf("NEXT UP printed %d rows and not the one live task:\n%s", n, got.stdout)
	}
	// The envelope says the same thing, so neither output can drift alone.
	envelope := m.run(t, "prime", "--json").assertCode(t, 0)
	data := envelopeOf(t, envelope.stdout)["data"].(map[string]any)
	if n := len(data["nextUp"].([]any)); n != 1 {
		t.Errorf("data.nextUp carries %d tasks and not the one live task", n)
	}
	counts := data["board"].(map[string]any)["countByStatus"].(map[string]any)
	if counts["To Do"].(float64) != 1 {
		t.Errorf("countByStatus counts the archived task: %v", counts["To Do"])
	}
}

// TestPrimeOnABoardWhoseOnlyTaskIsArchivedIsNotAnEmptyBoard is the other
// half of that rule, and the one place where "archived" and "empty" pull in
// opposite directions: the four blocks leave the archived tasks out, but
// "empty" means no task at all, archived ones included
// (docs/spec/cmd/prime.md#tablero-vacío).
func TestPrimeOnABoardWhoseOnlyTaskIsArchivedIsNotAnEmptyBoard(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")
	m.run(t, "new", "A task that was archived").assertCode(t, 0)
	m.execOnBoard(t, dir, `UPDATE task SET archived = 1 WHERE id = 'MYP-1'`)

	got := m.run(t, "prime").assertCode(t, 0)

	if strings.Contains(got.stdout, "THE BOARD IS EMPTY") {
		t.Errorf("a board with an archived task called itself empty:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "  To Do 0 | In Progress 0 | Done 0\n") {
		t.Errorf("the counts line counts the archived task:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, lastRuleOfTheMessage+"Pick one,") {
		t.Errorf("the blocks of tasks did not disappear cleanly:\n%s", got.stdout)
	}
}

// TestPrimeWritesNoneForAVocabularyThatIsEmpty is the line of
// docs/spec/cmd/prime.md#la-salida-literal that says a board takes no
// --type at all.
//
// The state is reached the way a caller reaches it, with `biso config set
// types ""` (docs/spec/cmd/config.md#comportamiento-caso-a-caso), which is
// the only door to it: `biso init` cannot produce it, because --types with
// an empty value is a usage error like any other flag given nothing
// (docs/spec/valores-de-entrada.md#el-valor-vacío). Until step 9 this test
// wrote the empty list straight into the configuration table, so the door
// itself went untested; now it is the test.
func TestPrimeWritesNoneForAVocabularyThatIsEmpty(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board",
		"--types", "").assertCode(t, 2)
	emptied := m.run(t, "config", "set", "types", "").assertCode(t, 0)
	assertEqual(t, emptied.stderr, "note: types =\n", "the note of an emptied vocabulary")
	assertEqual(t, emptied.stdout, "", "the standard output of a biso config set")

	got := m.run(t, "prime").assertCode(t, 0)

	if !strings.Contains(got.stdout, "  types       (none)\n") {
		t.Errorf("an empty vocabulary did not print (none):\n%s", boardBlockOf(got.stdout))
	}
	if !strings.Contains(got.stdout, "  priorities  high, medium, low\n") {
		t.Errorf("the vocabulary that is not empty changed:\n%s", boardBlockOf(got.stdout))
	}
}

// wordyBoard is a board with no task and forty statuses, forty types and
// forty priorities of long name: the shape whose summary is nothing but the
// BOARD block and which overflows the cap on that block alone.
func wordyBoard(t *testing.T) *machine {
	t.Helper()
	names := func(what string) string {
		var out []string
		for i := 1; i <= 40; i++ {
			out = append(out, fmt.Sprintf("A rather long %s name number %02d", what, i))
		}
		return strings.Join(out, ",")
	}
	statuses := names("status")
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board",
		"--statuses", statuses,
		"--initial-status", "A rather long status name number 01",
		"--active-status", "A rather long status name number 02",
		"--terminal-status", "A rather long status name number 03",
		"--types", names("type"),
		"--priorities", names("priority")).assertCode(t, 0)
	return m
}

// boardBlockOf is the BOARD block of a message, its blank line excluded.
func boardBlockOf(message string) string {
	at := strings.Index(message, "BOARD  ")
	if at < 0 {
		return ""
	}
	rest := message[at:]
	end := strings.Index(rest, "\n\n")
	if end < 0 {
		return rest
	}
	return rest[:end+1]
}
