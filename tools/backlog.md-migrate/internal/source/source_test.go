package source

import (
	"os"
	"path/filepath"
	"testing"
)

// findTask returns the task in tasks whose File matches name, failing the
// test if there is none.
func findTask(t *testing.T, tasks []Task, name string) Task {
	t.Helper()
	for _, task := range tasks {
		if task.File == name {
			return task
		}
	}
	t.Fatalf("no task with file %q, got files: %v", name, taskFiles(tasks))
	return Task{}
}

func taskFiles(tasks []Task) []string {
	var names []string
	for _, task := range tasks {
		names = append(names, task.File)
	}
	return names
}

// findingsFor returns every finding in findings whose File and Field match.
func findingsFor(findings []Finding, file, field string) []Finding {
	var out []Finding
	for _, f := range findings {
		if f.File == file && f.Field == field {
			out = append(out, f)
		}
	}
	return out
}

func readBoard(t *testing.T, dir string) Board {
	t.Helper()
	board, err := Read(filepath.Join("testdata", dir))
	if err != nil {
		t.Fatalf("Read(%q): %v", dir, err)
	}
	return board
}

func TestReadsANormalTaskWithAllFields(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-1 - Multi-field-task-FIX-1-mentions-itself.md")

	if task.Archived {
		t.Errorf("Archived = true, want false")
	}
	if task.ID != "FIX-1" {
		t.Errorf("ID = %q, want FIX-1", task.ID)
	}
	if task.Title != "Multi field task FIX-1 mentions itself" {
		t.Errorf("Title = %q", task.Title)
	}
	if task.Status != "To Do" {
		t.Errorf("Status = %q", task.Status)
	}
	wantAssignees := []string{"Sara Smith", "@ann"}
	if !equalSlices(task.Assignees, wantAssignees) {
		t.Errorf("Assignees = %v, want %v", task.Assignees, wantAssignees)
	}
	if task.CreatedDate != "2026-09-26 22:46" {
		t.Errorf("CreatedDate = %q", task.CreatedDate)
	}
	if task.UpdatedDate != "2026-09-26 22:47" {
		t.Errorf("UpdatedDate = %q", task.UpdatedDate)
	}
	if task.DueDate != "2026-10-01" {
		t.Errorf("DueDate = %q", task.DueDate)
	}
	wantLabels := []string{"backend", "with space"}
	if !equalSlices(task.Labels, wantLabels) {
		t.Errorf("Labels = %v, want %v", task.Labels, wantLabels)
	}
	if task.Milestone != "m-0" {
		t.Errorf("Milestone = %q", task.Milestone)
	}
	wantRefs := []string{"https://example.com/ref"}
	if !equalSlices(task.References, wantRefs) {
		t.Errorf("References = %v, want %v", task.References, wantRefs)
	}
	wantDocs := []string{"README.md"}
	if !equalSlices(task.Documentation, wantDocs) {
		t.Errorf("Documentation = %v, want %v", task.Documentation, wantDocs)
	}
	wantModified := []string{"src/main.go"}
	if !equalSlices(task.ModifiedFiles, wantModified) {
		t.Errorf("ModifiedFiles = %v, want %v", task.ModifiedFiles, wantModified)
	}
	if task.Priority != "high" {
		t.Errorf("Priority = %q", task.Priority)
	}
	if task.Type != "task" {
		t.Errorf("Type = %q", task.Type)
	}
	if task.Ordinal == nil || *task.Ordinal != 1000 {
		t.Errorf("Ordinal = %v, want 1000", task.Ordinal)
	}

	wantDescription := "Line one of the description.\n" +
		"Second line of the description with a heading below.\n" +
		"## Not a real section\n" +
		"More text after the fake heading."
	if task.Description != wantDescription {
		t.Errorf("Description = %q, want %q", task.Description, wantDescription)
	}
	if task.Plan != "A short plan" {
		t.Errorf("Plan = %q", task.Plan)
	}
	if task.Notes != "Some notes" {
		t.Errorf("Notes = %q", task.Notes)
	}
	if task.Summary != "Final summary text." {
		t.Errorf("Summary = %q", task.Summary)
	}

	wantAC := []Checkbox{
		{Number: 1, Checked: false, Text: "First criterion"},
		{Number: 2, Checked: false, Text: "Second criterion"},
	}
	if !equalCheckboxes(task.AcceptanceCriteria, wantAC) {
		t.Errorf("AcceptanceCriteria = %+v, want %+v", task.AcceptanceCriteria, wantAC)
	}

	if len(task.Comments) != 2 {
		t.Fatalf("len(Comments) = %d, want 2", len(task.Comments))
	}
	if task.Comments[0].Author != "@ann" || task.Comments[0].CreatedAt != "2026-09-26 22:47" || task.Comments[0].Body != "First comment body" {
		t.Errorf("Comments[0] = %+v", task.Comments[0])
	}
}

