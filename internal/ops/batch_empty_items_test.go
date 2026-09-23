package ops

import (
	"fmt"
	"strings"
	"testing"
)

// These are the rules of docs/spec/cmd/new.md#el-modo-lote and of
// docs/spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote about an
// element of a list that is empty or only spaces: it is dropped, it is not an
// error, the ones that stay are stored as they came, and `new --from` says how
// many it dropped from each list of each line. A null in the place of an
// element is another thing, a failure of the line.

// droppedWarnings is the warnings of a result that are of the empty-element
// kind, as "line/field/count" so that a test reads them at a glance.
func droppedWarnings(result *WriteResult) []string {
	var got []string
	for _, w := range result.Warnings {
		if w.Code == "imported_empty_dropped" {
			got = append(got, fmt.Sprintf("%v/%v/%v", w.Fields["line"], w.Fields["field"], w.Fields["count"]))
		}
	}
	return got
}

func warningCodes(result *WriteResult) string {
	codes := make([]string, 0, len(result.Warnings))
	for _, w := range result.Warnings {
		codes = append(codes, w.Code)
	}
	return strings.Join(codes, ",")
}

func TestBatchDropsEmptyReferencesAndCountsThem(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Holes","references":["","   ","ok"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != "ok" {
		t.Errorf("references = %q, want just ok", got)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/references/2" {
		t.Errorf("dropped warnings = %v", got)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("warnings = %v", result.Warnings)
	}
	w := result.Warnings[0]
	if w.Message != "line 1: 2 empty items dropped from references" {
		t.Errorf("the message = %q", w.Message)
	}
}

func TestBatchDropsASingleEmptyItemWithTheSingularWarning(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Hole","references":["a",""]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 1 || result.Warnings[0].Message != "line 1: 1 empty item dropped from references" {
		t.Errorf("warnings = %v", result.Warnings)
	}
}

func TestBatchAListOfOnlyEmptyItemsEndsUpEmptyAndDoesNotFail(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Only holes","references":["  "]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; len(got) != 0 {
		t.Errorf("references = %q, want none", got)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/references/1" {
		t.Errorf("dropped warnings = %v", got)
	}
}

func TestBatchTreatsEveryKindOfBlankAsEmpty(t *testing.T) {
	h := newHarness(t)
	// A tab, a line break and a no-break space, written as JSON escapes,
	// are all what strings.TrimSpace calls empty, which is the definition of
	// docs/spec/valores-de-entrada.md#el-valor-vacío.
	result, err := h.batch(`{"title":"Blanks","references":["\t","\n","\u00a0","\u2003 x"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; len(got) != 1 || got[0] != "\u2003 x" {
		t.Errorf("references = %q", got)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/references/3" {
		t.Errorf("dropped warnings = %v", got)
	}
}

func TestBatchKeepsTheElementsThatAreNotEmptyWithTheirSpaces(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"Spaces","references":[" a ","","b  "]}`); err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != " a |b  " {
		t.Errorf("references = %q, they must not be trimmed", got)
	}
}

func TestBatchDropsAnEmptyElementOfEveryListOfTokensAndIdentifiers(t *testing.T) {
	h := newHarness(t)
	h.create("First")
	result, err := h.batch(
		`{"title":"Tokens","assignees":["","@bob"],"labels":["","parser"," "],"dependencies":["","MYP-1"]}`)
	if err != nil {
		t.Fatalf("an empty element used to be a malformed label and is now dropped: %v", err)
	}
	task := h.load("MYP-2")
	if strings.Join(task.Assignees, "|") != "@bob" ||
		strings.Join(task.Labels, "|") != "parser" ||
		strings.Join(task.Dependencies, "|") != "MYP-1" {
		t.Errorf("assignees %q, labels %q, dependencies %q",
			task.Assignees, task.Labels, task.Dependencies)
	}
	want := "1/assignees/1,1/labels/2,1/dependencies/1"
	if got := droppedWarnings(result); strings.Join(got, ",") != want {
		t.Errorf("dropped warnings = %v, want %s", got, want)
	}
}

