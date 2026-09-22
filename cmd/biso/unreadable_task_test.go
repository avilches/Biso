package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// These tests are the one definition of an unreadable task of
// docs/spec/garantias.md#qué-se-comprueba, checked from outside the program:
// the same corrupted row is put next to healthy tasks and every reading
// command has to give the answer that page promises, with no command
// deciding for itself. Each corruption is an edit of the database, because
// that is the only way a task gets there.

// corruption is one way of leaving the task MYP-2 unreadable.
type corruption struct {
	name string
	// prepare runs commands of the program before the statements do, for the
	// corruptions that need a comment, a criterion or a question to exist.
	prepare func(t *testing.T, m *machine)
	sql     []string
	// field and given are what the error of `biso get` carries.
	field, given string
	// reason is the part of the message of `biso get`, after
	// "MYP-2 cannot be read: ", that is fixed text.
	reason string
	// vocabulary is a value outside status, type or priority, which is the
	// one kind `biso doctor` reports with a code of its own and the one a
	// write can repair.
	vocabulary bool
}

const notConfigured = ", which this board does not configure"
const notAnInstant = " is not an instant (YYYY-MM-DDTHH:MM:SSZ): "

func instantOf(field, given string) string {
	return field + notAnInstant + `"` + given + `"`
}

func withAComment(t *testing.T, m *machine) {
	t.Helper()
	m.run(t, "comment", "MYP-2", "a comment").assertCode(t, 0)
}

func withACriterion(t *testing.T, m *machine) {
	t.Helper()
	m.run(t, "set", "MYP-2", "--add-ac", "A criterion").assertCode(t, 0)
}

const aQuestion = `UPDATE task SET question_author = '@sara', question_body = 'Which one?', `

