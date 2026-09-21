package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/board"
)

// These tests walk, through the compiled program, the rules of
// docs/spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote: what `new
// --from` prints and answers when a list of a line holds an element that is
// empty or only spaces, what `init --from` does with the same line, and the
// guarantee of docs/spec/cmd/export.md that a board never holds such an
// element. The rule itself, list by list, is walked by
// internal/ops/batch_empty_items_test.go.

// emptyItemsWarning is one entry of data.warnings of the batch envelope for
// the warning of a dropped element, with the `field` the other two helper
// types of this package do not carry.
type emptyItemsWarning struct {
	Code  string `json:"code"`
	Line  int    `json:"line"`
	Field string `json:"field"`
	Count int    `json:"count"`
}

func TestBatchWarnsOfTheEmptyItemsWithTheLiteralOfTheSpecification(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, fixture(t, "new-batch-empty-items-input.txt"))

	got := m.run(t, "new", "--from", path).assertCode(t, 0)

	assertEqual(t, got.stderr, fixture(t, "new-batch-empty-items.txt"),
		"the warnings of a line with empty elements")
	assertEqual(t, got.stdout, "MYP-1\n", "the identifier the line created")
	// The task keeps the label and the reference that were not empty and
	// nothing else.
	shown := m.run(t, "get", "MYP-1").assertCode(t, 0).stdout
	if !strings.Contains(shown, "parser") || !strings.Contains(shown, "docs/bugs/BUG-02.md") {
		t.Errorf("the task lost what was not empty:\n%s", shown)
	}
	for _, line := range strings.Split(shown, "\n") {
		if strings.HasPrefix(line, "refs ") && strings.TrimSpace(strings.TrimPrefix(line, "refs")) != "docs/bugs/BUG-02.md" {
			t.Errorf("the references line is %q", line)
		}
	}
}

func TestBatchPreviewWarnsOfTheEmptyItemsLikeTheRealCall(t *testing.T) {
	m := batchBoard(t)
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, fixture(t, "new-batch-empty-items-input.txt"))

	got := m.run(t, "new", "--from", path, "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stderr,
		fixture(t, "new-batch-empty-items.txt")+
			"1 task would be created, nothing was written (--dry-run)\n",
		"the warnings of a preview")
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a preview left behind")
}

func TestBatchJSONCarriesTheEmptyItemsWarningWithItsFields(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"Clean"}`,
		`{"title":"Holes","labels":["parser",""],"references":["","docs/a.md","   "]}`,
		`{"title":"One hole","acceptanceCriteria":[""],"documentation":["","d1"]}`)

	// The raw envelope, to see the keys of the object and not only the ones
	// a struct of this test reads.
	got := m.run(t, "new", "--from", path, "--json").assertCode(t, 0)
	var envelope struct {
		Data struct {
			Warnings []map[string]any `json:"warnings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the envelope is not JSON: %v\n%s", err, got.stdout)
	}
	first := envelope.Data.Warnings[0]
	if len(first) != 4 || first["code"] != "imported_empty_dropped" || first["line"] != float64(2) ||
		first["field"] != "labels" || first["count"] != float64(1) {
		t.Errorf("the first warning object = %v", first)
	}

	var typed struct {
		Data struct {
			Warnings []emptyItemsWarning `json:"warnings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &typed); err != nil {
		t.Fatal(err)
	}
	want := []emptyItemsWarning{
		{"imported_empty_dropped", 2, "labels", 1},
		{"imported_empty_dropped", 2, "references", 2},
		{"imported_empty_dropped", 3, "acceptanceCriteria", 1},
		{"imported_empty_dropped", 3, "documentation", 1},
		// The line 3 merged one documentation item, which is the merge
		// warning that follows the ones of what was dropped.
		{"imported_documentation_merged", 3, "", 0},
	}
	if len(typed.Data.Warnings) != len(want) {
		t.Fatalf("warnings = %+v, want %+v", typed.Data.Warnings, want)
	}
	for i := range want {
		if want[i].Code == "imported_documentation_merged" {
			if typed.Data.Warnings[i].Code != want[i].Code || typed.Data.Warnings[i].Line != want[i].Line {
				t.Errorf("warning %d = %+v", i, typed.Data.Warnings[i])
			}
			continue
		}
		if typed.Data.Warnings[i] != want[i] {
			t.Errorf("warning %d = %+v, want %+v", i, typed.Data.Warnings[i], want[i])
		}
	}
	// The text for a person is still on stderr in the JSON mode.
	if !strings.Contains(got.stderr, "warning: line 2: 1 empty item dropped from labels\n") {
		t.Errorf("stderr = %q", got.stderr)
	}
}

func TestBatchWithAnInvalidLineWarnsOfNoEmptyItem(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"Holes","references":["","a"],"labels":[""]}`,
		`{"title":"Bad","references":["x",null]}`)

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	assertEqual(t, got.stderr,
		"error: 1 of 2 lines is invalid, nothing was written\n"+
			"  line 2: references.1: expected text, got null\n",
		"the failure of a batch with a null in the place of an element")
	if strings.Contains(got.stderr, "warning") {
		t.Errorf("an invalid batch warned:\n%s", got.stderr)
	}
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks an invalid batch left behind")

	asJSON := m.run(t, "new", "--from", path, "--json").assertCode(t, 7)
	if strings.Contains(asJSON.stdout+asJSON.stderr, "imported_empty_dropped") ||
		strings.Contains(asJSON.stderr, "warning") {
		t.Errorf("an invalid batch in JSON mode warned:\nstdout:\n%s\nstderr:\n%s", asJSON.stdout, asJSON.stderr)
	}
}

