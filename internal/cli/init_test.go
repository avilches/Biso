package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file walks the case table of docs/spec/cmd/init.md: what each flag
// combination is worth, and which exit code each ending answers with.

func TestInitWritesTheExclusionFileOfTheConfiguredSystem(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")
	assertEqual(t, m.read(filepath.Join(dir, ".gitignore")),
		"board.db\nboard.db-wal\nboard.db-shm\n",
		"the exclusion file biso init writes inside the board")
	assertEqual(t, m.read(filepath.Join(dir, "3f9a2b1c.id")),
		"{ \"storeVersion\": 1 }\n", "the identity marker")
}

func TestWithVcsNoneThereIsNoExclusionFileAndTheNoteSaysLess(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"), "{\"vcs\": \"none\"}\n")

	got := m.run("init", "My project", "--at", "board").assertCode(t, 0)

	dir := filepath.Join(m.dir, "board")
	if m.exists(filepath.Join(dir, ".gitignore")) {
		t.Error("with vcs none no exclusion file is written")
	}
	assertEqual(t, got.stderr,
		"note: the board lives inside this project. Ignore board/ and the board keeps\n"+
			"      its own history; version it and the snapshot travels with your code.\n"+
			"note: the location is stored as the relative path \"board\". A working copy\n"+
			"      outside this project will not have that folder while git ignores it, so\n"+
			"      it will not find the board: use an absolute --at if you work that way\n",
		"the notes without an exclusion file to name")
}

func TestAnAbsoluteAtOutsideTheProjectEmitsNoNote(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "My project", "--at", filepath.Join(m.home, "boards", "b")).assertCode(t, 0)
	assertEqual(t, got.stderr, "", "the notes of an absolute --at outside the project")
}

func TestTheStatusRolesAreSetTogetherWithTheStatuses(t *testing.T) {
	m := newMachine(t)
	for _, c := range []struct {
		argv []string
		code string
	}{
		{[]string{"init", "P", "--initial-status", "To Do"}, "invalid_status_roles"},
		{[]string{"init", "P", "--statuses", "A,B"}, "too_few_statuses"},
		{[]string{"init", "P", "--statuses", "A,B,C"}, "invalid_status_roles"},
		{[]string{"init", "P", "--statuses", "A,B,C", "--initial-status", "A",
			"--active-status", "B", "--terminal-status", "Z"}, "unknown_status_role"},
		{[]string{"init", "P", "--statuses", "A,B,C", "--initial-status", "A",
			"--active-status", "A", "--terminal-status", "C"}, "invalid_status_roles"},
	} {
		got := m.run(append(c.argv, "--json")...).assertCode(t, 2)
		assertErrorCode(t, got.stderr, c.code)
	}
}

func TestAStatusWithNoRoleIsFine(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "My project", "--statuses", "Ideas,To Do,In Progress,Done",
		"--initial-status", "Ideas", "--active-status", "In Progress",
		"--terminal-status", "Done").assertCode(t, 0)

	assertContains(t, got.stdout,
		"  statuses    Ideas (initial) | To Do | In Progress (active) | Done (terminal)")
}

func TestAPrefixIsLettersOnlyAndIsDerivedWhenItIsNotGiven(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "P", "--prefix", "MY-P", "--json").assertCode(t, 2)
	assertErrorCode(t, got.stderr, "invalid_prefix")

	got = m.run("init", "2026", "--json").assertCode(t, 2)
	assertErrorCode(t, got.stderr, "invalid_prefix")

	got = m.run("init", "mi-proyecto-2").assertCode(t, 0)
	assertContains(t, got.stdout, "  prefix      MIPROYECTO")
}