func TestDefinitionOfDoneReadSeparately(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-1 - Multi-field-task-FIX-1-mentions-itself.md")

	wantDOD := []Checkbox{{Number: 1, Checked: false, Text: "Done item one"}}
	if !equalCheckboxes(task.DefinitionOfDone, wantDOD) {
		t.Errorf("DefinitionOfDone = %+v, want %+v", task.DefinitionOfDone, wantDOD)
	}
	if len(task.AcceptanceCriteria) != 2 {
		t.Errorf("AcceptanceCriteria still has %d items, DefinitionOfDone must not merge into it", len(task.AcceptanceCriteria))
	}
}

func TestCommentWithoutAuthorHasNoAuthorAndNoFinding(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-1 - Multi-field-task-FIX-1-mentions-itself.md")

	if len(task.Comments) != 2 {
		t.Fatalf("len(Comments) = %d, want 2", len(task.Comments))
	}
	second := task.Comments[1]
	if second.Author != "" {
		t.Errorf("Comments[1].Author = %q, want empty", second.Author)
	}
	if second.CreatedAt != "2026-09-26 22:47" {
		t.Errorf("Comments[1].CreatedAt = %q", second.CreatedAt)
	}
	if second.Body != "Second comment without author" {
		t.Errorf("Comments[1].Body = %q", second.Body)
	}

	if fs := findingsFor(board.Findings, "fix-1 - Multi-field-task-FIX-1-mentions-itself.md", "created"); len(fs) != 0 {
		t.Errorf("a comment without author must not raise a finding, got %+v", fs)
	}
}

func TestACompletedTaskIsReadLikeANormalOne(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-4 - Task-to-complete.md")

	if task.Archived {
		t.Errorf("Archived = true for a completed task, want false")
	}
	if task.Title != "Task to complete" {
		t.Errorf("Title = %q", task.Title)
	}
	if task.Status != "Done" {
		t.Errorf("Status = %q, want Done", task.Status)
	}
}

func TestAnArchivedTaskIsMarkedArchived(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-4 - Task-to-archive.md")

	if !task.Archived {
		t.Errorf("Archived = false for a task read from archive/tasks/, want true")
	}
	if task.Title != "Task to archive" {
		t.Errorf("Title = %q", task.Title)
	}
}

func TestAMilestoneResolvesToItsTitle(t *testing.T) {
	board := readBoard(t, "backlog-board")

	if got := board.Milestones["m-0"]; got != "Release One" {
		t.Errorf(`Milestones["m-0"] = %q, want "Release One"`, got)
	}
	// m-1 sits in archive/milestones/, and a task (fix-2) still points to
	// it: both milestones/ and archive/milestones/ must be read.
	if got := board.Milestones["m-1"]; got != "Old Release" {
		t.Errorf(`Milestones["m-1"] = %q, want "Old Release"`, got)
	}

	task := findTask(t, board.Tasks, "fix-2 - Task-with-multiline-AC.md")
	if task.Milestone != "m-1" {
		t.Errorf("task Milestone = %q, want m-1", task.Milestone)
	}
}

func TestMultilineCriterionIsJoinedWithSpaceAndRaisesAFinding(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-2 - Task-with-multiline-AC.md")

	wantAC := []Checkbox{{Number: 1, Checked: false, Text: "Multi line criterion second physical line"}}
	if !equalCheckboxes(task.AcceptanceCriteria, wantAC) {
		t.Errorf("AcceptanceCriteria = %+v, want %+v", task.AcceptanceCriteria, wantAC)
	}

	fs := findingsFor(board.Findings, "fix-2 - Task-with-multiline-AC.md", "acceptanceCriteria")
	if len(fs) != 1 {
		t.Fatalf("findings for acceptanceCriteria = %+v, want exactly one", fs)
	}
}

func TestCheckedAndUncheckedCriteriaAndProjectAndDependencies(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-3 - Task-with-project-and-dependency.md")

	wantAC := []Checkbox{
		{Number: 1, Checked: true, Text: "Checked one"},
		{Number: 2, Checked: false, Text: "Checked two"},
	}
	if !equalCheckboxes(task.AcceptanceCriteria, wantAC) {
		t.Errorf("AcceptanceCriteria = %+v, want %+v", task.AcceptanceCriteria, wantAC)
	}
	if task.Project != "alpha" {
		t.Errorf("Project = %q, want alpha", task.Project)
	}
	if !equalSlices(task.Dependencies, []string{"FIX-1"}) {
		t.Errorf("Dependencies = %v, want [FIX-1]", task.Dependencies)
	}
}