var everyCorruption = []corruption{
	{name: "status outside the vocabulary", sql: []string{`UPDATE task SET status = 'Weird' WHERE id = 'MYP-2'`},
		field: "status", given: "Weird", reason: `its status is "Weird"` + notConfigured, vocabulary: true},
	{name: "empty status", sql: []string{`UPDATE task SET status = '' WHERE id = 'MYP-2'`},
		field: "status", given: "", reason: `its status is ""` + notConfigured, vocabulary: true},
	{name: "status spelled another way", sql: []string{`UPDATE task SET status = 'done' WHERE id = 'MYP-2'`},
		field: "status", given: "done", reason: `its status is "done"` + notConfigured, vocabulary: true},
	{name: "type outside the vocabulary", sql: []string{`UPDATE task SET type = 'epic' WHERE id = 'MYP-2'`},
		field: "type", given: "epic", reason: `its type is "epic"` + notConfigured, vocabulary: true},
	{name: "priority outside the vocabulary", sql: []string{`UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`},
		field: "priority", given: "urgent", reason: `its priority is "urgent"` + notConfigured, vocabulary: true},
	{name: "priority spelled another way", sql: []string{`UPDATE task SET priority = 'HIGH' WHERE id = 'MYP-2'`},
		field: "priority", given: "HIGH", reason: `its priority is "HIGH"` + notConfigured, vocabulary: true},

	{name: "due that is no date", sql: []string{`UPDATE task SET due = 'nope' WHERE id = 'MYP-2'`},
		field: "due", given: "nope", reason: `due is not a calendar day (YYYY-MM-DD): "nope"`},
	{name: "due without its zeros", sql: []string{`UPDATE task SET due = '2026-9-1' WHERE id = 'MYP-2'`},
		field: "due", given: "2026-9-1", reason: `due is not a calendar day (YYYY-MM-DD): "2026-9-1"`},
	{name: "due that is an instant", sql: []string{`UPDATE task SET due = '2026-09-21T10:00:00Z' WHERE id = 'MYP-2'`},
		field: "due", given: "2026-09-21T10:00:00Z", reason: `due is not a calendar day (YYYY-MM-DD): "2026-09-21T10:00:00Z"`},
	{name: "createdAt that is no date", sql: []string{`UPDATE task SET created_at = 'nope' WHERE id = 'MYP-2'`},
		field: "createdAt", given: "nope", reason: instantOf("createdAt", "nope")},
	{name: "empty createdAt", sql: []string{`UPDATE task SET created_at = '' WHERE id = 'MYP-2'`},
		field: "createdAt", given: "", reason: instantOf("createdAt", "")},
	{name: "createdAt with an offset", sql: []string{`UPDATE task SET created_at = '2026-09-21T10:00:00+02:00' WHERE id = 'MYP-2'`},
		field: "createdAt", given: "2026-09-21T10:00:00+02:00", reason: instantOf("createdAt", "2026-09-21T10:00:00+02:00")},
	{name: "updatedAt that is no day", sql: []string{`UPDATE task SET updated_at = '2026-13-45' WHERE id = 'MYP-2'`},
		field: "updatedAt", given: "2026-13-45", reason: instantOf("updatedAt", "2026-13-45")},
	{name: "empty updatedAt", sql: []string{`UPDATE task SET updated_at = '' WHERE id = 'MYP-2'`},
		field: "updatedAt", given: "", reason: instantOf("updatedAt", "")},
	{name: "leaseExpiresAt that is no date", sql: []string{`UPDATE task SET lease_expires_at = 'x' WHERE id = 'MYP-2'`},
		field: "leaseExpiresAt", given: "x", reason: instantOf("leaseExpiresAt", "x")},
	{name: "question asked at no date", sql: []string{aQuestion + `question_asked_at = 'x' WHERE id = 'MYP-2'`},
		field: "question.askedAt", given: "x", reason: instantOf("question.askedAt", "x")},
	{name: "question with no date", sql: []string{aQuestion + `question_asked_at = '' WHERE id = 'MYP-2'`},
		field: "question.askedAt", given: "", reason: instantOf("question.askedAt", "")},
	{name: "comment with no date", prepare: withAComment,
		sql:   []string{`UPDATE task_comment SET created_at = 'bad' WHERE task_id = 'MYP-2'`},
		field: "comment #1 createdAt", given: "bad", reason: instantOf("comment #1 createdAt", "bad")},
	{name: "comment with an empty date", prepare: withAComment,
		sql:   []string{`UPDATE task_comment SET created_at = '' WHERE task_id = 'MYP-2'`},
		field: "comment #1 createdAt", given: "", reason: instantOf("comment #1 createdAt", "")},

	{name: "archived that is no boolean", sql: []string{`UPDATE task SET archived = 7 WHERE id = 'MYP-2'`},
		field: "archived", given: "7", reason: `archived`},
	{name: "ordinal that is no number", sql: []string{`UPDATE task SET ordinal = 'abc' WHERE id = 'MYP-2'`},
		field: "ordinal", given: "abc", reason: `ordinal`},
	{name: "criterion counter that is no number", sql: []string{`UPDATE task SET next_criterion_key = 'x' WHERE id = 'MYP-2'`},
		field: "next_criterion_key", given: "x", reason: `next_criterion_key`},
	{name: "comment counter that is no number", sql: []string{`UPDATE task SET next_comment_key = 'x' WHERE id = 'MYP-2'`},
		field: "next_comment_key", given: "x", reason: `next_comment_key`},
	{name: "criterion that is no boolean", prepare: withACriterion,
		sql:   []string{`UPDATE task_criterion SET checked = 5 WHERE task_id = 'MYP-2'`},
		field: "criterion checked", given: "5", reason: `criterion checked`},
	{name: "criterion key that is no number", prepare: withACriterion,
		sql:   []string{`UPDATE task_criterion SET key = 'x' WHERE task_id = 'MYP-2'`},
		field: "criterion key", given: "x", reason: `criterion key`},
	{name: "comment key that is no number", prepare: withAComment,
		sql:   []string{`UPDATE task_comment SET key = 'x' WHERE task_id = 'MYP-2'`},
		field: "comment key", given: "x", reason: `comment key`},
}

// validOf maps the field of a vocabulary corruption to the valid values the
// error has to carry, which are the ones of a fresh board.
var validOf = map[string][]string{
	"status":   {"To Do", "In Progress", "Done"},
	"type":     {"task", "bug", "docs"},
	"priority": {"high", "medium", "low"},
}

// corrupted is the small board with MYP-2 broken the way c says. MYP-1, MYP-3
// and MYP-4 are the healthy tasks around it.
func corrupted(t *testing.T, c corruption) *machine {
	t.Helper()
	m := smallBoard(t)
	if c.prepare != nil {
		c.prepare(t, m)
	}
	for _, statement := range c.sql {
		m.execOnBoard(t, m.boardDir(t), statement)
	}
	return m
}

