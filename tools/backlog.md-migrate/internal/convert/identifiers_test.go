package convert

import (
	"fmt"
	"path/filepath"
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
// docs/especificacion.md, "Identificadores", point 4 (reassignment) and
// point 8's exact wording, for a simple id.
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

// TestIdentifiersASubtaskIsAlwaysReassignedAndLabeled covers
// docs/especificacion.md, "Identificadores", point 4 (subtasks always
// reassigned), point 8's subtask wording, and point 9 (the backlog.id::
// label).
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
// half of docs/especificacion.md, "Identificadores", point 9: a subtask
// whose own labels already carry a backlog.id key loses that source label
// (with a Finding), same as milestone/project already do in phase 4a.
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

// TestIdentifiersSkipsATaskAlreadyOnTheDestination covers
// docs/especificacion.md, "Identificadores", point 3: same title and same
// already-converted createdAt as an existing destination task means the
// source task is not re-emitted, and its id equates to whatever the
// destination already has, not a new number.
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

// TestIdentifiersAlreadyImportedComparisonUsesTheNaivelyRewrittenTitle
// covers docs/especificacion.md, "Identificadores", point 3's naive title
// rewrite, and its reasoning in docs/decisiones.md, section "Los
// identificadores conservan su número y cambian de prefijo", in the
// paragraph about the naive title comparison. TASK-2's raw title mentions
// TASK-1 ("Follow-up of TASK-1"); the destination already has a task whose
// title is the REWRITTEN form from an earlier run ("Follow-up of BISO-1").
// Comparing the raw title would never match this destination title, and
// TASK-2 would be duplicated; comparing the naively rewritten title does
// match, because TASK-1 is a simple id present in the batch, so its own
// number gets substituted.
//
// The destination also already has an unrelated task occupying number 1
// ("BISO-1"), which forces TASK-1 itself to collide and be reassigned to a
// different number in THIS run. The naive comparison must still use TASK-1's
// ORIGINAL number, 1, not its freshly reassigned one, or this test would not
// actually distinguish the naive rewrite from the definitive one.
func TestIdentifiersAlreadyImportedComparisonUsesTheNaivelyRewrittenTitle(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "one.md", ID: "TASK-1", Title: "Original"},
			Result: Result{CreatedAt: "2020-06-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "two.md", ID: "TASK-2", Title: "Follow-up of TASK-1"},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks: []destination.Task{
			{ID: "BISO-1", Title: "Unrelated", CreatedAt: "2099-01-01T00:00:00Z"},
			{ID: "BISO-2", Title: "Follow-up of BISO-1", CreatedAt: "2026-01-01T00:00:00Z"},
		},
	}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reassignedTask1ID := findByID(t, out, "TASK-1").ID
	if reassignedTask1ID == "BISO-1" {
		t.Fatalf("TASK-1 was not actually reassigned, this test proves nothing: ID = %q", reassignedTask1ID)
	}

	for _, i := range out {
		if i.SourceID == "TASK-2" {
			t.Fatalf("TASK-2 was emitted, want it skipped as already on the destination: %+v", i)
		}
	}
	want := source.Finding{File: "two.md", Field: "id", Message: "already on the destination, skipped"}
	found := false
	for _, f := range findings {
		if f == want {
			found = true
		}
	}
	if !found {
		t.Fatalf("findings = %v, want %+v", findings, want)
	}
}

// TestNaiveTitleLeavesAMentionOfAnUnknownIdUntouched covers
// docs/especificacion.md, "Identificadores", point 3's naive rewrite: a
// mention with the right shape that does not correspond to any task in the
// batch (TASK-99 is not one of the batch's ids at all here) is left
// untouched, the same way point 6 leaves an unresolved mention untouched in
// the definitive rewrite.
func TestNaiveTitleLeavesAMentionOfAnUnknownIdUntouched(t *testing.T) {
	pattern := mentionPattern("TASK")
	simpleSourceNumbers := map[int]bool{} // TASK-99 names no task in this batch.

	got := naiveTitle("Revert TASK-99", pattern, "TASK", "BISO", simpleSourceNumbers)

	want := "Revert TASK-99"
	if got != want {
		t.Errorf("got %q, want %q (TASK-99 is not a batch task, so it must be left untouched)", got, want)
	}
}

// TestNaiveTitleLeavesAMentionOfAKnownSubtaskUntouched covers
// docs/especificacion.md, "Identificadores", point 3's narrower final rule:
// a mention of a SUBTASK (TASK-1.1 here) is left untouched even though that
// exact subtask really is a task in the batch, because a subtask always
// gets a fresh number this naive rewrite cannot predict (point 4). This is
// the accepted limitation docs/decisiones.md documents for a title that
// mentions a subtask.
//
// simpleSourceNumbers deliberately has main number 1, as if a SIMPLE
// "TASK-1" also existed in the batch alongside the subtask "TASK-1.1": a
// version of naiveTitle that only checked the main number, without also
// checking that the candidate itself has no dot, would wrongly substitute
// this mention. An empty simpleSourceNumbers would not catch that mistake,
// since the number-not-found path and the is-a-subtask path would look the
// same.
func TestNaiveTitleLeavesAMentionOfAKnownSubtaskUntouched(t *testing.T) {
	pattern := mentionPattern("TASK")
	simpleSourceNumbers := map[int]bool{1: true}

	got := naiveTitle("Finish TASK-1.1", pattern, "TASK", "BISO", simpleSourceNumbers)

	want := "Finish TASK-1.1"
	if got != want {
		t.Errorf("got %q, want %q (TASK-1.1 is a subtask, point 3 never substitutes a subtask mention)", got, want)
	}
}