func TestTheDryRunOfInitWritesNothingAndSaysWhereItWouldHaveGone(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")

	got := m.run("init", "My project", "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stdout, "", "the standard output of a preview")
	assertEqual(t, got.stderr,
		"note: board would be created at "+
			filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")+" (--dry-run)\n",
		"the note of the preview")
	if m.exists(filepath.Join(m.dir, ".biso.json")) {
		t.Error("a preview wrote the pointer")
	}
	if m.exists(m.boardsRoot()) {
		t.Error("a preview created the boards root")
	}
}

func TestTheDryRunOfInitNamesTheAtItWasGivenInTheFormItWasGiven(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "My project", "--at", "board", "--dry-run").assertCode(t, 0)
	assertContains(t, got.stderr, "note: board would be created at board (--dry-run)")
}

func TestOverwriteConfigKeepsEveryKeyItIsNotGiven(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project", "--prefix", "MYP", "--types", "task,bug").assertCode(t, 0)

	got := m.run("init", "--overwrite-config", "--priorities", "urgent,normal").assertCode(t, 0)

	assertEqual(t, got.stdout,
		"Rewrote the configuration of board \"My project\"\n"+
			"  statuses    To Do (initial) | In Progress (active) | Done (terminal)\n"+
			"  types       task, bug\n"+
			"  priorities  urgent, normal\n"+
			"  prefix      MYP\n"+
			"This project now points at that board.\n",
		"the output of --overwrite-config")

	// And it survives, which is the half a rewrite in memory would not
	// prove.
	where := m.run("where").assertCode(t, 0)
	assertContains(t, where.stdout, "board    My project")
}

func TestOverwriteConfigWithAnAtThatNamesAnotherBoardIsAnError(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	got := m.run("init", "--overwrite-config", "--at", filepath.Join(m.home, "other")).assertCode(t, 2)
	assertContains(t, got.stderr, "error: this project already has board 3f9a2b1c")
}

func TestFromIsRejectedTogetherWithEveryVocabularyFlag(t *testing.T) {
	m := newMachine(t)
	for _, argv := range [][]string{
		{"init", "--from", "/tmp/x", "--prefix", "MYP"},
		{"init", "--prefix", "MYP", "--from", "/tmp/x"},
		{"init", "--from", "/tmp/x", "--overwrite-config"},
		{"init", "--overwrite-config", "--from", "/tmp/x"},
		{"init", "--from", "/tmp/x", "--statuses", "A,B,C"},
	} {
		got := m.run(append(argv, "--json")...).assertCode(t, 2)
		assertErrorCode(t, got.stderr, "incompatible_flags")
	}
	got := m.run("init", "My project", "--from", "/tmp/x", "--json").assertCode(t, 2)
	assertErrorCode(t, got.stderr, "incompatible_flags")
}

func TestTheGlobalFlagsThatDoNotApply(t *testing.T) {
	m := newMachine(t)

	// --dry-run over a read-only command, with the envelope to read the
	// code off.
	got := m.run("where", "--dry-run", "--json").assertCode(t, 2)
	assertErrorCode(t, got.stderr, "read_only_flag")

	// --print goes without --json, because the two are incompatible in any
	// call and that pair would be the error instead.
	for _, argv := range [][]string{{"where", "--print"}, {"init", "P", "--print"}} {
		got := m.run(argv...).assertCode(t, 2)
		assertEqual(t, got.stderr,
			"error: --print does not apply to a command that affects no task\n",
			"the message of --print where it has nothing to do")
	}
	got = m.run("where", "--dry-run").assertCode(t, 2)
	assertEqual(t, got.stderr,
		"error: --dry-run does not apply to a read-only command\n",
		"the message of --dry-run in a read-only command")
}

