package ops

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"biso/internal/model"
)

// These are the cases of the behaviour table of docs/spec/cmd/get.md that
// are about what the command answers. What the card looks like is checked
// against the specification's own block in cmd/biso/read_golden_test.go,
// over the compiled program.

func (h *harness) get(p GetParams) *GetResult {
	h.t.Helper()
	result, err := GetOn(h.b, h.env, p)
	if err != nil {
		h.t.Fatalf("biso get %s: %v", p.Ref, err)
	}
	return result
}

// TestGetHasEightSectionsAndNotNine is the correction the adversarial
// review of this step made: the card has meta, desc, ac, plan, notes,
// summary, comments and question, and nothing else.
func TestGetHasEightSectionsAndNotNine(t *testing.T) {
	want := []string{"meta", "desc", "ac", "plan", "notes", "summary", "comments", "question"}
	if !reflect.DeepEqual(Sections, want) {
		t.Errorf("the sections are %v and not the eight of the specification", Sections)
	}
}

// TestSectionsComeOutInTheOrderOfTheCard is the rule of
// docs/spec/cmd/get.md: never in the order they were asked for.
func TestSectionsComeOutInTheOrderOfTheCard(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	one := h.get(GetParams{Ref: id, Sections: []string{"plan", "ac"}})
	other := h.get(GetParams{Ref: id, Sections: []string{"ac", "plan"}})

	if !reflect.DeepEqual(one.Sections, []string{"ac", "plan"}) {
		t.Errorf("the sections came out as %v", one.Sections)
	}
	if !reflect.DeepEqual(one.Sections, other.Sections) {
		t.Errorf("the two orders answered %v and %v", one.Sections, other.Sections)
	}
}

// TestWithoutSectionEveryOneOfTheEightIsAnswered is what decides between
// printing `(empty)` and leaving a section out.
func TestWithoutSectionEveryOneOfTheEightIsAnswered(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	r := h.get(GetParams{Ref: id})

	if !r.WholeCard || !reflect.DeepEqual(r.Sections, Sections) {
		t.Errorf("a call with no --section answered %v, whole card %v", r.Sections, r.WholeCard)
	}
}

func TestASectionThatDoesNotExistIsAUsageErrorWithTheEightNames(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	_, err := GetOn(h.b, h.env, GetParams{Ref: id, Sections: []string{"dod"}})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "unknown_section" {
		t.Fatalf("error = %d/%s, want 2/unknown_section", e.ExitCode, e.Code)
	}
	if !reflect.DeepEqual(e.Valid, Sections) {
		t.Errorf("the valid values are %v and not the eight sections", e.Valid)
	}
}

// TestGetByTextSaysWhichTaskItMatched is the note of
// docs/spec/referencias.md#la-búsqueda-por-texto.
func TestGetByTextSaysWhichTaskItMatched(t *testing.T) {
	h := newHarness(t)
	h.create("Normalize CRLF in the diff")
	h.create("Something else")

	r := h.get(GetParams{Ref: "CRLF"})

	if r.Task.Task.ID != "MYP-1" {
		t.Errorf("the text resolved to %s", r.Task.Task.ID)
	}
	if !reflect.DeepEqual(r.Notes, []string{`"CRLF" matched MYP-1`}) {
		t.Errorf("the notes are %v", r.Notes)
	}
}

// TestGetByTextWithSeveralMatchesCarriesTheCandidatesAlreadyListed is the
// half of docs/spec/referencias.md this step closed: the error already
// knew its candidates, and now it also carries them in the order and with
// the limit of `biso ls`.
func TestGetByTextWithSeveralMatchesCarriesTheCandidatesAlreadyListed(t *testing.T) {
	h := newHarness(t)
	h.create("A task about CRLF", scalar("priority", "low"))
	h.create("Another about CRLF", scalar("priority", "high"))

	_, err := GetOn(h.b, h.env, GetParams{Ref: "CRLF"})

	var ambiguous *AmbiguousRef
	if !errors.As(err, &ambiguous) {
		t.Fatalf("the error is %v and not an ambiguous reference", err)
	}
	if ambiguous.Err.ExitCode != 5 || ambiguous.Err.Code != "ambiguous_reference" {
		t.Fatalf("error = %d/%s", ambiguous.Err.ExitCode, ambiguous.Err.Code)
	}
	if ambiguous.Listing == nil {
		t.Fatal("the candidates were not listed")
	}
	// In the order of a listing, which is the urgency and not the order
	// they were found in.
	if got := ids(ambiguous.Listing); !reflect.DeepEqual(got, []string{"MYP-2", "MYP-1"}) {
		t.Errorf("the candidates came out as %v", got)
	}
}

