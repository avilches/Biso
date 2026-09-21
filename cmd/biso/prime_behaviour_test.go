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

// These are the cases of `biso prime` that are not the literal message:
// the board with nothing in it, the board with no identity, the task that
// could not be read, the two ways of calling it wrong, and the cascade
// that trims the summary when it does not fit
// (docs/spec/presupuestos.md#el-presupuesto-de-tamaño).

func TestPrimeOnAnEmptyBoardReplacesTheFourBlocks(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)

	got := m.run(t, "prime").assertCode(t, 0)

	if !strings.Contains(got.stdout, fixture(t, "prime-empty.txt")) {
		t.Errorf("the empty board block is not in the message:\n%s", got.stdout)
	}
	for _, heading := range []string{
		"IN PROGRESS", "NEEDS ANSWER", "ASSIGNED TO YOU", "NEXT UP",
	} {
		if strings.Contains(got.stdout, heading) {
			t.Errorf("an empty board printed the heading %q:\n%s", heading, got.stdout)
		}
	}
	// The rest of the message does not change, the counts line included.
	if !strings.Contains(got.stdout, "  To Do 0 | In Progress 0 | Done 0\n") {
		t.Errorf("the counts line of an empty board is missing:\n%s", got.stdout)
	}
	if !strings.HasSuffix(got.stdout, primeClosingParagraph) {
		t.Errorf("the closing paragraph changed on an empty board:\n%s", got.stdout)
	}
	assertWithinBudget(t, got.stdout)
}

