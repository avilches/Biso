package ops

import (
	"encoding/json"
	"strings"
	"testing"
)

// These are the rules of docs/spec/cmd/new.md#el-modo-lote, one test per
// rule. They run over the board directly and not through the compiled
// program, because what they check is a rule of the format and not a line
// of output: the literal text of the batch lives in the golden tests of
// cmd/biso.

// batch runs `biso new --from` over the harness's board with the lines
// given, each one already a JSON object.
func (h *harness) batch(lines ...string) (*WriteResult, error) {
	h.t.Helper()
	return NewBatchOn(h.b, h.env, BatchParams{Content: strings.Join(lines, "\n") + "\n"})
}

// assertBatchFails runs a batch that must not pass validation and answers
// the one message its single failure carries.
func (h *harness) assertBatchFails(lines ...string) string {
	h.t.Helper()
	_, err := h.batch(lines...)
	e := specError(h.t, err)
	if e.ExitCode != 7 || e.Code != "batch_invalid" {
		h.t.Fatalf("error = %d/%s (%s), want 7/batch_invalid", e.ExitCode, e.Code, e.Message)
	}
	if len(e.Details) != 1 {
		h.t.Fatalf("the batch reported %d failures and one was expected: %v", len(e.Details), e.Detail)
	}
	return e.Details[0].Message
}