// TestTheCandidatesReachATaskInTheTerminalStatus is the one place the
// candidates of an ambiguous reference are not what `biso ls --search`
// would answer, and docs/spec/referencias.md#la-búsqueda-por-texto says so
// on purpose: resolving a reference searches the whole board, and the only
// thing it leaves out is an archived task. A listing leaves out the
// terminal status too, because that is the default value of its --status, and a
// finished task is still a task anyone can name.
func TestTheCandidatesReachATaskInTheTerminalStatus(t *testing.T) {
	h := newHarness(t)
	h.create("A task about CRLF")
	finished := h.create("Another about CRLF")
	h.set(finished, scalar("status", "Done"))

	// The listing does not have it, because its --status defaults to every
	// status but the terminal one.
	if got := ids(h.list(ListParams{Search: pointer("CRLF")})); !reflect.DeepEqual(got, []string{"MYP-1"}) {
		t.Fatalf("the listing answered %v, and a finished task is not in it by default", got)
	}

	_, err := GetOn(h.b, h.env, GetParams{Ref: "CRLF"})

	var ambiguous *AmbiguousRef
	if !errors.As(err, &ambiguous) {
		t.Fatalf("the error is %v, and two tasks carry the text", err)
	}
	if got := ids(ambiguous.Listing); !reflect.DeepEqual(got, []string{"MYP-1", finished}) {
		t.Errorf("the candidates are %v, and the finished task has to be among them", got)
	}
}