func TestAnErrorWithJSONTravelsInTheEnvelopeOnStderr(t *testing.T) {
	m := newMachine(t)

	got := m.run("where", "--json").assertCode(t, 20)

	assertEqual(t, got.stdout, "", "the standard output of a failed call")
	var envelope struct {
		SchemaVersion int    `json:"schemaVersion"`
		Kind          string `json:"kind"`
		GeneratedAt   string `json:"generatedAt"`
		Data          any    `json:"data"`
		Error         struct {
			ExitCode int    `json:"exitCode"`
			Code     string `json:"code"`
			Message  string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(got.stderr), &envelope); err != nil {
		t.Fatalf("stderr is not one JSON object: %v\n%s", err, got.stderr)
	}
	if envelope.SchemaVersion != 1 || envelope.Kind != "error" {
		t.Errorf("envelope = %+v, want schemaVersion 1 and kind error", envelope)
	}
	if envelope.GeneratedAt != "2026-09-06T09:12:04Z" {
		t.Errorf("generatedAt = %q", envelope.GeneratedAt)
	}
	if envelope.Data != nil {
		t.Error("an error envelope never carries data as well")
	}
	if envelope.Error.ExitCode != 20 || envelope.Error.Code != "no_board" {
		t.Errorf("error = %+v, want exit code 20 and code no_board", envelope.Error)
	}
}

func TestAnErrorBeforeTheCallIsAnalyzedStillAnswersJSON(t *testing.T) {
	m := newMachine(t)
	got := m.run("--json", "sett").assertCode(t, 2)
	assertErrorCode(t, got.stderr, "unknown_command")
}

func TestQuietPrintsNothingOnStandardOutputAndSuppressesTheNotes(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "My project", "--at", "board", "--quiet").assertCode(t, 0)
	assertEqual(t, got.stdout, "", "the standard output of a quiet init")
	assertEqual(t, got.stderr, "", "the notes of a quiet init")
}

func TestQuietDoesNotChangeTheOutputOfAReadOnlyCommand(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	got := m.run("where", "--quiet").assertCode(t, 0)
	assertContains(t, got.stdout, "id       3f9a2b1c")
}

func TestTheIdentityComesFromTheEnvironmentAndFromTheMachineFile(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"), "{\"me\": \"@sara\"}\n")
	m.run("init", "My project").assertCode(t, 0)

	got := m.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "me       @sara")

	m.env["BISO_ME"] = "@claude"
	got = m.run("where").assertCode(t, 0)
	assertContains(t, got.stdout, "me       @claude")
}

func TestAnUnknownKeyInTheMachineFileIsAnError(t *testing.T) {
	m := newMachine(t)
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"), "{\"boards_roots\": \"/tmp\"}\n")
	got := m.run("where", "--json").assertCode(t, 3)
	assertErrorCode(t, got.stderr, "bad_config_value")
}

func TestAnUnknownKeyInThePointerIsAnError(t *testing.T) {
	m := newMachine(t)
	m.writeFile(filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"where\": \"here\" }\n")
	got := m.run("where", "--json").assertCode(t, 3)
	assertErrorCode(t, got.stderr, "bad_config_value")
}

func TestAnUppercaseIdInThePointerIsInvalid(t *testing.T) {
	m := newMachine(t)
	m.writeFile(filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3F9A2B1C\" }\n")
	got := m.run("where", "--json").assertCode(t, 3)
	assertErrorCode(t, got.stderr, "bad_config_value")
}

func TestTheEnvironmentVariableForTheWorkingDirectoryAndItsFlag(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	far := m.at(filepath.Join(m.home, "elsewhere"))
	far.env = map[string]string{"BISO_CWD": m.dir}
	assertContains(t, far.run("where").assertCode(t, 0).stdout, "id       3f9a2b1c")

	// And the flag wins over the variable.
	far.env["BISO_CWD"] = far.dir
	assertContains(t, far.run("-C", m.dir, "where").assertCode(t, 0).stdout, "id       3f9a2b1c")
}

// assertErrorCode reads the code out of an error envelope, which is what a
// caller branches on without reading any prose.
func assertErrorCode(t *testing.T, stderr, want string) {
	t.Helper()
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(stderr), &envelope); err != nil {
		t.Fatalf("stderr is not one JSON object: %v\n%s", err, stderr)
	}
	if envelope.Error.Code != want {
		t.Errorf("error code = %q, want %q\n%s", envelope.Error.Code, want, stderr)
	}
}