func TestBatchIgnoresBlankAndCommentedLines(t *testing.T) {
	h := newHarness(t)
	result, err := NewBatchOn(h.b, h.env, BatchParams{
		Content: "# a header\n\n{\"title\":\"One\"}\n\n   \n# and a tail\n",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 1 {
		t.Errorf("the batch created %d tasks and the file had one", len(result.Tasks))
	}
}

func TestBatchAcceptsACriterionAsAStringOrAsAnObject(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(
		`{"title":"Mixed","acceptanceCriteria":["plain one",{"key":5,"text":"keyed","checked":true},"another"]}`,
	); err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	want := []struct {
		key     int
		text    string
		checked bool
	}{
		// The plain string takes the next free key of the counter, which
		// starts above the highest key the line brought.
		{6, "plain one", false},
		{5, "keyed", true},
		{7, "another", false},
	}
	if len(task.AcceptanceCriteria) != len(want) {
		t.Fatalf("criteria = %v", task.AcceptanceCriteria)
	}
	for i, c := range want {
		got := task.AcceptanceCriteria[i]
		if got.Key != c.key || got.Text != c.text || got.Checked != c.checked {
			t.Errorf("criterion %d = %+v, want %+v", i, got, c)
		}
	}
	if task.NextCriterionKey != 8 {
		t.Errorf("the counter is at %d, and it sits above the highest key", task.NextCriterionKey)
	}
}

func TestBatchRefusesARepeatedCriterionKey(t *testing.T) {
	h := newHarness(t)
	message := h.assertBatchFails(
		`{"title":"Twice","acceptanceCriteria":[{"key":2,"text":"a"},{"key":2,"text":"b"}]}`)
	if !strings.Contains(message, "key 2 appears twice") {
		t.Errorf("message = %q", message)
	}
}

func TestBatchTurnsDefinitionOfDoneIntoAcceptanceCriteria(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Normalize CRLF","acceptanceCriteria":[{"key":1,"text":"The diff ignores CRLF","checked":true}],` +
			`"definitionOfDone":[{"key":1,"text":"Reviewed","checked":false}]}`)
	if err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	if len(task.AcceptanceCriteria) != 2 {
		t.Fatalf("criteria = %v", task.AcceptanceCriteria)
	}
	// The key of the converted element is the next free one of the counter
	// and never the one it carried, which was already taken.
	if got := task.AcceptanceCriteria[1]; got.Key != 2 || got.Text != "Reviewed" || got.Checked {
		t.Errorf("the converted criterion = %+v", got)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Code != "imported_dod_merged" {
		t.Fatalf("warnings = %v, want one imported_dod_merged", result.Warnings)
	}
	if got := result.Warnings[0].Message; got != "line 1: 1 definition-of-done item imported as an acceptance criterion" {
		t.Errorf("the warning says %q", got)
	}
	// The keys the conversion created are the one thing `biso new`
	// announces about a key it assigned.
	if len(result.Tasks[0].AcAdded) != 1 || result.Tasks[0].AcAdded[0] != 2 {
		t.Errorf("acAdded = %v, want the key the conversion created", result.Tasks[0].AcAdded)
	}
}

func TestBatchDoesNotWarnForAnEmptyOrAbsentDefinitionOfDone(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Empty","definitionOfDone":[]}`,
		`{"title":"Null","definitionOfDone":null}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and nothing was converted", result.Warnings)
	}
}

func TestBatchAllowsARepeatedKeyInsideDefinitionOfDone(t *testing.T) {
	h := newHarness(t)
	// The two lists had counters of their own, so the same key twice inside
	// definitionOfDone says nothing: both are discarded without being read.
	if _, err := h.batch(
		`{"title":"Two of them","definitionOfDone":[{"key":1,"text":"a"},{"key":1,"text":"b"}]}`,
	); err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	if len(task.AcceptanceCriteria) != 2 ||
		task.AcceptanceCriteria[0].Key == task.AcceptanceCriteria[1].Key {
		t.Errorf("criteria = %v", task.AcceptanceCriteria)
	}
}

func TestBatchMergesDocumentationIntoReferences(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Two pointers","references":["docs/a.md","notes/b.md"],` +
			`"documentation":["docs/c.md","docs/a.md","https://example.com/d"]}`)
	if err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	// The documentation goes after the references, in the order it came,
	// and a value the references already had is kept once.
	want := []string{"docs/a.md", "notes/b.md", "docs/c.md", "https://example.com/d"}
	if strings.Join(task.References, "|") != strings.Join(want, "|") {
		t.Errorf("references = %v, want %v", task.References, want)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Code != "imported_documentation_merged" {
		t.Fatalf("warnings = %v, want one imported_documentation_merged", result.Warnings)
	}
	if got := result.Warnings[0].Message; got != "line 1: 3 documentation items imported as references" {
		t.Errorf("the warning says %q", got)
	}
	if got := result.Warnings[0].Fields; got["line"] != 1 || got["count"] != 3 {
		t.Errorf("the warning fields = %v", got)
	}
}

func TestBatchMergesOneDocumentationItemWithTheSingularWarning(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"One pointer","documentation":["docs/only.md"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; len(got) != 1 || got[0] != "docs/only.md" {
		t.Errorf("references = %v", got)
	}
	if len(result.Warnings) != 1 ||
		result.Warnings[0].Message != "line 1: 1 documentation item imported as a reference" {
		t.Errorf("warnings = %v", result.Warnings)
	}
}

func TestBatchDoesNotWarnForAnEmptyOrAbsentDocumentation(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Empty","documentation":[]}`,
		`{"title":"Null","documentation":null}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and nothing was merged", result.Warnings)
	}
}

func TestBatchMergesModifiedFilesIntoReferences(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Two files","references":["docs/a.md"],` +
			`"modifiedFiles":["internal/x.go","docs/a.md","internal/y.go"]}`)
	if err != nil {
		t.Fatal(err)
	}
	// The files go after the references, in the order they came, and a value
	// the references already had is kept once.
	want := []string{"docs/a.md", "internal/x.go", "internal/y.go"}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("references = %v, want %v", got, want)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Code != "imported_modified_files_merged" {
		t.Fatalf("warnings = %v, want one imported_modified_files_merged", result.Warnings)
	}
	if got := result.Warnings[0].Message; got != "line 1: 3 modified files imported as references" {
		t.Errorf("the warning says %q", got)
	}
	if got := result.Warnings[0].Fields; got["line"] != 1 || got["count"] != 3 {
		t.Errorf("the warning fields = %v", got)
	}
}

func TestBatchMergesOneModifiedFileWithTheSingularWarning(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"One file","modifiedFiles":["internal/only.go"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; len(got) != 1 || got[0] != "internal/only.go" {
		t.Errorf("references = %v", got)
	}
	if len(result.Warnings) != 1 ||
		result.Warnings[0].Message != "line 1: 1 modified file imported as a reference" {
		t.Errorf("warnings = %v", result.Warnings)
	}
}

func TestBatchDoesNotWarnForAnEmptyOrAbsentModifiedFiles(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Empty","modifiedFiles":[]}`,
		`{"title":"Null","modifiedFiles":null}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and nothing was merged", result.Warnings)
	}
}

func TestBatchMergesDocumentationBeforeModifiedFiles(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Both","references":["r1"],"modifiedFiles":["f1","d1"],` +
			`"documentation":["d1","d2"]}`)
	if err != nil {
		t.Fatal(err)
	}
	// references, then documentation, then modifiedFiles, whatever order the
	// keys were written in; a value seen before is kept once, where it was
	// first seen.
	want := []string{"r1", "d1", "d2", "f1"}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != strings.Join(want, "|") {
		t.Errorf("references = %v, want %v", got, want)
	}
	if len(result.Warnings) != 2 ||
		result.Warnings[0].Code != "imported_documentation_merged" ||
		result.Warnings[1].Code != "imported_modified_files_merged" {
		t.Fatalf("warnings = %v, want documentation first and modified files second", result.Warnings)
	}
	if got := result.Warnings[0].Fields["count"]; got != 2 {
		t.Errorf("the documentation count = %v, want 2", got)
	}
	if got := result.Warnings[1].Fields["count"]; got != 2 {
		t.Errorf("the modified files count = %v, want 2", got)
	}
}

func TestBatchKeepsTheKeysAndTheDatesOfTheComments(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"With comments","comments":[` +
		`{"key":3,"author":"@sara","createdAt":"2026-08-14T10:22:00Z","body":"First"},` +
		`{"author":"@sara","body":"No key and no date"}]}`); err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	if task.Comments[0].Key != 3 || task.Comments[0].CreatedAt.Format("2006") != "2026" {
		t.Errorf("the first comment = %+v", task.Comments[0])
	}
	if task.Comments[1].Key != 4 {
		t.Errorf("the second comment took the key %d and not the next free one", task.Comments[1].Key)
	}
	if !task.Comments[1].CreatedAt.Equal(writeClock) {
		t.Errorf("a comment with no date of its own was stamped %v", task.Comments[1].CreatedAt)
	}
}

func TestBatchRefusesAnIdentifierThatIsTakenOrOfAnotherBoard(t *testing.T) {
	h := newHarness(t)
	h.create("The first one")

	if got := h.assertBatchFails(`{"id":"MYP-1","title":"Again"}`); !strings.Contains(got, "already taken") {
		t.Errorf("message = %q", got)
	}
	if got := h.assertBatchFails(`{"id":"OTHER-5","title":"Elsewhere"}`); got !=
		`line 1: id "OTHER-5" does not match this board's task prefix "MYP"` {
		t.Errorf("message = %q", got)
	}
	if got := h.assertBatchFails(
		`{"id":"MYP-5","title":"a"}`, `{"id":"MYP-5","title":"b"}`,
	); got != `line 2: id "MYP-5" is already taken by line 1 of this file` {
		// The board has nothing to do with this collision, so the message
		// does not name it (docs/spec/cmd/new.md#el-modo-lote).
		t.Errorf("the same identifier twice in one file: %q", got)
	}
	if got := h.assertBatchFails(`{"id":"MYP-1","title":"Again"}`); got !=
		`line 1: id "MYP-1" is already taken on this board` {
		t.Errorf("an identifier the board already has: %q", got)
	}
}

