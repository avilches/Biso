package convert

import (
	"strings"
	"testing"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

var identifiersConfig = destination.Config{TaskPrefix: "BISO"}

// findByID finds the Identified for a given SourceID, failing the test if
// it is not there (a helper, not a rule this file is testing).
func findByID(t *testing.T, out []Identified, sourceID string) Identified {
	t.Helper()
	for _, i := range out {
		if i.SourceID == sourceID {
			return i
		}
	}
	t.Fatalf("no Identified with SourceID %q in %+v", sourceID, out)
	return Identified{}
}

func findingWithSubstring(findings []source.Finding, substr string) (source.Finding, bool) {
	for _, f := range findings {
		if strings.Contains(f.Message, substr) {
			return f, true
		}
	}
	return source.Finding{}, false
}

// TestIdentifiersASimpleFreeIdKeepsItsNumber covers docs/especificacion.md,
// "Identificadores", point 2: a simple id whose number is not taken on the
// destination keeps that number under the destination's prefix.
func TestIdentifiersASimpleFreeIdKeepsItsNumber(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "task-70.md", ID: "TASK-70", Title: "T"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
	if got := findByID(t, out, "TASK-70").ID; got != "BISO-70" {
		t.Errorf("ID = %q, want BISO-70", got)
	}
}

// TestIdentifiersASimpleIdCollidingWithTheDestinationIsReassigned covers
// docs task point 3 and point 8's exact wording for a simple id.
func TestIdentifiersASimpleIdCollidingWithTheDestinationIsReassigned(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "task-12.md", ID: "TASK-12", Title: "Colliding"}, Result: Result{}},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-12", Title: "Already here", CreatedAt: "2026-01-01T00:00:00Z"}},
	}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := findByID(t, out, "TASK-12").ID; got != "BISO-13" {
		t.Errorf("ID = %q, want BISO-13 (next free after the destination's max, 12)", got)
	}

	want := "TASK-12: id BISO-12 is taken on the destination, reassigned to BISO-13"
	f, ok := findingWithSubstring(findings, want)
	if !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, want)
	}
	if f.Field != "id" || f.File != "task-12.md" {
		t.Errorf("finding = %+v, want File=task-12.md Field=id", f)
	}
}

// TestIdentifiersASubtaskIsAlwaysReassignedAndLabeled covers docs task point
// 3 (subtasks always reassigned), point 8's subtask wording, and point 9
// (the backlog.id:: label).
func TestIdentifiersASubtaskIsAlwaysReassignedAndLabeled(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "task-56.1.md", ID: "XYZ-001.01", Title: "Sub"},
			Result: Result{Labels: []string{"backend"}},
		},
	}
	board := destination.Board{Config: destination.Config{TaskPrefix: "BISO"}}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "XYZ-001.01")
	if got.ID != "BISO-2" {
		t.Errorf("ID = %q, want BISO-2 (next free after max source main number, 1)", got.ID)
	}
	wantLabels := []string{"backend", "backlog.id::XYZ-001.01"}
	if !equalStrings(got.Labels, wantLabels) {
		t.Errorf("Labels = %v, want %v", got.Labels, wantLabels)
	}

	want := "XYZ-001.01: id BISO-2 assigned (subtask ids have no equivalent)"
	if _, ok := findingWithSubstring(findings, want); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, want)
	}
}

// TestIdentifiersDropsASourceBacklogIdLabelThatCollides covers the second
// half of docs task point 9: a subtask whose own labels already carry a
// backlog.id key loses that source label (with a Finding), same as
// milestone/project already do in phase 4a.
func TestIdentifiersDropsASourceBacklogIdLabelThatCollides(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "t.md", ID: "TASK-5.1", Title: "Sub"},
			Result: Result{Labels: []string{"backlog.id:stale-value"}},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "TASK-5.1")
	want := []string{"backlog.id::TASK-5.1"}
	if !equalStrings(got.Labels, want) {
		t.Errorf("Labels = %v, want %v (source label dropped)", got.Labels, want)
	}
	if _, ok := findingWithSubstring(findings, `"backlog.id:stale-value" collides with the derived backlog.id:: label, dropped`); !ok {
		t.Fatalf("findings = %v, want the collision finding", findings)
	}
}

