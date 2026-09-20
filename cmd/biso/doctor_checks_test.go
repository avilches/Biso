package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// This file is the rest of the table of
// docs/spec/cmd/doctor.md#qué-comprueba: the checks whose line no other test
// ever compared against the specification, the two forms of the three
// messages a repair rewrites, the two messages that have to say something
// else when the list they quote is empty, the order the report comes out in,
// and the one section of that page that exists only because a repair can
// fail halfway.
//
// Two rows of the table are not here and cannot be: duplicate identifiers
// and repeated criterion keys are unreachable with today's schema, which
// answers them by construction. They are checked all the same, because what
// is being asked is whether the data is sound and not whether this program
// is the one that wrote it.

// doctorCodes are the `code` of every finding of one report, in the order
// the report gives them. Reading them off the envelope and not off the text
// is what makes this a question about the order and not about the wording.
func doctorCodes(t *testing.T, m *machine, argv ...string) (problems, warnings []string) {
	t.Helper()
	got := m.run(t, append(argv, "--json")...)
	var envelope struct {
		Data struct {
			Problems []struct{ Code string } `json:"problems"`
			Warnings []struct{ Code string } `json:"warnings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the envelope of biso doctor is not JSON: %v\n%s", err, got.stdout)
	}
	for _, p := range envelope.Data.Problems {
		problems = append(problems, p.Code)
	}
	for _, w := range envelope.Data.Warnings {
		warnings = append(warnings, w.Code)
	}
	return problems, warnings
}

// assertReports fails unless the report carries that line, whole.
func assertReports(t *testing.T, m *machine, line string, argv ...string) {
	t.Helper()
	got := m.run(t, argv...)
	if !strings.Contains(got.stdout, line) {
		t.Errorf("biso %s did not report\n%s\nwhat it printed was:\n%s",
			strings.Join(argv, " "), line, got.stdout)
	}
}

// ------------------------------------------- the checks with no test yet

// TestDoctorReportsAnExtensionKeyTheBoardDoesNotDeclare is the row of an
// undeclared extension key, in its two forms: with a key declared, which
// the message quotes, and with none, which it cannot quote.
func TestDoctorReportsAnExtensionKeyTheBoardDoesNotDeclare(t *testing.T) {
	// With nothing declared, the message says so instead of printing a
	// pair of empty quotes. A board declares nothing by default.
	empty := oneTaskBoard(t)
	empty.execOnBoard(t, boardOf(empty),
		`INSERT INTO task_ext (task_id, key, value) VALUES ('MYP-1', 'trello.card', 'abc')`)
	empty.run(t, "doctor").assertCode(t, 6)
	assertReports(t, empty, strings.TrimSuffix(fixture(t, "doctor-ext-none.txt"), "\n"), "doctor")

	// And with a key declared, it quotes the list. It is another board,
	// because `biso config set extensions` refuses to leave behind the very
	// undeclared key this check reports, so the two states cannot be
	// reached one after the other on the same board.
	declared := oneTaskBoard(t)
	declared.run(t, "config", "set", "extensions", "jira.issue").assertCode(t, 0)
	declared.execOnBoard(t, boardOf(declared),
		`INSERT INTO task_ext (task_id, key, value) VALUES ('MYP-1', 'trello.card', 'abc')`)
	declared.run(t, "doctor").assertCode(t, 6)
	assertReports(t, declared,
		`  MYP-1  ext key "trello.card" is not declared, declared keys are "jira.issue"`, "doctor")
}

// TestDoctorReportsAValueOfAVocabularyTheBoardEmptied is the other message
// that quotes a list which can be emptied: `types` and `priorities` can both
// be left with nothing in them, and then there is no list to quote.
func TestDoctorReportsAValueOfAVocabularyTheBoardEmptied(t *testing.T) {
	m := oneTaskBoard(t)
	dir := boardOf(m)
	m.execOnBoard(t, dir, `UPDATE task SET type = 'bug' WHERE id = 'MYP-1'`)
	m.execOnBoard(t, dir, `UPDATE board_config SET value = '[]' WHERE key = 'types'`)

	m.run(t, "doctor").assertCode(t, 6)
	assertReports(t, m, strings.TrimSuffix(fixture(t, "doctor-vocabulary-none.txt"), "\n"), "doctor")
}

// TestDoctorReportsARoleThatNamesAStatusTheBoardDoesNotHave is the row of
// initial_status, active_status or terminal_status pointing outside
// `statuses`. `biso config set` will not let it happen, so a board in this
// state got there from outside.
func TestDoctorReportsARoleThatNamesAStatusTheBoardDoesNotHave(t *testing.T) {
	m := oneTaskBoard(t)
	m.execOnBoard(t, boardOf(m),
		`UPDATE board_config SET value = 'Doing' WHERE key = 'active_status'`)

	m.run(t, "doctor").assertCode(t, 6)
	assertReports(t, m,
		`  active_status "Doing" is not one of the configured statuses "To Do, In Progress, Done"`,
		"doctor")
}

// TestDoctorReportsStatusesThatCannotHoldThreeDistinctRoles is the other
// configuration row, which has two messages: too few statuses, and two roles
// on the same one.
func TestDoctorReportsStatusesThatCannotHoldThreeDistinctRoles(t *testing.T) {
	m := oneTaskBoard(t)
	m.execOnBoard(t, boardOf(m),
		`UPDATE board_config SET value = '["To Do","Done"]' WHERE key = 'statuses'`)
	m.execOnBoard(t, boardOf(m),
		`UPDATE board_config SET value = 'Done' WHERE key = 'active_status'`)

	m.run(t, "doctor").assertCode(t, 6)
	assertReports(t, m, "  statuses has 2 elements, at least 3 are required", "doctor")
	assertReports(t, m,
		`  active_status and terminal_status are both "Done", the three roles must be distinct`,
		"doctor")
}

// TestDoctorReportsADependencyCycleAndAParentCycle is the pair of rows that
// name the whole ring and report it once and not once per task in it.
func TestDoctorReportsADependencyCycleAndAParentCycle(t *testing.T) {
	m := oneTaskBoard(t)
	dir := boardOf(m)
	m.run(t, "new", "The second task").assertCode(t, 0)
	m.execOnBoard(t, dir,
		`INSERT INTO task_list_item (task_id, field, position, value)
		 VALUES ('MYP-1', 'dependencies', 0, 'MYP-2'),
		        ('MYP-2', 'dependencies', 0, 'MYP-1')`)
	m.execOnBoard(t, dir, `UPDATE task SET parent = 'MYP-2' WHERE id = 'MYP-1'`)
	m.execOnBoard(t, dir, `UPDATE task SET parent = 'MYP-1' WHERE id = 'MYP-2'`)

	got := m.run(t, "doctor").assertCode(t, 6)
	for _, line := range []string{
		"  MYP-1  MYP-1 is part of a dependency cycle: MYP-1 -> MYP-2 -> MYP-1",
		"  MYP-1  MYP-1 is part of a parent cycle: MYP-1 -> MYP-2 -> MYP-1",
	} {
		if !strings.Contains(got.stdout, line) {
			t.Errorf("the report does not carry\n%s\nwhat it printed was:\n%s", line, got.stdout)
		}
	}
	if strings.Count(got.stdout, "dependency cycle") != 1 {
		t.Errorf("the dependency cycle was reported more than once:\n%s", got.stdout)
	}
	if strings.Count(got.stdout, "parent cycle") != 1 {
		t.Errorf("the parent cycle was reported more than once:\n%s", got.stdout)
	}
}

// TestDoctorWarnsAboutAFilesystemWhereWALIsNotSafe is the one warning of the
// table that no test ever saw fire: every other one checked only the
// direction where nothing is wrong.
//
// The probe of
// docs/spec/cmd/doctor.md#el-sondeo-del-sistema-de-ficheros starts by
// writing a file of its own in the board directory, and a directory that
// refuses that is already a place where WAL cannot work, so taking the write
// permission off the directory is enough to make the probe fail for real.
// The two auxiliary files of the database are created first and left there,
// because SQLite needs to write to them and not to create them: that is what
// keeps the board openable while the directory itself is closed.
func TestDoctorWarnsAboutAFilesystemWhereWALIsNotSafe(t *testing.T) {
	skipIfRootCanWriteAnything(t)
	m := oneTaskBoard(t)
	dir := boardOf(m)
	closeTheDirectory(t, dir)

	got := m.run(t, "doctor").assertCode(t, 0)

	want := "  board directory " + quoted(dir) +
		" is on a filesystem where SQLite's WAL mode is not safe " +
		"(the byte-range lock or the shared mmap probe failed)"
	if !strings.Contains(got.stdout, want) {
		t.Errorf("the warning does not match the one of the specification:\n%s", got.stdout)
	}
	if strings.Contains(got.stdout, "Errors:") {
		t.Errorf("the probe reported an error and not a warning:\n%s", got.stdout)
	}
}

// quoted is a path the way the messages of this program write one.
func quoted(s string) string { return `"` + s + `"` }

// closeTheDirectory takes the write permission off a board's directory
// while leaving its database writable, which is what a read-only mount
// looks like from the inside of one file.
func closeTheDirectory(t *testing.T, dir string) {
	t.Helper()
	for _, name := range []string{"board.db-wal", "board.db-shm"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err != nil {
			if err := os.WriteFile(path, nil, 0o644); err != nil {
				t.Fatal(err)
			}
		}
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(dir, 0o755) })
}

// --------------------------------------------- the messages of a repair

// TestTheTwoFormsOfTheCounterAndTheMarkerAreTheOnesOfTheSpecification is the
// subsection of the two forms every repairable message has: one under
// `Errors:` while the problem is still there, and another in the group of
// what was repaired.
func TestTheTwoFormsOfTheCounterAndTheMarkerAreTheOnesOfTheSpecification(t *testing.T) {
	m := brokenBoard(t)
	dir := filepath.Join(m.dir, "board")
	id := markerNameOf(t, dir)
	if err := os.Remove(filepath.Join(dir, id+".id")); err != nil {
		t.Fatal(err)
	}
	// The marker is what answers which directory an identifier means, so a
	// board without one is only reachable from inside itself.
	inside := m.at(dir)

	found := inside.run(t, "doctor").assertCode(t, 6)
	for _, name := range []string{"doctor-highest-error.txt", "doctor-marker-error.txt"} {
		want := strings.ReplaceAll(fixture(t, name), "3f9a2b1c", id)
		if !strings.Contains(found.stdout, want) {
			t.Errorf("%s is not in the report:\n%s", name, found.stdout)
		}
	}

	// The board also carries a broken dependency that --fix cannot repair,
	// so the call still ends with 6: what is being read here is the two
	// blocks of the report and not its exit code.
	fixed := inside.run(t, "doctor", "--fix").assertCode(t, 6)
	for _, name := range []string{"doctor-highest-fixed.txt", "doctor-marker-fixed.txt"} {
		want := strings.ReplaceAll(fixture(t, name), "3f9a2b1c", id)
		if !strings.Contains(fixed.stdout, want) {
			t.Errorf("%s is not in the report of the repair:\n%s", name, fixed.stdout)
		}
	}
	// And neither of the two is listed twice: a repaired error leaves
	// `Errors:` and shows up only in the group of what was repaired.
	for _, name := range []string{"doctor-highest-error.txt", "doctor-marker-error.txt"} {
		gone := strings.ReplaceAll(fixture(t, name), "3f9a2b1c", id)
		if strings.Contains(fixed.stdout, strings.TrimSuffix(gone, "\n")+"\n") {
			t.Errorf("%s stayed under Errors: after being repaired:\n%s", name, fixed.stdout)
		}
	}
}

// TestTheCounterSaysSoWhenNoIdentifierWasEverRecorded is the edge of that
// same message: a counter at zero is not the identifier MYP-0, and the
// message does not invent one.
func TestTheCounterSaysSoWhenNoIdentifierWasEverRecorded(t *testing.T) {
	m := brokenBoard(t)
	m.execOnBoard(t, filepath.Join(m.dir, "board"), `UPDATE board_counter SET last_task_num = 0`)

	assertReports(t, m, strings.TrimSuffix(fixture(t, "doctor-highest-none.txt"), "\n"), "doctor")

	fixed := m.run(t, "doctor", "--fix").assertCode(t, 6)
	want := strings.TrimSuffix(fixture(t, "doctor-highest-none.txt"), "\n") + "; recorded MYP-52"
	if !strings.Contains(fixed.stdout, want) {
		t.Errorf("the repaired form does not carry the same sentence:\n%s", fixed.stdout)
	}
	if strings.Contains(fixed.stdout, "MYP-0") {
		t.Errorf("the report named an identifier that does not exist:\n%s", fixed.stdout)
	}
}

// ------------------------------------------------------- the order, and
// ------------------------------------------- a repair that fails halfway

// TestTheReportComesOutInTheOrderOfTheTableOfChecks is the rule of
// docs/spec/cmd/doctor.md#el-orden-en-que-sale-el-informe. The checks that
// look at a task are all asked in one pass over the board, so without that
// rule the report would come out grouped by task in one place and by check
// in another.
func TestTheReportComesOutInTheOrderOfTheTableOfChecks(t *testing.T) {
	m := brokenBoard(t)
	dir := filepath.Join(m.dir, "board")
	// One finding of five different checks, produced in an order that is
	// not the one of the table: the lease and the extension key are found
	// in the same pass over the tasks, and the role is found before it.
	m.execOnBoard(t, dir, `UPDATE task SET lease_expires_at = '2026-09-06T13:12:04Z',
		lease_holder = '@claude' WHERE id = 'MYP-41'`)
	m.execOnBoard(t, dir,
		`INSERT INTO task_ext (task_id, key, value) VALUES ('MYP-42', 'trello.card', 'abc')`)
	m.execOnBoard(t, dir,
		`UPDATE board_config SET value = 'Doing' WHERE key = 'active_status'`)
	m.write(t, filepath.Join(dir, ".gitignore"), "board.db\n")

	problems, warnings := doctorCodes(t, m, "doctor")

	assertOrder(t, problems, []string{
		"undeclared_extension_key",
		"status_role_unknown",
		"dependency_not_found",
		"lease_invariant",
		"highest_id_behind",
	}, "the errors")
	assertOrder(t, warnings, []string{
		"extra_root_unreadable",
		"ignore_file_mismatch",
	}, "the warnings")
}

// assertOrder fails unless got is exactly want, which is what "in the order
// of the table" means when every check fired once.
func assertOrder(t *testing.T, got, want []string, what string) {
	t.Helper()
	if strings.Join(got, ", ") != strings.Join(want, ", ") {
		t.Errorf("%s came out as\n  %s\nand the table asks for\n  %s",
			what, strings.Join(got, ", "), strings.Join(want, ", "))
	}
}

// TestARepairOfDataSurvivesAMarkerThatCannotBeWritten is the section
// docs/spec/cmd/doctor.md#atomicidad-de---fix-con-varias-reparaciones, which
// exists only because of this command: the data repairs go in one
// transaction and the marker is written afterwards and apart, because
// writing a file cannot be inside a transaction of SQLite. If the marker
// fails, what was repaired stays repaired, the call ends with the exit code
// of not being able to write, and the missing marker shows up again next
// time, because it is still true.
func TestARepairOfDataSurvivesAMarkerThatCannotBeWritten(t *testing.T) {
	skipIfRootCanWriteAnything(t)
	m := brokenBoard(t)
	dir := filepath.Join(m.dir, "board")
	id := markerNameOf(t, dir)
	if err := os.Remove(filepath.Join(dir, id+".id")); err != nil {
		t.Fatal(err)
	}
	inside := m.at(dir)
	// Two repairs of different kinds are pending, which is the case the
	// section is about: the counter, which is data, and the marker, which
	// is a file.
	inside.run(t, "doctor").assertCode(t, 6)
	closeTheDirectory(t, dir)

	got := inside.run(t, "doctor", "--fix").assertCode(t, 8)
	if !strings.Contains(got.stderr, filepath.Join(dir, id+".id")+" cannot be written: ") {
		t.Errorf("the error does not name the marker that could not be written:\n%s", got.stderr)
	}

	// The directory opens again, and what is left is exactly the repair
	// that never happened: the counter is right and the marker is not
	// there.
	if err := os.Chmod(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	after := inside.run(t, "doctor").assertCode(t, 6)
	if strings.Contains(after.stdout, "highest recorded id") {
		t.Errorf("the repair of the counter was undone by the marker failing:\n%s", after.stdout)
	}
	if !strings.Contains(after.stdout, "board directory has no <id>.id marker") {
		t.Errorf("the missing marker did not come back as an error:\n%s", after.stdout)
	}
}