func (m *machine) scalarOnBoard(t *testing.T, query string) string {
	t.Helper()
	db := openDatabase(t, m.boardDir(t))
	defer db.Close()
	var value string
	if err := db.QueryRow(query).Scan(&value); err != nil {
		t.Fatal(err)
	}
	return value
}

func decode(t *testing.T, text string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(text), &out); err != nil {
		t.Fatalf("not JSON: %v\n%s", err, text)
	}
	return out
}

func skippedOf(t *testing.T, stdout string) []any {
	t.Helper()
	data := decode(t, stdout)["data"].(map[string]any)
	skipped, ok := data["skipped"].([]any)
	if !ok {
		t.Fatalf("data.skipped is not a list:\n%s", stdout)
	}
	return skipped
}

const warningOfMYP2 = "warning: 1 task could not be read and was skipped: MYP-2\n"

func TestLsSkipsTheUnreadableTaskWhateverTheCorruption(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			ids := m.run(t, "ls", "--ids", "--any-status").assertCode(t, 0)
			assertEqual(t, ids.stdout, "MYP-1\nMYP-3\nMYP-4\n", "the rows of the listing")
			assertEqual(t, ids.stderr, warningOfMYP2, "the warning of the listing")

			count := m.run(t, "ls", "--count", "--any-status").assertCode(t, 0)
			assertEqual(t, count.stdout, "3\n", "the count of the listing")
			assertEqual(t, count.stderr, warningOfMYP2, "the warning of the count")

			asJSON := m.run(t, "ls", "--any-status", "--json").assertCode(t, 0)
			data := decode(t, asJSON.stdout)["data"].(map[string]any)
			if data["matched"] != float64(3) {
				t.Errorf("data.matched is %v and not 3", data["matched"])
			}
			if skipped := skippedOf(t, asJSON.stdout); len(skipped) != 1 || skipped[0] != "MYP-2" {
				t.Errorf("data.skipped is %v and not [MYP-2]", skipped)
			}
		})
	}
}

// TestLsWarnsAboutTheUnreadableTaskEvenIfTheFilterWouldNotHaveKeptIt is the
// rule that the warning names every unreadable task of the board: such a
// task cannot be compared with a filter, so nobody can say it does not match.
func TestLsWarnsAboutTheUnreadableTaskEvenIfTheFilterWouldNotHaveKeptIt(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			none := m.run(t, "ls", "--status", "Done").assertCode(t, 0)
			assertEqual(t, none.stdout, "", "the rows of a listing that matches nothing")
			assertEqual(t, none.stderr, warningOfMYP2+"note: no tasks match\n",
				"the note and the warning of a listing that matches nothing")

			byLabel := m.run(t, "ls", "--ids", "--label", "frontend").assertCode(t, 0)
			assertEqual(t, byLabel.stdout, "MYP-1\n", "the rows of a label that only MYP-1 has")
			assertEqual(t, byLabel.stderr, warningOfMYP2, "the warning of a listing MYP-2 could have matched")
		})
	}
}

// TestLsFindsNoUnreadableTaskWithAValueThatIsValidAndEmpty is the other side of
// the definition: an empty type, priority, due or lease is not a fault.
func TestLsKeepsATaskWhoseOptionalFieldsAreEmpty(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t),
		`UPDATE task SET type = '', priority = '', due = '', lease_expires_at = '' WHERE id = 'MYP-2'`)

	got := m.run(t, "ls", "--ids", "--any-status").assertCode(t, 0)

	assertEqual(t, got.stdout, "MYP-1\nMYP-2\nMYP-3\nMYP-4\n", "the rows with empty optional fields")
	assertEqual(t, got.stderr, "", "the standard error with empty optional fields")
}

