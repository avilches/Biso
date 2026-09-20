package ops

import (
	"strings"
	"testing"
	"time"
)

// These are the cases of the behaviour table of docs/spec/cmd/new.md, one
// test per row, plus the two rules that table states in prose: what
// --replace-*, --rm-* and --clear-* do on a task that does not exist yet,
// and what --start leaves behind.

func TestNewRefusesAnEmptyTitle(t *testing.T) {
	h := newHarness(t)
	for _, title := range []string{"", "   ", "\t"} {
		_, err := NewOn(h.b, h.env, NewParams{Title: title, HasTitle: title != ""})
		assertSpec(t, err, 2, "missing_title")
	}
}

func TestNewKeepsALongTitleWhole(t *testing.T) {
	h := newHarness(t)
	title := strings.Repeat("a very long title ", 40)

	id := h.create(title)

	if got := h.load(id).Title; got != title {
		t.Errorf("the title was stored with %d characters, and it had %d", len(got), len(title))
	}
}

func TestNewAcceptsARepeatedTitleWithoutWarning(t *testing.T) {
	h := newHarness(t)
	h.create("Normalize CRLF in the diff")

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "Normalize CRLF in the diff", HasTitle: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and two tasks may be called the same", result.Warnings)
	}
	if result.Tasks[0].ID == "MYP-1" {
		t.Errorf("the second task took the identifier of the first")
	}
}

func TestNewRejectsAValueOutsideAVocabulary(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct {
		change Change
		code   string
	}{
		{scalar("status", "Pending"), "unknown_status"},
		{scalar("type", "epic"), "unknown_type"},
		{scalar("priority", "urgent"), "unknown_priority"},
	} {
		_, err := NewOn(h.b, h.env, NewParams{
			Title: "A task", HasTitle: true, Changes: []Change{c.change},
		})
		assertSpec(t, err, 3, c.code)
	}
}

// The valid values travel in the error, because the message the caller reads
// lists them and a second call to find them out is exactly what the
// specification is avoiding.
func TestTheUnknownValueErrorCarriesTheVocabulary(t *testing.T) {
	h := newHarness(t)

	_, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{scalar("status", "Pending")},
	})

	e := specError(t, err)
	if strings.Join(e.Valid, ", ") != "To Do, In Progress, Done" {
		t.Errorf("valid = %v, want the three configured statuses", e.Valid)
	}
	if e.Given != "Pending" {
		t.Errorf("given = %q, want what was typed", e.Given)
	}
}

func TestNewValidatesDependenciesAsItWritesThem(t *testing.T) {
	h := newHarness(t)

	_, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{add("add-deps", "MYP-90")},
	})

	assertSpec(t, err, 4, "never_allocated")
}

func TestNewStoresADependencyByItsIdentifier(t *testing.T) {
	h := newHarness(t)
	first := h.create("Normalize CRLF in the diff")

	id := h.create("Depends on it", add("add-deps", "CRLF"))

	deps := h.load(id).Dependencies
	if len(deps) != 1 || deps[0] != first {
		t.Errorf("dependencies = %v, want the identifier %s resolved to", deps, first)
	}
}

func TestNewRejectsAParentThatDoesNotExist(t *testing.T) {
	h := newHarness(t)

	_, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{scalar("parent", "MYP-90")},
	})

	assertSpec(t, err, 4, "never_allocated")
}

func TestNewRejectsAnUndeclaredExtensionKey(t *testing.T) {
	h := newHarness(t)

	_, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{ext("jira.key", "X-1")},
	})

	assertSpec(t, err, 3, "unknown_extension_key")
}

func TestNewRejectsAMalformedDueDateAndWarnsAboutAPastOne(t *testing.T) {
	h := newHarness(t)

	_, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{scalar("due", "20/09/2026")},
	})
	assertSpec(t, err, 2, "invalid_date")

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Changes: []Change{scalar("due", "2026-01-01")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !warned(result, "due_in_past") {
		t.Errorf("a due date already gone by was accepted without a warning")
	}
}