func TestAnAtWhereNothingCanBeWrittenIsTheEnvironmentFailing(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("a directory's permissions do not stop the superuser")
	}
	m := newMachine(t).withIDs("3f9a2b1c")
	locked := filepath.Join(m.home, "locked")
	if err := os.MkdirAll(locked, 0o555); err != nil {
		t.Fatal(err)
	}

	got := m.run("init", "My project", "--at", filepath.Join(locked, "board"), "--json").assertCode(t, 8)
	assertErrorCode(t, got.stderr, "io_error")
}

func TestWhereDoesNotDodgeADatabaseItCannotRead(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	// Saying which board is in use means opening it, so a file that is not a
	// database at all is exit code 21 here like anywhere else
	// (docs/spec/cmd/where.md and
	// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
	board := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")
	m.writeFile(filepath.Join(board, "board.db"), "this is not a database")
	for _, name := range []string{"board.db-wal", "board.db-shm"} {
		os.Remove(filepath.Join(board, name))
	}

	got := m.run("where").assertCode(t, 21)
	assertEqual(t, got.stderr,
		"error: board 3f9a2b1c's database could not be read\n"+
			"hint: it did not open, or it failed its integrity check, and there is no automatic repair\n"+
			"hint: rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id\n",
		"the message of a database that cannot be read")
}

// TestABoardWhoseDatabaseCannotBeReadIsNotOneForInit walks the two rows of
// the case table of docs/spec/cmd/init.md for a destination "sin una base de
// datos legible", which
// docs/spec/garantias.md#el-segundo-caso-la-base-de-datos-que-no-se-puede-leer
// says does not count as an accessible board for this command: it is rebuilt
// in place, adopting the id of its marker, with exit code 0.
func TestABoardWhoseDatabaseCannotBeReadIsNotOneForInit(t *testing.T) {
	for _, c := range []struct {
		name string
		from func(m *machine, dir string) *machine
		argv []string
	}{
		{"with an explicit destination", func(m *machine, dir string) *machine { return m },
			[]string{"init", "My project", "--at"}},
		{"standing inside it", func(m *machine, dir string) *machine { return m.at(dir) },
			[]string{"init", "My project"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := newMachine(t).withIDs("ffffffff")
			dir := filepath.Join(m.home, "boards", "my-project-3f9a2b1c")
			m.writeFile(filepath.Join(dir, "3f9a2b1c.id"), "{ \"storeVersion\": 1 }\n")
			m.writeFile(filepath.Join(dir, "board.db"), "this is not a database")

			argv := c.argv
			if argv[len(argv)-1] == "--at" {
				argv = append(argv, dir)
			}
			got := c.from(m, dir).run(argv...).assertCode(t, 0)

			assertContains(t, got.stdout, "Created board \"My project\"")
			// The identity is the marker's, not a minted one, and the
			// database that could not be read is now one that can.
			where := c.from(m, dir).run("where").assertCode(t, 0)
			assertContains(t, where.stdout, "id       3f9a2b1c")
			assertContains(t, where.stdout, "path     "+dir)

			// And the project points at it: the board is under no root of
			// this machine, so the pointer carries its path.
			assertEqual(t, m.read(filepath.Join(c.from(m, dir).dir, ".biso.json")),
				"{ \"version\": 1, \"id\": \"3f9a2b1c\", \"path\": \""+dir+"\" }\n",
				"the pointer of a board rebuilt in place")
		})
	}
}