// TestNaiveTitleSubstitutesAMentionOfAKnownSimpleTask covers
// docs/especificacion.md, "Identificadores", point 3's one case that IS
// substituted: a mention whose number names a SIMPLE task present in the
// batch is rewritten to the destination's own prefix and that same number.
func TestNaiveTitleSubstitutesAMentionOfAKnownSimpleTask(t *testing.T) {
	pattern := mentionPattern("TASK")
	simpleSourceNumbers := map[int]bool{1: true}

	got := naiveTitle("Follow-up of TASK-1", pattern, "TASK", "BISO", simpleSourceNumbers)

	want := "Follow-up of BISO-1"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// TestIdentifiersDependencyOnASkippedTaskResolvesToItsExistingDestinationId
// covers this explicit case: a batch task depends on another source task
// that turned out to already be on the destination. The dependency must
// resolve to the id that task ALREADY has there, not be dropped and not get
// a freshly assigned number.
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

// TestIdentifiersReassignmentCeilingCountsSkippedTasksToo covers
// docs/especificacion.md, "Identificadores", point 4: the ceiling a
// reassigned number starts above must count the main number of EVERY
// source task in the batch, including one skipped for already being on the
// destination (TASK-500 here), not just the tasks this phase ends up
// reassigning. If the ceiling ignored skipped tasks, TASK-1.1 would wrongly
// be assigned a low number that could collide with a later import of the
// same source board.
func TestIdentifiersReassignmentCeilingCountsSkippedTasksToo(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "a.md", ID: "TASK-500", Title: "Repeat"},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "b.md", ID: "TASK-1.1", Title: "Sub"},
			Result: Result{CreatedAt: "2026-02-01T00:00:00Z"},
		},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-9", Title: "Repeat", CreatedAt: "2026-01-01T00:00:00Z"}},
	}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := findByID(t, out, "TASK-1.1").ID; got != "BISO-501" {
		t.Errorf("ID = %q, want BISO-501 (next free after 500, the skipped task's own main number)", got)
	}
}

// TestIdentifiersReassignsInNaturalOrderRegardlessOfBatchOrder covers
// docs/especificacion.md, "Identificadores", point 4's natural-order
// assignment rule: three subtasks that all need reassigning arrive in
// scrambled batch order, but their final numbers must still follow the
// ascending (main number, subtask number) order of their SOURCE ids, not
// the order they were given in.
func TestIdentifiersReassignsInNaturalOrderRegardlessOfBatchOrder(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "c.md", ID: "TASK-5.2", Title: "C"}, Result: Result{}},
		{Task: source.Task{File: "a.md", ID: "TASK-1.1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-5.1", Title: "B"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := findByID(t, out, "TASK-1.1").ID; got != "BISO-6" {
		t.Errorf("TASK-1.1 = %q, want BISO-6 (assigned first: smallest main number)", got)
	}
	if got := findByID(t, out, "TASK-5.1").ID; got != "BISO-7" {
		t.Errorf("TASK-5.1 = %q, want BISO-7 (assigned second: same main number as TASK-5.2, smaller subtask number)", got)
	}
	if got := findByID(t, out, "TASK-5.2").ID; got != "BISO-8" {
		t.Errorf("TASK-5.2 = %q, want BISO-8 (assigned last)", got)
	}
}

// TestIdentifiersASimpleIdSortsBeforeItsOwnSubtaskWhenBothAreReassigned
// covers docs/especificacion.md, "Identificadores", point 4's natural-order
// tie-break in the one case it can actually happen: a simple id that
// collides with the destination (TASK-1) and one of its own subtasks
// (TASK-1.0, always reassigned) share the same main number, 1, and both
// land in the reassigned group at once. A simple id must always be
// assigned the smaller of the two new numbers ("TASK-1" reads before
// "TASK-1.0"), regardless of the order the two arrive in the batch slice;
// giving the subtask first in the input here is what exposes a tie-break
// that falls back to batch order instead of being deterministic.
func TestIdentifiersASimpleIdSortsBeforeItsOwnSubtaskWhenBothAreReassigned(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "sub.md", ID: "TASK-1.0", Title: "Sub"}, Result: Result{}},
		{Task: source.Task{File: "simple.md", ID: "TASK-1", Title: "Simple"}, Result: Result{}},
	}
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-1", Title: "Unrelated", CreatedAt: "2020-01-01T00:00:00Z"}},
	}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := findByID(t, out, "TASK-1").ID; got != "BISO-2" {
		t.Errorf("TASK-1 = %q, want BISO-2 (the simple id, always assigned first on a tie)", got)
	}
	if got := findByID(t, out, "TASK-1.0").ID; got != "BISO-3" {
		t.Errorf("TASK-1.0 = %q, want BISO-3 (its own subtask, assigned second)", got)
	}
}