// TestGetOfAnArchivedTaskSaysSo is the fourth row of the table.
func TestGetOfAnArchivedTaskSaysSo(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	if _, err := h.b.Store.Exec(`UPDATE task SET archived = 1 WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}

	r := h.get(GetParams{Ref: id})

	if !reflect.DeepEqual(r.Notes, []string{id + " is archived"}) {
		t.Errorf("the notes are %v", r.Notes)
	}
}

// TestGetOfAnUnreadableTaskIsExitCodeThree is the targeted read of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// which is the opposite ending from the one a listing takes.
func TestGetOfAnUnreadableTaskIsExitCodeThree(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	if _, err := h.b.Store.Exec(
		`UPDATE task SET priority = 'urgent' WHERE id = ?`, id); err != nil {
		t.Fatal(err)
	}

	_, err := GetOn(h.b, h.env, GetParams{Ref: id})

	e := specError(t, err)
	if e.ExitCode != 3 || e.Code != "undecodable_task" {
		t.Fatalf("error = %d/%s, want 3/undecodable_task", e.ExitCode, e.Code)
	}
}

// TestGetComputesTheDerivedFieldsOfTheCard is what the card and the
// envelope both need and no task carries by itself.
func TestGetComputesTheDerivedFieldsOfTheCard(t *testing.T) {
	h := newHarness(t)
	blocker := h.create("The blocker")
	h.create("Depends on it", add("add-deps", blocker))

	r := h.get(GetParams{Ref: blocker})

	if !reflect.DeepEqual(r.Task.Blocks, []string{"MYP-2"}) {
		t.Errorf("blocks = %v", r.Task.Blocks)
	}
	if r.Task.Blocked || r.Task.Waiting || r.Task.LeaseExpired {
		t.Errorf("a plain task answered blocked %v, waiting %v, lease expired %v",
			r.Task.Blocked, r.Task.Waiting, r.Task.LeaseExpired)
	}
}

// TestExplainUrgencyIsOnlyComputedWhenItIsAskedFor is the one key a flag
// governs in the whole JSON contract.
func TestExplainUrgencyIsOnlyComputedWhenItIsAskedFor(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", scalar("priority", "high"))

	if h.get(GetParams{Ref: id}).Task.Breakdown != nil {
		t.Error("a call without --explain-urgency carries the breakdown")
	}
	r := h.get(GetParams{Ref: id, ExplainUrgency: true})
	if r.Task.Breakdown == nil {
		t.Fatal("a call with --explain-urgency does not carry the breakdown")
	}
	if r.Task.Breakdown.Total != r.Task.Urgency {
		t.Errorf("the breakdown totals %v and the field says %v",
			r.Task.Breakdown.Total, r.Task.Urgency)
	}
}

// TestGetWithNoReferenceIsAUsageError, which is the only thing this command
// really needs.
func TestGetWithNoReferenceIsAUsageError(t *testing.T) {
	h := newHarness(t)

	_, err := GetOn(h.b, h.env, GetParams{})

	assertSpec(t, err, 2, "missing_ref")
}

// TestEmptySectionKnowsWhichSectionsHaveNothingInThem is what the text
// output asks it, and `meta` is never one of them.
func TestEmptySectionKnowsWhichSectionsHaveNothingInThem(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("append-desc", "Something"))
	task := h.load(id)

	for _, section := range Sections {
		empty := EmptySection(task, section)
		want := section != SectionMeta && section != SectionDesc
		if empty != want {
			t.Errorf("%s answered empty %v", section, empty)
		}
	}
}

// TestTheErrorOfAnAmbiguousReferenceStillPrintsAsAnOrdinaryError is what
// lets every layer that only knows how to print a *model.Error keep working
// unchanged.
func TestTheErrorOfAnAmbiguousReferenceStillPrintsAsAnOrdinaryError(t *testing.T) {
	h := newHarness(t)
	h.create("A task about CRLF")
	h.create("Another about CRLF")

	_, err := GetOn(h.b, h.env, GetParams{Ref: "CRLF"})

	var e *model.Error
	if !errors.As(err, &e) {
		t.Fatalf("the error does not unwrap to a *model.Error: %v", err)
	}
	if !strings.Contains(e.Message, "matches 2 tasks") {
		t.Errorf("the message is %q", e.Message)
	}
}

// TestClosureCountsAreAlwaysPresentButTheListsOnlyWithTheFlag is the split
// docs/spec/cmd/get.md#salida draws between blockedByCount/unblocksCount,
// which every call answers, and the full lists of
// docs/spec/cmd/get.md#--closure, which only --closure fills.
func TestClosureCountsAreAlwaysPresentButTheListsOnlyWithTheFlag(t *testing.T) {
	h := newHarness(t)
	blocker := h.create("The blocker")
	h.create("Depends on it", add("add-deps", blocker))

	without := h.get(GetParams{Ref: blocker})
	if without.Task.Closure != nil {
		t.Error("the closure lists were computed without --closure")
	}
	if without.Task.UnblocksCount != 1 {
		t.Errorf("unblocksCount = %d, want 1", without.Task.UnblocksCount)
	}

	with := h.get(GetParams{Ref: blocker, Closure: true})
	if with.Task.Closure == nil {
		t.Fatal("the closure lists are missing with --closure")
	}
	if !reflect.DeepEqual(with.Task.Closure.Unblocks, []string{"MYP-2"}) {
		t.Errorf("unblocks = %v", with.Task.Closure.Unblocks)
	}
	if len(with.Task.Closure.BlockedBy) != 0 {
		t.Errorf("blockedBy = %v, want none", with.Task.Closure.BlockedBy)
	}
}

// TestClosureWalksThroughACycleWithoutHangingOrRepeating is the first of
// the two mandatory edge cases of TASK-78: a cycle in the dependency graph
// never makes the walk fail or hang, and the task the closure was asked
// for never appears in its own closure even though the cycle reaches it
// (docs/spec/cmd/get.md#--closure).
//
// The cycle is closed by writing straight to the store, because the write
// path already refuses to create one (internal/ops/consistency.go,
// checkGraph): the only way a real board carries one is damage from
// outside the program, which is exactly the case this test stands in for.
func TestClosureWalksThroughACycleWithoutHangingOrRepeating(t *testing.T) {
	h := newHarness(t)
	a := h.create("A")
	b := h.create("B", add("add-deps", a))
	c := h.create("C", add("add-deps", b))
	// A -> C closes the cycle A -> C -> B -> A: A already depends on
	// nothing, and now it depends on C, which depends on B, which depends
	// on A.
	task := h.load(a)
	task.Dependencies = []string{c}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatalf("closing the cycle: %v", err)
	}

	r := h.get(GetParams{Ref: a, Closure: true})

	want := []string{b, c}
	if !reflect.DeepEqual(r.Task.Closure.BlockedBy, want) {
		t.Errorf("blockedBy = %v, want %v", r.Task.Closure.BlockedBy, want)
	}
	if !reflect.DeepEqual(r.Task.Closure.Unblocks, want) {
		t.Errorf("unblocks = %v, want %v", r.Task.Closure.Unblocks, want)
	}
	if r.Task.BlockedByCount != 2 || r.Task.UnblocksCount != 2 {
		t.Errorf("counts = %d/%d, want 2/2", r.Task.BlockedByCount, r.Task.UnblocksCount)
	}
}

// TestClosureExcludesADependencyThatPointsAtNoTask is the second mandatory
// edge case: a dependency that names an identifier the board does not have
// adds nothing to the closure and the walk does not try to go past it
// (docs/spec/cmd/get.md#--closure). --add-deps itself refuses a reference
// that does not resolve, so the dangling value is written straight to the
// store, the same way an external edit to the database would leave one.
func TestClosureExcludesADependencyThatPointsAtNoTask(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	task := h.load(id)
	task.Dependencies = []string{"MYP-999"}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatalf("writing the dangling dependency: %v", err)
	}

	r := h.get(GetParams{Ref: id, Closure: true})

	if len(r.Task.Closure.BlockedBy) != 0 {
		t.Errorf("blockedBy = %v, want none", r.Task.Closure.BlockedBy)
	}
	if r.Task.BlockedByCount != 0 {
		t.Errorf("blockedByCount = %d, want 0", r.Task.BlockedByCount)
	}
}
