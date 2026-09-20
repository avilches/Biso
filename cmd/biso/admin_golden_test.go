package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// These are the golden tests of the four commands of the last step:
// `biso archive`, `biso config`, `biso doctor` and `biso help`. Like every
// other golden test of this package they run the compiled program and
// compare what the process wrote with the blocks transcribed from the
// specification, character for character.
//
// `biso board` is not here and never will be under this name: it is out of
// the scope of version 1.0 by an explicit decision, and the only thing of
// it that exists is its help text, which `biso help board` answers.

// ---------------------------------------------------------------- help

// TestTheHelpOfTheAdministrativeCommands is the other half of
// TestHelpTextsAreTheOnesOfTheSpecification, for the commands whose help
// arrived with this step. The two doors are checked together on purpose:
// `biso <cmd> --help` and `biso help <cmd>` promise the same text, and a
// help that drifted between them would be invisible from either one alone.
func TestTheHelpOfTheAdministrativeCommands(t *testing.T) {
	m := newMachine(t)

	for _, c := range []struct{ command, want string }{
		{"archive", "archive-help.txt"},
		{"config", "config-help.txt"},
		{"doctor", "doctor-help.txt"},
		{"help", "help-help.txt"},
	} {
		want := fixture(t, c.want)
		got := m.run(t, c.command, "--help").assertCode(t, 0)
		assertEqual(t, got.stdout, want, "biso "+c.command+" --help")
		if got.stderr != "" {
			t.Errorf("biso %s --help wrote to stderr: %s", c.command, got.stderr)
		}
		through := m.run(t, "help", c.command).assertCode(t, 0)
		assertEqual(t, through.stdout, want, "biso help "+c.command)
	}
}

// TestHelpAnswersForCommandsThatDoNotExistYet is the promise of the catalog:
// every name `biso help all` prints answers `biso help <name>`, whether or
// not this build carries the logic behind it. `board` is out of scope for
// 1.0 and `export` and `snapshot` belong to another step, and the help of a
// command is part of the interface all the same.
func TestHelpAnswersForCommandsThatDoNotExistYet(t *testing.T) {
	m := newMachine(t)

	for _, name := range []string{"board", "export", "snapshot"} {
		got := m.run(t, "help", name).assertCode(t, 0)
		if !strings.HasPrefix(got.stdout, "Usage: biso "+name) {
			t.Errorf("biso help %s = %q", name, got.stdout)
		}
	}
}

func TestHelpWithNoArgumentIsTheTopLevelHelp(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "top-help.txt"), "biso help")
}

// TestHelpAllPrintsTheTopLevelHelpAndTheAdministrationBlock is the output of
// docs/spec/cmd/help.md#salida-de-biso-help-all: the top-level help, and
// then the nine administrative commands, with the blank line that separates
// every block of that text from the next.
func TestHelpAllPrintsTheTopLevelHelpAndTheAdministrationBlock(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help", "all").assertCode(t, 0)

	assertEqual(t, got.stdout,
		fixture(t, "top-help.txt")+"\n"+fixture(t, "help-all.txt"), "biso help all")
}

// TestHelpOfSeveralCommandsPrintsThemInOrder is the reason this command
// takes several names at once: a caller that wants the detail of three
// commands pays one invocation and not three.
func TestHelpOfSeveralCommandsPrintsThemInOrder(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help", "finish", "note").assertCode(t, 0)

	assertEqual(t, got.stdout,
		fixture(t, "finish-help.txt")+"\n"+fixture(t, "note-help.txt"),
		"biso help finish note")
}

// TestHelpOfANameThatDoesNotExistPrintsNothingAtAll is the all or nothing of
// docs/spec/cmd/help.md#varios-comandos-en-una-sola-llamada: the help of the
// names that do exist earlier in the list is not printed either.
func TestHelpOfANameThatDoesNotExistPrintsNothingAtAll(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help", "finish", "fnish", "note").assertCode(t, 4)

	assertEqual(t, "$ biso help fnish\n"+got.stderr, fixture(t, "help-unknown.txt"),
		"the refusal of a command name that does not exist")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso help")
}