// TestIdentifiersRewritesAParentThatWasReassigned covers
// docs/especificacion.md, "Identificadores", point 7: a parent naming a
// task that itself gets reassigned (TASK-1.1, a subtask, always reassigned
// per point 4) must be rewritten to that task's NEW final id, not left
// pointing at a stale one.
func TestIdentifiersRewritesAParentThatWasReassigned(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "parent.md", ID: "TASK-1.1", Title: "Parent"}, Result: Result{}},
		{
			Task:   source.Task{File: "child.md", ID: "TASK-2", Title: "Child", ParentTaskID: "TASK-1.1"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	reassignedParentID := findByID(t, out, "TASK-1.1").ID
	if got := findByID(t, out, "TASK-2").Parent; got != reassignedParentID {
		t.Errorf("Parent = %q, want %q (TASK-1.1's own reassigned id)", got, reassignedParentID)
	}
}

// TestIdentifiersRewritesMentionsAcrossFieldsInOnePass covers
// docs/especificacion.md, "Identificadores", point 5 over every text field
// it names, in a single batch that genuinely mixes a conserved id (TASK-99,
// whose number is free on the destination) and a reassigned one (TASK-12,
// whose number is already taken by an existing destination task): every
// mention of either one, in the title, the description, an acceptance
// criterion, and a comment, must resolve to the right final id.
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
	board := destination.Board{
		Config: identifiersConfig,
		Tasks:  []destination.Task{{ID: "BISO-12", Title: "Unrelated", CreatedAt: "2020-01-01T00:00:00Z"}},
	}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reassignedID := findByID(t, out, "TASK-12").ID
	if reassignedID == "BISO-12" {
		t.Fatalf("TASK-12 was not actually reassigned, this test proves nothing: ID = %q", reassignedID)
	}
	if got := findByID(t, out, "TASK-99").ID; got != "BISO-99" {
		t.Fatalf("TASK-99 = %q, want BISO-99 (conserved, its number was free)", got)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want exactly 1 (the TASK-12 reassignment): %v", len(findings), findings)
	}

	got := findByID(t, out, "TASK-12")
	wantTitle := fmt.Sprintf("Fix %s follow-up", reassignedID)
	if got.Title != wantTitle {
		t.Errorf("Title = %q, want %q", got.Title, wantTitle)
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

// TestIdentifiersRewritesMentionsInVariousTextContexts covers
// docs/especificacion.md, "Identificadores", point 5's rule that a mention
// is rewritten anywhere in the text, code blocks included: a mention
// surrounded by parentheses, backticks, a multi-line code fence, or
// immediately followed by a sentence-ending period, is still found and
// substituted, because none of those characters are letters, digits, '_',
// or '-'.
func TestIdentifiersRewritesMentionsInVariousTextContexts(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "target.md", ID: "TASK-5", Title: "Target"}, Result: Result{}},
		{Task: source.Task{File: "sub.md", ID: "TASK-9.1", Title: "Sub"}, Result: Result{}},
		{
			Task: source.Task{
				File: "c.md", ID: "TASK-1",
				Description: "See (TASK-5) and `TASK-5` for context.\n" +
					"```\nrelates to TASK-5 in code\n```\n" +
					"Blocked by TASK-9.1.",
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// TASK-9.1 is a subtask, so it is always reassigned (point 4), which
	// itself raises exactly one Finding (point 8): that is expected here
	// and unrelated to what this test actually checks, mention rewriting.
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want exactly 1 (TASK-9.1's own reassignment): %v", len(findings), findings)
	}

	subID := findByID(t, out, "TASK-9.1").ID
	want := "See (BISO-5) and `BISO-5` for context.\n" +
		"```\nrelates to BISO-5 in code\n```\n" +
		"Blocked by " + subID + "."
	got := findByID(t, out, "TASK-1").Description
	if got != want {
		t.Errorf("Description = %q, want %q", got, want)
	}
}

// TestIdentifiersRewritesAMentionByCanonicalNumberIgnoringLeadingZeros
// covers docs/especificacion.md, "Identificadores", point 5's canonical-form
// lookup and docs/decisiones.md's paragraph on mentions and canonical form:
// a source task recorded as "TASK-001" is mentioned elsewhere in the batch
// as "TASK-1", without the leading zero. The definitive rewrite must still
// resolve that mention to TASK-001's final id, exactly the case a
// string-exact lookup in the equivalence table would miss.
func TestIdentifiersRewritesAMentionByCanonicalNumberIgnoringLeadingZeros(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "target.md", ID: "TASK-001", Title: "Target"}, Result: Result{}},
		{
			Task:   source.Task{File: "mentioner.md", ID: "TASK-2", Title: "See TASK-1 for details"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0 (TASK-1 must resolve, not be reported as unrecognized): %v", len(findings), findings)
	}

	targetID := findByID(t, out, "TASK-001").ID
	got := findByID(t, out, "TASK-2").Title
	want := fmt.Sprintf("See %s for details", targetID)
	if got != want {
		t.Errorf("Title = %q, want %q (TASK-1 must resolve to TASK-001's final id by canonical number)", got, want)
	}
}

// TestIdentifiersRewritesAPaddedMentionOfAnUnpaddedSourceTask covers the
// other direction of the same canonical-form lookup: the source task itself
// is recorded WITHOUT padding ("TASK-1"), but the text mentions it WITH
// padding ("TASK-001"). This is the direction that actually proves the
// lookup canonicalizes the mention, not just the table: the equivalence
// table's canonical key for "TASK-1" is already "TASK-1" (canonicalIDKey
// always strips padding), so a mention that happens to already be
// unpadded (like TASK-1 mentioning TASK-1) would still match even a plain
// exact-string lookup against that table by coincidence; only a PADDED
// mention forces the lookup itself to canonicalize before searching.
func TestIdentifiersRewritesAPaddedMentionOfAnUnpaddedSourceTask(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "target.md", ID: "TASK-1", Title: "Target"}, Result: Result{}},
		{
			Task:   source.Task{File: "mentioner.md", ID: "TASK-2", Title: "See TASK-001 for details"},
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

	targetID := findByID(t, out, "TASK-1").ID
	got := findByID(t, out, "TASK-2").Title
	want := fmt.Sprintf("See %s for details", targetID)
	if got != want {
		t.Errorf("Title = %q, want %q (a padded mention TASK-001 must resolve to the unpadded source task TASK-1)", got, want)
	}
}