func TestBatchAListOfOnlyEmptyLabelsIsNotAMalformedLabel(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"a","labels":[""]}`, `{"title":"b","assignees":[" "]}`,
		`{"title":"c","dependencies":[""]}`); err != nil {
		t.Fatal(err)
	}
}

func TestBatchDropsEmptyCriteriaInEveryForm(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Criteria","acceptanceCriteria":[` +
		`"", "   ", {"text":""}, {"text":"  "}, {}, {"key":9,"checked":true},` +
		`{"text":"first"}, "second", {"key":4,"text":"third","checked":true}]}`)
	if err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	got := []string{}
	for _, c := range task.AcceptanceCriteria {
		got = append(got, fmt.Sprintf("%d:%s:%v", c.Key, c.Text, c.Checked))
	}
	// The dropped one that carried the key 9 reserves nothing: the highest
	// key that counts is 4, so the two without a key take 5 and 6.
	want := "5:first:false,6:second:false,4:third:true"
	if strings.Join(got, ",") != want {
		t.Errorf("criteria = %v, want %s", got, want)
	}
	if task.NextCriterionKey != 7 {
		t.Errorf("the counter is at %d, want 7", task.NextCriterionKey)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/acceptanceCriteria/6" {
		t.Errorf("dropped warnings = %v", got)
	}
}

func TestBatchADroppedCriterionDoesNotCountForARepeatedKey(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(
		`{"title":"Same key","acceptanceCriteria":[{"key":1,"text":""},{"key":1,"text":"x"}]}`); err != nil {
		t.Fatalf("the dropped one holds no key: %v", err)
	}
	task := h.load("MYP-1")
	if len(task.AcceptanceCriteria) != 1 || task.AcceptanceCriteria[0].Key != 1 ||
		task.AcceptanceCriteria[0].Text != "x" || task.NextCriterionKey != 2 {
		t.Errorf("criteria = %+v, counter %d", task.AcceptanceCriteria, task.NextCriterionKey)
	}
	// And two kept ones with the same key still fail.
	message := h.assertBatchFails(
		`{"title":"Twice","acceptanceCriteria":[{"key":1,"text":"a"},{"key":1,"text":""},{"key":1,"text":"b"}]}`)
	if !strings.Contains(message, "key 1 appears twice") {
		t.Errorf("message = %q", message)
	}
}

func TestBatchADroppedCriterionDoesNotRaiseTheCounter(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"Counter","acceptanceCriteria":[{"key":5,"text":""},"x"]}`); err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	if len(task.AcceptanceCriteria) != 1 || task.AcceptanceCriteria[0].Key != 1 || task.NextCriterionKey != 2 {
		t.Errorf("criteria = %+v, counter %d", task.AcceptanceCriteria, task.NextCriterionKey)
	}
}

func TestBatchADroppedCriterionWithAKeyBelowOneIsNotAnError(t *testing.T) {
	h := newHarness(t)
	if _, err := h.batch(`{"title":"Zero","acceptanceCriteria":[{"key":0,"text":" "},"x"]}`); err != nil {
		t.Fatalf("a dropped element is not judged: %v", err)
	}
}

func TestBatchDefinitionOfDoneDropsEmptyItemsBeforeConverting(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Dod","definitionOfDone":["",{"text":"  ","checked":true},"real"]}`)
	if err != nil {
		t.Fatal(err)
	}
	task := h.load("MYP-1")
	if len(task.AcceptanceCriteria) != 1 || task.AcceptanceCriteria[0].Key != 1 ||
		task.AcceptanceCriteria[0].Text != "real" || task.AcceptanceCriteria[0].Checked {
		t.Errorf("criteria = %+v", task.AcceptanceCriteria)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/definitionOfDone/2" {
		t.Errorf("dropped warnings = %v", got)
	}
	// The conversion warning counts what was converted, in the singular.
	if len(result.Warnings) != 2 || result.Warnings[1].Code != "imported_dod_merged" ||
		result.Warnings[1].Message != "line 1: 1 definition-of-done item imported as an acceptance criterion" {
		t.Errorf("warnings = %+v", result.Warnings)
	}
	if got := result.Tasks; len(got) != 1 || len(got[0].AcAdded) != 1 || got[0].AcAdded[0] != 1 {
		t.Errorf("ac_added = %+v", got)
	}
}