func TestGetOfTheUnreadableTaskIsExitCodeThreeWithTheExactReason(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			for _, argv := range [][]string{
				{"get", "MYP-2"},
				{"get", "MYP-2", "--section", "plan"},
			} {
				got := m.run(t, argv...).assertCode(t, 3)
				assertEqual(t, got.stdout, "", "the card of an unreadable task")
				if !strings.HasPrefix(got.stderr, "error: MYP-2 cannot be read: ") ||
					!strings.Contains(got.stderr, c.reason) {
					t.Errorf("the message is %q and not one that says %q", got.stderr, c.reason)
				}
				if strings.Count(strings.TrimSuffix(got.stderr, "\n"), "\n") != 0 {
					t.Errorf("the error is more than one line: %q", got.stderr)
				}
			}

			asJSON := m.run(t, "get", "MYP-2", "--json").assertCode(t, 3)
			object := decode(t, asJSON.stderr)["error"].(map[string]any)
			if object["code"] != "undecodable_task" || object["exitCode"] != float64(3) {
				t.Errorf("the error object is %v", object)
			}
			if object["field"] != c.field || object["given"] != c.given {
				t.Errorf("field and given are %v and %v, not %q and %q",
					object["field"], object["given"], c.field, c.given)
			}
			valid, hasValid := object["valid"]
			if c.vocabulary {
				list, _ := valid.([]any)
				want := validOf[c.field]
				if !hasValid || len(list) != len(want) {
					t.Fatalf("valid is %v and not %v", valid, want)
				}
				for i := range want {
					if list[i] != want[i] {
						t.Errorf("valid is %v and not %v", list, want)
					}
				}
			} else if hasValid {
				t.Errorf("a field that is no vocabulary carries valid: %v", valid)
			}
		})
	}
}

func TestGetOfAnotherTaskPrintsItsCardAndWarns(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			got := m.run(t, "get", "MYP-1", "--section", "meta").assertCode(t, 0)

			if !strings.HasPrefix(got.stdout, "MYP-1  Task 1\n") {
				t.Errorf("the card is %q", got.stdout)
			}
			assertEqual(t, got.stderr, warningOfMYP2, "the warning of the card of another task")
		})
	}
}

// TestATextThatOnlyTheUnreadableTaskMatchesFindsNothing is the row about
// resolving a reference by text: the task does not take part, and the
// warning names it.
func TestATextThatOnlyTheUnreadableTaskMatchesFindsNothing(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET title = 'Zebra', status = 'Weird' WHERE id = 'MYP-2'`)

	got := m.run(t, "get", "Zebra").assertCode(t, 4)

	if !strings.Contains(got.stderr, warningOfMYP2) {
		t.Errorf("the warning does not name the task that was left out: %q", got.stderr)
	}
}

func TestExportWritesOnlyTheReadableTasksAndAnswersSix(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			got := m.run(t, "export").assertCode(t, 6)

			assertEqual(t, got.stderr, warningOfMYP2, "the warning of the export")
			assertEqual(t, exportedIDs(got.stdout), "MYP-1,MYP-3,MYP-4", "the tasks that were written")

			out := filepath.Join(m.home, "backup.ndjson")
			toFile := m.run(t, "export", "--out", out).assertCode(t, 6)
			assertEqual(t, exportedIDs(m.read(t, out)), "MYP-1,MYP-3,MYP-4", "the tasks that were written to a file")
			if !strings.Contains(toFile.stderr, warningOfMYP2) {
				t.Errorf("the warning is missing with --out: %q", toFile.stderr)
			}

			// An unreadable task does not stop being one because a filter
			// would not have kept it.
			filtered := m.run(t, "export", "--status", "Done").assertCode(t, 6)
			assertEqual(t, filtered.stdout, "", "the tasks of a filter that keeps none")
			assertEqual(t, filtered.stderr, warningOfMYP2, "the warning of a filtered export")
			// `biso export` never accepts --json, its output is already
			// NDJSON (docs/spec/cmd/export.md#el-rechazo-de---json);
			// data.skipped of a set read is checked through `biso snapshot`
			// and `biso ls` instead.
		})
	}
}

func exportedIDs(ndjson string) string {
	var ids []string
	for _, line := range strings.Split(strings.TrimSpace(ndjson), "\n") {
		if line == "" {
			continue
		}
		var task struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal([]byte(line), &task); err != nil {
			return "not JSON: " + line
		}
		ids = append(ids, task.ID)
	}
	return strings.Join(ids, ",")
}

func TestSnapshotWritesOnlyTheReadableTasksAndTheCopyImportsWhole(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)
			dir := m.boardDir(t)

			got := m.run(t, "snapshot", "--vcs", "none").assertCode(t, 6)

			assertEqual(t, got.stderr, warningOfMYP2, "the warning of the snapshot")
			assertEqual(t, exportedIDs(m.read(t, filepath.Join(dir, "snapshot.ndjson"))),
				"MYP-1,MYP-3,MYP-4", "the tasks of snapshot.ndjson")
			if !strings.Contains(got.stdout, "(3 tasks)") {
				t.Errorf("the output does not count the tasks that were written: %q", got.stdout)
			}

			// What was written is read back whole, which is the promise
			// the copy exists for.
			restored := newMachine(t)
			restored.env["BISO_ME"] = "@claude"
			restored.run(t, "init", "--at", filepath.Join(restored.home, "restored"), "--from", dir).assertCode(t, 0)
			assertEqual(t, restored.run(t, "ls", "--ids", "--any-status").assertCode(t, 0).stdout,
				"MYP-1\nMYP-3\nMYP-4\n", "the tasks of the restored board")

			asJSON := m.run(t, "snapshot", "--vcs", "none", "--json").assertCode(t, 6)
			if skipped := skippedOf(t, asJSON.stdout); len(skipped) != 1 || skipped[0] != "MYP-2" {
				t.Errorf("data.skipped is %v and not [MYP-2]", skipped)
			}
		})
	}
}

// TestSnapshotWithAVersionControlSystemStillRecordsTheRevision is the row that
// says the two files are written and the revision is saved with the tasks that
// were written, and not with all of them.
func TestSnapshotWithAVersionControlSystemStillRecordsTheRevision(t *testing.T) {
	m, dir := snapshotBoard(t)
	m.execOnBoard(t, dir, `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "snapshot").assertCode(t, 6)

	if !strings.Contains(got.stdout, "Committed") {
		t.Errorf("no revision was recorded:\n%s", got.stdout)
	}
	if message := revisionMessage(t, dir); message != "biso snapshot: 2 tasks" {
		t.Errorf("the revision message is %q and not the count of the tasks it holds", message)
	}
}