// TestPrimeOnABoardWhoseTasksAreAllFinished is the case next to the one
// above: the board is not empty, so it does not print the empty-board
// block, and none of the four blocks has a row, so none of them prints
// (docs/spec/cmd/prime.md#tablero-vacío).
func TestPrimeOnABoardWhoseTasksAreAllFinished(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "Something that is done").assertCode(t, 0)
	m.run(t, "finish", "MYP-1", "--no-checks").assertCode(t, 0)

	got := m.run(t, "prime").assertCode(t, 0)

	if strings.Contains(got.stdout, "THE BOARD IS EMPTY") {
		t.Errorf("a board with a finished task called itself empty:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "  To Do 0 | In Progress 0 | Done 1\n") {
		t.Errorf("the finished task is not in the counts line:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "decided you should do.\n\nPick one,") {
		t.Errorf("the blocks of tasks did not disappear cleanly:\n%s", got.stdout)
	}
}

// TestPrimeWithoutAnIdentitySaysHowToSetOne is the line that replaces the
// identity, and the reason it travels inside the message instead of going
// to stderr as a note: `biso prime` writes nothing on stderr at all.
func TestPrimeWithoutAnIdentitySaysHowToSetOne(t *testing.T) {
	m, _ := primeBoard(t)
	delete(m.env, "BISO_ME")

	got := m.run(t, "prime").assertCode(t, 0)

	if !strings.Contains(got.stdout, fixture(t, "prime-identity.txt")) {
		t.Errorf("the hint about BISO_ME is not in the message:\n%s", got.stdout)
	}
	if strings.Contains(got.stdout, "ASSIGNED TO YOU") {
		t.Errorf("ASSIGNED TO YOU was printed with no identity:\n%s", got.stdout)
	}
	if got.stderr != "" {
		t.Errorf("biso prime wrote to stderr: %q", got.stderr)
	}
	// The two tasks that were assigned did not disappear: with no identity
	// they are part of NEXT UP like any other, so MYP-61 is now there.
	if !strings.Contains(blockText(got.stdout, "NEXT UP"), "MYP-61") {
		t.Errorf("the assigned tasks did not fall into NEXT UP:\n%s", got.stdout)
	}
}

// TestPrimeWithNoIdentityLeavesBoardMeNull is the same case in the
// envelope: there is no hint there, only the null
// (docs/spec/cmd/prime.md#la-salida-literal).
func TestPrimeWithNoIdentityLeavesBoardMeNull(t *testing.T) {
	m, _ := primeBoard(t)
	delete(m.env, "BISO_ME")

	got := m.run(t, "prime", "--json").assertCode(t, 0)

	board := envelopeOf(t, got.stdout)["data"].(map[string]any)["board"].(map[string]any)
	if board["me"] != nil {
		t.Errorf("board.me is %#v and not null", board["me"])
	}
	if !strings.Contains(got.stdout, `"assignedToYou": []`) {
		t.Errorf("assignedToYou is not the empty list:\n%s", got.stdout)
	}
}

// TestPrimeSkipsATaskItCannotReadInsideTheMessage is the exception
// docs/spec/cmd/prime.md declares against
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar:
// the warning is a line of the message on stdout, and stderr stays empty.
func TestPrimeSkipsATaskItCannotReadInsideTheMessage(t *testing.T) {
	m, dir := primeBoard(t)
	m.execOnBoard(t, dir, `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-44'`)

	got := m.run(t, "prime").assertCode(t, 0)

	if !strings.Contains(got.stdout, fixture(t, "prime-unreadable.txt")) {
		t.Errorf("the unreadable line is not in the message:\n%s", got.stdout)
	}
	if got.stderr != "" {
		t.Errorf("biso prime wrote the warning to stderr: %q", got.stderr)
	}
	if strings.Contains(got.stdout, "MYP-44") {
		t.Errorf("the unreadable task was listed anyway:\n%s", got.stdout)
	}
	// The line sits at the end of the BOARD block, under `you are`, and
	// the counts line no longer counts that task.
	if !strings.Contains(got.stdout, "  you are     @claude\n  unreadable  ") {
		t.Errorf("the unreadable line is not the last of the BOARD block:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "  To Do 53 | In Progress 4 | Done 190\n") {
		t.Errorf("the counts line still counts the unreadable task:\n%s", got.stdout)
	}
	assertWithinBudget(t, got.stdout)
}

// TestPrimeNamesTheSkippedTasksInTheEnvelope is the other half: the text
// says how many, `data.skipped` says which ones.
func TestPrimeNamesTheSkippedTasksInTheEnvelope(t *testing.T) {
	m, dir := primeBoard(t)
	m.execOnBoard(t, dir, `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-44'`)

	got := m.run(t, "prime", "--json").assertCode(t, 0)

	skipped := envelopeOf(t, got.stdout)["data"].(map[string]any)["skipped"]
	assertJSONEqual(t, skipped, []any{"MYP-44"}, "data.skipped")
}

func TestPrimeRejectsANegativeLimit(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime", "--limit", "-1").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "prime-limit-negative.txt"),
		"the message of a negative --limit")
	assertEqual(t, got.stdout, "", "the standard output of a rejected biso prime")
}

// TestPrimeFullDoesNotCombineWithJSON is the one restriction --json has
// here, and the rejection is plain text even though --json was written,
// because --json is the invalid part of the call
// (docs/spec/cmd/prime.md#parámetros).
func TestPrimeFullDoesNotCombineWithJSON(t *testing.T) {
	m, _ := primeBoard(t)

	for _, argv := range [][]string{
		{"prime", "--full", "--json"},
		{"prime", "--json", "--full"},
	} {
		got := m.run(t, argv...).assertCode(t, 2)

		assertEqual(t, got.stderr, fixture(t, "prime-full-json.txt"),
			"the message of "+strings.Join(argv, " "))
		if strings.Contains(got.stderr, "\"schemaVersion\"") {
			t.Errorf("the rejection came out as a JSON envelope: %q", got.stderr)
		}
	}
}

func TestPrimeWithoutABoardIsExitCodeTwenty(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "prime").assertCode(t, 20)

	if !strings.Contains(got.stderr, "no board here, and none configured for this project") {
		t.Errorf("the no-board message is %q", got.stderr)
	}
	assertEqual(t, got.stdout, "", "the standard output of a biso prime with no board")
}

// TestPrimeLimitSharesItsRowsBetweenTwoBlocks is the rule of
// docs/spec/cmd/prime.md#la-salida-literal: --limit is rows across
// ASSIGNED TO YOU and NEXT UP together, in that order of preference.
func TestPrimeLimitSharesItsRowsBetweenTwoBlocks(t *testing.T) {
	m, _ := primeBoard(t)

	// Three rows: the two assigned tasks first, one left for NEXT UP.
	got := m.run(t, "prime", "--limit", "3").assertCode(t, 0)

	if n := len(blockOf(got.stdout, "ASSIGNED TO YOU")); n != 2 {
		t.Errorf("ASSIGNED TO YOU printed %d lines and not two:\n%s", n, got.stdout)
	}
	// One row and the count line the two blocks share.
	if n := len(blockOf(got.stdout, "NEXT UP")); n != 2 {
		t.Errorf("NEXT UP printed %d lines and not a row plus its count:\n%s", n, got.stdout)
	}
	if !strings.Contains(got.stdout, "  51 more not shown: `biso ls --not-active --not-waiting`\n") {
		t.Errorf("the shared count line is wrong:\n%s", got.stdout)
	}
}