// TestIdentifiersRewritesAPaddedSubtaskMentionOfAnUnpaddedSubtask covers the
// same canonical-form lookup for a subtask, in the direction that actually
// distinguishes it from a literal-string match by coincidence: the source
// subtask is recorded WITHOUT padding ("TASK-1.2"), but the text mentions it
// WITH padding on the subtask number ("TASK-1.02"). A version that looked up
// the mention's own literal text in a canonically-keyed index (instead of
// canonicalizing the mention before searching) would miss this, the same
// way TestIdentifiersRewritesAPaddedMentionOfAnUnpaddedSourceTask catches
// that gap for a simple id: TASK-1.02's canonical key and TASK-1.2's
// canonical key must be the SAME key before the lookup even happens, or
// this never matches.
func TestIdentifiersRewritesAPaddedSubtaskMentionOfAnUnpaddedSubtask(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "sub.md", ID: "TASK-1.2", Title: "Sub"}, Result: Result{}},
		{
			Task:   source.Task{File: "mentioner.md", ID: "TASK-3", Title: "Blocked by TASK-1.02"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// The only Finding expected here is TASK-1.2's own reassignment (every
	// subtask is always reassigned, point 4); the mention itself must not
	// add an "unresolved mention" Finding.
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want exactly 1 (TASK-1.2's own reassignment): %v", len(findings), findings)
	}

	subID := findByID(t, out, "TASK-1.2").ID
	got := findByID(t, out, "TASK-3").Title
	want := fmt.Sprintf("Blocked by %s", subID)
	if got != want {
		t.Errorf("Title = %q, want %q (a padded mention TASK-1.02 must resolve to the unpadded subtask TASK-1.2)", got, want)
	}
}

// TestRewriteMentionsSubstitutesFromTheOriginalTextOnly is a direct,
// deterministic proof of docs/especificacion.md, "Identificadores", point
// 5's single-pass substitution requirement, exercised at rewriteMentions
// itself: TASK-1's equivalent is TASK-2, and TASK-2's own equivalent is
// TASK-3. If substitution rescanned its own output, the "TASK-2" written in
// place of "TASK-1" would be found again and turned into "TASK-3",
// corrupting the result.
//
// This exact shape of coincidence, a task's equivalent looking like another
// task's own original id, is not just a hypothetical for this isolated
// unit test: TestIdentifiersMentionsOfASkippedTaskAndAReassignedTaskNeverCrossSubstitute
// reproduces it end to end, through Identifiers itself, using a destination
// that shares the source's own prefix.
func TestRewriteMentionsSubstitutesFromTheOriginalTextOnly(t *testing.T) {
	pattern := mentionPattern("TASK")
	equivalentsByCanonical := map[string]string{
		"TASK-1": "TASK-2",
		"TASK-2": "TASK-3",
	}

	got, unresolved, _ := rewriteMentions("See TASK-1 and TASK-2", pattern, "TASK", equivalentsByCanonical, nil)

	want := "See TASK-2 and TASK-3"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if unresolved != 0 {
		t.Errorf("unresolved = %d, want 0", unresolved)
	}
}

// TestIdentifiersMentionsOfASkippedTaskAndAReassignedTaskNeverCrossSubstitute
// reproduces, end to end through Identifiers, the coincidence
// TestRewriteMentionsSubstitutesFromTheOriginalTextOnly proves at the helper
// level: it really happens when the destination shares the source's own
// prefix. Task A is skipped because it is already on the destination, and
// its equivalent is the id it ALREADY has there, "TASK-5". Task B's own
// source id happens to BE "TASK-5" too (the "5" is the destination's
// existing task number, unrelated to Task B), which collides with that same
// destination task and gets Task B reassigned to a new number. A mention of
// Task A's original id and a mention of Task B's original id, side by side
// in Task C's title, must each resolve independently, straight from the
// original text: applying the equivalence table with separate sequential
// replacements instead of a single pass would resolve "TASK-77" to "TASK-5"
// first, then wrongly catch that very "TASK-5" a second time while applying
// Task B's own key ("TASK-5" -> its reassigned id).
func TestIdentifiersMentionsOfASkippedTaskAndAReassignedTaskNeverCrossSubstitute(t *testing.T) {
	sharedPrefixConfig := destination.Config{TaskPrefix: "TASK"}
	batch := []TaskInput{
		{
			Task:   source.Task{File: "a.md", ID: "TASK-77", Title: "Repeat"},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "b.md", ID: "TASK-5", Title: "B"},
			Result: Result{CreatedAt: "2026-02-01T00:00:00Z"},
		},
		{
			Task: source.Task{
				File: "c.md", ID: "TASK-6",
				Title: "Mentions TASK-77 and TASK-5",
			},
			Result: Result{CreatedAt: "2026-03-01T00:00:00Z"},
		},
	}
	board := destination.Board{
		Config: sharedPrefixConfig,
		Tasks:  []destination.Task{{ID: "TASK-5", Title: "Repeat", CreatedAt: "2026-01-01T00:00:00Z"}},
	}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	reassignedB := findByID(t, out, "TASK-5").ID
	if reassignedB == "TASK-5" {
		t.Fatalf("Task B was not actually reassigned, this test proves nothing: ID = %q", reassignedB)
	}

	want := fmt.Sprintf("Mentions TASK-5 and %s", reassignedB)
	got := findByID(t, out, "TASK-6").Title
	if got != want {
		t.Errorf(
			"Title = %q, want %q (TASK-77 resolves to the skipped task's existing id, "+
				"TASK-5 resolves to Task B's own reassigned id, independently)",
			got, want,
		)
	}
}