func TestBatchNullInThePlaceOfAnElementIsTheMessageOfTheSpecification(t *testing.T) {
	m := batchBoard(t)
	lines := make([]string, 0, 7)
	for i := 1; i <= 6; i++ {
		lines = append(lines, `{"title":"Fine"}`)
	}
	lines = append(lines, `{"title":"Bad","references":["docs/a.md",null]}`)
	path := writeBatch(t, m, lines...)

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	if !strings.Contains(got.stderr, "  "+fixture(t, "new-batch-null-element.txt")) {
		t.Errorf("the failure is not the block of the specification:\n%s", got.stderr)
	}
	// It is an invalid_line of exit code 3 inside the exit code 7.
	asJSON := m.run(t, "new", "--from", path, "--json").assertCode(t, 7)
	var envelope struct {
		Error struct {
			Details []struct {
				ExitCode int    `json:"exitCode"`
				Code     string `json:"code"`
				Message  string `json:"message"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal([]byte(asJSON.stderr), &envelope); err != nil {
		t.Fatalf("the error envelope is not JSON: %v\nstdout:\n%s\nstderr:\n%s", err, asJSON.stdout, asJSON.stderr)
	}
	if d := envelope.Error.Details; len(d) != 1 || d[0].Code != "invalid_line" || d[0].ExitCode != 3 ||
		d[0].Message != "line 7: references.1: expected text, got null" {
		t.Errorf("the details = %+v", d)
	}
}

func TestBatchPrintWithAnEmptyItemsBatchStillFailsWithNoChange(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m, `{"title":"Holes","references":["","a"]}`)

	m.run(t, "new", "--from", path, "--print").assertCode(t, 2)

	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a refused call left behind")
}

// restoredSnapshot builds a snapshot whose tasks file is the one given, over
// the real files of a board so that everything else about it is valid.
func restoredSnapshot(t *testing.T, tasks string) (*machine, string) {
	t.Helper()
	m, dir := snapshotDir(t)
	edited := filepath.Join(m.home, "edited-snapshot")
	copySnapshot(t, m, dir, edited)
	m.write(t, filepath.Join(edited, board.SnapshotTasksFile), tasks)
	return m, edited
}

func TestRestoreDropsEmptyItemsWithoutWarning(t *testing.T) {
	m, edited := restoredSnapshot(t,
		`{"title":"Holes","labels":["","parser"],"references":["","a"," "],`+
			`"acceptanceCriteria":["","x"],"documentation":["","d"],`+
			`"modifiedFiles":[" "],"definitionOfDone":["", "dod"]}`+"\n")
	into := filepath.Join(m.home, "restored-holes")

	got := m.restoreInto(t, edited, into).assertCode(t, 0)

	if strings.Contains(got.stderr, "warning") || strings.Contains(got.stderr, "imported_") {
		t.Errorf("a restore warned:\n%s", got.stderr)
	}
	restored := m.at(into)
	shown := restored.run(t, "get", "MYP-1").assertCode(t, 0).stdout
	for _, line := range strings.Split(shown, "\n") {
		if strings.HasPrefix(line, "refs ") {
			assertEqual(t, strings.TrimSpace(strings.TrimPrefix(line, "refs")), "a, d",
				"the references of the restored task")
		}
	}
	// The criteria are two, `x` and the converted `dod`, with no gap.
	exported := restored.run(t, "export").assertCode(t, 0).stdout
	var task struct {
		Labels             []string `json:"labels"`
		References         []string `json:"references"`
		AcceptanceCriteria []struct {
			Key  int    `json:"key"`
			Text string `json:"text"`
		} `json:"acceptanceCriteria"`
	}
	if err := json.Unmarshal([]byte(strings.TrimSpace(exported)), &task); err != nil {
		t.Fatal(err)
	}
	if strings.Join(task.Labels, "|") != "parser" || strings.Join(task.References, "|") != "a|d" ||
		len(task.AcceptanceCriteria) != 2 || task.AcceptanceCriteria[0].Key != 1 ||
		task.AcceptanceCriteria[0].Text != "x" || task.AcceptanceCriteria[1].Key != 2 ||
		task.AcceptanceCriteria[1].Text != "dod" {
		t.Errorf("the restored task = %+v", task)
	}
	restored.run(t, "doctor").assertCode(t, 0)
}

func TestRestorePreviewOfEmptyItemsWarnsOfNothingEither(t *testing.T) {
	m, edited := restoredSnapshot(t, `{"title":"Holes","references":["","a"]}`+"\n")
	into := filepath.Join(m.home, "previewed-holes")

	got := m.restoreInto(t, edited, into, "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stderr,
		"1 task would be created, nothing was written (--dry-run)\n",
		"the preview of a restore with empty elements")
	if _, err := os.Stat(into); err == nil {
		t.Errorf("a preview created %s", into)
	}
}

func TestRestoreRefusesANullInThePlaceOfAnElement(t *testing.T) {
	m, edited := restoredSnapshot(t,
		`{"title":"Fine"}`+"\n"+`{"title":"Bad","references":[null]}`+"\n")
	into := filepath.Join(m.home, "not-restored-null")

	got := m.restoreInto(t, edited, into).assertCode(t, 7)

	if !strings.Contains(got.stderr, "line 2: references.0: expected text, got null") {
		t.Errorf("the failure does not name the element:\n%s", got.stderr)
	}
	if _, err := os.Stat(into); err == nil {
		t.Errorf("a failed restore left %s behind", into)
	}
}

// assertNoEmptyElement walks every list of every line of an export and fails
// on an element that is empty or only spaces: a text, or the text of a
// criterion, which is the one object whose text is what must not be empty.
func assertNoEmptyElement(t *testing.T, export string) {
	t.Helper()
	for _, line := range strings.Split(strings.TrimSpace(export), "\n") {
		var task map[string]any
		if err := json.Unmarshal([]byte(line), &task); err != nil {
			t.Fatal(err)
		}
		for key, value := range task {
			list, ok := value.([]any)
			if !ok {
				continue
			}
			for i, element := range list {
				var text string
				switch e := element.(type) {
				case string:
					text = e
				case map[string]any:
					if key != "acceptanceCriteria" {
						continue
					}
					text, _ = e["text"].(string)
				default:
					continue
				}
				if strings.TrimSpace(text) == "" {
					t.Errorf("%s.%d of the export is empty: %s", key, i, line)
				}
			}
		}
	}
}

const (
	// withHoles has an empty or blank element in every list of the format, and
	// withoutHoles is the same lines with them taken out. Both carry their own
	// dates so that two imports of them differ in nothing the clock decides.
	withHoles = `{"id":"MYP-1","title":"One","createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-14T10:20:00Z",` +
		`"assignees":["","@sara"],"labels":[" ","parser"],"references":["","docs/a.md","  "],` +
		`"acceptanceCriteria":["",{"key":9,"text":" ","checked":true},{"key":2,"text":"kept","checked":true},"plain"],` +
		`"definitionOfDone":["","dod one"],"documentation":["","doc one"],"modifiedFiles":["   ","file one"]}` + "\n" +
		`{"id":"MYP-2","title":"Two","createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-14T10:20:00Z",` +
		`"dependencies":["","MYP-1"],"references":[""]}` + "\n"
	withoutHoles = `{"id":"MYP-1","title":"One","createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-14T10:20:00Z",` +
		`"assignees":["@sara"],"labels":["parser"],"references":["docs/a.md"],` +
		`"acceptanceCriteria":[{"key":2,"text":"kept","checked":true},"plain"],` +
		`"definitionOfDone":["dod one"],"documentation":["doc one"],"modifiedFiles":["file one"]}` + "\n" +
		`{"id":"MYP-2","title":"Two","createdAt":"2026-08-14T10:20:00Z","updatedAt":"2026-08-14T10:20:00Z",` +
		`"dependencies":["MYP-1"]}` + "\n"
)

func importInto(t *testing.T, name, content string) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, append([]string{"init", name}, vocabulary...)...).assertCode(t, 0)
	path := filepath.Join(m.dir, "in.ndjson")
	m.write(t, path, content)
	m.run(t, "new", "--from", path).assertCode(t, 0)
	return m
}

// TestABatchWithEmptyItemsLeavesTheBoardOfTheSameBatchWithoutThem is what
// "dropped" means: the board that comes out is the one the same lines without
// the empty elements give, in every table and every column.
func TestABatchWithEmptyItemsLeavesTheBoardOfTheSameBatchWithoutThem(t *testing.T) {
	holes := importInto(t, "Holes", withHoles)
	clean := importInto(t, "Clean", withoutHoles)

	assertEqual(t,
		dumpDatabase(t, holes.boardDir(t), taskTables...),
		dumpDatabase(t, clean.boardDir(t), taskTables...),
		"the tasks of the board a batch with empty elements produced")
}

// TestAnImportedBatchWithEmptyItemsExportsAndReimportsIdentically is the
// symmetry of docs/spec/cmd/export.md for the case the export could break:
// an element that the import dropped is not in what is exported, and
// importing that export again changes nothing.
func TestAnImportedBatchWithEmptyItemsExportsAndReimportsIdentically(t *testing.T) {
	source := importInto(t, "Holes", withHoles)
	dump := source.run(t, "export").assertCode(t, 0).stdout
	assertNoEmptyElement(t, dump)

	destination := importInto(t, "Reimported", dump)

	assertEqual(t,
		dumpDatabase(t, destination.boardDir(t), taskTables...),
		dumpDatabase(t, source.boardDir(t), taskTables...),
		"the tasks of the board the export of a batch with empty elements gave")
	assertEqual(t, destination.run(t, "export").assertCode(t, 0).stdout, dump,
		"the export of the reimported board")
}

func TestTheExportOfTheRichBoardHasNoEmptyElement(t *testing.T) {
	assertNoEmptyElement(t, sourceBoard(t).run(t, "export").assertCode(t, 0).stdout)
}

// TestAFlagThatAddsStillRefusesWhatABatchDrops is the other half of the same
// rule: `--add-refs ""` warns and stores nothing, so the two ways in agree.
func TestAFlagThatAddsStillRefusesWhatABatchDrops(t *testing.T) {
	m := batchBoard(t)
	m.run(t, "new", "Task", "--add-refs", "a").assertCode(t, 0)

	got := m.run(t, "set", "MYP-1", "--add-refs", "", "--add-labels", "x").assertCode(t, 0)

	if !strings.Contains(got.stderr, "warning: --add-refs: empty value, nothing was added") {
		t.Errorf("stderr = %q", got.stderr)
	}
	assertNoEmptyElement(t, m.run(t, "export").assertCode(t, 0).stdout)
}
