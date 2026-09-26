package convert

import (
	"strings"
	"testing"
)

func TestMilestoneSlugsOfAKnownMilestone(t *testing.T) {
	slugs, findings := MilestoneSlugs(map[string]string{"m-4": "Puesta en uso"})
	if slugs["m-4"] != "puesta-en-uso" {
		t.Errorf("slugs[m-4] = %q, want puesta-en-uso", slugs["m-4"])
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}

// TestMilestoneSlugsFallsBackToTheIdWhenTheTitleSlugifiesToNothing covers
// docs task point 4, step 3: a milestone whose title has no letter or
// digit at all uses its own id as the slug, with a finding.
func TestMilestoneSlugsFallsBackToTheIdWhenTheTitleSlugifiesToNothing(t *testing.T) {
	slugs, findings := MilestoneSlugs(map[string]string{"m-9": "..."})
	if slugs["m-9"] != "m-9" {
		t.Errorf("slugs[m-9] = %q, want m-9", slugs["m-9"])
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, "m-9") {
		t.Errorf("finding = %+v, want it to mention m-9", findings[0])
	}
}

// TestMilestoneSlugsReportsACollisionBetweenTwoDistinctMilestones covers
// docs task point 4, step 4: two distinct milestone ids that produce the
// same slug raise a finding.
func TestMilestoneSlugsReportsACollisionBetweenTwoDistinctMilestones(t *testing.T) {
	slugs, findings := MilestoneSlugs(map[string]string{
		"m-1": "Sprint 3",
		"m-2": "Sprint  3",
	})
	if slugs["m-1"] != "sprint-3" || slugs["m-2"] != "sprint-3" {
		t.Fatalf("slugs = %v, want both m-1 and m-2 to be sprint-3", slugs)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, "m-1") || !strings.Contains(findings[0].Message, "m-2") {
		t.Errorf("finding = %+v, want it to mention both m-1 and m-2", findings[0])
	}
}

// TestScopedLabelsForAKnownMilestone checks that a task's own labels come
// first, followed by the milestone:: label.
func TestScopedLabelsForAKnownMilestone(t *testing.T) {
	slugs := map[string]string{"m-4": "puesta-en-uso"}
	result, findings := ScopedLabels("t.md", []string{"backend"}, "m-4", "", slugs)

	want := []string{"backend", "milestone::puesta-en-uso"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}

// TestScopedLabelsForAnUnknownMilestoneUsesTheId covers a task referencing
// a milestone id MilestoneSlugs never saw (its file was not found while
// reading the board): falls back to the id, with a finding.
func TestScopedLabelsForAnUnknownMilestoneUsesTheId(t *testing.T) {
	result, findings := ScopedLabels("t.md", nil, "m-99", "", map[string]string{})

	want := []string{"milestone::m-99"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, "m-99") {
		t.Errorf("finding = %+v, want it to mention m-99", findings[0])
	}
}

// TestScopedLabelsRemovesAColldingSourceMilestoneLabel covers the encargo's
// explicit example: a task with labels: ["milestone::ya-tenia-esto"] and
// milestone: m-4 drops the source label, with a finding, and ends up with
// the derived milestone::puesta-en-uso label instead.
func TestScopedLabelsRemovesAColldingSourceMilestoneLabel(t *testing.T) {
	slugs := map[string]string{"m-4": "puesta-en-uso"}
	result, findings := ScopedLabels("t.md", []string{"milestone::ya-tenia-esto"}, "m-4", "", slugs)

	want := []string{"milestone::puesta-en-uso"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
	if !strings.Contains(findings[0].Message, "milestone::ya-tenia-esto") {
		t.Errorf("finding = %+v, want it to mention the dropped label", findings[0])
	}
}

// TestScopedLabelsKeyComparisonIsCaseFoldedOnly checks
// docs/spec/valores-de-entrada.md, "La clave se compara plegada": a source
// label with key "Milestone" (different case) still collides, but a label
// like "milestoned:x", whose key is not exactly "milestone" even
// case-folded, does not.
func TestScopedLabelsKeyComparisonIsCaseFoldedOnly(t *testing.T) {
	slugs := map[string]string{"m-4": "puesta-en-uso"}
	result, findings := ScopedLabels("t.md", []string{"Milestone:old", "milestoned:x"}, "m-4", "", slugs)

	want := []string{"milestoned:x", "milestone::puesta-en-uso"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1 (only Milestone:old collides): %v", len(findings), findings)
	}
}

// TestScopedLabelsForAProject checks that project's label comes after
// milestone's, using the same slug algorithm.
func TestScopedLabelsForAProject(t *testing.T) {
	result, findings := ScopedLabels("t.md", nil, "", "Alpha Team", nil)

	want := []string{"project::alpha-team"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}

// TestScopedLabelsOrdersMilestoneBeforeProject checks the final ordering
// docs task point 4 mandates: origin labels, then milestone::, then
// project::.
func TestScopedLabelsOrdersMilestoneBeforeProject(t *testing.T) {
	slugs := map[string]string{"m-4": "puesta-en-uso"}
	result, _ := ScopedLabels("t.md", []string{"backend"}, "m-4", "alpha", slugs)

	want := []string{"backend", "milestone::puesta-en-uso", "project::alpha"}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
}

// TestScopedLabelsOfAProjectThatSlugifiesToNothingFallsBackToTheRawValue
// documents this phase's own decision for the case
// docs/especificacion.md leaves open for "project" (no id-like fallback
// exists the way it does for milestone): fall back to the raw value,
// with a finding.
func TestScopedLabelsOfAProjectThatSlugifiesToNothingFallsBackToTheRawValue(t *testing.T) {
	result, findings := ScopedLabels("t.md", nil, "", "...", nil)

	want := []string{"project::..."}
	if !equalStrings(result, want) {
		t.Fatalf("result = %v, want %v", result, want)
	}
	if len(findings) != 1 {
		t.Fatalf("got %d findings, want 1: %v", len(findings), findings)
	}
}

// TestATaskWithNeitherMilestoneNorProjectGetsNoLabelsAndNoFinding checks
// the encargo's explicit carve-out: absent milestone/project is not a
// finding.
func TestATaskWithNeitherMilestoneNorProjectGetsNoLabelsAndNoFinding(t *testing.T) {
	result, findings := ScopedLabels("t.md", []string{"backend"}, "", "", nil)
	if !equalStrings(result, []string{"backend"}) {
		t.Fatalf("result = %v, want [backend]", result)
	}
	if len(findings) != 0 {
		t.Fatalf("got %d findings, want 0: %v", len(findings), findings)
	}
}
