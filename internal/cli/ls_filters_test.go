package cli

import "testing"

// This file is the CLI-layer wiring of TASK-76: the flags docs/spec/cmd/ls.md
// adds to `biso ls`, parsed off the real table of Commands() and not the
// stripped-down one testCommands() builds for the generic parser rules. What
// each filter does to a listing is internal/ops's job
// (internal/ops/list_test.go); this file only covers what belongs to the
// command line: the flags exist, they read into ops.Filters, and --root and
// --parent refuse to share a call.

// TestRootConflictsWithParent is docs/spec/cmd/ls.md#comportamiento-caso-a-caso:
// the two are incompatible by construction, for any value of --parent, named
// in the table's own order whichever of the two was typed first.
func TestRootConflictsWithParent(t *testing.T) {
	for _, argv := range [][]string{
		{"ls", "--root", "--parent", "MYP-1"},
		{"ls", "--parent", "MYP-1", "--root"},
	} {
		_, err := Parse(argv, Commands(), Env{})
		e := wantError(t, err, 2, "incompatible_flags")
		if e.Message != "--parent and --root cannot be used together" {
			t.Errorf("parse(%q) said %q", argv, e.Message)
		}
	}
}

// TestTheNewLsFiltersReadIntoTheirFields is the parsed call carrying the
// eight new filters into the struct internal/cli/list.go's filtersOf and
// listParams build ops.Filters and ops.ListParams from.
func TestTheNewLsFiltersReadIntoTheirFields(t *testing.T) {
	p, err := Parse([]string{
		"ls",
		"--not-type", "docs", "--not-priority", "low",
		"--not-label", "blocked", "--not-assignee", "@sara",
		"--author", "@avilches", "--root",
		"--created-after", "2026-09-01", "--created-before", "2026-09-08",
		"--updated-after", "2026-09-01", "--updated-before", "2026-09-08",
		"--ref", "internal/ops/write.go", "--not-ref", "docs/bugs",
		"--sort", "priority",
	}, Commands(), Env{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	f := filtersOf(p)
	if got := f.NotType; len(got) != 1 || got[0] != "docs" {
		t.Errorf("--not-type read as %v", got)
	}
	if got := f.NotPriority; len(got) != 1 || got[0] != "low" {
		t.Errorf("--not-priority read as %v", got)
	}
	if got := f.NotLabel; len(got) != 1 || got[0] != "blocked" {
		t.Errorf("--not-label read as %v", got)
	}
	if got := f.NotAssignee; len(got) != 1 || got[0] != "@sara" {
		t.Errorf("--not-assignee read as %v", got)
	}
	if got := f.Author; len(got) != 1 || got[0] != "@avilches" {
		t.Errorf("--author read as %v", got)
	}
	if !f.Root {
		t.Error("--root did not set Root")
	}
	if got := f.Ref; len(got) != 1 || got[0] != "internal/ops/write.go" {
		t.Errorf("--ref read as %v", got)
	}
	if got := f.NotRef; len(got) != 1 || got[0] != "docs/bugs" {
		t.Errorf("--not-ref read as %v", got)
	}

	params, err := listParams(p)
	if err != nil {
		t.Fatalf("listParams: %v", err)
	}
	for _, c := range []struct {
		name string
		got  *string
		want string
	}{
		{"--created-after", params.CreatedAfter, "2026-09-01"},
		{"--created-before", params.CreatedBefore, "2026-09-08"},
		{"--updated-after", params.UpdatedAfter, "2026-09-01"},
		{"--updated-before", params.UpdatedBefore, "2026-09-08"},
	} {
		if c.got == nil || *c.got != c.want {
			t.Errorf("%s read as %v, want %q", c.name, c.got, c.want)
		}
	}
	if params.Sort != "priority" {
		t.Errorf("--sort priority read as %q", params.Sort)
	}
}

// TestANegationHasNoAnyStatusStyleConflict is the explicit rule of
// docs/spec/cmd/ls.md#parámetros: none of the four new negations declares an
// incompatibility of its own, unlike --any-status with --not-status.
func TestANegationHasNoAnyStatusStyleConflict(t *testing.T) {
	_, err := Parse([]string{
		"ls",
		"--type", "bug", "--not-type", "docs",
		"--priority", "high", "--not-priority", "low",
		"--label", "frontend", "--not-label", "backend",
		"--assignee", "@claude", "--not-assignee", "@sara",
	}, Commands(), Env{})
	if err != nil {
		t.Errorf("a positive filter with its cousin negation should not conflict: %v", err)
	}
}

// TestCreatedAfterRejectsAMalformedDate is the same error --due-before
// already gives, replicated for the four new date filters.
func TestCreatedAfterRejectsAMalformedDate(t *testing.T) {
	p, err := Parse([]string{"ls", "--created-after", "not-a-date"}, Commands(), Env{})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	_, err = listParams(p)
	e := wantError(t, err, 2, "invalid_date")
	if e.Field != "created-after" {
		t.Errorf("field is %q", e.Field)
	}
}