func TestBatchReservesAnExplicitIdentifierSoItIsNeverHandedOutAgain(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"id":"MYP-40","title":"Far ahead"}`); err != nil {
		t.Fatal(err)
	}
	if got := h.create("The next one"); got != "MYP-41" {
		t.Errorf("the next identifier was %s, and the batch had reserved up to MYP-40", got)
	}
}

func TestBatchRefusesAnUnknownKeyAndADerivedField(t *testing.T) {
	h := newHarness(t)
	for _, line := range []string{
		`{"title":"a","trelloCard":"5f2a8c1e"}`,
		`{"title":"a","urgency":19.0}`,
		`{"title":"a","acDone":1}`,
		`{"title":"a","leaseExpired":false}`,
	} {
		if got := h.assertBatchFails(line); !strings.Contains(got, "unknown key") {
			t.Errorf("%s: message = %q", line, got)
		}
	}
}

func TestBatchTreatsNullAsAbsentInAScalarAndAsAFailureInAList(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"Nulls","due":null,"ordinal":null,"parent":null,"question":null}`); err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{
		`{"title":"a","labels":null}`,
		`{"title":"a","ext":null}`,
		`{"title":"a","acceptanceCriteria":null}`,
		`{"title":"a","comments":null}`,
	} {
		if got := h.assertBatchFails(line); !strings.Contains(got, "is null") {
			t.Errorf("%s: message = %q", line, got)
		}
	}
}