// TestAWarningIsNeverLostWhenTheCallFails is the rule of
// docs/spec/salida-y-terminal.md#notas-y-avisos, "nunca se suprime", applied
// to the call that does not reach its own output: without --json the warning
// is printed as text before the error, and with --json it folds into the same
// envelope (docs/spec/contrato-json.md#los-errores-en-json).
func TestAWarningIsNeverLostWhenTheCallFails(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	// A repeated value warns, and this second init fails with board_exists.
	got := m.run("init", "Another", "--types", "task,task").assertCode(t, 2)
	assertContains(t, got.stderr, "warning: --types: \"task\" given twice, kept once")
	assertContains(t, got.stderr, "error: this project already has board 3f9a2b1c")

	got = m.run("init", "Another", "--types", "task,task", "--json").assertCode(t, 2)
	var envelope struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		Warnings []struct {
			Code  string `json:"code"`
			Flag  string `json:"flag"`
			Value string `json:"value"`
		} `json:"warnings"`
	}
	if err := json.Unmarshal([]byte(got.stderr), &envelope); err != nil {
		t.Fatalf("stderr is not one JSON object: %v\n%s", err, got.stderr)
	}
	if envelope.Error.Code != "board_exists" {
		t.Errorf("error code = %q", envelope.Error.Code)
	}
	if len(envelope.Warnings) != 1 || envelope.Warnings[0].Code != "duplicate_flag_value" ||
		envelope.Warnings[0].Flag != "--types" || envelope.Warnings[0].Value != "task" {
		t.Errorf("warnings = %+v, want the duplicate value of --types", envelope.Warnings)
	}
	if strings.Contains(got.stderr, "warning: ") {
		t.Errorf("with --json the text of a warning does not travel beside the envelope:\n%s", got.stderr)
	}
}

// TestAWarningOfACallThatDoesNotEvenParseIsPrinted is the same rule one step
// earlier: the analysis itself fails, and the warnings it had already found
// do not disappear with it.
func TestAWarningOfACallThatDoesNotEvenParseIsPrinted(t *testing.T) {
	m := newMachine(t)
	got := m.run("init", "P", "--types", "task,task", "--nope").assertCode(t, 2)
	assertContains(t, got.stderr, "warning: --types: \"task\" given twice, kept once")
	assertContains(t, got.stderr, "error: unknown flag: --nope")
}

// TestTheNotesTravelAsTextWithJSONAsWell is the sibling case of the warning
// above, decided the other way round by
// docs/spec/salida-y-terminal.md#notas-y-avisos: a note has no place in any
// envelope, so --json does not suppress it and it stays on stderr.
func TestTheNotesTravelAsTextWithJSONAsWell(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	got := m.run("init", "My project", "--at", "board", "--json").assertCode(t, 0)
	assertContains(t, got.stderr, "note: the board lives inside this project.")
	assertContains(t, got.stderr, "note: the location is stored as the relative path \"board\".")
}

// TestTheDryRunOfOverwriteConfigDoesNotSayTheBoardWouldBeCreated is the note
// of a preview that creates nothing, fixed by docs/spec/cmd/init.md.
func TestTheDryRunOfOverwriteConfigDoesNotSayTheBoardWouldBeCreated(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	got := m.run("init", "--overwrite-config", "--types", "task", "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stderr,
		"note: the configuration of board 3f9a2b1c at "+dir+"\n"+
			"      would be rewritten, and no task would change (--dry-run)\n",
		"the note of a preview of --overwrite-config")

	// And it was a preview: the configuration is the one it had.
	assertContains(t, m.run("where").assertCode(t, 0).stdout, "board    My project")
	got = m.run("init", "--overwrite-config", "--types", "task").assertCode(t, 0)
	assertContains(t, got.stdout, "  types       task")
}