func TestSubtaskHasParentTaskID(t *testing.T) {
	board := readBoard(t, "backlog-board")
	task := findTask(t, board.Tasks, "fix-1.1 - Subtask-of-FIX-1.md")

	if task.ParentTaskID != "FIX-1" {
		t.Errorf("ParentTaskID = %q, want FIX-1", task.ParentTaskID)
	}
}

func TestDraftsDocsAndDecisionsFoldersProduceACountFinding(t *testing.T) {
	board := readBoard(t, "backlog-board")

	for _, field := range []string{"drafts", "docs", "decisions"} {
		fs := findingsFor(board.Findings, "-", field)
		if len(fs) != 1 {
			t.Errorf("findings for field %q = %+v, want exactly one", field, fs)
		}
	}
}

func TestEmptyOrAbsentDraftsDocsDecisionsProduceNoFinding(t *testing.T) {
	dir := t.TempDir()
	backlogDir := filepath.Join(dir, "backlog")
	mustMkdirAll(t, filepath.Join(backlogDir, "tasks"))
	copyFixtureFile(t,
		filepath.Join("testdata", "backlog-board", "tasks", "fix-3 - Task-with-project-and-dependency.md"),
		filepath.Join(backlogDir, "tasks", "task.md"),
	)
	// docs/ exists but is empty; drafts/ and decisions/ do not exist at
	// all. Neither shape should produce a finding.
	mustMkdirAll(t, filepath.Join(backlogDir, "docs"))

	board, err := Read(backlogDir)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	for _, field := range []string{"drafts", "docs", "decisions"} {
		if fs := findingsFor(board.Findings, "-", field); len(fs) != 0 {
			t.Errorf("field %q: got findings %+v, want none", field, fs)
		}
	}
}

func TestUnrecognizedFrontmatterKeyProducesAFindingAndIsNotKept(t *testing.T) {
	board := readBoard(t, "unknown-fields")
	task := findTask(t, board.Tasks, "quirks.md")

	if task.Title != "Task with a quirk the CLI cannot produce on its own" {
		t.Errorf("Title = %q", task.Title)
	}

	fs := findingsFor(board.Findings, "quirks.md", "custom_field")
	if len(fs) != 1 {
		t.Fatalf("findings for custom_field = %+v, want exactly one", fs)
	}
	if fs[0].Message == "" {
		t.Errorf("finding message is empty")
	}
}

func TestUnrecognizedBodySectionProducesAFinding(t *testing.T) {
	board := readBoard(t, "unknown-fields")

	fs := findingsFor(board.Findings, "quirks.md", "REVIEW")
	if len(fs) != 1 {
		t.Fatalf("findings for REVIEW section = %+v, want exactly one", fs)
	}
}

func TestInvalidDateShapeProducesAFindingAndTheFieldIsOmitted(t *testing.T) {
	board := readBoard(t, "unknown-fields")
	task := findTask(t, board.Tasks, "quirks.md")

	if task.DueDate != "" {
		t.Errorf("DueDate = %q, want empty because 01/10/2026 does not match either date shape", task.DueDate)
	}
	fs := findingsFor(board.Findings, "quirks.md", "due_date")
	if len(fs) != 1 {
		t.Fatalf("findings for due_date = %+v, want exactly one", fs)
	}
}

func TestCommentWithoutCreatedIsKeptWithoutADateAndRaisesAFinding(t *testing.T) {
	board := readBoard(t, "unknown-fields")
	task := findTask(t, board.Tasks, "quirks.md")

	if len(task.Comments) != 1 {
		t.Fatalf("len(Comments) = %d, want 1", len(task.Comments))
	}
	comment := task.Comments[0]
	if comment.Author != "@ann" {
		t.Errorf("Author = %q, want @ann", comment.Author)
	}
	if comment.CreatedAt != "" {
		t.Errorf("CreatedAt = %q, want empty", comment.CreatedAt)
	}
	if comment.Body != "A comment with no created line at all." {
		t.Errorf("Body = %q", comment.Body)
	}

	fs := findingsFor(board.Findings, "quirks.md", "created")
	if len(fs) != 1 {
		t.Fatalf("findings for created = %+v, want exactly one", fs)
	}
}

