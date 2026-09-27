package convert

import (
	"encoding/json"
	"strings"
	"testing"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

// findLine finds the Line with the given final id, failing the test if it is
// not there (a helper, not a rule this file is testing).
func findLine(t *testing.T, lines []Line, id string) Line {
	t.Helper()
	for _, l := range lines {
		if l.ID == id {
			return l
		}
	}
	var ids []string
	for _, l := range lines {
		ids = append(ids, l.ID)
	}
	t.Fatalf("no Line with id %q, got: %v", id, ids)
	return Line{}
}

// TestAssembleCombinesEveryFieldSource covers the task-70 phase 5 brief's
// pipeline step 4: a single Line combines fields from four different
// places (Identified, the matching Result, AssignOrdinals, and the original
// source.Task) that no earlier phase ever joins together on its own. This
// is the one test that would catch a field pulled from the wrong source, or
// forgotten entirely, especially References/Documentation/ModifiedFiles/
// Archived, which the specification's own field-mapping table never lists
// because no earlier phase touches them at all.
func TestAssembleCombinesEveryFieldSource(t *testing.T) {
	ordinal := 1000.0
	sourceBoard := source.Board{
		Tasks: []source.Task{
			{
				File:          "task-1 - Fix parser.md",
				Archived:      true,
				ID:            "TASK-1",
				Title:         "Fix parser",
				Status:        "Done",
				Type:          "bug",
				Priority:      "high",
				Assignees:     []string{"ann"},
				Labels:        []string{"backend"},
				CreatedDate:   "2026-01-01 10:00",
				UpdatedDate:   "2026-01-02 11:00",
				DueDate:       "2026-01-05",
				References:    []string{"docs/a.md"},
				Documentation: []string{"docs/b.md"},
				ModifiedFiles: []string{"src/a.go"},
				Ordinal:       &ordinal,
				Description:   "Some description",
				Plan:          "Some plan",
				Notes:         "Some notes",
				Summary:       "Some summary",
				AcceptanceCriteria: []source.Checkbox{
					{Number: 1, Text: "Works", Checked: true},
				},
				Comments: []source.Comment{
					{Author: "@ann", CreatedAt: "2026-01-01 10:00", Body: "hi"},
				},
			},
		},
	}
	destBoard := destination.Board{
		Config: destination.Config{
			TaskPrefix: "BISO",
			Statuses:   []string{"Done"},
			Types:      []string{"bug"},
			Priorities: []string{"high"},
		},
	}

	lines, findings, err := Assemble(sourceBoard, destBoard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1", len(lines))
	}

	got := lines[0]

	// From Identified (phase 4b).
	if got.ID != "BISO-1" {
		t.Errorf("ID = %q, want BISO-1", got.ID)
	}
	if got.Title != "Fix parser" {
		t.Errorf("Title = %q, want %q", got.Title, "Fix parser")
	}
	if got.Description != "Some description" {
		t.Errorf("Description = %q", got.Description)
	}
	if got.Plan != "Some plan" {
		t.Errorf("Plan = %q", got.Plan)
	}
	if got.Notes != "Some notes" {
		t.Errorf("Notes = %q", got.Notes)
	}
	if got.Summary != "Some summary" {
		t.Errorf("Summary = %q", got.Summary)
	}
	if !equalStrings(got.Labels, []string{"backend"}) {
		t.Errorf("Labels = %v, want [backend]", got.Labels)
	}
	if len(got.AcceptanceCriteria) != 1 || got.AcceptanceCriteria[0] != (AcceptanceCriterion{Key: 1, Text: "Works", Checked: true}) {
		t.Errorf("AcceptanceCriteria = %+v", got.AcceptanceCriteria)
	}
	if len(got.Comments) != 1 || got.Comments[0] != (CommentLine{Author: "@ann", CreatedAt: "2026-01-01T10:00:00Z", Body: "hi"}) {
		t.Errorf("Comments = %+v", got.Comments)
	}

	// From Result (phase 4a).
	if got.Status != "Done" {
		t.Errorf("Status = %q, want Done", got.Status)
	}
	if got.Type != "bug" {
		t.Errorf("Type = %q, want bug", got.Type)
	}
	if got.Priority != "high" {
		t.Errorf("Priority = %q, want high", got.Priority)
	}
	if !equalStrings(got.Assignees, []string{"ann"}) {
		t.Errorf("Assignees = %v, want [ann]", got.Assignees)
	}
	if got.Due != "2026-01-05" {
		t.Errorf("Due = %q, want 2026-01-05 (verbatim, never UTC-converted)", got.Due)
	}
	if got.CreatedAt != "2026-01-01T10:00:00Z" {
		t.Errorf("CreatedAt = %q", got.CreatedAt)
	}
	if got.UpdatedAt != "2026-01-02T11:00:00Z" {
		t.Errorf("UpdatedAt = %q", got.UpdatedAt)
	}

	// From AssignOrdinals (phase 4c).
	if got.Ordinal == "" {
		t.Errorf("Ordinal is empty, want a manual order key (the source task had an ordinal)")
	}

	// Straight from the original source.Task, untouched by any earlier
	// phase.
	if !equalStrings(got.References, []string{"docs/a.md"}) {
		t.Errorf("References = %v, want [docs/a.md]", got.References)
	}
	if !equalStrings(got.Documentation, []string{"docs/b.md"}) {
		t.Errorf("Documentation = %v, want [docs/b.md]", got.Documentation)
	}
	if !equalStrings(got.ModifiedFiles, []string{"src/a.go"}) {
		t.Errorf("ModifiedFiles = %v, want [src/a.go]", got.ModifiedFiles)
	}
	if !got.Archived {
		t.Errorf("Archived = false, want true (the source task came from archive/tasks/)")
	}
}

// TestAssembleOmitsOrdinalForATaskWithNoSourceOrdinal covers the brief's
// explicit warning: a task with no ordinal at all must have its ordinal key
// omitted entirely, not written as an empty string or null.
func TestAssembleOmitsOrdinalForATaskWithNoSourceOrdinal(t *testing.T) {
	sourceBoard := source.Board{
		Tasks: []source.Task{
			{File: "t.md", ID: "TASK-1", Title: "Bare task"},
		},
	}
	destBoard := destination.Board{Config: destination.Config{TaskPrefix: "BISO"}}

	lines, _, err := Assemble(sourceBoard, destBoard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findLine(t, lines, "BISO-1")
	if got.Ordinal != "" {
		t.Errorf("Ordinal = %q, want empty (no source ordinal at all)", got.Ordinal)
	}

	data, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	if strings.Contains(string(data), `"ordinal"`) {
		t.Errorf("marshaled line = %s, must not mention the \"ordinal\" key at all", data)
	}
}

// TestAssembleSortsByTheSourceIdsNumberAscending covers
// docs/especificacion.md, "La salida": lines come out ordered by the source
// id's own number, ascending, using the SAME natural order identifiers.go
// and ordinal.go already use (main number first, a simple id before any
// subtask of the same number). The batch is deliberately given out of
// order, and the destination assigns final ids in a completely different
// order than the source numbers (TASK-20 keeps its own number and sorts
// last, while the subtask TASK-1.5 is reassigned to whatever the next free
// destination number is and must still sort first).
func TestAssembleSortsByTheSourceIdsNumberAscending(t *testing.T) {
	sourceBoard := source.Board{
		Tasks: []source.Task{
			{File: "b.md", ID: "TASK-20", Title: "Twenty"},
			{File: "c.md", ID: "TASK-1.5", Title: "Subtask of one"},
			{File: "a.md", ID: "TASK-3", Title: "Three"},
		},
	}
	destBoard := destination.Board{Config: destination.Config{TaskPrefix: "BISO"}}

	lines, _, err := Assemble(sourceBoard, destBoard)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lines) != 3 {
		t.Fatalf("got %d lines, want 3", len(lines))
	}

	var gotTitles []string
	for _, l := range lines {
		gotTitles = append(gotTitles, l.Title)
	}
	want := []string{"Subtask of one", "Three", "Twenty"}
	if !equalStrings(gotTitles, want) {
		t.Errorf("Title order = %v, want %v (ascending by source id number: 1, 3, 20)", gotTitles, want)
	}
}

// TestAssemblePropagatesTheIdentifiersFatalError covers the brief's pipeline
// step 2: when Identifiers fails its point-1 validation, Assemble returns
// that error and nothing else, since there is no batch left to assemble.
func TestAssemblePropagatesTheIdentifiersFatalError(t *testing.T) {
	sourceBoard := source.Board{
		Tasks: []source.Task{
			{File: "t.md", ID: "not-an-id", Title: "Malformed"},
		},
	}
	destBoard := destination.Board{Config: destination.Config{TaskPrefix: "BISO"}}

	lines, findings, err := Assemble(sourceBoard, destBoard)
	if err == nil {
		t.Fatal("got nil error, want one naming the malformed id")
	}
	if lines != nil || findings != nil {
		t.Errorf("lines = %v, findings = %v, want both nil on a fatal error", lines, findings)
	}
	if !strings.Contains(err.Error(), "not-an-id") {
		t.Errorf("error = %v, want it to mention the malformed id", err)
	}
}

// TestAssembleIsDeterministic covers the brief's own requirement: two runs
// of Assemble over the exact same input boards must produce byte-for-byte
// identical NDJSON.
func TestAssembleIsDeterministic(t *testing.T) {
	buildBoards := func() (source.Board, destination.Board) {
		ordinal1 := 500.0
		ordinal2 := 10.0
		sourceBoard := source.Board{
			Milestones: map[string]string{"m-1": "Alpha Phase"},
			Tasks: []source.Task{
				{
					File: "task-2.md", ID: "TASK-2", Title: "Second", Milestone: "m-1",
					Dependencies: []string{"TASK-1"}, Ordinal: &ordinal2,
				},
				{
					File: "task-1.md", ID: "TASK-1", Title: "First",
					AcceptanceCriteria: []source.Checkbox{{Number: 1, Text: "Done", Checked: false}},
					Ordinal:            &ordinal1,
				},
				{File: "task-3.1.md", ID: "TASK-3.1", Title: "Sub of three", ParentTaskID: "TASK-3"},
				{File: "task-3.md", ID: "TASK-3", Title: "Three"},
			},
		}
		destBoard := destination.Board{
			Config: destination.Config{TaskPrefix: "BISO"},
			Tasks:  []destination.Task{{ID: "BISO-9", Title: "Existing", CreatedAt: "2020-01-01T00:00:00Z", Ordinal: "5"}},
		}
		return sourceBoard, destBoard
	}

	encode := func(lines []Line) string {
		var b strings.Builder
		enc := json.NewEncoder(&b)
		for _, line := range lines {
			if err := enc.Encode(line); err != nil {
				t.Fatalf("json encode: %v", err)
			}
		}
		return b.String()
	}

	sb1, db1 := buildBoards()
	lines1, findings1, err := Assemble(sb1, db1)
	if err != nil {
		t.Fatalf("unexpected error (run 1): %v", err)
	}
	sb2, db2 := buildBoards()
	lines2, findings2, err := Assemble(sb2, db2)
	if err != nil {
		t.Fatalf("unexpected error (run 2): %v", err)
	}

	ndjson1 := encode(lines1)
	ndjson2 := encode(lines2)
	if ndjson1 != ndjson2 {
		t.Fatalf("two runs produced different NDJSON:\nrun 1: %s\nrun 2: %s", ndjson1, ndjson2)
	}
	if len(findings1) != len(findings2) {
		t.Fatalf("two runs produced a different number of findings: %d vs %d", len(findings1), len(findings2))
	}
	for i := range findings1 {
		if findings1[i] != findings2[i] {
			t.Fatalf("finding %d differs: %+v vs %+v", i, findings1[i], findings2[i])
		}
	}
}

// TestLineFieldOrderMatchesTheSpecification covers docs/especificacion.md,
// "La salida" ("las claves de cada linea en un orden fijo") and the task-70
// phase 5 brief's exact list: a Line with every field filled in must
// serialize with its keys in exactly that order. This is a literal,
// byte-for-byte comparison against the brief's own key list, deliberately
// not reusing Assemble so that a change to Line's field order is caught even
// if the assembly pipeline's own logic did not change.
func TestLineFieldOrderMatchesTheSpecification(t *testing.T) {
	line := Line{
		ID:                 "BISO-1",
		Title:              "T",
		Status:             "Done",
		Type:               "bug",
		Priority:           "high",
		Assignees:          []string{"ann"},
		Labels:             []string{"backend"},
		Dependencies:       []string{"BISO-2"},
		Parent:             "BISO-3",
		Ordinal:            "i",
		Due:                "2026-01-05",
		Documentation:      []string{"docs/b.md"},
		References:         []string{"docs/a.md"},
		ModifiedFiles:      []string{"src/a.go"},
		Archived:           true,
		CreatedAt:          "2026-01-01T10:00:00Z",
		UpdatedAt:          "2026-01-02T11:00:00Z",
		Description:        "D",
		Plan:               "P",
		Notes:              "N",
		Summary:            "S",
		AcceptanceCriteria: []AcceptanceCriterion{{Key: 1, Text: "Works", Checked: true}},
		Comments:           []CommentLine{{Author: "@ann", CreatedAt: "2026-01-01T10:00:00Z", Body: "hi"}},
	}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	want := `{"id":"BISO-1","title":"T","status":"Done","type":"bug","priority":"high",` +
		`"assignees":["ann"],"labels":["backend"],"dependencies":["BISO-2"],"parent":"BISO-3",` +
		`"ordinal":"i","due":"2026-01-05","documentation":["docs/b.md"],"references":["docs/a.md"],` +
		`"modifiedFiles":["src/a.go"],"archived":true,"createdAt":"2026-01-01T10:00:00Z",` +
		`"updatedAt":"2026-01-02T11:00:00Z","description":"D","plan":"P","notes":"N","summary":"S",` +
		`"acceptanceCriteria":[{"key":1,"text":"Works","checked":true}],` +
		`"comments":[{"author":"@ann","createdAt":"2026-01-01T10:00:00Z","body":"hi"}]}`

	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}
}

