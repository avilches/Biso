package main

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"biso/internal/cli"
	"biso/internal/model"
)

// These are the golden tests of `biso prime`: the compiled program runs on
// the board of docs/spec/cmd/prime.md and what it printed is compared,
// character for character, with the block transcribed from that page.
//
// The board is built to be the board of that example and the message is
// not adjusted to whatever came out. In particular the order of every
// block is the order of the urgency and never an ordinal put there to
// force it: MYP-11 blocks MYP-40 and MYP-19 blocks MYP-33, which is what
// puts MYP-52 above MYP-40 and MYP-61 above MYP-33.

// The nine tasks the example names, and the composition of the rest.
const (
	primeBoardTasks = 248
	primeBoardDone  = 190
)

// primeBoard is the board of docs/spec/cmd/prime.md#la-salida-literal.
func primeBoard(t *testing.T) (*machine, string) {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board",
		"--types", "idea,memory,task,bug,docs").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")

	// The two hundred and forty eight tasks are seeded with one statement
	// and not with two hundred and forty eight calls: they are a fixture
	// and not a use of the program, and the nine the example names are the
	// only ones whose fields the message shows. They are all created today
	// so that the age term of the urgency is zero for every one of them,
	// which is what keeps the order of the blocks a property of the board
	// and not of the day the suite runs.
	today := time.Now().UTC().Format(model.InstantLayout)
	m.execOnBoard(t, dir, `
		WITH RECURSIVE n(i) AS (
			SELECT 1 UNION ALL SELECT i + 1 FROM n WHERE i < ?
		)
		INSERT INTO task (
			id, num, title, status, type, priority, parent, author, due,
			ordinal, description, plan, notes, summary, created_at, updated_at,
			archived, lease_expires_at, lease_holder, question_author,
			question_asked_at, question_body, next_criterion_key, next_comment_key
		)
		SELECT 'MYP-' || i, i, 'Task ' || i, 'To Do', '', 'low', '', '', '',
			NULL, '', '', '', '', ?, ?, 0, '', '', '', '', '', 1, 1
		FROM n`, primeBoardTasks, today, today)
	m.execOnBoard(t, dir, `UPDATE board_counter SET last_task_num = ?`, primeBoardTasks)

	named := map[int]bool{7: true, 11: true, 19: true, 33: true, 40: true,
		44: true, 52: true, 60: true, 61: true}
	var done []string
	for i := primeBoardTasks; i >= 1 && len(done) < primeBoardDone; i-- {
		if !named[i] {
			done = append(done, "MYP-"+strconv.Itoa(i))
		}
	}
	m.execOnBoard(t, dir, fmt.Sprintf(
		`UPDATE task SET status = 'Done' WHERE id IN ('%s')`, strings.Join(done, "','")))

	m.run(t, "set", "MYP-7", "--title", "Crash on an empty repository",
		"--type", "bug", "--priority", "high", "--due", "2026-09-08",
		"--add-ac", "It does not crash", "--add-ac", "There is a test",
		"--add-ac", "The message names the repository",
		"--add-ac", "The exit code is 2").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--title", "Normalize CRLF in the diff",
		"--type", "bug", "--priority", "high", "--status", "In Progress",
		"--add-assignees", "@claude",
		"--add-ac", "The diff ignores CRLF",
		"--add-ac", "There is a test that covers it").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1").assertCode(t, 0)
	m.run(t, "set", "MYP-19", "--title", "Retry the upload on 5xx",
		"--type", "task", "--priority", "high",
		"--add-ac", "It retries three times",
		"--add-ac", "It gives up with a message").assertCode(t, 0)
	m.run(t, "set", "MYP-33", "--title", "Add a retry counter to the upload log",
		"--type", "task", "--priority", "medium", "--add-assignees", "@claude",
		"--add-deps", "MYP-19",
		"--add-ac", "The counter is written", "--add-ac", "It resets on success",
		"--add-ac", "There is a test").assertCode(t, 0)
	m.run(t, "set", "MYP-33", "--check-ac", "1").assertCode(t, 0)
	m.run(t, "set", "MYP-40", "--title", "Split the config loader",
		"--type", "task", "--priority", "medium", "--status", "In Progress",
		"--add-assignees", "@claude", "--add-deps", "MYP-11",
		"--add-ac", "The loader is split",
		"--add-ac", "Every caller still compiles").assertCode(t, 0)
	m.run(t, "set", "MYP-44", "--title", "Wrong column width on narrow ttys",
		"--type", "bug", "--priority", "low",
		"--add-ac", "The width is right on eighty columns").assertCode(t, 0)
	m.run(t, "set", "MYP-52", "--title", "Document the release checklist",
		"--type", "task", "--priority", "low", "--status", "In Progress",
		"--add-ac", "The checklist is written").assertCode(t, 0)
	m.run(t, "set", "MYP-60", "--title", "Confirm the retry budget for the upload endpoint",
		"--type", "task", "--priority", "high", "--status", "In Progress",
		"--add-assignees", "@claude",
		"--add-ac", "The budget is agreed", "--add-ac", "It is written down").assertCode(t, 0)
	m.run(t, "ask", "MYP-60",
		"Should the retry budget be shared with the download endpoint or kept separate?").
		assertCode(t, 0)
	m.run(t, "set", "MYP-61", "--title", "Rewrite the install section",
		"--type", "docs", "--priority", "medium", "--add-assignees", "@claude",
		"--add-ac", "The section is rewritten").assertCode(t, 0)

	// The expired lease of MYP-52 goes in last and with a statement: no
	// command of biso hands a lease to somebody else, and any later write
	// on that task would empty it, because the task is assigned to nobody
	// (docs/spec/lease.md).
	m.execOnBoard(t, dir,
		`UPDATE task SET lease_expires_at = ?, lease_holder = ? WHERE id = 'MYP-52'`,
		"2026-09-05T09:00:00Z", "@bob")
	return m, dir
}

func TestPrimePrintsTheMessageOfTheSpecification(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "prime-output.txt"), "the message of biso prime")
	if got.stderr != "" {
		t.Errorf("biso prime wrote to stderr: %q", got.stderr)
	}
}

// TestPrimeFitsInItsBudget is the cap of
// docs/spec/presupuestos.md#el-presupuesto-de-tamaño, which is the one
// number docs/spec/estabilidad.md freezes. It is checked over the board of
// the example and, in the test below, over a board that makes the summary
// overflow.
func TestPrimeFitsInItsBudget(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
}

func TestPrimeEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime", "--json").assertCode(t, 0)

	assertSameShape(t, got.stdout, fixture(t, "prime-json.txt"))
}

func TestPrimeFullAddsTheFlagsOfNewAndSet(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime", "--full").assertCode(t, 0)

	message := fixture(t, "prime-output.txt")
	assertEqual(t, got.stdout, message+"\n"+fixture(t, "prime-full.txt"),
		"the message of biso prime --full")
}

// assertWithinBudget checks the hard cap of
// docs/spec/presupuestos.md#el-presupuesto-de-tamaño against the message a
// real process printed, which is the only place the number can be measured
// the way whoever reads it measures it.
func assertWithinBudget(t *testing.T, message string) {
	t.Helper()
	if len(message) > cli.PrimeBudget {
		t.Errorf("the message measures %d bytes and the cap is %d (docs/spec/presupuestos.md#el-presupuesto-de-tamaño)",
			len(message), cli.PrimeBudget)
	}
}