func revisionMessage(t *testing.T, dir string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "-1", "--format=%s")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git log in %s: %v", dir, err)
	}
	return strings.TrimSpace(string(out))
}

func TestPrimeLeavesTheUnreadableTaskOutOfTheMessage(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			got := m.run(t, "prime").assertCode(t, 0)

			assertEqual(t, got.stderr, "", "prime writes nothing to stderr")
			if !strings.Contains(got.stdout, "  unreadable  1 task could not be read and was skipped\n") {
				t.Errorf("the unreadable line is missing:\n%s", got.stdout)
			}
			if strings.Contains(got.stdout, "MYP-2") {
				t.Errorf("the unreadable task is in the message:\n%s", got.stdout)
			}
			if !strings.Contains(got.stdout, "  To Do 3 | In Progress 0 | Done 0\n") {
				t.Errorf("the counts line counts the unreadable task:\n%s", got.stdout)
			}
			assertWithinBudget(t, got.stdout)

			asJSON := m.run(t, "prime", "--json").assertCode(t, 0)
			if skipped := skippedOf(t, asJSON.stdout); len(skipped) != 1 || skipped[0] != "MYP-2" {
				t.Errorf("data.skipped is %v and not [MYP-2]", skipped)
			}
			counts := decode(t, asJSON.stdout)["data"].(map[string]any)["board"].(map[string]any)["countByStatus"].(map[string]any)
			if counts["To Do"] != float64(3) {
				t.Errorf("the count of To Do is %v and not 3", counts["To Do"])
			}
		})
	}
}

func TestDoctorReportsTheUnreadableTaskWithTheCodeOfItsKind(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			got := m.run(t, "doctor", "--json").assertCode(t, 6)

			wantCode := "task_unreadable"
			if c.vocabulary {
				wantCode = "value_not_configured"
			}
			var found map[string]any
			for _, p := range decode(t, got.stdout)["data"].(map[string]any)["problems"].([]any) {
				problem := p.(map[string]any)
				if problem["task"] == "MYP-2" {
					if found != nil {
						t.Errorf("MYP-2 is reported twice: %v and %v", found, problem)
					}
					found = problem
				}
			}
			if found == nil {
				t.Fatalf("MYP-2 is not reported:\n%s", got.stdout)
			}
			if found["code"] != wantCode {
				t.Errorf("the code is %v and not %s", found["code"], wantCode)
			}
			message, _ := found["message"].(string)
			if c.vocabulary {
				if !strings.HasPrefix(message, c.field+` "`+c.given+`" is not one of the configured `) {
					t.Errorf("the message is %q", message)
				}
			} else if want := `task "MYP-2" could not be parsed: `; !strings.HasPrefix(message, want) ||
				!strings.Contains(message, c.reason) {
				t.Errorf("the message is %q and not one that starts with %q and says %q", message, want, c.reason)
			}
		})
	}
}