// TestPrimeLimitZeroLeavesOnlyTheCountLine is the end of that same rule.
func TestPrimeLimitZeroLeavesOnlyTheCountLine(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime", "--limit", "0").assertCode(t, 0)

	for _, heading := range []string{"ASSIGNED TO YOU", "NEXT UP"} {
		if strings.Contains(got.stdout, heading) {
			t.Errorf("--limit 0 printed the heading %q:\n%s", heading, got.stdout)
		}
	}
	if !strings.Contains(got.stdout,
		"\n  54 more not shown: `biso ls --not-active --not-waiting`\n\nPick one,") {
		t.Errorf("the lone count line is not where the two blocks would have gone:\n%s",
			got.stdout)
	}
	// The other two blocks are untouched: --limit never reached them.
	if n := len(blockOf(got.stdout, "IN PROGRESS")); n != 4 {
		t.Errorf("--limit 0 touched IN PROGRESS: %d lines\n%s", n, got.stdout)
	}
}

// TestPrimeLimitDoesNotReachTheEnvelopesOtherTwoBlocks is the same thing
// in JSON: --limit governs two of the four lists and hiddenCount says what
// it left out.
func TestPrimeLimitDoesNotReachTheEnvelopesOtherTwoBlocks(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime", "--limit", "0", "--json").assertCode(t, 0)

	data := envelopeOf(t, got.stdout)["data"].(map[string]any)
	for name, want := range map[string]int{
		"inProgress": 3, "needsAnswer": 1, "assignedToYou": 0, "nextUp": 0,
	} {
		if n := len(data[name].([]any)); n != want {
			t.Errorf("data.%s has %d tasks and not %d", name, n, want)
		}
	}
	if data["hiddenCount"].(float64) != 54 {
		t.Errorf("hiddenCount is %v and not 54", data["hiddenCount"])
	}
}

// The four tests below walk the cascade of
// docs/spec/presupuestos.md#el-presupuesto-de-tamaño one step at a time,
// each over a board of three hundred tasks shaped so that the summary
// overflows by just enough to reach that step and no further. Together
// they check the order, which is the whole content of that list: a
// cascade that trimmed IN PROGRESS first would fit in the cap just as
// well and would throw away the one block that says somebody is working.

// TestTheCascadeTrimsNextUpFirst is step 1.
func TestTheCascadeTrimsNextUpFirst(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 2, waiting: 2, assigned: 4})

	got := m.run(t, "prime", "--limit", "30").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	// --limit 30 gives ASSIGNED TO YOU its four rows and leaves twenty six
	// for NEXT UP, which would be twenty seven lines with the count one.
	if n := len(blockOf(got.stdout, "NEXT UP")); n >= 27 {
		t.Errorf("NEXT UP kept its %d lines although the summary did not fit:\n%s",
			n, got.stdout)
	}
	if !strings.Contains(got.stdout, "more not shown: `biso ls --not-active --not-waiting`") {
		t.Errorf("NEXT UP did not say what the cascade took:\n%s", got.stdout)
	}
	if n := len(blockOf(got.stdout, "ASSIGNED TO YOU")); n != 4 {
		t.Errorf("ASSIGNED TO YOU lost rows before NEXT UP was empty: %d lines\n%s", n, got.stdout)
	}
	if n := len(blockOf(got.stdout, "NEEDS ANSWER")); n != 4 {
		t.Errorf("NEEDS ANSWER was trimmed too early: %d lines\n%s", n, got.stdout)
	}
	if n := len(blockOf(got.stdout, "IN PROGRESS")); n != 2 {
		t.Errorf("IN PROGRESS was trimmed too early: %d lines\n%s", n, got.stdout)
	}
}

