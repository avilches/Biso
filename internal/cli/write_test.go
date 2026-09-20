package cli

import (
	"strings"
	"testing"

	"biso/internal/ops"
)

// These tests drive Run in this process, over a board this same process
// creates, which is what lets them fix the clock. What the compiled program
// prints, character for character, is checked in cmd/biso instead.

// board creates a board in the machine's project and answers the machine,
// so that a test of a writing command starts from something to write on.
func (m *machine) board() *machine {
	m.t.Helper()
	m.env["BISO_ME"] = "@claude"
	if out := m.withIDs("3f9a2b1c").run("init", "My project", "--prefix", "MYP"); out.code != 0 {
		m.t.Fatalf("biso init: %d\n%s", out.code, out.stderr)
	}
	return m
}

func TestTheStatusLineHasItsFourPieces(t *testing.T) {
	for _, c := range []struct {
		name string
		task ops.TaskWrite
		want string
	}{
		{
			name: "a task with no criteria prints only the id, the status and the urgency",
			task: ops.TaskWrite{ID: "MYP-11", Status: "In Progress", Urgency: 19.0},
			want: "MYP-11  In Progress  urgency 19.0",
		},
		{
			name: "a task with criteria adds their progress",
			task: ops.TaskWrite{ID: "MYP-11", Status: "In Progress", AcDone: 1, AcTotal: 2, Urgency: 19.0},
			want: "MYP-11  In Progress  ac 1/2  urgency 19.0",
		},
		{
			name: "the keys created by the call close the line",
			task: ops.TaskWrite{
				ID: "MYP-11", Status: "In Progress", AcDone: 1, AcTotal: 4,
				Urgency: 19.0, AcAdded: []int{4, 5},
			},
			want: "MYP-11  In Progress  ac 1/4  urgency 19.0  added ac #4, #5",
		},
		{
			name: "and the word archived closes it when the task ends up archived",
			task: ops.TaskWrite{ID: "MYP-11", Status: "Done", Urgency: 0.0, Archived: true},
			want: "MYP-11  Done  urgency 0.0  archived",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			assertEqual(t, statusLine(c.task), c.want, "the status line")
		})
	}
}

// docs/spec/contrato-json.md#números-fechas-y-ausencias: the urgency is a
// decimal with exactly one digit after the point, which the default
// marshalling of a whole number would turn into an integer.
func TestTheUrgencyKeepsItsSingleDecimal(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)

	out := m.run("set", "MYP-1", "--add-labels", "parser", "--json").assertCode(t, 0)

	if !strings.Contains(out.stdout, `"urgency": 1.8`) {
		t.Errorf("the envelope does not carry the urgency with one decimal:\n%s", out.stdout)
	}
}

func TestTheWriteEnvelopeIsTheOneOfTheSpecification(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task", "--add-ac", "First", "--add-ac", "Second").assertCode(t, 0)

	out := m.run("set", "MYP-1", "--add-labels", "parser", "--check-ac", "1", "--json").
		assertCode(t, 0)

	for _, fragment := range []string{
		`"kind": "task.write"`,
		`"id": "MYP-1"`,
		`"status": "To Do"`,
		`"acDone": 1`,
		`"acTotal": 2`,
		`"changed"`,
		`"acAdded": []`,
		`"warnings": []`,
	} {
		if !strings.Contains(out.stdout, fragment) {
			t.Errorf("the envelope has no %s:\n%s", fragment, out.stdout)
		}
	}
}

// docs/spec/cmd/set.md#el-esquema-json: acAdded carries the keys the call
// created, in the order it created them, and it is there in `biso new` too
// even though the text output of that command never prints them.
func TestNewCarriesTheKeysItCreatedInTheEnvelope(t *testing.T) {
	m := newMachine(t).board()

	out := m.run("new", "A task", "--add-ac", "First", "--add-ac", "Second", "--json").
		assertCode(t, 0)

	if !strings.Contains(out.stdout, "\"acAdded\": [\n          1,\n          2\n        ]") {
		t.Errorf("acAdded is not the two keys it created:\n%s", out.stdout)
	}
}

func TestQuietReducesAWriteToTheIdentifiers(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "First").assertCode(t, 0)
	m.run("new", "Second").assertCode(t, 0)

	out := m.run("set", "MYP-1", "MYP-2", "--add-labels", "parser", "--quiet").assertCode(t, 0)

	assertEqual(t, out.stdout, "MYP-1\nMYP-2\n", "the output of --quiet")
}

// docs/spec/salida-y-terminal.md#notas-y-avisos: a warning is never
// suppressed, and one the call had already earned before failing travels
// with the error instead of disappearing with it.
func TestAFailedWriteStillPrintsTheWarningsItHadEarned(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)

	out := m.run("set", "MYP-1", "--due", "2020-01-01", "--status", "Pending").assertCode(t, 3)

	if !strings.Contains(out.stderr, "warning: --due 2020-01-01 is in the past") {
		t.Errorf("the warning was lost with the error:\n%s", out.stderr)
	}
	if !strings.Contains(out.stderr, `error: unknown status: "Pending"`) {
		t.Errorf("the error is not the one of the specification:\n%s", out.stderr)
	}
}

