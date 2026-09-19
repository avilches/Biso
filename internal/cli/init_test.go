package cli

import (
	"encoding/json"
	"path/filepath"
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