// TestIdentifiersSkipsATaskAlreadyOnTheDestination covers docs task point 3
// (the "ya está en el destino" rule, numbered 2 in the encargo's own
// numbering): same title and same already-converted createdAt as an
// existing destination task means it is not re-emitted, and its id equates
// to whatever the destination already has, not a new number.
func TestIdentifiersSkipsATaskAlreadyOnTheDestination(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "task-9.md", ID: "TASK-9", Title: "Repeat"},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-40", Title: "Repeat", CreatedAt: "2026-01-01T00:00:00Z"}},
	}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("out = %v, want no rows (the task was skipped)", out)
	}
	want := source.Finding{File: "task-9.md", Field: "id", Message: "already on the destination, skipped"}
	if len(findings) != 1 || findings[0] != want {
		t.Fatalf("findings = %v, want exactly [%+v]", findings, want)
	}
}

// TestIdentifiersDependencyOnASkippedTaskResolvesToItsExistingDestinationId
// covers the encargo's explicit case: a batch task depends on another
// source task that turned out to already be on the destination. The
// dependency must resolve to the id that task ALREADY has there, not be
// dropped and not get a freshly assigned number.
func TestIdentifiersDependencyOnASkippedTaskResolvesToItsExistingDestinationId(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "old.md", ID: "TASK-1", Title: "Repeat"},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{
			Task: source.Task{
				File: "new.md", ID: "TASK-2", Title: "Depends on the repeat",
				Dependencies: []string{"TASK-1"},
			},
			Result: Result{CreatedAt: "2026-02-01T00:00:00Z"},
		},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-40", Title: "Repeat", CreatedAt: "2026-01-01T00:00:00Z"}},
	}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 1 {
		t.Fatalf("out = %v, want exactly 1 row (TASK-1 was skipped)", out)
	}
	got := findByID(t, out, "TASK-2")
	want := []string{"BISO-40"}
	if !equalStrings(got.Dependencies, want) {
		t.Errorf("Dependencies = %v, want %v", got.Dependencies, want)
	}
}