// TestIdentifiersDoesNotTouchABranchNameLookingMention covers
// docs/especificacion.md, "Identificadores", point 5's explicit
// "TASK-10-modelo" example: not preceded by a letter/digit/_/-, but
// followed by '-' then a letter, so it is left alone.
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

// TestIdentifiersDoesNotTouchASubtaskPrefixLookAlike covers
// docs/especificacion.md, "Identificadores", point 5's other explicit
// example: "SUBTASK-12" is never touched, even when the batch has a real
// "TASK-12", because the 'T' of "TASK-12" is preceded by the letter 'B' of
// "SUB".
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
// docs/especificacion.md, "Identificadores", point 6: a mention with the
// right shape that names no task in the batch is left as-is and reported
// once per file and field, with a count, not once per mention.
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

// TestIdentifiersAWrongCaseMentionIsNeverSubstitutedAndIsGrouped covers
// docs/especificacion.md, "Identificadores", point 6's "Xyz-002"/"task-12"
// example: the pattern only matches exact uppercase, so a case variant is
// never a real match, but it still falls into the same grouped finding as
// an unresolved mention.
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

// TestIdentifiersRecognizesAParentByCanonicalNumberIgnoringLeadingZeros
// covers docs/especificacion.md, "Identificadores", point 7's full canonical
// resolution: a parent_task_id of "TASK-1" actually names a source task
// recorded as "TASK-001". They must be recognized as the same task, with no
// "parent not found" Finding.
func TestIdentifiersRecognizesAParentByCanonicalNumberIgnoringLeadingZeros(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "parent.md", ID: "TASK-001", Title: "Parent"}, Result: Result{}},
		{
			Task:   source.Task{File: "child.md", ID: "TASK-2", Title: "Child", ParentTaskID: "TASK-1"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0 (TASK-1 must recognize TASK-001, not be dropped as unknown): %v", len(findings), findings)
	}

	parentID := findByID(t, out, "TASK-001").ID
	if got := findByID(t, out, "TASK-2").Parent; got != parentID {
		t.Errorf("Parent = %q, want %q (parent_task_id: TASK-1 must recognize TASK-001 by canonical number)", got, parentID)
	}
}

// TestIdentifiersRecognizesADependencyWithALowercasePrefix covers point 7's
// other half: unlike a free-text mention (point 5), which is case-sensitive
// on purpose to avoid matching a branch name by accident, a dependency
// field is never free text, so its prefix is compared without regard to
// case. A dependency written "task-1" (lowercase) must still recognize a
// source task recorded as "TASK-001".
func TestIdentifiersRecognizesADependencyWithALowercasePrefix(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "target.md", ID: "TASK-001", Title: "Target"}, Result: Result{}},
		{
			Task: source.Task{
				File: "b.md", ID: "TASK-2", Title: "B",
				Dependencies: []string{"task-1"},
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0 (a dependency's prefix case must not matter): %v", len(findings), findings)
	}

	targetID := findByID(t, out, "TASK-001").ID
	want := []string{targetID}
	if got := findByID(t, out, "TASK-2").Dependencies; !equalStrings(got, want) {
		t.Errorf("Dependencies = %v, want %v (lowercase task-1 must still recognize TASK-001)", got, want)
	}
}

// TestIdentifiersDeduplicatesDependenciesThatResolveToTheSameFinalId covers
// point 7's deduplication: "TASK-1" and "TASK-001" in the same
// Dependencies list resolve to the same final id by canonical form, so only
// the first is kept.
func TestIdentifiersDeduplicatesDependenciesThatResolveToTheSameFinalId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "target.md", ID: "TASK-1", Title: "Target"}, Result: Result{}},
		{
			Task: source.Task{
				File: "b.md", ID: "TASK-2", Title: "B",
				Dependencies: []string{"TASK-1", "TASK-001"},
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	targetID := findByID(t, out, "TASK-1").ID
	want := []string{targetID}
	if got := findByID(t, out, "TASK-2").Dependencies; !equalStrings(got, want) {
		t.Errorf("Dependencies = %v, want %v (TASK-1 and TASK-001 resolve to the same task, kept once)", got, want)
	}
}

// TestIdentifiersDropsAnUnknownParentWithAFinding covers
// docs/especificacion.md, "Identificadores", point 7, for parent: a parent
// with the shape of an id that does not correspond to any task in the
// batch, not even by canonical form, is still dropped with a Finding.
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

// TestIdentifiersDropsOneUnknownDependencyKeepingTheRest covers
// docs/especificacion.md, "Identificadores", point 7, for dependencies:
// only the unresolvable element is removed, TASK-999 has the shape of an id
// but does not correspond to any task in the batch, not even by canonical
// form.
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

// TestIdentifiersRejectsAMalformedId covers docs/especificacion.md,
// "Identificadores", point 1: an id with the wrong shape aborts with an
// error, not a Finding.
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

// TestIdentifiersRejectsTwoDifferentPrefixes covers docs/especificacion.md,
// "Identificadores", point 1: two source tasks whose ids do not share a
// prefix abort with an error.
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

// TestIdentifiersRejectsADuplicateId covers docs/especificacion.md,
// "Identificadores", point 1: two tasks with the exact same id abort with
// an error.
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

// TestIdentifiersRejectsIdsThatCanonicalizeToTheSameSimpleId covers
// docs/especificacion.md, "Identificadores", point 1's last paragraph:
// "TASK-1" and "TASK-001" are literally different strings but the same id
// once leading zeros are ignored, so this aborts exactly like an exact
// duplicate would.
func TestIdentifiersRejectsIdsThatCanonicalizeToTheSameSimpleId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-001", Title: "B"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one about TASK-1 and TASK-001 being the same id")
	}
	if !strings.Contains(err.Error(), "TASK-1") || !strings.Contains(err.Error(), "TASK-001") {
		t.Errorf("error = %v, want it to mention both TASK-1 and TASK-001", err)
	}
}