// docs/spec/cmd/new.md: on a task that does not exist yet there is nothing
// to replace and nothing to remove, so --replace-* leaves the list the same
// as --add-* would and --rm-* has nothing to act on.
func TestOnANewTaskReplaceBehavesLikeAddAndRemoveDoesNothing(t *testing.T) {
	h := newHarness(t)

	id := h.create("A task", replace("replace-labels", "parser"), remove("rm-labels", "urgent"))

	assertLabels(t, h.load(id), "parser")
}

func TestOnANewTaskEveryClearWarnsAndDoesNothing(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true,
		Changes: []Change{clear("clear-labels"), add("add-labels", "parser")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "clear_on_new_task") {
		t.Errorf("--clear-labels on a new task did not warn: %v", result.Warnings)
	}
	// And it did nothing, so the label added in the same call survives even
	// though clearing comes first in the order of application.
	assertLabels(t, h.load(result.Tasks[0].ID), "parser")
}

// docs/spec/modelo-de-datos/autor.md: the author is the caller's identity
// unless --author says otherwise, and a call with no identity creates the
// task without one, with no warning.
func TestTheAuthorOfANewTask(t *testing.T) {
	h := newHarness(t)

	if got := h.load(h.create("A task")).Author; got != "@claude" {
		t.Errorf("author = %q, want the caller's identity", got)
	}
	if got := h.load(h.create("A task", scalar("author", "@sara"))).Author; got != "@sara" {
		t.Errorf("author = %q, want the one --author gave", got)
	}

	anonymous := h.as("")
	result, err := NewOn(anonymous.b, anonymous.env, NewParams{Title: "A task", HasTitle: true})
	if err != nil {
		t.Fatal(err)
	}
	if got := h.load(result.Tasks[0].ID).Author; got != "" {
		t.Errorf("author = %q, want none", got)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("a task without an author warned: %v", result.Warnings)
	}
}

// docs/spec/cmd/new.md: --start without any identity and without -a leaves
// the task active and unassigned, with a note and WITHOUT a lease, because
// there is nobody to attribute one to.
func TestStartWithoutAnIdentityLeavesTheTaskUnassignedAndWithoutALease(t *testing.T) {
	h := newHarness(t).as("")

	result, err := NewOn(h.b, h.env, NewParams{Title: "A task", HasTitle: true, Start: true})
	if err != nil {
		t.Fatal(err)
	}

	task := h.load(result.Tasks[0].ID)
	if task.Status != "In Progress" {
		t.Errorf("status = %q, want the active one", task.Status)
	}
	if len(task.Assignees) != 0 {
		t.Errorf("assignees = %v, want nobody", task.Assignees)
	}
	if !task.LeaseExpiresAt.IsZero() || task.LeaseHolder != "" {
		t.Errorf("the task took a lease with no identity to hold it")
	}
	if !noted(result, "no identity configured, task left unassigned") {
		t.Errorf("notes = %v, want the one of the specification", result.Notes)
	}
}

// docs/spec/cmd/new.md: with -a @sara and another identity configured, the
// task is assigned to @sara and the lease is the caller's, because whoever
// takes it is whoever writes and not whoever is in assignees.
func TestStartAssignsToWhoWasNamedAndLeasesToWhoCalls(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Start: true,
		Changes: []Change{add("add-assignees", "@sara")},
	})
	if err != nil {
		t.Fatal(err)
	}

	task := h.load(result.Tasks[0].ID)
	if strings.Join(task.Assignees, ",") != "@sara" {
		t.Errorf("assignees = %v, want @sara alone", task.Assignees)
	}
	if task.LeaseHolder != "@claude" {
		t.Errorf("leaseHolder = %q, want the caller", task.LeaseHolder)
	}
	if !task.LeaseExpiresAt.Equal(writeClock.Add(240 * time.Minute)) {
		t.Errorf("leaseExpiresAt = %v, want now plus the board's lease_minutes", task.LeaseExpiresAt)
	}
}

