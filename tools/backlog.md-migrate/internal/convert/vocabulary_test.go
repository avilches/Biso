package convert

import (
	"strings"
	"testing"
)

// TestVocabularyMatchingExamples reproduces, character for character, the
// example table of docs/spec/vocabularios.md, "El algoritmo de
// coincidencia", against the same three configured statuses that table
// uses.
func TestVocabularyMatchingExamples(t *testing.T) {
	configured := []string{"To Do", "In Progress", "Done"}

	matches := []string{"To Do", "todo", "TODO", "To-Do", "TO_DO", "to  do"}
	for _, v := range matches {
		match, ok := MatchVocabulary(v, configured)
		if !ok || match != "To Do" {
			t.Errorf("MatchVocabulary(%q, ...) = %q, %v, want \"To Do\", true", v, match, ok)
		}
	}

	noMatches := []string{"To Do.", "To.Do"}
	for _, v := range noMatches {
		match, ok := MatchVocabulary(v, configured)
		if ok {
			t.Errorf("MatchVocabulary(%q, ...) = %q, true, want ok=false", v, match)
		}
	}
}

// TestMatchFieldReportsUnmatchedValueWithTheConfiguredList checks that a
// value matching nothing raises a Finding carrying the original value and
// the full configured list.
func TestMatchFieldReportsUnmatchedValueWithTheConfiguredList(t *testing.T) {
	configured := []string{"To Do", "In Progress", "Done"}

	match, findings := MatchField("t.md", "status", "Pending", configured)
	if match != "" {
		t.Fatalf("MatchField = %q, want empty", match)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	f := findings[0]
	if f.File != "t.md" || f.Field != "status" {
		t.Errorf("Finding = %+v, want File=t.md Field=status", f)
	}
	wantSubstrings := []string{"Pending", "To Do", "In Progress", "Done"}
	for _, want := range wantSubstrings {
		if !strings.Contains(f.Message, want) {
			t.Errorf("Finding.Message = %q, want it to contain %q", f.Message, want)
		}
	}
}

// TestMatchFieldOnAnEmptyValueIsNotAFinding checks that an absent source
// value (the task simply has no status/type/priority) is not attempted at
// all: no match, and no Finding, since there was nothing to match.
func TestMatchFieldOnAnEmptyValueIsNotAFinding(t *testing.T) {
	match, findings := MatchField("t.md", "status", "", []string{"To Do", "Done"})
	if match != "" {
		t.Errorf("match = %q, want empty", match)
	}
	if len(findings) != 0 {
		t.Errorf("got %d findings, want 0: %v", len(findings), findings)
	}
}

// TestTwoConfiguredValuesThatNormalizeTheSameAreAmbiguous covers docs task
// point 1's requested case: two DISTINCT configured values ("high" and
// "HIGH") that normalize identically must make coincidir() report no
// result, not pick one arbitrarily.
func TestTwoConfiguredValuesThatNormalizeTheSameAreAmbiguous(t *testing.T) {
	configured := []string{"high", "HIGH", "low"}

	match, ok := MatchVocabulary("High", configured)
	if ok {
		t.Errorf("MatchVocabulary(\"High\", ...) = %q, true, want ok=false (ambiguous)", match)
	}

	// "low" alone still resolves fine: the ambiguity is specific to the
	// "high"/"HIGH" pair, not a blanket failure of the whole list.
	match, ok = MatchVocabulary("LOW", configured)
	if !ok || match != "low" {
		t.Errorf("MatchVocabulary(\"LOW\", ...) = %q, %v, want \"low\", true", match, ok)
	}
}

// TestExactMatchWinsWithoutNormalizing checks step "a" of coincidir():
// an exact configured value wins even when a different configured value
// would also match once normalized, and even before any normalization is
// attempted.
func TestExactMatchWinsWithoutNormalizing(t *testing.T) {
	configured := []string{"To Do", "To-Do"}

	match, ok := MatchVocabulary("To Do", configured)
	if !ok || match != "To Do" {
		t.Errorf("MatchVocabulary(\"To Do\", ...) = %q, %v, want \"To Do\", true", match, ok)
	}
	match, ok = MatchVocabulary("To-Do", configured)
	if !ok || match != "To-Do" {
		t.Errorf("MatchVocabulary(\"To-Do\", ...) = %q, %v, want \"To-Do\", true", match, ok)
	}
}

// TestDuplicateConfiguredValuesAreNotAnAmbiguity checks that the exact same
// configured value repeated does not trigger the "two or more distinct
// matches" case: it is the same value, counted once.
func TestDuplicateConfiguredValuesAreNotAnAmbiguity(t *testing.T) {
	configured := []string{"To Do", "To Do", "Done"}
	match, ok := MatchVocabulary("todo", configured)
	if !ok || match != "To Do" {
		t.Errorf("MatchVocabulary(\"todo\", ...) = %q, %v, want \"To Do\", true", match, ok)
	}
}