// TestTheCascadeTrimsAssignedToYouSecond is step 2.
func TestTheCascadeTrimsAssignedToYouSecond(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 2, waiting: 2, assigned: 20})

	got := m.run(t, "prime", "--limit", "30").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	assigned := len(blockOf(got.stdout, "ASSIGNED TO YOU"))
	if assigned == 0 || assigned >= 20 {
		t.Errorf("ASSIGNED TO YOU kept %d of its 20 rows, so this is not step 2:\n%s",
			assigned, got.stdout)
	}
	if n := len(blockOf(got.stdout, "NEEDS ANSWER")); n != 4 {
		t.Errorf("NEEDS ANSWER was trimmed before ASSIGNED TO YOU ran out: %d lines\n%s",
			n, got.stdout)
	}
	if n := len(blockOf(got.stdout, "IN PROGRESS")); n != 2 {
		t.Errorf("IN PROGRESS was trimmed too early: %d lines\n%s", n, got.stdout)
	}
}

// TestTheCascadeTrimsNeedsAnswerThird is step 3.
func TestTheCascadeTrimsNeedsAnswerThird(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 2, waiting: 40, assigned: 0})

	got := m.run(t, "prime", "--limit", "0").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	// Each parked task takes two lines, its row and its question.
	answers := len(blockOf(got.stdout, "NEEDS ANSWER"))
	if answers == 0 || answers >= 80 {
		t.Errorf("NEEDS ANSWER kept %d lines of its 80, so this is not step 3:\n%s",
			answers, got.stdout)
	}
	if n := len(blockOf(got.stdout, "IN PROGRESS")); n != 2 {
		t.Errorf("IN PROGRESS was trimmed before NEEDS ANSWER ran out: %d lines\n%s",
			n, got.stdout)
	}
	if !strings.Contains(got.stdout, "more not shown: `biso ls --waiting`") {
		t.Errorf("NEEDS ANSWER did not say what it left out:\n%s", got.stdout)
	}
}

// TestTheCascadeTrimsInProgressLastAndLeavesTheCountLines is step 4 and
// the fifth and last step: every other block is down to its single count
// line, and IN PROGRESS keeps whatever still fits.
func TestTheCascadeTrimsInProgressLastAndLeavesTheCountLines(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 60, waiting: 40, assigned: 60})

	got := m.run(t, "prime", "--limit", "30").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	for _, heading := range []string{"NEEDS ANSWER", "ASSIGNED TO YOU", "NEXT UP"} {
		if strings.Contains(got.stdout, heading) {
			t.Errorf("%s kept a row while IN PROGRESS was being trimmed:\n%s",
				heading, got.stdout)
		}
	}
	inProgress := len(blockOf(got.stdout, "IN PROGRESS"))
	if inProgress == 0 || inProgress >= 60 {
		t.Errorf("IN PROGRESS kept %d of its 60 rows:\n%s", inProgress, got.stdout)
	}
	// Every block that lost rows says so, with the filter of `biso ls`
	// that answers it whole (docs/spec/cmd/prime.md#el-recorte-en-cascada).
	for _, line := range []string{
		"more not shown: `biso ls --active`",
		"more not shown: `biso ls --waiting`",
		"more not shown: `biso ls --not-active --not-waiting`",
	} {
		if !strings.Contains(got.stdout, line) {
			t.Errorf("the message does not carry %q:\n%s", line, got.stdout)
		}
	}
}

// TestTheCascadeNeverTouchesTheFixedPartOrTheEnvelope is the other half of
// that budget: the part that teaches how to use the board is never given
// up, and `--json` is not trimmed at all.
func TestTheCascadeNeverTouchesTheFixedPartOrTheEnvelope(t *testing.T) {
	m := blockBoard(t, blockSizes{inProgress: 60, waiting: 40, assigned: 60})

	got := m.run(t, "prime", "--limit", "30").assertCode(t, 0)

	for _, block := range []string{"COMMANDS  (", "FIELD FLAGS  (", "RULES  (", "Pick one,"} {
		if !strings.Contains(got.stdout, block) {
			t.Errorf("the cascade ate the fixed part: %q is missing:\n%s", block, got.stdout)
		}
	}
	envelope := m.run(t, "prime", "--limit", "30", "--json").assertCode(t, 0)
	data := envelopeOf(t, envelope.stdout)["data"].(map[string]any)
	if n := len(data["inProgress"].([]any)); n != 60 {
		t.Errorf("data.inProgress was trimmed: %d tasks", n)
	}
	if n := len(data["needsAnswer"].([]any)); n != 40 {
		t.Errorf("data.needsAnswer was trimmed: %d tasks", n)
	}
}