func TestHelpAllCannotBeCombinedWithACommandName(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help", "all", "finish").assertCode(t, 2)

	assertEqual(t, got.stderr, fixture(t, "help-all-combined.txt"),
		"the refusal of all combined with a command name")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso help")
}

// TestHelpEnvelopeMatchesTheSchemaOfTheSpecification is the one key of that
// envelope: the command list, never the prose help.
func TestHelpEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := newMachine(t)

	got := m.run(t, "help", "prime", "new", "--json").assertCode(t, 0)

	assertSameJSON(t, got.stdout, fixture(t, "help-json.txt"))
}

// TestHelpWorksWithoutABoard is the line of the page that says so, and it
// is not free: every other command of this step answers 20 here.
func TestHelpWorksWithoutABoard(t *testing.T) {
	m := newMachine(t)

	m.run(t, "help").assertCode(t, 0)
	m.run(t, "help", "all").assertCode(t, 0)
	m.run(t, "help", "doctor").assertCode(t, 0)
	m.run(t, "help", "--json").assertCode(t, 0)
	m.run(t, "doctor").assertCode(t, 20)
}

// TestThereIsNoDeleteCommand is the refusal of
// docs/spec/cmd/archive.md#biso-delete-no-existe-y-su-ausencia-está-especificada,
// which answers what to do instead of dumping the list of commands.
func TestThereIsNoDeleteCommand(t *testing.T) {
	m := newMachine(t)

	for _, name := range []string{"delete", "rm", "remove"} {
		got := m.run(t, name, "MYP-11").assertCode(t, 2)
		assertEqual(t, got.stderr, fixture(t, "archive-no-delete.txt"), "biso "+name)
		assertEqual(t, got.stdout, "", "the standard output of biso "+name)
	}
}

// ------------------------------------------------------------- archive

// archivedExampleBoard is the board of the examples with MYP-11 closed and
// its two criteria checked, which is the `Done  ac 2/2  urgency 0.0` of the
// status line of docs/spec/cmd/archive.md#salida.
func archivedExampleBoard(t *testing.T) *machine {
	t.Helper()
	m := exampleBoard(t)
	m.run(t, "set", "MYP-11",
		"--add-ac", "The parser accepts CRLF",
		"--add-ac", "Dates keep their time zone").assertCode(t, 0)
	m.run(t, "set", "MYP-11", "--check-ac", "1", "--check-ac", "2").assertCode(t, 0)
	m.run(t, "set", "MYP-12", "--status", "Done").assertCode(t, 0)
	m.run(t, "finish", "MYP-11").assertCode(t, 0)
	return m
}

func TestArchivePrintsTheStatusLineOfTheSpecification(t *testing.T) {
	m := archivedExampleBoard(t)

	got := m.run(t, "archive", "MYP-11").assertCode(t, 0)

	assertEqual(t, got.stdout, fixture(t, "archive-status-line.txt"),
		"the status line of biso archive")
}

func TestArchiveDryRunPrintsTheSameLineUnderItsHeader(t *testing.T) {
	m := archivedExampleBoard(t)

	got := m.run(t, "archive", "MYP-11", "--dry-run").assertCode(t, 0)

	assertEqual(t, "$ biso archive MYP-11 --dry-run\n"+got.stdout,
		fixture(t, "archive-dry-run.txt"), "the preview of biso archive")
}

// -------------------------------------------------------------- config

// configuredBoard is the board of the example of
// docs/spec/cmd/config.md#salida: the vocabularies and the prefix that
// listing prints.
func configuredBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "TASK",
		"--types", "idea,memory,task,bug,docs",
		"--extensions", "trello.card").assertCode(t, 0)
	return m
}

// TestConfigPrintsTheOutputOfTheSpecification is the twenty keys, always,
// in the order of the table of keys, and the one value a `get` answers.
func TestConfigPrintsTheOutputOfTheSpecification(t *testing.T) {
	m := configuredBoard(t)

	one := m.run(t, "config", "get", "statuses").assertCode(t, 0)
	all := m.run(t, "config", "list").assertCode(t, 0)

	assertEqual(t,
		"$ biso config get statuses\n"+one.stdout+"\n$ biso config list\n"+all.stdout,
		fixture(t, "config-output.txt"), "the output of biso config")
}