func TestBatchChecksTheTwoHalvesOfTheLeaseInvariant(t *testing.T) {
	h := newHarness(t)
	if got := h.assertBatchFails(
		`{"title":"Not active","leaseHolder":"@sara","leaseExpiresAt":"2126-09-08T14:00:00Z"}`,
	); got != "line 1: leaseHolder on a task that is not both active and assigned" {
		t.Errorf("message = %q", got)
	}
	if got := h.assertBatchFails(
		`{"title":"Half","status":"In Progress","assignees":["@sara"],"leaseHolder":"@sara"}`,
	); got != "line 1: leaseHolder given without leaseExpiresAt; the two go together" {
		t.Errorf("message = %q", got)
	}
	if _, err := h.batch(
		`{"title":"Whole","status":"In Progress","assignees":["@sara"],` +
			`"leaseHolder":"@sara","leaseExpiresAt":"2126-09-08T14:00:00Z"}`,
	); err != nil {
		t.Fatal(err)
	}
}

func TestBatchValidatesTheWholeFileBeforeWritingAnything(t *testing.T) {
	h := newHarness(t)
	_, err := h.batch(
		`{"title":"Good"}`,
		`{"title":""}`,
		`{"title":"Also good"}`,
		`{"title":"a","status":"Pendiente"}`)
	e := specError(t, err)
	if e.ExitCode != 7 || len(e.Details) != 2 {
		t.Fatalf("error = %d/%s with %d details", e.ExitCode, e.Code, len(e.Details))
	}
	all, _, listErr := h.b.Tasks.All()
	if listErr != nil {
		t.Fatal(listErr)
	}
	if len(all) != 0 {
		t.Errorf("the board has %d tasks, and a failed batch writes nothing", len(all))
	}
}

func TestBatchResolvesADependencyThatALaterLineCreates(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(
		`{"id":"MYP-1","title":"Depends on the next one","dependencies":["MYP-2"]}`,
		`{"id":"MYP-2","title":"The next one"}`,
	); err != nil {
		t.Fatal(err)
	}
	if got := h.assertBatchFails(`{"title":"Nowhere","dependencies":["MYP-90"]}`); !strings.Contains(
		got, "not on this board and not in this file") {
		t.Errorf("message = %q", got)
	}
}