// TestAWriteAimedAtTheUnreadableTaskFailsAndWritesNothing is the row of the
// table about targeted writes, for every verb of the cycle that cannot
// repair the fault it is aimed at.
func TestAWriteAimedAtTheUnreadableTaskFailsAndWritesNothing(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)
			before := m.dumpOfTheBoard(t)

			for _, argv := range [][]string{
				{"set", "MYP-2", "--title", "Another title"},
				{"set", "MYP-2", "--title", "Another title", "--dry-run"},
				{"note", "MYP-2", "a note"},
				{"comment", "MYP-2", "another comment"},
				{"archive", "MYP-2"},
				{"ask", "MYP-2", "a question?"},
				{"answer", "MYP-2", "an answer"},
			} {
				got := m.run(t, argv...)
				if got.code != 3 {
					t.Errorf("biso %s exited %d and not 3\n%s", strings.Join(argv, " "), got.code, got.stderr)
				}
				if !strings.Contains(got.stderr, "MYP-2 cannot be read: ") {
					t.Errorf("biso %s does not give the reason: %q", strings.Join(argv, " "), got.stderr)
				}
			}
			assertEqual(t, m.dumpOfTheBoard(t), before, "the database after the refused writes")
		})
	}
}

func (m *machine) dumpOfTheBoard(t *testing.T) string {
	t.Helper()
	return dumpDatabase(t, m.boardDir(t), taskTables...)
}

func TestAWriteThatLeavesTheVocabularyValueRepairedIsApplied(t *testing.T) {
	cases := []struct {
		name  string
		sql   string
		argv  []string
		check string
		want  string
	}{
		{"set the priority", `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`,
			[]string{"set", "MYP-2", "--priority", "medium"},
			`SELECT priority FROM task WHERE id = 'MYP-2'`, "medium"},
		{"clear the priority", `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`,
			[]string{"set", "MYP-2", "--clear-priority"},
			`SELECT priority FROM task WHERE id = 'MYP-2'`, ""},
		{"set the type", `UPDATE task SET type = 'epic' WHERE id = 'MYP-2'`,
			[]string{"set", "MYP-2", "--type", "bug"},
			`SELECT type FROM task WHERE id = 'MYP-2'`, "bug"},
		{"set the status", `UPDATE task SET status = 'Weird' WHERE id = 'MYP-2'`,
			[]string{"set", "MYP-2", "--status", "To Do"},
			`SELECT status FROM task WHERE id = 'MYP-2'`, "To Do"},
		{"set the status of an empty one", `UPDATE task SET status = '' WHERE id = 'MYP-2'`,
			[]string{"set", "MYP-2", "--status", "To Do"},
			`SELECT status FROM task WHERE id = 'MYP-2'`, "To Do"},
		{"start it", `UPDATE task SET status = 'Weird' WHERE id = 'MYP-2'`,
			[]string{"start", "MYP-2"},
			`SELECT status FROM task WHERE id = 'MYP-2'`, "In Progress"},
		{"finish it", `UPDATE task SET status = 'Weird' WHERE id = 'MYP-2'`,
			[]string{"finish", "MYP-2"},
			`SELECT status FROM task WHERE id = 'MYP-2'`, "Done"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := smallBoard(t)
			m.execOnBoard(t, m.boardDir(t), c.sql)
			before := m.dumpOfTheBoard(t)

			dry := m.run(t, append(append([]string{}, c.argv...), "--dry-run")...)
			if dry.code != 0 {
				t.Fatalf("the preview exited %d:\n%s", dry.code, dry.stderr)
			}
			assertEqual(t, m.dumpOfTheBoard(t), before, "the database after the preview")

			m.run(t, c.argv...).assertCode(t, 0)

			assertEqual(t, m.scalarOnBoard(t, c.check), c.want, "the repaired value")
			assertEqual(t, m.run(t, "ls", "--ids", "--any-status", "--sort", "id").assertCode(t, 0).stdout,
				"MYP-1\nMYP-2\nMYP-3\nMYP-4\n", "the tasks once it is repaired")
			m.run(t, "doctor").assertCode(t, 0)
		})
	}
}

