package convert

import (
	"strings"
	"testing"
)

func equalStrings(a, b []string) bool {
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

// TestSpacesBecomeHyphensWithAFinding covers docs/decisiones.md, "Los
// espacios de una etiqueta o un asignado se convierten en guiones"'s two
// named examples: "with space" and "Sara Smith".
func TestSpacesBecomeHyphensWithAFinding(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "labels", []string{"with space", "Sara Smith"})

	want := []string{"with-space", "Sara-Smith"}
	if !equalStrings(cleaned, want) {
		t.Fatalf("cleaned = %v, want %v", cleaned, want)
	}
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2: %v", len(findings), findings)
	}
	for i, orig := range []string{"with space", "Sara Smith"} {
		if !strings.Contains(findings[i].Message, orig) || !strings.Contains(findings[i].Message, want[i]) {
			t.Errorf("finding %d = %+v, want it to mention %q and %q", i, findings[i], orig, want[i])
		}
		if findings[i].File != "t.md" || findings[i].Field != "labels" {
			t.Errorf("finding %d = %+v, want File=t.md Field=labels", i, findings[i])
		}
	}
}

// TestAValueWithNoSpaceIsNotAFinding checks the explicit carve-out of
// docs/decisiones.md, "Los espacios de una etiqueta o un asignado se
// convierten en guiones": a value with no whitespace at all does not raise
// the space-conversion finding, even though it passes through
// collapseWhitespaceToHyphens.
func TestAValueWithNoSpaceIsNotAFinding(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "labels", []string{"backend"})
	if !equalStrings(cleaned, []string{"backend"}) {
		t.Fatalf("cleaned = %v, want [backend]", cleaned)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}

// TestAValueOutsideTheAlphabetIsDropped covers docs/decisiones.md, "Los
// espacios de una etiqueta o un asignado se convierten en guiones"'s "a/b"
// example: a character outside the token alphabet drops the whole value,
// with a finding.
func TestAValueOutsideTheAlphabetIsDropped(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "labels", []string{"a/b", "ok"})

	if !equalStrings(cleaned, []string{"ok"}) {
		t.Fatalf("cleaned = %v, want [ok]", cleaned)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, "a/b") {
		t.Errorf("finding = %+v, want it to mention a/b", findings[0])
	}
}

// TestCollidingValuesAfterCleanupAreMergedWithoutAnExtraFinding covers
// docs/decisiones.md, "Los espacios de una etiqueta o un asignado se
// convierten en guiones"'s "a b" / "a-b" example: two values of the same
// list that become equal after cleanup collapse to one, and produce only
// the finding already raised for the conversion that caused the collision
// (no separate "duplicate" finding).
func TestCollidingValuesAfterCleanupAreMergedWithoutAnExtraFinding(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "labels", []string{"a b", "a-b"})

	if !equalStrings(cleaned, []string{"a-b"}) {
		t.Fatalf("cleaned = %v, want [a-b]", cleaned)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want exactly 1 (the space conversion): %v", len(findings), findings)
	}
}

// TestTokenAlphabetRuleAppliesTheSameToAssignees checks
// docs/decisiones.md, "Se aplica también a los asignados": the same
// cleanup runs for assignees, only the finding's Field differs.
func TestTokenAlphabetRuleAppliesTheSameToAssignees(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "assignees", []string{"Sara Smith", "c!"})

	if !equalStrings(cleaned, []string{"Sara-Smith"}) {
		t.Fatalf("cleaned = %v, want [Sara-Smith]", cleaned)
	}
	if len(findings) != 2 {
		t.Fatalf("got %d findings, want 2: %v", len(findings), findings)
	}
	for _, f := range findings {
		if f.Field != "assignees" {
			t.Errorf("finding = %+v, want Field=assignees", f)
		}
	}
}

// TestColonIsAllowedInTheTokenAlphabet checks that ':' (needed for scoped
// labels) is not treated as an alphabet violation.
func TestColonIsAllowedInTheTokenAlphabet(t *testing.T) {
	cleaned, findings := CleanTokenList("t.md", "labels", []string{"milestone::sprint-3"})
	if !equalStrings(cleaned, []string{"milestone::sprint-3"}) {
		t.Fatalf("cleaned = %v, want [milestone::sprint-3]", cleaned)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}