func TestBatchDefinitionOfDoneOfOnlyEmptyItemsConvertsNothing(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Dod","definitionOfDone":["", {"text":" "}]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").AcceptanceCriteria; len(got) != 0 {
		t.Errorf("criteria = %+v", got)
	}
	if warningCodes(result) != "imported_empty_dropped" {
		t.Errorf("warnings = %s, want only the dropped one", warningCodes(result))
	}
	if len(result.Tasks) == 1 && len(result.Tasks[0].AcAdded) != 0 {
		t.Errorf("ac_added = %v, nothing was converted", result.Tasks[0].AcAdded)
	}
}

func TestBatchDocumentationAndModifiedFilesDropEmptyItemsAndCountTheRest(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Pointers","references":["r1"],"documentation":["","d1"],"modifiedFiles":["  ","m1"]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != "r1|d1|m1" {
		t.Errorf("references = %q", got)
	}
	// Each key has its warning of dropped items, and then each merge warns
	// with what it merged, in the singular.
	wantCodes := "imported_empty_dropped,imported_empty_dropped," +
		"imported_documentation_merged,imported_modified_files_merged"
	if warningCodes(result) != wantCodes {
		t.Fatalf("warnings = %s, want %s", warningCodes(result), wantCodes)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/documentation/1,1/modifiedFiles/1" {
		t.Errorf("dropped warnings = %v", got)
	}
	if result.Warnings[2].Message != "line 1: 1 documentation item imported as a reference" ||
		result.Warnings[3].Message != "line 1: 1 modified file imported as a reference" {
		t.Errorf("merge warnings = %q, %q", result.Warnings[2].Message, result.Warnings[3].Message)
	}
}