func TestBrokenFrontmatterIsSkippedAndTheRestOfTheBatchIsStillRead(t *testing.T) {
	board := readBoard(t, "broken-frontmatter")

	if len(board.Tasks) != 1 {
		t.Fatalf("len(Tasks) = %d (%v), want exactly 1 (the good sibling)", len(board.Tasks), taskFiles(board.Tasks))
	}
	if board.Tasks[0].File != "good-sibling.md" {
		t.Errorf("the surviving task is %q, want good-sibling.md", board.Tasks[0].File)
	}
	if board.Tasks[0].ID != "FIX-3" {
		t.Errorf("ID = %q, want FIX-3", board.Tasks[0].ID)
	}

	for _, name := range []string{"broken-quote.md", "broken-duplicate-key.md"} {
		fs := findingsFor(board.Findings, name, "frontmatter")
		if len(fs) != 1 {
			t.Fatalf("findings for %q field frontmatter = %+v, want exactly one", name, fs)
		}
		if fs[0].Message == "" {
			t.Errorf("%s: finding message is empty, want the real YAML parse error", name)
		}
	}
}

func TestUnclosedDescriptionMarkerIsAbsentAndRaisesAFindingWithoutAffectingSiblings(t *testing.T) {
	board := readBoard(t, "unclosed-sections")
	task := findTask(t, board.Tasks, "unclosed-description.md")

	if task.Description != "" {
		t.Errorf("Description = %q, want empty: an unclosed section must be treated as absent", task.Description)
	}

	fs := findingsFor(board.Findings, "unclosed-description.md", "DESCRIPTION")
	if len(fs) != 1 {
		t.Fatalf("findings for DESCRIPTION = %+v, want exactly one", fs)
	}
	wantMessage := "DESCRIPTION section marker was never closed"
	if fs[0].Message != wantMessage {
		t.Errorf("finding message = %q, want %q", fs[0].Message, wantMessage)
	}

	wantAC := []Checkbox{{Number: 1, Checked: false, Text: "First criterion"}}
	if !equalCheckboxes(task.AcceptanceCriteria, wantAC) {
		t.Errorf("AcceptanceCriteria = %+v, want %+v, a sibling section must still be read", task.AcceptanceCriteria, wantAC)
	}
	wantPlan := "A short plan that must still be read even though Description above it is\nbroken."
	if task.Plan != wantPlan {
		t.Errorf("Plan = %q, want %q", task.Plan, wantPlan)
	}
}

func TestUnclosedUnknownNameMarkerRaisesBothFindingsAndLeavesSiblingsAlone(t *testing.T) {
	board := readBoard(t, "unclosed-sections")
	task := findTask(t, board.Tasks, "unclosed-unknown-name.md")

	unknown := findingsFor(board.Findings, "unclosed-unknown-name.md", "REVIEW")
	if len(unknown) != 2 {
		t.Fatalf("findings for REVIEW = %+v, want exactly two (unrecognized body section, and unclosed)", unknown)
	}
	var gotUnrecognized, gotUnclosed bool
	for _, f := range unknown {
		switch f.Message {
		case "unrecognized body section":
			gotUnrecognized = true
		case "REVIEW section marker was never closed":
			gotUnclosed = true
		}
	}
	if !gotUnrecognized || !gotUnclosed {
		t.Errorf("findings for REVIEW = %+v, want one %q and one %q", unknown, "unrecognized body section", "REVIEW section marker was never closed")
	}

	wantDescription := "A normal description, unaffected by the broken Review section below."
	if task.Description != wantDescription {
		t.Errorf("Description = %q, want %q", task.Description, wantDescription)
	}
}

func TestAllMarkersClosedRaisesNoUnclosedFinding(t *testing.T) {
	board := readBoard(t, "unclosed-sections")
	task := findTask(t, board.Tasks, "all-markers-closed.md")

	for _, f := range board.Findings {
		if f.File != "all-markers-closed.md" {
			continue
		}
		t.Errorf("unexpected finding for a fully closed body: %+v", f)
	}

	if task.Description != "A normal, fully closed description." {
		t.Errorf("Description = %q", task.Description)
	}
	wantAC := []Checkbox{{Number: 1, Checked: false, Text: "First criterion"}}
	if !equalCheckboxes(task.AcceptanceCriteria, wantAC) {
		t.Errorf("AcceptanceCriteria = %+v, want %+v", task.AcceptanceCriteria, wantAC)
	}
	if task.Plan != "A short plan." {
		t.Errorf("Plan = %q", task.Plan)
	}
}

// --- test helpers ---

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func equalCheckboxes(a, b []Checkbox) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func mustMkdirAll(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", dir, err)
	}
}

func copyFixtureFile(t *testing.T, src, dst string) {
	t.Helper()
	content, err := os.ReadFile(src)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", src, err)
	}
	if err := os.WriteFile(dst, content, 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", dst, err)
	}
}