// TestOverwriteConfigWritesThePointerWhenThereIsNone keeps the last line of
// the output honest: docs/spec/cmd/init.md prints "This project now points at
// that board." always, so a call that leaves nothing pointing at it would be
// saying something false.
func TestOverwriteConfigWritesThePointerWhenThereIsNone(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	// Standing inside the board, where the first way finds it and no
	// pointer of this project is anywhere above.
	inside := m.at(dir)
	got := inside.run("init", "--overwrite-config", "--types", "task").assertCode(t, 0)
	assertContains(t, got.stdout, "This project now points at that board.")

	// The board sits directly under a root, so the pointer needs no path:
	// the search by marker finds it.
	assertEqual(t, m.read(filepath.Join(dir, ".biso.json")),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n",
		"the pointer written by --overwrite-config")

	// A board outside every root needs its own path instead.
	other := m.at(filepath.Join(m.home, "other-project")).withIDs("7a1b2c3d")
	board := filepath.Join(m.home, "outside")
	other.run("init", "Other project", "--at", board).assertCode(t, 0)
	if err := os.Remove(filepath.Join(other.dir, ".biso.json")); err != nil {
		t.Fatal(err)
	}
	m.at(board).run("init", "--overwrite-config", "--types", "task").assertCode(t, 0)
	assertEqual(t, m.read(filepath.Join(board, ".biso.json")),
		"{ \"version\": 1, \"id\": \"7a1b2c3d\", \"path\": \""+board+"\" }\n",
		"the pointer of a board that no root holds")
}

// TestTheStatusRoleErrorsNameTheFlagTheyBlame is the rule of the table of
// docs/spec/contrato-json.md#los-errores-en-json: an error of code 2 that
// names a flag carries field and given, and the only one exempt is
// incompatible_flags.
func TestTheStatusRoleErrorsNameTheFlagTheyBlame(t *testing.T) {
	m := newMachine(t)
	for _, c := range []struct {
		argv  []string
		field string
		given string
	}{
		{[]string{"init", "P", "--active-status", "Doing"}, "active-status", "Doing"},
		{[]string{"init", "P", "--statuses", "A,B,C"}, "statuses", "A,B,C"},
	} {
		got := m.run(append(c.argv, "--json")...).assertCode(t, 2)
		var envelope struct {
			Error struct {
				Code  string `json:"code"`
				Field string `json:"field"`
				Given string `json:"given"`
			} `json:"error"`
		}
		if err := json.Unmarshal([]byte(got.stderr), &envelope); err != nil {
			t.Fatalf("stderr is not one JSON object: %v\n%s", err, got.stderr)
		}
		if envelope.Error.Field != c.field || envelope.Error.Given != c.given {
			t.Errorf("%v gave field %q and given %q, want %q and %q",
				c.argv, envelope.Error.Field, envelope.Error.Given, c.field, c.given)
		}
	}
}

// TestVcsCustomOnlyBelongsToTheCustomMode is the row of the table of
// docs/spec/invocacion.md#configuración-de-máquina that gives vcs_custom to
// the custom mode alone, and every key of that file that does not fit is
// exit code 3.
func TestVcsCustomOnlyBelongsToTheCustomMode(t *testing.T) {
	m := newMachine(t)
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"),
		"{\"vcs\": \"git\", \"vcs_custom\": {\"commit\": [\"jj\", \"commit\"]}}\n")

	got := m.run("where", "--json").assertCode(t, 3)
	assertErrorCode(t, got.stderr, "bad_config_value")

	// With the mode it belongs to, the same file is fine.
	m.writeFile(filepath.Join(m.home, ".biso", "config.json"),
		"{\"vcs\": \"custom\", \"vcs_custom\": {\"commit\": [\"jj\", \"commit\"]}}\n")
	m.run("where").assertCode(t, 20)
}

// TestAMalformedPointerIsAnErrorEvenInsideABoard is the rule of
// docs/spec/resolucion-del-tablero.md#cómo-se-lee-el-puntero: a pointer that
// cannot be read is exit code 3 and never something discarded in silence,
// and the working directory being a board of its own does not change it.
func TestAMalformedPointerIsAnErrorEvenInsideABoard(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)
	dir := filepath.Join(m.boardsRoot(), "my-project-3f9a2b1c")

	inside := m.at(dir)
	inside.writeFile(filepath.Join(dir, ".biso.json"), "{ \"version\": 1, \"id\": \"nope\" }\n")

	got := inside.run("where", "--json").assertCode(t, 3)
	assertErrorCode(t, got.stderr, "bad_config_value")
}