// TestIdentifiersRewritesMentionsAcrossFieldsInOnePass covers docs task
// point 5 over every text field it names, in a single batch that mixes a
// conserved id and a reassigned one.
func TestIdentifiersRewritesMentionsAcrossFieldsInOnePass(t *testing.T) {
	batch := []TaskInput{
		{
			Task: source.Task{
				File: "task-12.md", ID: "TASK-12", Title: "Fix TASK-12 follow-up",
				Description: "See TASK-99 for background.",
				AcceptanceCriteria: []source.Checkbox{
					{Number: 1, Text: "Depends on TASK-99 being done"},
				},
				Comments: []source.Comment{{Author: "@ann", Body: "Blocked by TASK-99"}},
			},
			Result: Result{
				AcceptanceCriteria: []source.Checkbox{{Number: 1, Text: "Depends on TASK-99 being done"}},
				Comments:           []Comment{{Author: "@ann", Body: "Blocked by TASK-99"}},
			},
		},
		{
			Task:   source.Task{File: "task-99.md", ID: "TASK-99", Title: "Background task"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}

	got := findByID(t, out, "TASK-12")
	if got.Title != "Fix BISO-12 follow-up" {
		t.Errorf("Title = %q", got.Title)
	}
	if got.Description != "See BISO-99 for background." {
		t.Errorf("Description = %q", got.Description)
	}
	if len(got.AcceptanceCriteria) != 1 || got.AcceptanceCriteria[0].Text != "Depends on BISO-99 being done" {
		t.Errorf("AcceptanceCriteria = %+v", got.AcceptanceCriteria)
	}
	if len(got.Comments) != 1 || got.Comments[0].Body != "Blocked by BISO-99" || got.Comments[0].Author != "@ann" {
		t.Errorf("Comments = %+v", got.Comments)
	}
}

// TestRewriteMentionsSubstitutesFromTheOriginalTextOnly is a direct,
// deterministic proof of docs task point 5's "en una sola pasada"
// requirement, exercised at rewriteMentions itself: TASK-1's equivalent is
// TASK-2, and TASK-2's own equivalent is TASK-3. If substitution rescanned
// its own output, the "TASK-2" written in place of "TASK-1" would be found
// again and turned into "TASK-3", corrupting the result. This exact
// coincidence cannot arise from a real Identifiers batch (a reassigned
// number is always chosen above every source main number, so it can never
// equal another source id's own number), which is why it is proven here
// against the helper directly rather than through the full pipeline.
func TestRewriteMentionsSubstitutesFromTheOriginalTextOnly(t *testing.T) {
	pattern := mentionPattern("TASK")
	equivalents := map[string]string{
		"TASK-1": "TASK-2",
		"TASK-2": "TASK-3",
	}

	got, unresolved := rewriteMentions("See TASK-1 and TASK-2", pattern, "TASK", equivalents)

	want := "See TASK-2 and TASK-3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if unresolved != 0 {
		t.Errorf("unresolved = %d, want 0", unresolved)
	}
}

// TestIdentifiersDoesNotTouchABranchNameLookingMention covers docs task
// point 5's explicit "TASK-10-modelo" example: not preceded by a
// letter/digit/_/-, but followed by '-' then a letter, so it is left alone.
func TestIdentifiersDoesNotTouchABranchNameLookingMention(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "t.md", ID: "TASK-10", Title: "Branch TASK-10-modelo lives here"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
	got := findByID(t, out, "TASK-10")
	if got.Title != "Branch TASK-10-modelo lives here" {
		t.Errorf("Title = %q, want it unchanged", got.Title)
	}
}

// TestIdentifiersDoesNotTouchASubtaskPrefixLookAlike covers docs task point
// 5's other explicit example: "SUBTASK-12" is never touched, even when the
// batch has a real "TASK-12", because the 'T' of "TASK-12" is preceded by
// the letter 'B' of "SUB".
func TestIdentifiersDoesNotTouchASubtaskPrefixLookAlike(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-12", Title: "Real task"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-13", Title: "Mentions SUBTASK-12 here"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
	got := findByID(t, out, "TASK-13")
	if got.Title != "Mentions SUBTASK-12 here" {
		t.Errorf("Title = %q, want it unchanged", got.Title)
	}
}

// TestIdentifiersAnUnresolvedMentionIsLeftAndGroupedByFileAndField covers
// docs task point 6: a mention with the right shape that names no task in
// the batch is left as-is and reported once per file and field, with a
// count, not once per mention.
func TestIdentifiersAnUnresolvedMentionIsLeftAndGroupedByFileAndField(t *testing.T) {
	batch := []TaskInput{
		{
			Task: source.Task{
				File: "t.md", ID: "TASK-1",
				Title:       "See TASK-404",
				Description: "Also see TASK-404 and TASK-405",
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "TASK-1")
	if got.Title != "See TASK-404" || got.Description != "Also see TASK-404 and TASK-405" {
		t.Errorf("text was changed: Title=%q Description=%q", got.Title, got.Description)
	}

	titleFinding, ok := findingWithSubstring(findings, "1 mention")
	if !ok {
		t.Fatalf("findings = %v, want one for title with count 1", findings)
	}
	_ = titleFinding
	var titleFindings, descriptionFindings int
	for _, f := range findings {
		switch f.Field {
		case "title":
			titleFindings++
			if !strings.Contains(f.Message, "1 mention") {
				t.Errorf("title finding = %+v, want count 1", f)
			}
		case "description":
			descriptionFindings++
			if !strings.Contains(f.Message, "2 mention") {
				t.Errorf("description finding = %+v, want count 2", f)
			}
		}
	}
	if titleFindings != 1 {
		t.Errorf("got %d title findings, want exactly 1 (grouped)", titleFindings)
	}
	if descriptionFindings != 1 {
		t.Errorf("got %d description findings, want exactly 1 (grouped)", descriptionFindings)
	}
}

// TestIdentifiersAWrongCaseMentionIsNeverSubstitutedAndIsGrouped covers docs
// task point 6's "Xyz-002"/"task-12" example: the pattern only matches
// exact uppercase, so a case variant is never a real match, but it still
// falls into the same grouped finding as an unresolved mention.
func TestIdentifiersAWrongCaseMentionIsNeverSubstitutedAndIsGrouped(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-2", Title: "Real"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-3", Title: "See task-2 and Task-2"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "TASK-3")
	if got.Title != "See task-2 and Task-2" {
		t.Errorf("Title = %q, want it unchanged (wrong case is never substituted)", got.Title)
	}
	f, ok := findingWithSubstring(findings, "2 mention")
	if !ok || f.Field != "title" || f.File != "b.md" {
		t.Fatalf("findings = %v, want a title finding on b.md with count 2", findings)
	}
}

// TestIdentifiersDropsAnUnknownParentWithAFinding covers docs task point 7
// for parent.
func TestIdentifiersDropsAnUnknownParentWithAFinding(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "t.md", ID: "TASK-2", Title: "Orphan", ParentTaskID: "TASK-999"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "TASK-2")
	if got.Parent != "" {
		t.Errorf("Parent = %q, want empty", got.Parent)
	}
	f, ok := findingWithSubstring(findings, `"TASK-999"`)
	if !ok || f.Field != "parent" || f.File != "t.md" {
		t.Fatalf("findings = %v, want a parent finding naming TASK-999", findings)
	}
}

// TestIdentifiersDropsOneUnknownDependencyKeepingTheRest covers docs task
// point 7 for dependencies: only the unresolvable element is removed.
func TestIdentifiersDropsOneUnknownDependencyKeepingTheRest(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{
			Task: source.Task{
				File: "b.md", ID: "TASK-2", Title: "B",
				Dependencies: []string{"TASK-1", "TASK-999"},
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := findByID(t, out, "TASK-2")
	if want := []string{"BISO-1"}; !equalStrings(got.Dependencies, want) {
		t.Errorf("Dependencies = %v, want %v", got.Dependencies, want)
	}
	f, ok := findingWithSubstring(findings, `"TASK-999"`)
	if !ok || f.Field != "dependencies" || f.File != "b.md" {
		t.Fatalf("findings = %v, want a dependencies finding naming TASK-999", findings)
	}
}

// TestIdentifiersRejectsAMalformedId covers docs task point 1: an id with
// the wrong shape aborts with an error, not a Finding.
func TestIdentifiersRejectsAMalformedId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "t.md", ID: "not-an-id", Title: "T"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one naming the malformed id")
	}
	if !strings.Contains(err.Error(), "not-an-id") {
		t.Errorf("error = %v, want it to mention the malformed id", err)
	}
}

// TestIdentifiersRejectsTwoDifferentPrefixes covers docs task point 1: two
// source tasks whose ids do not share a prefix abort with an error.
func TestIdentifiersRejectsTwoDifferentPrefixes(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "XYZ-2", Title: "B"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one about mismatched prefixes")
	}
	if !strings.Contains(err.Error(), "TASK") || !strings.Contains(err.Error(), "XYZ") {
		t.Errorf("error = %v, want it to mention both prefixes", err)
	}
}

// TestIdentifiersRejectsADuplicateId covers docs task point 1: two tasks
// with the exact same id abort with an error.
func TestIdentifiersRejectsADuplicateId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-1", Title: "B"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one about the duplicate id")
	}
	if !strings.Contains(err.Error(), "TASK-1") {
		t.Errorf("error = %v, want it to mention the duplicate id TASK-1", err)
	}
}