// TestIdentifiersRejectsSubtaskIdsThatCanonicalizeToTheSameId covers
// docs/especificacion.md, "Identificadores", point 1's last paragraph for a
// subtask: "TASK-1.2" and "TASK-1.02" are the same id once leading zeros
// are ignored on both the main and the subtask number.
func TestIdentifiersRejectsSubtaskIdsThatCanonicalizeToTheSameId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1.2", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-1.02", Title: "B"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one about TASK-1.2 and TASK-1.02 being the same id")
	}
	if !strings.Contains(err.Error(), "TASK-1.2") || !strings.Contains(err.Error(), "TASK-1.02") {
		t.Errorf("error = %v, want it to mention both TASK-1.2 and TASK-1.02", err)
	}
}

// TestIdentifiersNeverTreatsASimpleIdAndItsSubtaskAsTheSameId covers
// docs/especificacion.md, "Identificadores", point 1's explicit carve-out:
// "TASK-1" and "TASK-1.2" share the same main number but are never the same
// id, since one is simple and the other is a subtask. Both must validate
// successfully, side by side in the same batch.
func TestIdentifiersNeverTreatsASimpleIdAndItsSubtaskAsTheSameId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-1.2", Title: "B"}, Result: Result{}},
	}
	out, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err != nil {
		t.Fatalf("unexpected error: %v, want TASK-1 and TASK-1.2 to both be valid, distinct ids", err)
	}
	if len(out) != 2 {
		t.Fatalf("out = %v, want 2 tasks", out)
	}
}

// ---------------------------------------------------------------------------
// docs/especificacion.md, "Identificadores", point 1's id-reuse-after-
// archiving case (Backlog.md handing an archived task's number to the next
// one it creates), and points 4, 5, and 7's handling of it.
//
// Two copies that share a reused id also share the exact literal SourceID
// string (Backlog.md hands the reused task the very same id), so findByID
// cannot tell them apart. These tests index into Identifiers' own output
// slice directly instead: Identifiers documents that it returns one
// Identified per non-skipped batch entry IN THE SAME RELATIVE ORDER as
// batch, and none of these tests' tasks are skipped (point 3), so out[i]
// is always batch[i]'s own result.
// ---------------------------------------------------------------------------

// TestIdentifiersAnActiveAndAnArchivedCopyOfTheSameIdIsNotFatal covers point
// 1's central case: exactly one non-archived copy of a shared simple id is
// never fatal. The active copy conserves its number under the normal rule
// (point 2, free here), and the archived copy is reassigned with the new
// "id reused after archiving" reason, never the ordinary collision wording.
func TestIdentifiersAnActiveAndAnArchivedCopyOfTheSameIdIsNotFatal(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-2", Title: "Active"}, Result: Result{}},
		{Task: source.Task{File: "archived.md", ID: "TASK-2", Title: "Archived", Archived: true}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v, want the shared id to be tolerated (only one copy is non-archived)", err)
	}
	if len(out) != 2 {
		t.Fatalf("out = %v, want both copies emitted", out)
	}

	active, archived := out[0], out[1]
	if active.ID != "BISO-2" {
		t.Errorf("active.ID = %q, want BISO-2 (conserved, its number is free)", active.ID)
	}
	if archived.ID == "BISO-2" || archived.ID == "" {
		t.Fatalf("archived.ID = %q, want a freshly reassigned id, distinct from the active copy's", archived.ID)
	}

	want := fmt.Sprintf("TASK-2: id %s assigned (id reused after archiving)", archived.ID)
	if _, ok := findingWithSubstring(findings, want); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, want)
	}
	if _, ok := findingWithSubstring(findings, "is taken on the destination"); ok {
		t.Errorf("findings = %v, want no ordinary collision wording for the archived copy", findings)
	}
}

// TestIdentifiersMentionResolvesToTheNonArchivedCopyOfASharedId covers
// docs/especificacion.md, "Identificadores", point 5's first ambiguous case:
// a free-text mention of a shared id with exactly one non-archived copy
// resolves to that copy's own final id, with the exact Finding wording point
// 5 fixes.
func TestIdentifiersMentionResolvesToTheNonArchivedCopyOfASharedId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-2", Title: "Active"}, Result: Result{}},
		{Task: source.Task{File: "archived.md", ID: "TASK-2", Title: "Archived", Archived: true}, Result: Result{}},
		{Task: source.Task{File: "c.md", ID: "TASK-9", Title: "Mentions TASK-2"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	activeID := out[0].ID
	if activeID != "BISO-2" {
		t.Fatalf("active.ID = %q, want BISO-2", activeID)
	}

	mentioner := out[2]
	want := fmt.Sprintf("Mentions %s", activeID)
	if mentioner.Title != want {
		t.Errorf("Title = %q, want %q", mentioner.Title, want)
	}

	wantFinding := fmt.Sprintf("TASK-2: mention resolved to %s, the non-archived task sharing this id", activeID)
	if _, ok := findingWithSubstring(findings, wantFinding); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, wantFinding)
	}
}

