package convert

import (
	"testing"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

var testConfig = destination.Config{
	TaskPrefix: "BISO",
	Statuses:   []string{"To Do", "In Progress", "Done"},
	Types:      []string{"bug", "feature"},
	Priorities: []string{"low", "high"},
}

func TestTaskConvertsVocabularyLabelsAndDates(t *testing.T) {
	task := source.Task{
		File:        "fix-1 - Title.md",
		Status:      "todo",
		Type:        "bug",
		Priority:    "HIGH",
		Labels:      []string{"with space", "backend"},
		Assignees:   []string{"Sara Smith"},
		CreatedDate: "2026-09-20 22:08",
		DueDate:     "2026-10-01",
		Milestone:   "m-4",
	}
	milestoneSlugs := map[string]string{"m-4": "puesta-en-uso"}

	result, findings := Task(task, milestoneSlugs, testConfig)

	if result.Status != "To Do" {
		t.Errorf("Status = %q, want To Do", result.Status)
	}
	if result.Type != "bug" {
		t.Errorf("Type = %q, want bug", result.Type)
	}
	if result.Priority != "high" {
		t.Errorf("Priority = %q, want high", result.Priority)
	}
	wantLabels := []string{"with-space", "backend", "milestone::puesta-en-uso"}
	if !equalStrings(result.Labels, wantLabels) {
		t.Errorf("Labels = %v, want %v", result.Labels, wantLabels)
	}
	wantAssignees := []string{"Sara-Smith"}
	if !equalStrings(result.Assignees, wantAssignees) {
		t.Errorf("Assignees = %v, want %v", result.Assignees, wantAssignees)
	}
	if result.CreatedAt != "2026-09-20T22:08:00Z" {
		t.Errorf("CreatedAt = %q", result.CreatedAt)
	}
	if result.UpdatedAt != result.CreatedAt {
		t.Errorf("UpdatedAt = %q, want it to equal CreatedAt %q", result.UpdatedAt, result.CreatedAt)
	}
	if result.Due != "2026-10-01" {
		t.Errorf("Due = %q, want 2026-10-01 unconverted", result.Due)
	}

	// Two findings are expected: "with space" -> "with-space" in labels,
	// and "Sara Smith" -> "Sara-Smith" in assignees. Status, type,
	// priority, and the milestone all matched cleanly.
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2: %v", len(findings), findings)
	}
	fieldsSeen := map[string]bool{}
	for _, f := range findings {
		fieldsSeen[f.Field] = true
	}
	if !fieldsSeen["labels"] || !fieldsSeen["assignees"] {
		t.Errorf("findings = %v, want one for labels and one for assignees", findings)
	}
}

func TestTaskUpdatedAtIsConvertedWhenPresent(t *testing.T) {
	task := source.Task{
		File:        "t.md",
		CreatedDate: "2026-09-20",
		UpdatedDate: "2026-09-21 10:00",
	}

	result, _ := Task(task, nil, testConfig)

	if result.CreatedAt != "2026-09-20T00:00:00Z" {
		t.Errorf("CreatedAt = %q", result.CreatedAt)
	}
	if result.UpdatedAt != "2026-09-21T10:00:00Z" {
		t.Errorf("UpdatedAt = %q", result.UpdatedAt)
	}
}

func TestTaskCommentDatesAreConverted(t *testing.T) {
	task := source.Task{
		File:        "t.md",
		CreatedDate: "2026-09-20",
		Comments: []source.Comment{
			{Author: "@ann", CreatedAt: "2026-09-20 22:08", Body: "First comment"},
			{Author: "", CreatedAt: "", Body: "No date"},
		},
	}

	result, _ := Task(task, nil, testConfig)

	if len(result.Comments) != 2 {
		t.Fatalf("got %d comments, want 2", len(result.Comments))
	}
	if result.Comments[0].CreatedAt != "2026-09-20T22:08:00Z" {
		t.Errorf("Comments[0].CreatedAt = %q", result.Comments[0].CreatedAt)
	}
	if result.Comments[0].Author != "@ann" || result.Comments[0].Body != "First comment" {
		t.Errorf("Comments[0] = %+v", result.Comments[0])
	}
	if result.Comments[1].CreatedAt != "" {
		t.Errorf("Comments[1].CreatedAt = %q, want empty (no invented date)", result.Comments[1].CreatedAt)
	}
}

func TestTaskMergesAcceptanceCriteriaAndDefinitionOfDone(t *testing.T) {
	task := source.Task{
		File:        "t.md",
		CreatedDate: "2026-09-20",
		AcceptanceCriteria: []source.Checkbox{
			{Number: 1, Checked: true, Text: "First"},
		},
		DefinitionOfDone: []source.Checkbox{
			{Number: 1, Checked: false, Text: "Reviewed"},
		},
	}

	result, _ := Task(task, nil, testConfig)

	want := []source.Checkbox{
		{Number: 1, Checked: true, Text: "First"},
		{Number: 2, Checked: false, Text: "Reviewed #dod"},
	}
	if len(result.AcceptanceCriteria) != len(want) {
		t.Fatalf("got %v, want %v", result.AcceptanceCriteria, want)
	}
	for i := range want {
		if result.AcceptanceCriteria[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i, result.AcceptanceCriteria[i], want[i])
		}
	}
}

func TestTaskWithUnmatchedVocabularyOmitsTheFieldAndReports(t *testing.T) {
	task := source.Task{
		File:        "t.md",
		Status:      "Pending",
		CreatedDate: "2026-09-20",
	}

	result, findings := Task(task, nil, testConfig)

	if result.Status != "" {
		t.Errorf("Status = %q, want empty", result.Status)
	}
	if len(findings) != 1 || findings[0].Field != "status" {
		t.Fatalf("findings = %v, want exactly one with Field=status", findings)
	}
}