// TestConfigNotes is the note of a correct `set`, in its three shapes: a
// scalar, a list, and a list emptied.
func TestConfigNotes(t *testing.T) {
	for _, c := range []struct {
		argv    []string
		fixture string
	}{
		{[]string{"config", "set", "lease_minutes", "45"}, "config-note-scalar.txt"},
		{[]string{"config", "set", "statuses", "To Do,In Progress,Done,Blocked"},
			"config-note-list.txt"},
		{[]string{"config", "set", "types", ""}, "config-note-empty.txt"},
		{[]string{"config", "set", "lease_minutes", "45", "--dry-run"},
			"config-dry-run-scalar.txt"},
		{[]string{"config", "set", "statuses", "To Do,In Progress,Done,Blocked", "--dry-run"},
			"config-dry-run-list.txt"},
		{[]string{"config", "set", "types", "", "--dry-run"}, "config-dry-run-empty.txt"},
	} {
		m := configuredBoard(t)
		got := m.run(t, c.argv...).assertCode(t, 0)

		assertEqual(t, "$ biso "+shellArgs(c.argv)+"\n"+got.stderr, fixture(t, c.fixture),
			"the note of biso "+strings.Join(c.argv, " "))
		// A correct `set` writes nothing on stdout: its whole answer is
		// the note (docs/spec/cmd/config.md#comportamiento-caso-a-caso).
		assertEqual(t, got.stdout, "", "the standard output of a biso config set")
	}
}

// shellArgs writes a call the way the examples of the specification write
// it, with the one value that carries spaces in quotes.
func shellArgs(argv []string) string {
	out := make([]string, 0, len(argv))
	for _, arg := range argv {
		if arg == "" || strings.Contains(arg, " ") {
			arg = `"` + arg + `"`
		}
		out = append(out, arg)
	}
	return strings.Join(out, " ")
}

func TestConfigOfAKeyThatDoesNotExistSuggestsTheClosest(t *testing.T) {
	m := configuredBoard(t)

	got := m.run(t, "config", "get", "urgency.pryority").assertCode(t, 4)

	assertEqual(t, "$ biso config get urgency.pryority\n"+got.stderr,
		fixture(t, "config-unknown-key.txt"), "the refusal of a key that does not exist")
	assertEqual(t, got.stdout, "", "the standard output of a failed biso config get")
}

// TestConfigJSONOnlyAppliesToList is plain text on stderr and not the error
// envelope, because it is --json itself that is the invalid part of the
// call (docs/spec/cmd/config.md#parámetros).
func TestConfigJSONOnlyAppliesToList(t *testing.T) {
	m := configuredBoard(t)

	for _, argv := range [][]string{
		{"config", "get", "statuses", "--json"},
		{"config", "set", "lease_minutes", "45", "--json"},
	} {
		got := m.run(t, argv...).assertCode(t, 2)
		assertEqual(t, got.stderr, fixture(t, "config-json-flag.txt"),
			"the refusal of --json in biso "+strings.Join(argv, " "))
	}
}

func TestConfigEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := configuredBoard(t)

	got := m.run(t, "config", "list", "--json").assertCode(t, 0)

	assertSameJSON(t, got.stdout, specBlock(t, "cmd/config.md", "El esquema JSON", 0))
}

// -------------------------------------------------------------- doctor

// TestDoctorPrintsTheReportOfTheSpecification builds the board the example
// of docs/spec/cmd/doctor.md#salida describes: MYP-40 depends on a task
// that does not exist, the counter says MYP-40 while the tasks go up to
// MYP-52, and the machine declares a further boards root that cannot be
// read.
//
// The three findings are one of each kind on purpose: an error that --fix
// cannot repair, an error it can, and a warning.
func TestDoctorPrintsTheReportOfTheSpecification(t *testing.T) {
	m := brokenBoard(t)

	got := m.run(t, "doctor", "--fix").assertCode(t, 6)

	assertEqual(t, got.stdout, m.substituted(fixture(t, "doctor-output.txt")),
		"the report of biso doctor --fix")
	assertEqual(t, got.stderr, "", "the standard error of biso doctor")

	// The repair really landed: the next identifier the board hands out is
	// the one after the highest task, and not the one after the counter it
	// had before.
	assertEqual(t, m.run(t, "new", "A task after the repair").assertCode(t, 0).stdout,
		"MYP-53\n", "the identifier handed out after the counter was repaired")
}