// TestIdentifiersThreeArchivedCopiesMentionResolvesToTheMostRecentDate
// covers docs/especificacion.md, "Identificadores", point 5's second
// ambiguous case: with no non-archived copy at all, all three are
// reassigned, and a mention resolves to the copy with the most recent
// created_date, with the exact "archived task selected among those sharing
// this id" wording.
func TestIdentifiersThreeArchivedCopiesMentionResolvesToTheMostRecentDate(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "old.md", ID: "TASK-2", Title: "Old", Archived: true},
			Result: Result{CreatedAt: "2020-01-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "mid.md", ID: "TASK-2", Title: "Mid", Archived: true},
			Result: Result{CreatedAt: "2022-06-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "new.md", ID: "TASK-2", Title: "New", Archived: true},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{Task: source.Task{File: "mentioner.md", ID: "TASK-9", Title: "Mentions TASK-2"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 4 {
		t.Fatalf("out = %v, want 4 rows, all three copies reassigned and emitted", out)
	}

	oldID, midID, newID := out[0].ID, out[1].ID, out[2].ID
	ids := map[string]bool{oldID: true, midID: true, newID: true}
	if len(ids) != 3 {
		t.Fatalf("the three archived copies must get three DISTINCT ids: got %q, %q, %q", oldID, midID, newID)
	}
	for _, id := range []string{oldID, midID, newID} {
		f := fmt.Sprintf("TASK-2: id %s assigned (id reused after archiving)", id)
		if _, ok := findingWithSubstring(findings, f); !ok {
			t.Errorf("findings = %v, want one containing %q", findings, f)
		}
	}

	mentioner := out[3]
	want := fmt.Sprintf("Mentions %s", newID)
	if mentioner.Title != want {
		t.Errorf("Title = %q, want %q (must resolve to the most recently created copy, new.md)", mentioner.Title, want)
	}

	wantFinding := fmt.Sprintf("TASK-2: mention resolved to %s, the archived task selected among those sharing this id", newID)
	if _, ok := findingWithSubstring(findings, wantFinding); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, wantFinding)
	}
}

// TestIdentifiersTwoArchivedCopiesWithTheSameExactDateTieBreakIsStable
// covers docs/especificacion.md, "Identificadores", point 4's second tie-
// break level: when two archived copies of a shared id tie on the exact
// same created_date, the synthetic Archived+File key decides, and it must
// give the same answer every time Identifiers runs over the same batch.
func TestIdentifiersTwoArchivedCopiesWithTheSameExactDateTieBreakIsStable(t *testing.T) {
	batch := []TaskInput{
		{
			Task:   source.Task{File: "a.md", ID: "TASK-2", Title: "A", Archived: true},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "z.md", ID: "TASK-2", Title: "Z", Archived: true},
			Result: Result{CreatedAt: "2026-01-01T00:00:00Z"},
		},
		{Task: source.Task{File: "mentioner.md", ID: "TASK-9", Title: "Mentions TASK-2"}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	var mentionsAcrossRuns []string
	for run := 0; run < 5; run++ {
		out, _, err := Identifiers(batch, board)
		if err != nil {
			t.Fatalf("run %d: unexpected error: %v", run, err)
		}
		mentionsAcrossRuns = append(mentionsAcrossRuns, out[2].Title)
	}
	for i, got := range mentionsAcrossRuns {
		if got != mentionsAcrossRuns[0] {
			t.Fatalf("run %d resolved the tie differently (%q) than run 0 (%q), want a stable, deterministic pick", i, got, mentionsAcrossRuns[0])
		}
	}
}

// TestIdentifiersSharedSubtaskIdIsAlwaysReassignedNeverIdReusedAfterArchiving
// covers docs/especificacion.md, "Identificadores", point 4's explicit
// carve-out: a SUBTASK id shared by a non-archived and an archived copy
// reassigns BOTH, neither with "id reused after archiving" (that reason only
// ever applies to a SIMPLE shared id, point 8).
func TestIdentifiersSharedSubtaskIdIsAlwaysReassignedNeverIdReusedAfterArchiving(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-5.1", Title: "Active sub"}, Result: Result{}},
		{Task: source.Task{File: "archived.md", ID: "TASK-5.1", Title: "Archived sub", Archived: true}, Result: Result{}},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("out = %v, want both copies emitted", out)
	}

	activeID, archivedID := out[0].ID, out[1].ID
	if activeID == archivedID || activeID == "" || archivedID == "" {
		t.Fatalf("both subtask copies must get their own distinct reassigned id: got %q and %q", activeID, archivedID)
	}

	for _, id := range []string{activeID, archivedID} {
		f := fmt.Sprintf("id %s assigned (subtask ids have no equivalent)", id)
		if _, ok := findingWithSubstring(findings, f); !ok {
			t.Errorf("findings = %v, want one containing %q", findings, f)
		}
	}
	if _, ok := findingWithSubstring(findings, "id reused after archiving"); ok {
		t.Errorf("findings = %v, want no \"id reused after archiving\" wording for a subtask", findings)
	}
}

// TestIdentifiersActiveAndTwoArchivedCopiesShareAnId covers
// docs/especificacion.md, "Identificadores", point 4: the active copy
// conserves its number, and the two archived copies are reassigned in the
// point-4 tie-break order (here, by created_date ascending), each getting
// its own consecutive number.
func TestIdentifiersActiveAndTwoArchivedCopiesShareAnId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-3", Title: "Active"}, Result: Result{}},
		{
			Task:   source.Task{File: "older.md", ID: "TASK-3", Title: "Older", Archived: true},
			Result: Result{CreatedAt: "2020-01-01T00:00:00Z"},
		},
		{
			Task:   source.Task{File: "newer.md", ID: "TASK-3", Title: "Newer", Archived: true},
			Result: Result{CreatedAt: "2021-01-01T00:00:00Z"},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, _, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out) != 3 {
		t.Fatalf("out = %v, want all three copies emitted", out)
	}

	active, older, newer := out[0], out[1], out[2]
	if active.ID != "BISO-3" {
		t.Errorf("active.ID = %q, want BISO-3 (conserved)", active.ID)
	}
	if older.ID != "BISO-4" {
		t.Errorf("older.ID = %q, want BISO-4 (reassigned first: older created_date)", older.ID)
	}
	if newer.ID != "BISO-5" {
		t.Errorf("newer.ID = %q, want BISO-5 (reassigned second: newer created_date)", newer.ID)
	}
}