// The equivalence the specification demands: `biso new "X" --start` has to
// leave the same task as `biso new "X"` followed by `biso start` would, so
// it claims the lease and does not merely set the status.
func TestStartClaimsTheLeaseForTheCaller(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{Title: "A task", HasTitle: true, Start: true})
	if err != nil {
		t.Fatal(err)
	}

	task := h.load(result.Tasks[0].ID)
	if strings.Join(task.Assignees, ",") != "@claude" {
		t.Errorf("assignees = %v, want the caller", task.Assignees)
	}
	if task.LeaseHolder != "@claude" || task.LeaseExpiresAt.IsZero() {
		t.Errorf("the lease was not claimed: holder %q, expires %v",
			task.LeaseHolder, task.LeaseExpiresAt)
	}
}

func TestStartWarnsAboutUnfinishedDependencies(t *testing.T) {
	h := newHarness(t)
	blocker := h.create("The blocker")

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, Start: true,
		Changes: []Change{add("add-deps", blocker)},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "unresolved_dependencies") {
		t.Errorf("starting a blocked task did not warn: %v", result.Warnings)
	}
}

// docs/spec/cmd/new.md#salida: the default answer is the identifier, and the
// keys of the criteria the call created are never announced in it, because
// a task is born with none and the caller can count them. They do travel in
// the envelope, which is the same for every command of kind task.write.
func TestNewNumbersItsCriteriaFromOneAndCarriesTheKeysInTheResult(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true,
		Changes: []Change{add("add-ac", "First"), add("add-ac", "Second")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tasks[0].AcAdded) != 2 ||
		result.Tasks[0].AcAdded[0] != 1 || result.Tasks[0].AcAdded[1] != 2 {
		t.Errorf("acAdded = %v, want the keys 1 and 2 in the order written", result.Tasks[0].AcAdded)
	}
	criteria := h.load(result.Tasks[0].ID).AcceptanceCriteria
	if len(criteria) != 2 || criteria[0].Text != "First" || criteria[1].Text != "Second" {
		t.Errorf("the criteria were not stored in the order they were written: %v", criteria)
	}
}

// docs/spec/cmd/new.md#--dry-run-sobre-una-sola-tarea: nothing is written,
// no identifier is spent, and a failure keeps its own specific code instead
// of becoming the 7 of the batch.
func TestNewDryRunWritesNothingAndSpendsNoIdentifier(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, DryRun: true,
		Changes: []Change{add("add-labels", "parser")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Tasks) != 0 {
		t.Errorf("a preview answered %d tasks, and it created none", len(result.Tasks))
	}

	last, err := h.b.Tasks.LastAllocated()
	if err != nil {
		t.Fatal(err)
	}
	if last != 0 {
		t.Errorf("the counter moved to %d, and a preview spends no identifier", last)
	}
	if id := h.create("The first real one"); id != "MYP-1" {
		t.Errorf("the first real task got %s, so the preview had spent MYP-1", id)
	}
}

func TestNewDryRunKeepsTheSpecificCodeAndNeverTheSeven(t *testing.T) {
	h := newHarness(t)

	for _, c := range []struct {
		change Change
		code   int
		id     string
	}{
		{scalar("status", "Pending"), 3, "unknown_status"},
		{scalar("parent", "MYP-90"), 4, "never_allocated"},
		{scalar("due", "nope"), 2, "invalid_date"},
	} {
		_, err := NewOn(h.b, h.env, NewParams{
			Title: "A task", HasTitle: true, DryRun: true, Changes: []Change{c.change},
		})
		assertSpec(t, err, c.code, c.id)
	}
}

func TestNewDryRunStillEmitsTheWarningsTheRealCallWould(t *testing.T) {
	h := newHarness(t)

	result, err := NewOn(h.b, h.env, NewParams{
		Title: "A task", HasTitle: true, DryRun: true,
		Changes: []Change{scalar("due", "2026-01-01")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "due_in_past") {
		t.Errorf("a preview said less than the call it simulates: %v", result.Warnings)
	}
}