// TestTheFixedPartOfTheMessageIsAlwaysTheSame is what makes the arithmetic
// of docs/spec/presupuestos.md#el-presupuesto-de-tamaño exact: two boards
// with nothing in common print the same fixed part, byte for byte, and it
// fits in its half of the budget on its own.
func TestTheFixedPartOfTheMessageIsAlwaysTheSame(t *testing.T) {
	one, _ := primeBoard(t)
	other := blockBoard(t, blockSizes{inProgress: 2, waiting: 2, assigned: 4})

	first := fixedPartOf(one.run(t, "prime").assertCode(t, 0).stdout)
	second := fixedPartOf(other.run(t, "prime").assertCode(t, 0).stdout)

	assertEqual(t, second, first, "the fixed part of the message")
	if len(first) > cli.PrimeFixedBudget {
		t.Errorf("the fixed part measures %d bytes and its half of the budget is %d",
			len(first), cli.PrimeFixedBudget)
	}
}

// TestTheBudgetConstantsAreTheNumbersOfTheSpecification keeps the two caps
// of the message from drifting: the hard cap of 5,504 bytes is the only
// number the stability contract freezes, and the fixed part is 3,840 of it
// (docs/spec/presupuestos.md#el-presupuesto-de-tamaño). Raising either to
// make room for a gloss would pass every other test in this file.
func TestTheBudgetConstantsAreTheNumbersOfTheSpecification(t *testing.T) {
	assertEqual(t, fmt.Sprint(cli.PrimeBudget), "5504", "the hard cap of the startup message")
	assertEqual(t, fmt.Sprint(cli.PrimeFixedBudget), "3840", "the cap of its fixed part")
}

// blockSizes is how many tasks each of the first three blocks gets on a
// board of three hundred; the rest fall into NEXT UP.
type blockSizes struct{ inProgress, waiting, assigned int }

const blockBoardTasks = 300

// blockBoard seeds a board of three hundred tasks split into the four
// blocks, with titles about as long as a real one, which is what makes the
// summary overflow.
//
// It writes them with one statement per group and not with three hundred
// calls to the program: they are a fixture, and what is being checked is
// the shape of one message.
func blockBoard(t *testing.T, sizes blockSizes) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")

	now := time.Now().UTC().Format(model.InstantLayout)
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
		SELECT 'MYP-' || i, i,
			'A task with a title about as long as a real one gets, the one numbered ' || i,
			'To Do', 'task', 'medium', '', '', '',
			NULL, '', '', '', '', ?, ?, 0, '', '', '', '', '', 1, 1
		FROM n`, blockBoardTasks, now, now)
	m.execOnBoard(t, dir, `UPDATE board_counter SET last_task_num = ?`, blockBoardTasks)

	ids := func(from, count int) string {
		var out []string
		for i := from; i < from+count; i++ {
			out = append(out, "MYP-"+strconv.Itoa(i))
		}
		return "'" + strings.Join(out, "','") + "'"
	}
	assign := func(from, count int) {
		if count == 0 {
			return
		}
		var values []string
		for i := from; i < from+count; i++ {
			values = append(values,
				fmt.Sprintf("('MYP-%d', 'assignees', 0, '@claude')", i))
		}
		m.execOnBoard(t, dir,
			"INSERT INTO task_list_item (task_id, field, position, value) VALUES "+
				strings.Join(values, ", "))
	}

	if sizes.inProgress > 0 {
		m.execOnBoard(t, dir, fmt.Sprintf(
			`UPDATE task SET status = 'In Progress' WHERE id IN (%s)`,
			ids(1, sizes.inProgress)))
		assign(1, sizes.inProgress)
	}
	if sizes.waiting > 0 {
		m.execOnBoard(t, dir, fmt.Sprintf(`UPDATE task SET status = 'In Progress',
				question_author = '@claude', question_asked_at = ?, question_body = ?
			WHERE id IN (%s)`, ids(sizes.inProgress+1, sizes.waiting)),
			now, "Is this the question that parks the task, or is there another one?")
		assign(sizes.inProgress+1, sizes.waiting)
	}
	assign(sizes.inProgress+sizes.waiting+1, sizes.assigned)
	return m
}

// primeClosingParagraph is the end of every message, whatever the board
// (docs/spec/cmd/prime.md#lo-que-no-depende-del-tablero).
const primeClosingParagraph = "Pick one, `biso start <ref> --append-plan \"...\"`, work, " +
	"`biso note <ref> \"...\"` as you go,\n" +
	"and close with `biso finish <ref> --check-ac all --append-summary \"...\"`.\n" +
	"That is the loop. Create a task when the work needs planning or review; do small\n" +
	"edits directly.\n"