// TestAWriteThatDoesNotRepairTheVocabularyValueFailsEvenWithADryRun is the
// other half: the preview answers what the real call answers.
func TestAWriteThatDoesNotRepairTheVocabularyValueFailsEvenWithADryRun(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	for _, argv := range [][]string{
		{"set", "MYP-2", "--title", "Other"},
		{"set", "MYP-2", "--status", "Done"},
		{"set", "MYP-2", "--type", "bug"},
		{"start", "MYP-2"},
		{"finish", "MYP-2"},
	} {
		for _, dry := range []bool{false, true} {
			if dry {
				argv = append(append([]string{}, argv...), "--dry-run")
			}
			got := m.run(t, argv...)
			if got.code != 3 || !strings.Contains(got.stderr, `MYP-2 cannot be read: its priority is "urgent"`) {
				t.Errorf("biso %s exited %d: %s", strings.Join(argv, " "), got.code, got.stderr)
			}
		}
	}
	assertEqual(t, m.scalarOnBoard(t, `SELECT priority FROM task WHERE id = 'MYP-2'`), "urgent",
		"the priority after every refusal")
}

// TestAWriteAimedAtAnotherTaskWorksAndNamesTheUnreadableOne is the row about
// the other tasks: only the task the call names decides.
func TestAWriteAimedAtAnotherTaskWorksAndNamesTheUnreadableOne(t *testing.T) {
	for _, c := range everyCorruption {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			m := corrupted(t, c)

			got := m.run(t, "set", "MYP-1", "--title", "Renamed").assertCode(t, 0)

			assertEqual(t, got.stderr, warningOfMYP2, "the warning of a write to another task")
			assertEqual(t, m.scalarOnBoard(t, `SELECT title FROM task WHERE id = 'MYP-1'`), "Renamed", "the title")
		})
	}
}

// TestRepairingTheTaskDoesNotWarnAboutItAsIfItWereStillBroken is a corner of
// the same row: the warning names the tasks the call left alone.
func TestRepairingTheTaskDoesNotWarnAboutItAsIfItWereStillBroken(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "set", "MYP-2", "--priority", "medium").assertCode(t, 0)

	assertEqual(t, got.stderr, "", "the standard error of a repair")
}

// TestADateThatIsNotADateHasNoRepairThroughAWrite is the row that says the
// fault is not one a command can fix: the row does not even load.
func TestADateThatIsNotADateHasNoRepairThroughAWrite(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET created_at = 'nope' WHERE id = 'MYP-2'`)

	got := m.run(t, "set", "MYP-2", "--due", "2026-01-01").assertCode(t, 3)

	if !strings.Contains(got.stderr, "MYP-2 cannot be read: createdAt is not an instant") {
		t.Errorf("the reason is %q", got.stderr)
	}
}

// TestDeclaringTheValueMakesTheTaskReadable is the second remedy of the page:
// what was wrong was the configuration, and the task was right.
func TestDeclaringTheValueMakesTheTaskReadable(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)
	m.run(t, "get", "MYP-2").assertCode(t, 3)

	m.run(t, "config", "set", "priorities", "high,medium,low,urgent").assertCode(t, 0)

	m.run(t, "get", "MYP-2", "--section", "meta").assertCode(t, 0)
	assertEqual(t, m.run(t, "ls", "--ids", "--sort", "id").assertCode(t, 0).stdout,
		"MYP-1\nMYP-2\nMYP-3\nMYP-4\n", "the listing once the value is declared")
}

func TestABoardWithOnlyAnUnreadableTaskIsNotAnEmptyBoard(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "Only one").assertCode(t, 0)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET status = 'Weird' WHERE id = 'MYP-1'`)

	ls := m.run(t, "ls").assertCode(t, 0)
	assertEqual(t, ls.stdout, "", "the rows of a board with one unreadable task")
	assertEqual(t, ls.stderr, strings.Replace(warningOfMYP2, "MYP-2", "MYP-1", 1)+"note: no tasks match\n",
		"the note and the warning")

	prime := m.run(t, "prime").assertCode(t, 0)
	if strings.Contains(prime.stdout, "THE BOARD IS EMPTY") {
		t.Errorf("prime says the board is empty:\n%s", prime.stdout)
	}
	if !strings.Contains(prime.stdout, "  unreadable  1 task could not be read and was skipped\n") ||
		!strings.Contains(prime.stdout, "  To Do 0 | In Progress 0 | Done 0\n") {
		t.Errorf("prime does not describe a board with one skipped task:\n%s", prime.stdout)
	}
}