// TestIdentifiersParentResolvesToTheNonArchivedCopyOfASharedId covers
// docs/especificacion.md, "Identificadores", point 7's reuse of point 5's
// ambiguous resolution for parent_task_id.
func TestIdentifiersParentResolvesToTheNonArchivedCopyOfASharedId(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-2", Title: "Active"}, Result: Result{}},
		{Task: source.Task{File: "archived.md", ID: "TASK-2", Title: "Archived", Archived: true}, Result: Result{}},
		{
			Task:   source.Task{File: "child.md", ID: "TASK-9", Title: "Child", ParentTaskID: "TASK-2"},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	activeID := out[0].ID
	child := out[2]
	if child.Parent != activeID {
		t.Errorf("Parent = %q, want %q (the non-archived copy)", child.Parent, activeID)
	}

	want := fmt.Sprintf("TASK-2: parent resolved to %s, the non-archived task sharing this id", activeID)
	if _, ok := findingWithSubstring(findings, want); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, want)
	}
}

// TestIdentifiersTwoDependenciesResolvingToTheSameSharedIdAreDeduplicated
// covers docs/especificacion.md, "Identificadores", point 7's deduplication
// rule combined with point 1's reuse case: two Dependencies elements both
// naming the shared id "TASK-2" resolve to the SAME final id (the
// non-archived copy's), so only one survives, and the ambiguous-resolution
// Finding is grouped into one line with a count of 2.
func TestIdentifiersTwoDependenciesResolvingToTheSameSharedIdAreDeduplicated(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "active.md", ID: "TASK-2", Title: "Active"}, Result: Result{}},
		{Task: source.Task{File: "archived.md", ID: "TASK-2", Title: "Archived", Archived: true}, Result: Result{}},
		{
			Task: source.Task{
				File: "b.md", ID: "TASK-9", Title: "B",
				Dependencies: []string{"TASK-2", "TASK-002"},
			},
			Result: Result{},
		},
	}
	board := destination.Board{Config: identifiersConfig}

	out, findings, err := Identifiers(batch, board)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	activeID := out[0].ID
	dependent := out[2]
	want := []string{activeID}
	if !equalStrings(dependent.Dependencies, want) {
		t.Errorf("Dependencies = %v, want %v (both elements resolve to the same non-archived copy, kept once)", dependent.Dependencies, want)
	}

	wantFinding := fmt.Sprintf("2 dependency(s) of TASK-2 resolve to %s, the non-archived task sharing this id", activeID)
	if _, ok := findingWithSubstring(findings, wantFinding); !ok {
		t.Fatalf("findings = %v, want one containing %q", findings, wantFinding)
	}
}

// TestIdentifiersTwoNonArchivedCopiesOfASharedIdRemainFatal covers
// docs/especificacion.md, "Identificadores", point 1's own carve-out for the
// carve-out: MORE THAN ONE non-archived copy of the same id is still the one
// fatal case, with the exact same error wording as before this file started
// tolerating a single archived copy.
func TestIdentifiersTwoNonArchivedCopiesOfASharedIdRemainFatal(t *testing.T) {
	batch := []TaskInput{
		{Task: source.Task{File: "a.md", ID: "TASK-1", Title: "A"}, Result: Result{}},
		{Task: source.Task{File: "b.md", ID: "TASK-1", Title: "B"}, Result: Result{}},
	}
	_, _, err := Identifiers(batch, destination.Board{Config: identifiersConfig})
	if err == nil {
		t.Fatal("got nil error, want one about the two non-archived copies sharing an id")
	}
	if !strings.Contains(err.Error(), "TASK-1") {
		t.Errorf("error = %v, want it to mention TASK-1", err)
	}
}

// TestAssembleNoLongerFailsOnAnArchivedAndANonArchivedCopyOfTheSameId is an
// end-to-end smoke test through Assemble (phase 5), reusing the existing
// "FIX-4" fixture (internal/source/testdata/backlog-board): one copy in
// completed/ (non-archived) and one in archive/tasks/ (archived), both with
// the literal id "FIX-4", reproducing the exact real-world sequence
// docs/decisiones.md describes (archive a task, let Backlog.md reuse its
// number). Before this fix, validateSourceShape rejected this as two files
// sharing an id; this test only confirms that fatal error is gone, not the
// full field-level correctness of every Line Assemble produces for this
// board.
func TestAssembleNoLongerFailsOnAnArchivedAndANonArchivedCopyOfTheSameId(t *testing.T) {
	sourceBoard, err := source.Read(filepath.Join("..", "source", "testdata", "backlog-board"))
	if err != nil {
		t.Fatalf("source.Read: %v", err)
	}
	destBoard := destination.Board{Config: identifiersConfig}

	if _, _, err := Assemble(sourceBoard, destBoard); err != nil {
		t.Fatalf("Assemble returned an error, want none (the two FIX-4 copies must not be treated as a fatal duplicate id): %v", err)
	}
}