// TestDoctorDryRunReportsWhatItWouldFixAndWritesNothing is the pure preview:
// the same report with its last block in the conditional, the same exit
// code, and a second run that finds the very same thing because nothing was
// written.
func TestDoctorDryRunReportsWhatItWouldFixAndWritesNothing(t *testing.T) {
	m := brokenBoard(t)

	got := m.run(t, "doctor", "--fix", "--dry-run").assertCode(t, 6)

	assertEqual(t, "$ biso doctor --fix --dry-run\n"+got.stdout,
		m.substituted(fixture(t, "doctor-dry-run.txt")),
		"the dry report of biso doctor --fix --dry-run")

	again := m.run(t, "doctor", "--fix", "--dry-run").assertCode(t, 6)
	assertEqual(t, again.stdout, got.stdout, "the report of a second preview")
}

func TestDoctorEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m := brokenBoard(t)

	got := m.run(t, "doctor", "--fix", "--json").assertCode(t, 6)

	assertSameJSON(t, got.stdout, m.substituted(fixture(t, "doctor-json.txt")))
}

// brokenBoard is the board of those three examples. Its tasks are written
// straight into the database because no call of the program can produce
// them: `biso set` refuses a dependency on a task that does not exist, and
// nothing walks a counter backwards.
func brokenBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	// The machine declares a boards root that is not there, which is the
	// warning of the report. The board itself lives inside the project, so
	// that root is never walked to find it.
	root := filepath.Join(m.home, "Volumes", "disco", "boards")
	m.write(t, filepath.Join(m.home, ".biso", "config.json"),
		`{"boards_extra_roots": ["`+root+`"], "vcs": "none"}`+"\n")
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)

	dir := filepath.Join(m.dir, "board")
	m.execOnBoard(t, dir, `
		WITH RECURSIVE n(i) AS (
			SELECT 40 UNION ALL SELECT i + 1 FROM n WHERE i < 52
		)
		INSERT INTO task (
			id, num, title, status, type, priority, parent, author, due,
			ordinal, description, plan, notes, summary, created_at, updated_at,
			archived, lease_expires_at, lease_holder, question_author,
			question_asked_at, question_body, next_criterion_key, next_comment_key
		)
		SELECT 'MYP-' || i, i, 'Task ' || i, 'To Do', 'task', 'medium', '', '', '',
			i, '', '', '', '', '2026-09-06T09:12:04Z', '2026-09-06T09:12:04Z',
			0, '', '', '', '', '', 1, 1
		FROM n`)
	// MYP-40 depends on a task that does not exist, which no call of the
	// program can leave behind: `biso set` refuses the dependency in the
	// first place.
	m.execOnBoard(t, dir,
		`INSERT INTO task_list_item (task_id, field, position, value)
		 VALUES ('MYP-40', 'dependencies', 0, 'MYP-99')`)
	// And the counter stayed at MYP-40 while the tasks went up to MYP-52,
	// which nothing walks backwards either.
	m.execOnBoard(t, dir, `UPDATE board_counter SET last_task_num = 40`)
	return m
}

// TestDoctorOnACleanBoardFindsNothing is the first row of the table of
// cases, and the one this suite would miss the most: a report that named
// something on a board with nothing wrong would be noise on every call.
func TestDoctorOnACleanBoardFindsNothing(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "A task").assertCode(t, 0)

	got := m.run(t, "doctor").assertCode(t, 0)

	assertEqual(t, got.stdout, "no problems found\n", "the report of a clean board")
	assertEqual(t, got.stderr, "", "the standard error of biso doctor")

	withJSON := m.run(t, "doctor", "--json").assertCode(t, 0)
	assertSameJSON(t, withJSON.stdout, `{
		"schemaVersion": 1, "kind": "doctor",
		"generatedAt": "2026-09-06T09:12:04Z",
		"data": {"problems": [], "warnings": [], "fixed": []}}`)
}
