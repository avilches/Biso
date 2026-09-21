package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// These tests walk, through the compiled program, the rules of
// docs/spec/cmd/new.md#el-modo-lote about the two keys that `new --from`
// accepts and folds into `references`, `documentation` and `modifiedFiles`:
// what a value of the wrong type answers, that a line that is invalid for
// another reason does not warn about a merge that never happens, that a
// preview warns exactly like the real call, and the shape of the two
// warnings. The rule of the merge itself (order, repeated values) is walked
// by internal/ops/batch_test.go.

// writeBatch writes the lines as an NDJSON file in the machine's project and
// answers its path.
func writeBatch(t *testing.T, m *machine, lines ...string) string {
	t.Helper()
	path := filepath.Join(m.dir, "tasks.ndjson")
	m.write(t, path, strings.Join(lines, "\n")+"\n")
	return path
}

// batchWarning is one entry of data.warnings of the batch envelope, for the
// two warnings these tests look at.
type batchWarning struct {
	Code  string `json:"code"`
	Line  int    `json:"line"`
	Count int    `json:"count"`
}

// batchWarnings runs `new --from <path> --json` (plus the extra flags) and
// answers the warnings of the envelope, in the order it lists them.
func batchWarnings(t *testing.T, m *machine, path string, extra ...string) []batchWarning {
	t.Helper()
	argv := append([]string{"new", "--from", path, "--json"}, extra...)
	got := m.run(t, argv...).assertCode(t, 0)
	var envelope struct {
		Data struct {
			Warnings []batchWarning `json:"warnings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the envelope of a batch is not JSON: %v\n%s", err, got.stdout)
	}
	return envelope.Data.Warnings
}

// TestBatchRefusesAPointerKeyOfTheWrongType is the first row of that
// rule: documentation and modifiedFiles are lists of text, and anything else
// is a validation failure of its line, listed with the rest and creating
// nothing.
func TestBatchRefusesAPointerKeyOfTheWrongType(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"A","documentation":"docs/a.md"}`,
		`{"title":"B","documentation":{"path":"docs/b.md"}}`,
		`{"title":"C","documentation":["docs/c.md",3]}`,
		`{"title":"D","modifiedFiles":"internal/d.go"}`,
		`{"title":"E","modifiedFiles":{"path":"internal/e.go"}}`,
		`{"title":"F","modifiedFiles":["internal/f.go",7]}`)

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	assertEqual(t, got.stderr,
		"error: 6 of 6 lines are invalid, nothing was written\n"+
			"  line 1: documentation: expected a list, got string\n"+
			"  line 2: documentation: expected a list, got object\n"+
			"  line 3: documentation.1: expected text, got number\n"+
			"  line 4: modifiedFiles: expected a list, got string\n"+
			"  line 5: modifiedFiles: expected a list, got object\n"+
			"  line 6: modifiedFiles.1: expected text, got number\n",
		"the failures of a batch with a pointer key of the wrong type")
	assertEqual(t, got.stdout, "", "the standard output of a batch that failed")
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a batch with a bad pointer key left behind")
}

// TestBatchDoesNotWarnAboutAMergeOfALineThatIsInvalidForAnotherReason is the
// second row: the warning belongs to a merge that happened, and a batch with
// an invalid line writes nothing, so it answers the failures and no
// warning, not even about the lines that were fine.
func TestBatchDoesNotWarnAboutAMergeOfALineThatIsInvalidForAnotherReason(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"Fine","documentation":["docs/a.md"]}`,
		`{"title":"Wrong status","status":"Nope","documentation":["docs/b.md"],"modifiedFiles":["internal/b.go"]}`)

	got := m.run(t, "new", "--from", path).assertCode(t, 7)

	assertEqual(t, got.stderr,
		"error: 1 of 2 lines is invalid, nothing was written\n"+
			"  line 2: unknown status: \"Nope\" (valid: To Do, In Progress, Done)\n",
		"the failures of a batch with an invalid line that also had documentation")
	if strings.Contains(got.stderr, "warning") || strings.Contains(got.stdout, "imported_") {
		t.Errorf("a batch that wrote nothing warned about a merge:\n%s%s", got.stderr, got.stdout)
	}
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks an invalid batch left behind")
}

// TestBatchDryRunWarnsAboutTheMergeAndWritesNothing is the third row: a
// preview names the same warnings the real call does, and leaves the board as
// it was.
func TestBatchDryRunWarnsAboutTheMergeAndWritesNothing(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"A","documentation":["docs/a.md","docs/b.md"],"modifiedFiles":["internal/a.go"]}`)

	preview := m.run(t, "new", "--from", path, "--dry-run").assertCode(t, 0)

	assertEqual(t, preview.stderr,
		"warning: line 1: 2 documentation items imported as references\n"+
			"warning: line 1: 1 modified file imported as a reference\n"+
			"1 task would be created, nothing was written (--dry-run)\n",
		"the preview of a batch with the two pointer keys")
	assertEqual(t, preview.stdout, "", "the standard output of a preview")
	assertEqual(t, m.run(t, "ls", "--count").assertCode(t, 0).stdout, "0\n",
		"the tasks a preview left behind")

	// The real call says the same two lines of warning.
	applied := m.run(t, "new", "--from", path).assertCode(t, 0)
	assertEqual(t, applied.stderr,
		"warning: line 1: 2 documentation items imported as references\n"+
			"warning: line 1: 1 modified file imported as a reference\n",
		"the warnings of the real call")
}

