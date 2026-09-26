package convert

import (
	"testing"

	"backlog.md-migrate/internal/source"
)

func TestMergeDefinitionOfDoneContinuesTheKeySequence(t *testing.T) {
	criteria := []source.Checkbox{
		{Number: 1, Checked: true, Text: "First"},
		{Number: 3, Checked: false, Text: "Third"},
	}
	dod := []source.Checkbox{
		{Number: 1, Checked: false, Text: "Write docs"},
		{Number: 2, Checked: true, Text: "Ship it"},
	}

	got := MergeDefinitionOfDone(criteria, dod)

	want := []source.Checkbox{
		{Number: 1, Checked: true, Text: "First"},
		{Number: 3, Checked: false, Text: "Third"},
		{Number: 4, Checked: false, Text: "Write docs #dod"},
		{Number: 5, Checked: true, Text: "Ship it #dod"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("item %d = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestMergeDefinitionOfDoneStartsAtOneWithNoAcceptanceCriteria(t *testing.T) {
	dod := []source.Checkbox{{Number: 1, Checked: true, Text: "Deployed"}}

	got := MergeDefinitionOfDone(nil, dod)

	want := []source.Checkbox{{Number: 1, Checked: true, Text: "Deployed #dod"}}
	if len(got) != 1 || got[0] != want[0] {
		t.Fatalf("got %v, want %v", got, want)
	}
}

func TestMergeDefinitionOfDoneOfATaskWithNoDodLeavesCriteriaUnchanged(t *testing.T) {
	criteria := []source.Checkbox{{Number: 1, Checked: true, Text: "First"}}

	got := MergeDefinitionOfDone(criteria, nil)

	if len(got) != 1 || got[0] != criteria[0] {
		t.Fatalf("got %v, want %v unchanged", got, criteria)
	}
}