func TestBatchPreviewWritesNothingAndCountsTheLines(t *testing.T) {
	h := newHarness(t)
	result, err := NewBatchOn(h.b, h.env, BatchParams{
		Content: "{\"title\":\"One\"}\n{\"title\":\"Two\"}\n", DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Previewed != 2 || len(result.Tasks) != 0 {
		t.Errorf("result = %+v, want two previewed and none created", result)
	}
	all, _, listErr := h.b.Tasks.All()
	if listErr != nil {
		t.Fatal(listErr)
	}
	if len(all) != 0 {
		t.Errorf("a preview left %d tasks behind", len(all))
	}
}

// TestBatchRefusesNullInEveryListTheFormatWrites takes the keys to try from
// an exported line instead of from a list written here: every key whose
// value comes out as [] or {} is a key whose null is a failure
// (docs/spec/cmd/new.md#el-modo-lote), so a list field added to the format
// later is covered by this test the day it is added.
func TestBatchRefusesNullInEveryListTheFormatWrites(t *testing.T) {
	h := newHarness(t)
	h.create("One of everything")
	all, _, err := h.b.Tasks.All()
	if err != nil {
		t.Fatal(err)
	}
	line, err := encodeTask(all[0])
	if err != nil {
		t.Fatal(err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(line, &fields); err != nil {
		t.Fatal(err)
	}

	tried := 0
	for key, value := range fields {
		if string(value) != "[]" && string(value) != "{}" {
			continue
		}
		tried++
		nulled, err := json.Marshal(map[string]json.RawMessage{
			"title": json.RawMessage(`"Nulled"`), key: json.RawMessage("null"),
		})
		if err != nil {
			t.Fatal(err)
		}
		if got := h.assertBatchFails(string(nulled)); !strings.Contains(got, key+" is null") {
			t.Errorf("%s: message = %q", key, got)
		}
	}
	if tried != len(listKeys) {
		t.Errorf("the export wrote %d empty lists and the format declares %d: %v",
			tried, len(listKeys), listKeys)
	}
}

// TestBatchNamesTheCodeOfEveryFailure is the other half of the messages the
// tests above check: the stable `code` each one carries, which is what a
// caller branches on without reading prose
// (docs/spec/contrato-json.md#los-identificadores-de-error). Three of them
// are exit code 2 and the fourth is 3, and the difference is the point: a
// key the format does not have is a file written wrong, while a line that
// cannot be read as a task is data this board cannot interpret.
func TestBatchNamesTheCodeOfEveryFailure(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"id":"MYP-1","title":"Taken"}`); err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		line     string
		code     string
		exitCode int
	}{
		{`{"title":"a","trelloCard":"5f2a8c1e"}`, "unknown_key", 2},
		{`{"id":"MYP-1","title":"a"}`, "id_taken", 2},
		{`{"title":"a","leaseHolder":"@sara"}`, "invalid_lease", 2},
		{`{"title":"a","labels":null}`, "invalid_line", 3},
		{`{"title":"a","createdAt":"yesterday"}`, "invalid_line", 3},
	} {
		_, err := h.batch(c.line)
		e := specError(t, err)
		if len(e.Details) != 1 {
			t.Fatalf("%s: %d failures", c.line, len(e.Details))
		}
		if e.Details[0].Code != c.code || e.Details[0].ExitCode != c.exitCode {
			t.Errorf("%s: code = %d/%s, want %d/%s (%s)", c.line,
				e.Details[0].ExitCode, e.Details[0].Code, c.exitCode, c.code,
				e.Details[0].Message)
		}
	}
}

// TestBatchReportsItsFailuresInTheOrderOfTheFile pins the order of the
// block of docs/spec/cmd/new.md#el-modo-lote, which is the order of the
// file. A failure of the graph is the one that can break it: it is found
// after every line has been read, not while reading its own.
func TestBatchReportsItsFailuresInTheOrderOfTheFile(t *testing.T) {
	h := newHarness(t)
	_, err := h.batch(
		`{"title":"","status":"To Do"}`,
		`{"title":"Fine"}`,
		`{"title":"Nowhere","parent":"MYP-900"}`,
		`{"title":"Also fine"}`,
		`{"title":"Bad status","status":"Pendiente"}`,
		`{"title":"No dependency","dependencies":["MYP-800"]}`,
	)
	e := specError(t, err)
	var lines []string
	for _, d := range e.Details {
		lines = append(lines, strings.SplitN(d.Message, ":", 2)[0])
	}
	want := []string{"line 1", "line 3", "line 5", "line 6"}
	if strings.Join(lines, ", ") != strings.Join(want, ", ") {
		t.Errorf("the failures came in the order %v and the file's is %v:\n%s",
			lines, want, strings.Join(e.Detail, "\n"))
	}
}