func TestSeveralUnreadableTasksAreNamedOnceAndInNumericOrder(t *testing.T) {
	m := newMachine(t)
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	for i := 0; i < 11; i++ {
		m.run(t, "new", "Task").assertCode(t, 0)
	}
	// One of them for its vocabulary and the other for a date, so that the
	// list is put together from two kinds and not sorted by text.
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-10'`)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET created_at = 'nope' WHERE id = 'MYP-9'`)

	got := m.run(t, "ls", "--count").assertCode(t, 0)

	assertEqual(t, got.stdout, "9\n", "the count")
	assertEqual(t, got.stderr, "warning: 2 tasks could not be read and were skipped: MYP-9, MYP-10\n",
		"the warning of two tasks of different kinds")
}

// TestAnUnreadableTaskThatIsFinishedAndArchivedIsStillNamed is the last of the
// cases the warning covers: `biso ls` hides finished and archived tasks by
// default, and a task it cannot read is not hidden by that.
func TestAnUnreadableTaskThatIsFinishedAndArchivedIsStillNamed(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t),
		`UPDATE task SET status = 'Done', archived = 1, priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "ls", "--ids").assertCode(t, 0)

	assertEqual(t, got.stdout, "MYP-1\nMYP-3\nMYP-4\n", "the rows")
	assertEqual(t, got.stderr, warningOfMYP2, "the warning")

	prime := m.run(t, "prime").assertCode(t, 0)
	if !strings.Contains(prime.stdout, "  unreadable  1 task could not be read and was skipped\n") {
		t.Errorf("prime does not count the archived task that cannot be read:\n%s", prime.stdout)
	}
}

func TestATaskThatDependsOnAnUnreadableOneIsNotBlockedByIt(t *testing.T) {
	m := smallBoard(t)
	m.run(t, "set", "MYP-1", "--add-deps", "MYP-2").assertCode(t, 0)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`)

	got := m.run(t, "get", "MYP-1", "--section", "meta", "--json").assertCode(t, 0)

	task := decode(t, got.stdout)["data"].(map[string]any)["task"].(map[string]any)
	if task["blocked"] != false {
		t.Errorf("MYP-1 is blocked by a task that cannot be read: %v", task["blocked"])
	}
	if !strings.Contains(got.stderr, "MYP-2") {
		t.Errorf("the warning does not name the task: %q", got.stderr)
	}
}

// TestARowOfTheWrongTypeNeverEndsTheReadWithExitCodeOne is the case that
// aborted every reading command before there was one definition: the task
// is skipped and the exit code is not the one of an unexpected failure.
func TestARowOfTheWrongTypeNeverEndsTheReadWithExitCodeOne(t *testing.T) {
	m := smallBoard(t)
	m.execOnBoard(t, m.boardDir(t), `UPDATE task SET archived = 7 WHERE id = 'MYP-2'`)

	for _, argv := range [][]string{
		{"ls"}, {"prime"}, {"get", "MYP-1"}, {"export"}, {"snapshot", "--vcs", "none"}, {"doctor"},
	} {
		got := m.run(t, argv...)
		if got.code == 1 {
			t.Errorf("biso %s ended with exit code 1:\n%s", strings.Join(argv, " "), got.stderr)
		}
	}
}

// TestTheDefaultExampleBoardOfPrimeStaysByteIdentical is the reminder that the
// one definition does not touch the startup message of a healthy board.
func TestTheStartupMessageOfAHealthyBoardDoesNotDependOnTheCheck(t *testing.T) {
	m, _ := primeBoard(t)

	got := m.run(t, "prime").assertCode(t, 0)

	assertWithinBudget(t, got.stdout)
	if strings.Contains(got.stdout, "unreadable") {
		t.Errorf("a healthy board mentions unreadable tasks:\n%s", got.stdout)
	}
}