// The same rule with --json: stderr never mixes the text of a warning with
// the JSON object of an error, so the two travel in one envelope
// (docs/spec/contrato-json.md#los-errores-en-json).
func TestAFailedWriteFoldsItsWarningsIntoTheErrorEnvelope(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)

	out := m.run("set", "MYP-1", "--due", "2020-01-01", "--status", "Pending", "--json").
		assertCode(t, 3)

	if strings.Contains(out.stderr, "warning: ") {
		t.Errorf("stderr mixes the text of a warning with a JSON object:\n%s", out.stderr)
	}
	if !strings.Contains(out.stderr, `"code": "due_in_past"`) {
		t.Errorf("the warning is not inside the envelope:\n%s", out.stderr)
	}
}

func TestNewRefusesStartTogetherWithAStatus(t *testing.T) {
	m := newMachine(t).board()

	out := m.run("new", "A task", "--start", "--status", "Done").assertCode(t, 2)

	if !strings.Contains(out.stderr, "cannot be used together") {
		t.Errorf("the message is not the one of a pair of incompatible flags:\n%s", out.stderr)
	}
	// And the same call written the other way round fails with the same
	// text, because a message never depends on how it was typed.
	other := m.run("new", "A task", "--status", "Done", "--start").assertCode(t, 2)
	assertEqual(t, other.stderr, out.stderr, "the message of the pair, written the other way round")
}

func TestSetRefusesIdAndMatchTogether(t *testing.T) {
	m := newMachine(t).board()

	m.run("set", "MYP-1", "--id", "--match", "--add-labels", "x").assertCode(t, 2)
}

// changesOf is the one place a flag of the parser becomes a change of the
// engine, so these are the three things it has to carry across and that no
// other test would notice.
func TestChangesCarryTheEmptyReplaceAndTheCommentAuthor(t *testing.T) {
	p, err := Parse(
		[]string{"set", "MYP-1", "--replace-labels=", "--comment", "hi", "--comment-author", "@sara"},
		Commands(), Env{})
	if err != nil {
		t.Fatal(err)
	}

	var sawReplace, sawAuthor bool
	for _, c := range changesOf(p) {
		if c.Flag == "replace-labels" && c.Step == ops.StepReplace && c.Value == "" {
			sawReplace = true
		}
		if c.Flag == "comment-author" && c.Value == "@sara" {
			sawAuthor = true
		}
	}
	if !sawReplace {
		t.Errorf("--replace-labels with an empty value did not reach the engine")
	}
	if !sawAuthor {
		t.Errorf("--comment-author did not reach the engine")
	}
}

// The steps of the parser and the steps of the engine are the same section
// of the specification, and changesOf converts one into the other by casting
// the number. If either table were reordered, this is where it would show.
func TestTheStepsOfBothLayersLineUp(t *testing.T) {
	for _, c := range []struct {
		parser Category
		engine ops.Step
	}{
		{Clear, ops.StepClear},
		{Replace, ops.StepReplace},
		{Remove, ops.StepRemove},
		{Add, ops.StepAdd},
		{ExtKey, ops.StepExt},
		{Scalar, ops.StepScalar},
		{CheckAC, ops.StepCheckAC},
		{CommentDate, ops.StepCommentDate},
		{AddComment, ops.StepComment},
	} {
		if ops.Step(c.parser) != c.engine {
			t.Errorf("the parser's step %d is the engine's %d, and they have to be the same",
				c.parser, c.engine)
		}
	}
}

// TestEveryFieldFlagOfTheSpecificationIsInTheTable walks the families of
// docs/spec/familias-de-flags.md and asks the table for each one by name. A
// flag the specification lists and the table forgot is a flag no command
// accepts, which is the failure this catches.
func TestEveryFieldFlagOfTheSpecificationIsInTheTable(t *testing.T) {
	wanted := []string{
		"add-labels", "rm-labels", "clear-labels", "replace-labels",
		"add-assignees", "rm-assignees", "clear-assignees", "replace-assignees",
		"add-refs", "rm-refs", "clear-refs", "replace-refs",
		"add-docs", "rm-docs", "clear-docs", "replace-docs",
		"add-deps", "rm-deps", "clear-deps", "replace-deps",
		"add-files", "rm-files", "clear-files", "replace-files",
		"add-ac", "rm-ac", "clear-acs", "check-ac", "uncheck-ac",
		"append-desc", "append-plan", "append-note", "append-summary",
		"clear-desc", "clear-plan", "clear-notes", "clear-summary",
		"ext", "rm-ext", "clear-ext",
		"title", "status", "type", "priority", "parent", "due", "ordinal", "author",
		"clear-type", "clear-priority", "clear-parent", "clear-due",
		"clear-ordinal", "clear-author",
		"comment", "comment-author", "rm-comment", "set-comment-date",
	}
	for _, name := range []string{"new", "set"} {
		command := commandNamed(t, name)
		for _, flag := range wanted {
			if lookupLong(command, flag) == nil {
				t.Errorf("biso %s does not accept --%s", name, flag)
			}
		}
	}
}

func commandNamed(t *testing.T, name string) *CommandSpec {
	t.Helper()
	commands := Commands()
	for i := range commands {
		if commands[i].Name == name {
			return &commands[i]
		}
	}
	t.Fatalf("there is no command called %q", name)
	return nil
}
