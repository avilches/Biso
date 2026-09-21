package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// These are the case tables of the four pages of the last step, the rows
// whose answer is a behaviour and not a literal block: what archiving twice
// does, what a configuration change is refused for, and what `biso doctor`
// finds and repairs.

// taskField reads one key of a task's record through the program, which is
// how a test asks what was really written without opening the database.
func taskField(t *testing.T, m *machine, id, key string) any {
	t.Helper()
	got := m.run(t, "get", id, "--json").assertCode(t, 0)
	var envelope struct {
		Data struct {
			Task map[string]any `json:"task"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the envelope of biso get is not JSON: %v\n%s", err, got.stdout)
	}
	return envelope.Data.Task[key]
}

// ------------------------------------------------------------- archive

// oneTaskBoard is the smallest board a test of this file needs.
func oneTaskBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "The parser drops the CRLF").assertCode(t, 0)
	return m
}

// TestArchivingATaskThatIsAlreadyArchivedChangesNothing is the first row of
// the case table: exit code 0, a note, and nothing written.
func TestArchivingATaskThatIsAlreadyArchivedChangesNothing(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "archive", "MYP-1").assertCode(t, 0)
	before := taskField(t, m, "MYP-1", "updatedAt")

	got := m.run(t, "archive", "MYP-1").assertCode(t, 0)

	assertEqual(t, got.stderr, "note: MYP-1 was already archived\n",
		"the note of archiving twice")
	if after := taskField(t, m, "MYP-1", "updatedAt"); after != before {
		t.Errorf("archiving twice touched the task: updatedAt %v became %v", before, after)
	}
}

// TestUnarchivingATaskThatIsNotArchivedChangesNothing is the same row in
// the other direction, with the note the page writes out in full.
func TestUnarchivingATaskThatIsNotArchivedChangesNothing(t *testing.T) {
	m := oneTaskBoard(t)

	got := m.run(t, "archive", "MYP-1", "--unarchive").assertCode(t, 0)

	assertEqual(t, got.stderr, "note: MYP-1 was not archived\n",
		"the note of unarchiving a task that was on the board")
	if archived := taskField(t, m, "MYP-1", "archived"); archived != false {
		t.Errorf("the task came back archived: %v", archived)
	}
}

// TestUnarchivingGivesBackTheStatusAndNoLease is the row of --unarchive: the
// task returns with the status it had, and with no lease, because archiving
// emptied it and nothing claims one here.
func TestUnarchivingGivesBackTheStatusAndNoLease(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "start", "MYP-1").assertCode(t, 0)
	if holder := taskField(t, m, "MYP-1", "leaseHolder"); holder != "@claude" {
		t.Fatalf("biso start left leaseHolder %v", holder)
	}

	m.run(t, "archive", "MYP-1").assertCode(t, 0)
	if holder := taskField(t, m, "MYP-1", "leaseHolder"); holder != nil {
		t.Errorf("archiving left the lease behind: %v", holder)
	}
	if when := taskField(t, m, "MYP-1", "leaseExpiresAt"); when != nil {
		t.Errorf("archiving left leaseExpiresAt behind: %v", when)
	}

	m.run(t, "archive", "MYP-1", "--unarchive").assertCode(t, 0)
	if status := taskField(t, m, "MYP-1", "status"); status != "In Progress" {
		t.Errorf("the task came back with status %v, and not the one it had", status)
	}
	if holder := taskField(t, m, "MYP-1", "leaseHolder"); holder != nil {
		t.Errorf("unarchiving handed out a lease: %v", holder)
	}
}

// TestArchivingWarnsAboutEveryUnfinishedTaskThatDependsOnIt never refuses:
// archiving unblocks them in the same read, and whether that is what the
// caller meant is the caller's business.
func TestArchivingWarnsAboutEveryUnfinishedTaskThatDependsOnIt(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "new", "Waits for the parser", "--add-deps", "MYP-1").assertCode(t, 0)
	m.run(t, "new", "Also waits", "--add-deps", "MYP-1").assertCode(t, 0)
	m.run(t, "new", "Waited and finished", "--add-deps", "MYP-1").assertCode(t, 0)
	m.run(t, "set", "MYP-4", "--status", "Done").assertCode(t, 0)

	got := m.run(t, "archive", "MYP-1").assertCode(t, 0)

	assertEqual(t, got.stderr,
		"warning: MYP-1 is a dependency of MYP-2, which is not finished\n"+
			"warning: MYP-1 is a dependency of MYP-3, which is not finished\n",
		"the warnings of archiving a task others are waiting on")
}

// TestArchivingAnUnfinishedTaskUnblocksItsDependents is the paragraph of the
// page: for blocked and blocks, an archived task counts as finished, and
// nobody has to write anything for that to be true.
func TestArchivingAnUnfinishedTaskUnblocksItsDependents(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "new", "Waits for the parser", "--add-deps", "MYP-1").assertCode(t, 0)
	if blocked := taskField(t, m, "MYP-2", "blocked"); blocked != true {
		t.Fatalf("MYP-2 was not blocked to begin with: %v", blocked)
	}

	m.run(t, "archive", "MYP-1").assertCode(t, 0)

	if blocked := taskField(t, m, "MYP-2", "blocked"); blocked != false {
		t.Errorf("archiving the dependency left MYP-2 blocked: %v", blocked)
	}
}

// TestSeveralReferencesAreAllOrNothing is the last row of the case table.
func TestSeveralReferencesAreAllOrNothing(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "new", "A second task").assertCode(t, 0)

	m.run(t, "archive", "MYP-1", "MYP-99").assertCode(t, 4)

	if archived := taskField(t, m, "MYP-1", "archived"); archived != false {
		t.Errorf("a call that failed archived MYP-1 anyway: %v", archived)
	}
}

// -------------------------------------------------------------- config

// TestConfigRefusesAValueOutsideItsType is the exit code 3 of the case
// table: the value arrived well formed and the key does not take it.
func TestConfigRefusesAValueOutsideItsType(t *testing.T) {
	m := oneTaskBoard(t)

	for _, c := range []struct{ key, value, says string }{
		{"finish_strict", "maybe", "finish_strict is true or false"},
		{"lease_minutes", "0", "greater than zero"},
		{"lease_minutes", "many", "greater than zero"},
		{"initial_status", "Somewhere else", "unknown status: \"Somewhere else\""},
		{"project_name", "", "project_name cannot be empty"},
		{"urgency.due", "soon", "urgency.due is a decimal number"},
	} {
		got := m.run(t, "config", "set", c.key, c.value).assertCode(t, 3)
		if !strings.Contains(got.stderr, c.says) {
			t.Errorf("biso config set %s %q said %q, and it has to say what it expected",
				c.key, c.value, got.stderr)
		}
	}
	// A negative number is the same refusal, reached past the -- that
	// stops the command line from reading it as a flag
	// (docs/spec/valores-de-entrada.md#cómo-se-lee-la-línea-de-comandos).
	negative := m.run(t, "config", "set", "lease_minutes", "--", "-1").assertCode(t, 3)
	if !strings.Contains(negative.stderr, "greater than zero") {
		t.Errorf("biso config set lease_minutes -- -1 said %q", negative.stderr)
	}
}

// TestConfigRefusesAChangeThatWouldLeaveTheBoardInconsistent is the exit
// code 6 of the same table, which is the half that has to read the tasks.
func TestConfigRefusesAChangeThatWouldLeaveTheBoardInconsistent(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "set", "MYP-1", "--type", "task", "--priority", "medium").assertCode(t, 0)

	for _, c := range []struct {
		key, value, says string
	}{
		{"statuses", "In Progress,Done", "at least 3 are required"},
		{"statuses", "Somewhere,In Progress,Done", `"To Do" is the initial_status`},
		{"types", "bug,docs", `type "task" is used by 1 task: MYP-1`},
		{"priorities", "high,low", `priority "medium" is used by 1 task: MYP-1`},
		{"active_status", "To Do", "initial_status and active_status would both be"},
		{"task_prefix", "OTHER", "cannot change on a board that has already handed out"},
	} {
		got := m.run(t, "config", "set", c.key, c.value).assertCode(t, 6)
		if !strings.Contains(got.stderr, c.says) {
			t.Errorf("biso config set %s %q said %q", c.key, c.value, got.stderr)
		}
	}
}

// TestConfigEmptiesAVocabularyThatNobodyUses is the row that says so, and
// its consequence: the board stops taking --type at all.
func TestConfigEmptiesAVocabularyThatNobodyUses(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)

	m.run(t, "config", "set", "types", "").assertCode(t, 0)

	assertEqual(t, m.run(t, "config", "get", "types").assertCode(t, 0).stdout, "\n",
		"the value of an emptied vocabulary")
	m.run(t, "new", "A task", "--type", "task").assertCode(t, 3)
	m.run(t, "new", "A task").assertCode(t, 0)
}

// TestChangingTheTaskPrefixOnAnEmptyBoardWorks is the other half of the row
// that refuses it: while there is no identifier, there is nothing for the
// prefix to contradict.
func TestChangingTheTaskPrefixOnAnEmptyBoardWorks(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)

	m.run(t, "config", "set", "task_prefix", "other").assertCode(t, 0)

	assertEqual(t, m.run(t, "config", "get", "task_prefix").assertCode(t, 0).stdout,
		"OTHER\n", "a prefix is stored in upper case")
	assertEqual(t, m.run(t, "new", "A task").assertCode(t, 0).stdout, "OTHER-1\n",
		"the identifier the board hands out after the change")
}

// TestRenamingABoardMovesNothing is the promise the page makes twice: no
// configuration change touches the file system, and none recalculates the
// prefix.
func TestRenamingABoardMovesNothing(t *testing.T) {
	m := oneTaskBoard(t)
	dir := filepath.Join(m.dir, "board")
	before := taskField(t, m, "MYP-1", "updatedAt")

	m.run(t, "config", "set", "project_name", "Another name entirely").assertCode(t, 0)

	if _, err := os.Stat(dir); err != nil {
		t.Errorf("renaming the board moved its folder: %v", err)
	}
	assertEqual(t, m.run(t, "config", "get", "task_prefix").assertCode(t, 0).stdout,
		"MYP\n", "the prefix after a rename")
	if after := taskField(t, m, "MYP-1", "updatedAt"); after != before {
		t.Errorf("a configuration change touched a task: updatedAt %v became %v", before, after)
	}
	m.run(t, "where").assertCode(t, 0)
}

// TestConfigSetOfTheSameValueStillCompletes is the row that says so: it is
// not an error and it is not silent, it is a `set` with its note.
func TestConfigSetOfTheSameValueStillCompletes(t *testing.T) {
	m := oneTaskBoard(t)

	got := m.run(t, "config", "set", "project_name", "My project").assertCode(t, 0)

	assertEqual(t, got.stderr, "note: project_name = My project\n",
		"the note of setting a value to what it already was")
}

// TestConfigSubcommandsAreJudgedHere is the usage half of the command,
// which the generic table of commands.go carries none of because the three
// subcommands are positional.
func TestConfigSubcommandsAreJudgedHere(t *testing.T) {
	m := oneTaskBoard(t)

	for _, c := range []struct {
		argv []string
		says string
	}{
		{[]string{"config"}, "needs one of get, set or list"},
		{[]string{"config", "listen"}, `unknown config subcommand: "listen"`},
		{[]string{"config", "get"}, "biso config get needs a key"},
		{[]string{"config", "set", "lease_minutes"}, "biso config set needs a value"},
		{[]string{"config", "set", "lease_minutes", "45", "60"}, "unexpected argument: 60"},
		{[]string{"config", "list", "extra"}, "unexpected argument: extra"},
		{[]string{"config", "list", "--dry-run"}, "--dry-run does not apply"},
		{[]string{"config", "get", "statuses", "--dry-run"}, "--dry-run does not apply"},
		{[]string{"config", "list", "--print"}, "--print does not apply"},
	} {
		got := m.run(t, c.argv...).assertCode(t, 2)
		if !strings.Contains(got.stderr, c.says) {
			t.Errorf("biso %s said %q", strings.Join(c.argv, " "), got.stderr)
		}
	}
}

// -------------------------------------------------------------- doctor

// TestDoctorFindsAndRepairsALeaseThatNoTaskJustifies is the one row that
// repairs in a single direction: the two lease fields are what is left
// over, and the status and the people assigned are the data.
func TestDoctorFindsAndRepairsALeaseThatNoTaskJustifies(t *testing.T) {
	m := leftOverLeaseBoard(t)

	found := m.run(t, "doctor").assertCode(t, 6)
	if !strings.Contains(found.stdout, fixture(t, "doctor-lease-error.txt")) {
		t.Errorf("the report did not carry the line of the specification:\n%s", found.stdout)
	}

	fixed := m.run(t, "doctor", "--fix").assertCode(t, 0)
	if !strings.Contains(fixed.stdout, fixture(t, "doctor-lease-fixed.txt")) {
		t.Errorf("the repair did not carry the line of the specification:\n%s", fixed.stdout)
	}
	if holder := taskField(t, m, "MYP-52", "leaseHolder"); holder != nil {
		t.Errorf("--fix left leaseHolder behind: %v", holder)
	}
	if when := taskField(t, m, "MYP-52", "leaseExpiresAt"); when != nil {
		t.Errorf("--fix left leaseExpiresAt behind: %v", when)
	}
	// And the status and the people assigned are untouched, which is the
	// whole reason the repair is allowed without a decision.
	if status := taskField(t, m, "MYP-52", "status"); status != "To Do" {
		t.Errorf("--fix moved the task to %v", status)
	}
	m.run(t, "doctor").assertCode(t, 0)
}

// leftOverLeaseBoard is a board with the invariant of
// docs/spec/lease.md#el-vaciado / broken: a lease on a task that is neither
// active nor assigned. No write of biso can produce it, which is why the
// page calls it external damage.
func leftOverLeaseBoard(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "The parser drops the CRLF").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")
	m.execOnBoard(t, dir, `UPDATE task SET id = 'MYP-52', num = 52 WHERE id = 'MYP-1'`)
	m.execOnBoard(t, dir, `UPDATE board_counter SET last_task_num = 52`)
	m.execOnBoard(t, dir, `UPDATE task SET lease_expires_at = '2026-09-06T13:12:04Z',
		lease_holder = '@claude' WHERE id = 'MYP-52'`)
	return m
}

// TestDoctorWritesAMarkerThatIsMissingAndRefusesToPickBetweenTwoIdentities
// is the pair of rows that look alike and repair the opposite way.
func TestDoctorWritesAMarkerThatIsMissingAndRefusesToPickBetweenTwoIdentities(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	dir := filepath.Join(m.dir, "board")
	id := markerNameOf(t, dir)

	if err := os.Remove(filepath.Join(dir, id+".id")); err != nil {
		t.Fatal(err)
	}
	// The call is made from inside the board, which is the one way of
	// reaching a board with no marker: a pointer cannot name it, because
	// the marker is what answers which directory an identifier means
	// (docs/spec/resolucion-del-tablero.md#cómo-se-busca-el-tablero). That
	// way exists precisely so that this repair has something to repair.
	inside := m.at(dir)
	found := inside.run(t, "doctor").assertCode(t, 6)
	if !strings.Contains(found.stdout, "board directory has no <id>.id marker") {
		t.Errorf("the missing marker was not reported:\n%s", found.stdout)
	}
	inside.run(t, "doctor", "--fix").assertCode(t, 0)
	if _, err := os.Stat(filepath.Join(dir, id+".id")); err != nil {
		t.Errorf("--fix did not write the marker back: %v", err)
	}
	// And now that it is back, the project's pointer resolves again.
	m.run(t, "doctor").assertCode(t, 0)

	// A marker that names another identifier loses something either way,
	// so it is decided by hand and --fix leaves it alone.
	if err := os.Rename(filepath.Join(dir, id+".id"),
		filepath.Join(dir, "a1b2c3d4.id")); err != nil {
		t.Fatal(err)
	}
	stillBroken := inside.run(t, "doctor", "--fix").assertCode(t, 6)
	if !strings.Contains(stillBroken.stdout, `marker file names id "a1b2c3d4"`) {
		t.Errorf("the mismatched marker was not reported:\n%s", stillBroken.stdout)
	}
}

// markerNameOf is the identifier of the board in that directory, read off
// the one marker file that is there.
func markerNameOf(t *testing.T, dir string) string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".id") {
			return strings.TrimSuffix(e.Name(), ".id")
		}
	}
	t.Fatalf("%s has no marker", dir)
	return ""
}

// TestDoctorReportsATaskItCannotReadAndCarriesOn is the row that separates
// one bad task from a database that does not open: the first is a finding,
// the second is the end of the command.
func TestDoctorReportsATaskItCannotReadAndCarriesOn(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	m.run(t, "new", "A task that stays readable").assertCode(t, 0)
	m.run(t, "new", "A task that will not parse").assertCode(t, 0)
	m.execOnBoard(t, filepath.Join(m.dir, "board"),
		`UPDATE task SET created_at = 'the day before yesterday' WHERE id = 'MYP-2'`)

	got := m.run(t, "doctor").assertCode(t, 6)

	if !strings.Contains(got.stdout, `MYP-2  task "MYP-2" could not be parsed`) {
		t.Errorf("the unreadable task was not reported:\n%s", got.stdout)
	}
	if strings.Contains(got.stdout, "MYP-1") {
		t.Errorf("the readable task was reported too:\n%s", got.stdout)
	}
}

// TestDoctorAbortsOnADatabaseItCannotRead is the other half of that row:
// exit code 21, no report at all, and the message of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar.
func TestDoctorAbortsOnADatabaseItCannotRead(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	db := filepath.Join(m.dir, "board", "board.db")
	if err := os.WriteFile(db, []byte("this is not a database at all"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := m.run(t, "doctor").assertCode(t, 21)

	assertEqual(t, got.stdout, "", "the standard output of a doctor that aborted")
	if !strings.Contains(got.stderr, "error:") {
		t.Errorf("stderr = %q", got.stderr)
	}
}

// TestDoctorReportsAValueTheBoardNoLongerConfigures is the row that mirrors
// the refusal of `biso config set`: what that command will not let happen
// is what this one reports when it happened anyway.
func TestDoctorReportsAValueTheBoardNoLongerConfigures(t *testing.T) {
	m := oneTaskBoard(t)
	m.execOnBoard(t, filepath.Join(m.dir, "board"),
		`UPDATE task SET status = 'Blocked' WHERE id = 'MYP-1'`)

	got := m.run(t, "doctor").assertCode(t, 6)

	want := `MYP-1  status "Blocked" is not one of the configured statuses ` +
		`"To Do, In Progress, Done"`
	if !strings.Contains(got.stdout, want) {
		t.Errorf("the report does not carry the message of the specification:\n%s", got.stdout)
	}
}

// TestDoctorWarnsAboutAnExclusionFileLeftOverFromGit is the one case of a
// stale exclusion file that can be recognized with certainty. It is a
// warning and not an error, so the board is still healthy and the exit code
// is still 0.
func TestDoctorWarnsAboutAnExclusionFileLeftOverFromGit(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "board").assertCode(t, 0)
	if _, err := os.Stat(filepath.Join(m.dir, "board", ".gitignore")); err != nil {
		t.Fatalf("biso init wrote no .gitignore to leave behind: %v", err)
	}
	m.write(t, filepath.Join(m.home, ".biso", "config.json"), `{"vcs": "none"}`+"\n")

	got := m.run(t, "doctor").assertCode(t, 0)

	if !strings.Contains(got.stdout, fixture(t, "doctor-ignore-file.txt")) {
		t.Errorf("the report does not carry the warning of the specification:\n%s", got.stdout)
	}
	if strings.Contains(got.stdout, "Errors:") {
		t.Errorf("a warning was reported as an error:\n%s", got.stdout)
	}
}

// TestDoctorIsReadOnlyWithoutFix is the line of the page: --print and
// --dry-run are bad usage there, the same as in any other read-only
// command, and --fix is what turns it into a writing one.
func TestDoctorIsReadOnlyWithoutFix(t *testing.T) {
	m := oneTaskBoard(t)

	m.run(t, "doctor", "--dry-run").assertCode(t, 2)
	m.run(t, "doctor", "--print").assertCode(t, 2)
	m.run(t, "doctor", "--fix", "--dry-run").assertCode(t, 0)
	m.run(t, "doctor", "--fix", "--print").assertCode(t, 2)
}

// TestArchivingATaskSomebodyElseHoldsWarnsAndClearsTheLease is the row of
// the case table about a live lease of another identity: the warning of
// docs/spec/salida-y-terminal.md#notas-y-avisos comes out, and the two
// fields are emptied in that same write whoever they belonged to
// (docs/spec/lease.md#el-vaciado).
func TestArchivingATaskSomebodyElseHoldsWarnsAndClearsTheLease(t *testing.T) {
	m := oneTaskBoard(t)
	m.env["BISO_ME"] = "@sara"
	m.run(t, "start", "MYP-1").assertCode(t, 0)
	m.env["BISO_ME"] = "@claude"

	got := m.run(t, "archive", "MYP-1").assertCode(t, 0)

	if !strings.Contains(got.stderr, "warning: MYP-1's lease is held by @sara until ") {
		t.Errorf("archiving a task somebody else holds said %q", got.stderr)
	}
	if holder := taskField(t, m, "MYP-1", "leaseHolder"); holder != nil {
		t.Errorf("the lease of another identity survived archiving: %v", holder)
	}
}

// TestUnarchivingByTextReachesTheArchivedTask is the one exception of
// docs/spec/referencias.md#la-búsqueda-por-texto: a text reference looks
// only at the tasks on the board, and with --unarchive it looks at the
// whole one, because what that call names is off the board by definition.
func TestUnarchivingByTextReachesTheArchivedTask(t *testing.T) {
	m := oneTaskBoard(t)
	m.run(t, "archive", "MYP-1").assertCode(t, 0)

	// The ordinary scope does not reach it, which is what makes the
	// exception necessary rather than convenient.
	m.run(t, "archive", "the parser").assertCode(t, 4)

	got := m.run(t, "archive", "the parser", "--unarchive").assertCode(t, 0)
	if !strings.Contains(got.stderr, `note: "the parser" matched MYP-1`) {
		t.Errorf("the text reference of --unarchive said %q", got.stderr)
	}
	if archived := taskField(t, m, "MYP-1", "archived"); archived != false {
		t.Errorf("the task is still archived: %v", archived)
	}

	// And naming by text a task that is already on the board is the same
	// idempotent note it would get by identifier, which is why the scope
	// is the whole board and not the archived half.
	again := m.run(t, "archive", "the parser", "--unarchive").assertCode(t, 0)
	if !strings.Contains(again.stderr, "note: MYP-1 was not archived") {
		t.Errorf("unarchiving by text a task on the board said %q", again.stderr)
	}
}