func TestBatchDocumentationOfOnlyEmptyItemsMergesNothingAndSaysNoMerge(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(
		`{"title":"Blank pointers","references":["r1"],"documentation":["","  "],"modifiedFiles":[""]}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").References; strings.Join(got, "|") != "r1" {
		t.Errorf("references = %q", got)
	}
	if warningCodes(result) != "imported_empty_dropped,imported_empty_dropped" {
		t.Errorf("warnings = %s", warningCodes(result))
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/documentation/2,1/modifiedFiles/1" {
		t.Errorf("dropped warnings = %v", got)
	}
}

func TestBatchWarnsOfTheEmptyItemsBeforeTheMergesInAFixedOrder(t *testing.T) {
	h := newHarness(t)
	h.create("First")
	// The keys are written in the opposite order of the fixed one, on two
	// lines, and the second line has nothing to drop.
	result, err := h.batch(
		`{"title":"All","modifiedFiles":["","m"],"documentation":["","d"],"definitionOfDone":["","x"],`+
			`"acceptanceCriteria":["","y"],"references":["","r"],"dependencies":["","MYP-1"],`+
			`"labels":["","l"],"assignees":["","@a"]}`,
		`{"title":"Clean","documentation":["d2"]}`,
		`{"title":"Holes again","references":[""]}`)
	if err != nil {
		t.Fatal(err)
	}
	wantDropped := "1/assignees/1,1/labels/1,1/dependencies/1,1/references/1," +
		"1/acceptanceCriteria/1,1/definitionOfDone/1,1/documentation/1,1/modifiedFiles/1,3/references/1"
	if got := droppedWarnings(result); strings.Join(got, ",") != wantDropped {
		t.Errorf("dropped warnings =\n%v\nwant\n%s", got, wantDropped)
	}
	wantCodes := strings.Repeat("imported_empty_dropped,", 8) +
		"imported_dod_merged,imported_documentation_merged,imported_modified_files_merged," +
		"imported_documentation_merged," + // line 2
		"imported_empty_dropped" // line 3
	if warningCodes(result) != wantCodes {
		t.Errorf("warnings = %s\nwant       %s", warningCodes(result), wantCodes)
	}
}

func TestBatchRefusesANullInThePlaceOfAnElement(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct{ line, message string }{
		{`{"title":"a","references":["x",null]}`, "references.1: expected text, got null"},
		{`{"title":"a","labels":[null]}`, "labels.0: expected text, got null"},
		{`{"title":"a","assignees":[null]}`, "assignees.0: expected text, got null"},
		{`{"title":"a","dependencies":["MYP-1",null]}`, "dependencies.1: expected text, got null"},
		{`{"title":"a","documentation":[null]}`, "documentation.0: expected text, got null"},
		{`{"title":"a","modifiedFiles":["f",null]}`, "modifiedFiles.1: expected text, got null"},
		{`{"title":"a","definitionOfDone":[null]}`, "definitionOfDone.0: expected text or an object, got null"},
		{`{"title":"a","acceptanceCriteria":[null]}`, "acceptanceCriteria.0: expected text or an object, got null"},
		{`{"title":"a","acceptanceCriteria":["x",{"key":2,"text":null}]}`, "acceptanceCriteria.1.text: expected text, got null"},
		{`{"title":"a","definitionOfDone":[{"text":null}]}`, "definitionOfDone.0.text: expected text, got null"},
		{`{"title":"a","comments":[{"author":"@sara","body":"First"},null]}`, "comments.1: expected an object, got null"},
	} {
		if got := h.assertBatchFails(c.line); got != "line 1: "+c.message {
			t.Errorf("%s\n  message = %q\n  want      %q", c.line, got, "line 1: "+c.message)
		}
		_, err := h.batch(c.line)
		if e := specError(t, err).Details[0]; e.Code != "invalid_line" || e.ExitCode != 3 {
			t.Errorf("%s: %d/%s, want 3/invalid_line", c.line, e.ExitCode, e.Code)
		}
	}
}

// TestBatchRefusesAnInvalidComment is
// docs/decisiones/detalles.md#un-comentario-vacío-o-null-en-un-lote-es-un-fallo-de-validación:
// a comment does not follow the rule of the empty element, it fails the
// line instead, whether the element is not an object at all or is an
// object whose body is empty, only spaces, or missing.
func TestBatchRefusesAnInvalidComment(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct{ line, message string }{
		{`{"title":"a","comments":[null]}`, "comments.0: expected an object, got null"},
		{`{"title":"a","comments":["x"]}`, "comments.0: expected an object, got string"},
		{`{"title":"a","comments":[7]}`, "comments.0: expected an object, got number"},
		{`{"title":"a","comments":[{"author":"@sara","body":""}]}`, "comments.0: comment body cannot be empty"},
		{`{"title":"a","comments":[{"author":"@sara","body":"   "}]}`, "comments.0: comment body cannot be empty"},
		{`{"title":"a","comments":[{"author":"@sara"}]}`, "comments.0: comment body cannot be empty"},
		{`{"title":"a","comments":[{"body":"ok"},{"author":"@sara","body":""}]}`, "comments.1: comment body cannot be empty"},
	} {
		if got := h.assertBatchFails(c.line); got != "line 1: "+c.message {
			t.Errorf("%s\n  message = %q\n  want      %q", c.line, got, "line 1: "+c.message)
		}
		_, err := h.batch(c.line)
		if e := specError(t, err).Details[0]; e.Code != "invalid_line" || e.ExitCode != 3 {
			t.Errorf("%s: %d/%s, want 3/invalid_line", c.line, e.ExitCode, e.Code)
		}
	}
	// A comment whose body has real content is written as it came, without
	// trimming, exactly like the flag (docs/spec/valores-de-entrada.md#el-valor-vacío).
	if _, err := h.batch(`{"title":"Good","comments":[{"author":"@sara","body":" real "}]}`); err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").Comments[0].Body; got != " real " {
		t.Errorf("body = %q, want the untrimmed text", got)
	}
}

func TestBatchRefusesAnElementThatIsNotTextWhateverItIs(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct{ line, message string }{
		{`{"title":"a","references":[1]}`, "references.0: expected text, got number"},
		{`{"title":"a","acceptanceCriteria":[7]}`, "acceptanceCriteria.0: expected text or an object, got number"},
		{`{"title":"a","acceptanceCriteria":[true]}`, "acceptanceCriteria.0: expected text or an object, got bool"},
		{`{"title":"a","definitionOfDone":[["x"]]}`, "definitionOfDone.0: expected text or an object, got array"},
	} {
		if got := h.assertBatchFails(c.line); got != "line 1: "+c.message {
			t.Errorf("%s\n  message = %q\n  want      %q", c.line, got, "line 1: "+c.message)
		}
	}
}

func TestBatchAnAbsentTextIsAnEmptyOneAndNullIsNot(t *testing.T) {
	h := newHarness(t)
	// Not writing "text" is not writing null: the element has no text, so it
	// is empty and it is dropped. `documentation: null` is the key absent.
	result, err := h.batch(`{"title":"a","acceptanceCriteria":[{"checked":true}],"documentation":null}`)
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load("MYP-1").AcceptanceCriteria; len(got) != 0 {
		t.Errorf("criteria = %+v", got)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/acceptanceCriteria/1" {
		t.Errorf("dropped warnings = %v", got)
	}
}

func TestBatchAnInvalidBatchEmitsNoDroppedWarning(t *testing.T) {
	h := newHarness(t)
	_, err := h.batch(
		`{"title":"Holes","references":["","a"]}`,
		`{"title":"Bad","references":[null]}`)
	e := specError(t, err)
	if e.ExitCode != 7 || len(e.Details) != 1 {
		t.Fatalf("error = %+v", e)
	}
	all, _, listErr := h.b.Tasks.All()
	if listErr != nil {
		t.Fatal(listErr)
	}
	if len(all) != 0 {
		t.Errorf("an invalid batch wrote %d tasks", len(all))
	}
}

func TestBatchPreviewWarnsOfTheDroppedItemsAndWritesNothing(t *testing.T) {
	h := newHarness(t)
	result, err := NewBatchOn(h.b, h.env, BatchParams{
		Content: `{"title":"Holes","references":["","a"]}` + "\n", DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := droppedWarnings(result); strings.Join(got, ",") != "1/references/1" {
		t.Errorf("dropped warnings = %v", got)
	}
	all, _, listErr := h.b.Tasks.All()
	if listErr != nil {
		t.Fatal(listErr)
	}
	if len(all) != 0 {
		t.Errorf("a preview left %d tasks behind", len(all))
	}
}

func TestBatchTheFieldOfTheWarningIsTheKeyOfTheFile(t *testing.T) {
	h := newHarness(t)
	result, err := h.batch(`{"title":"Holes","acceptanceCriteria":[""]}`)
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 1 {
		t.Fatalf("warnings = %v", result.Warnings)
	}
	w := result.Warnings[0]
	if w.Code != "imported_empty_dropped" || w.Fields["field"] != "acceptanceCriteria" ||
		w.Fields["count"] != 1 || w.Fields["line"] != 1 || len(w.Fields) != 3 {
		t.Errorf("warning = %+v", w)
	}
}