// TestLineOmitsEveryEmptyField covers docs/especificacion.md, "La salida":
// a field that is absent or empty (an empty string, an empty list, or
// false) never appears in the line at all, not as null and not as [].
func TestLineOmitsEveryEmptyField(t *testing.T) {
	line := Line{ID: "BISO-1", Title: "Bare task"}

	data, err := json.Marshal(line)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}

	want := `{"id":"BISO-1","title":"Bare task"}`
	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}
}

// TestCommentLineOmitsAuthorAndCreatedAtWhenAbsent covers
// docs/especificacion.md, "Comentarios": a comment without an author, or
// without a created date, is valid, and the brief's own instruction that
// both keys are then omitted from the comment object, not written as empty
// strings.
func TestCommentLineOmitsAuthorAndCreatedAtWhenAbsent(t *testing.T) {
	data, err := json.Marshal(CommentLine{Body: "No author, no date"})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	want := `{"body":"No author, no date"}`
	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}
}

// TestAcceptanceCriterionAlwaysWritesChecked covers the brief's own
// clarification: unlike every other optional field, "checked" is never
// omitted even when false, because a false value is meaningful (unchecked),
// not absent.
func TestAcceptanceCriterionAlwaysWritesChecked(t *testing.T) {
	data, err := json.Marshal(AcceptanceCriterion{Key: 3, Text: "Not done yet", Checked: false})
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	want := `{"key":3,"text":"Not done yet","checked":false}`
	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}
}