// lastRuleOfTheMessage is where the fixed part ends and the blocks of
// tasks begin.
const lastRuleOfTheMessage = "     assigned to you is one a person decided you should do.\n\n"

// blockOf answers the lines of one block of the message, its heading
// excluded and its count line included, and nothing when that block did
// not print a heading at all.
func blockOf(message, heading string) []string {
	body := blockText(message, heading)
	if body == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(body, "\n"), "\n")
}

// blockText is the same thing unsplit.
func blockText(message, heading string) string {
	at := strings.Index(message, "\n"+heading)
	if at < 0 {
		return ""
	}
	rest := message[at+1:]
	rest = rest[strings.Index(rest, "\n")+1:]
	return rest[:strings.Index(rest, "\n\n")+1]
}

// fixedPartOf is the title line, COMMANDS, FIELD FLAGS, RULES and the
// closing paragraph, each with the blank line that belongs to it: the
// message minus the BOARD block and the four blocks of tasks
// (docs/spec/presupuestos.md#el-presupuesto-de-tamaño).
func fixedPartOf(message string) string {
	title := message[:strings.Index(message, "\n")+2]
	commands := strings.Index(message, "COMMANDS  (")
	rules := strings.Index(message, lastRuleOfTheMessage) + len(lastRuleOfTheMessage)
	return title + message[commands:rules] + message[strings.Index(message, "Pick one,"):]
}

// TestPrimeCutsTheQuestionAtAHundredCells is the rule NEEDS ANSWER
// borrows from the column algorithm: the body of the question is one line,
// its real newlines turned into spaces, cut to a hundred cells with the
// same three dots a title gets (docs/spec/cmd/prime.md#la-salida-literal).
func TestPrimeCutsTheQuestionAtAHundredCells(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "A parked task", "--start").assertCode(t, 0)
	m.run(t, "ask", "MYP-1",
		strings.Repeat("a", 60)+"\n"+strings.Repeat("b", 60)).assertCode(t, 0)

	got := m.run(t, "prime").assertCode(t, 0)

	lines := blockOf(got.stdout, "NEEDS ANSWER")
	if len(lines) != 2 {
		t.Fatalf("NEEDS ANSWER printed %d lines and not a row plus its question:\n%s",
			len(lines), got.stdout)
	}
	want := "    " + strings.Repeat("a", 60) + " " + strings.Repeat("b", 36) + "..."
	assertEqual(t, lines[1], want, "the question of a parked task")
}

// TestPrimeKeepsTheSpacesOfAQuestion is the other half of that rule: the
// line breaks become a space and nothing else is touched, so a question
// written with two spaces after a full stop, or with an indented second
// line, reaches the message as its author typed it
// (docs/spec/cmd/prime.md#la-salida-literal).
func TestPrimeKeepsTheSpacesOfAQuestion(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "A parked task", "--start").assertCode(t, 0)
	m.run(t, "ask", "MYP-1", "Two  spaces here.  And a tab\there.\n    An indented line.").
		assertCode(t, 0)

	got := m.run(t, "prime").assertCode(t, 0)

	lines := blockOf(got.stdout, "NEEDS ANSWER")
	if len(lines) != 2 {
		t.Fatalf("NEEDS ANSWER printed %d lines and not a row plus its question:\n%s",
			len(lines), got.stdout)
	}
	assertEqual(t, lines[1],
		"    Two  spaces here.  And a tab\there.     An indented line.",
		"the question of a parked task")
}