// TestBatchWarningsOfThePointerKeysInJSON is the fourth row: the two
// warnings carry code, line and count in the envelope, on a preview as on
// the real call, and the text says the singular or the plural after the
// count.
func TestBatchWarningsOfThePointerKeysInJSON(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"One each","documentation":["docs/a.md"],"modifiedFiles":["internal/a.go"]}`,
		`{"title":"Several","documentation":["docs/b.md","docs/c.md"],"modifiedFiles":["internal/b.go","internal/c.go","internal/d.go"]}`)
	want := []batchWarning{
		{"imported_documentation_merged", 1, 1},
		{"imported_modified_files_merged", 1, 1},
		{"imported_documentation_merged", 2, 2},
		{"imported_modified_files_merged", 2, 3},
	}

	for _, extra := range [][]string{{"--dry-run"}, {}} {
		got := batchWarnings(t, m, path, extra...)
		if len(got) != len(want) {
			t.Fatalf("%v: %d warnings %v, want %v", extra, len(got), got, want)
		}
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%v: warning %d = %+v, want %+v", extra, i, got[i], want[i])
			}
		}
	}

	// The keys of one warning are exactly these three.
	raw := m.run(t, "new", "--from", path, "--json", "--dry-run").assertCode(t, 0).stdout
	var envelope struct {
		Data struct {
			Warnings []map[string]any `json:"warnings"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		t.Fatal(err)
	}
	for _, w := range envelope.Data.Warnings {
		if len(w) != 3 || w["code"] == nil || w["line"] == nil || w["count"] == nil {
			t.Errorf("a warning carries %v and not code, line and count", w)
		}
	}

	// The text of the singular and the plural of both.
	text := m.run(t, "new", "--from", path, "--dry-run").assertCode(t, 0)
	assertEqual(t, text.stderr,
		"warning: line 1: 1 documentation item imported as a reference\n"+
			"warning: line 1: 1 modified file imported as a reference\n"+
			"warning: line 2: 2 documentation items imported as references\n"+
			"warning: line 2: 3 modified files imported as references\n"+
			"2 tasks would be created, nothing was written (--dry-run)\n",
		"the text of the warnings, in the singular and in the plural")
}

// TestBatchMergesReferencesDocumentationAndModifiedFilesInOrder is the
// fifth row: a line with the three keys, written in the opposite order,
// ends with references first, then documentation, then modifiedFiles, and
// gets the two warnings in that order, the one of documentation first.
func TestBatchMergesReferencesDocumentationAndModifiedFilesInOrder(t *testing.T) {
	m := batchBoard(t)
	path := writeBatch(t, m,
		`{"title":"All three","modifiedFiles":["f1","d1"],"documentation":["d1","d2"],"references":["r1"]}`)

	got := batchWarnings(t, m, path)

	want := []batchWarning{
		{"imported_documentation_merged", 1, 2},
		{"imported_modified_files_merged", 1, 2},
	}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("warnings = %+v, want %+v", got, want)
	}
	refs := ""
	for _, line := range strings.Split(m.run(t, "get", "MYP-1").assertCode(t, 0).stdout, "\n") {
		if strings.HasPrefix(line, "refs ") {
			refs = strings.TrimSpace(strings.TrimPrefix(line, "refs"))
		}
	}
	assertEqual(t, refs, "r1, d1, d2, f1", "the references of the task")
}
