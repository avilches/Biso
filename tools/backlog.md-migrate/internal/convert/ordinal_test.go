package convert

import (
	"testing"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

func float64Ptr(v float64) *float64 {
	return &v
}

// TestKeyBetweenTheElevenRequiredCases reproduces, character for character,
// every row of docs/spec/modelo-de-datos/orden-manual.md, "El algoritmo del
// punto medio"'s required table. before or after being "" stands for
// "nada" (absent), exactly as KeyBetween's own doc comment defines.
func TestKeyBetweenTheElevenRequiredCases(t *testing.T) {
	cases := []struct {
		name   string
		before string
		after  string
		want   string
	}{
		{"nada and nada", "", "", "i"},
		{"nada and i", "", "i", "9"},
		{"i and nada", "i", "", "r"},
		{"i and r", "i", "r", "m"},
		{"a and c", "a", "c", "b"},
		{"a and b", "a", "b", "ai"},
		{"9 and a", "9", "a", "9i"},
		{"mm and mn", "mm", "mn", "mmi"},
		{"i and j", "i", "j", "ii"},
		{"zz and nada", "zz", "", "zzi"},
		{"nada and 01", "", "01", "00i"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := KeyBetween(c.before, c.after)
			if got != c.want {
				t.Errorf("KeyBetween(%q, %q) = %q, want %q", c.before, c.after, got, c.want)
			}
		})
	}
}

func TestDestinationAnchorIsEmptyWhenNoTaskHasAKey(t *testing.T) {
	board := destination.Board{Tasks: []destination.Task{
		{ID: "TASK-1", Ordinal: ""},
		{ID: "TASK-2", Ordinal: ""},
	}}
	if got := DestinationAnchor(board); got != "" {
		t.Errorf("DestinationAnchor = %q, want empty", got)
	}
}

func TestDestinationAnchorIsTheGreatestExistingKey(t *testing.T) {
	board := destination.Board{Tasks: []destination.Task{
		{ID: "TASK-1", Ordinal: "m"},
		{ID: "TASK-2", Ordinal: ""},
		{ID: "TASK-3", Ordinal: "b"},
		{ID: "TASK-4", Ordinal: "zzi"},
	}}
	if got := DestinationAnchor(board); got != "zzi" {
		t.Errorf("DestinationAnchor = %q, want %q", got, "zzi")
	}
}

func TestAssignOrdinalsOnAnEmptyDestinationGivesTheFirstTaskI(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-1", Ordinal: float64Ptr(100)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1", len(got))
	}
	if got[0] != "i" {
		t.Errorf(`got[0] = %q, want "i"`, got[0])
	}
}

func TestAssignOrdinalsChainsOnAnEmptyDestination(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-1", Ordinal: float64Ptr(100)},
		{ID: "TASK-2", Ordinal: float64Ptr(200)},
		{ID: "TASK-3", Ordinal: float64Ptr(300)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	// docs/spec/modelo-de-datos/orden-manual.md's own worked chain:
	// KeyBetween("", "") = "i", KeyBetween("i", "") = "r",
	// KeyBetween("r", "") = KeyBetween's own next step.
	want1 := KeyBetween("", "")
	want2 := KeyBetween(want1, "")
	want3 := KeyBetween(want2, "")

	if got[0] != want1 || got[1] != want2 || got[2] != want3 {
		t.Fatalf("got = %#v, want TASK-1=%q TASK-2=%q TASK-3=%q", got, want1, want2, want3)
	}
	if !(got[0] < got[1] && got[1] < got[2]) {
		t.Errorf("keys are not in ascending code point order: %#v", got)
	}
}

func TestAssignOrdinalsOnADestinationWithAnExistingKeyAnchorsAfterIt(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-1", Ordinal: float64Ptr(1)},
		{ID: "TASK-2", Ordinal: float64Ptr(2)},
	}
	board := destination.Board{Tasks: []destination.Task{
		{ID: "OLD-1", Ordinal: "m"},
	}}

	got := AssignOrdinals(tasks, board)

	want1 := KeyBetween("m", "")
	want2 := KeyBetween(want1, "")

	if got[0] != want1 {
		t.Errorf(`got[0] = %q, want %q`, got[0], want1)
	}
	if got[1] != want2 {
		t.Errorf(`got[1] = %q, want %q`, got[1], want2)
	}
	// The anchor itself is never reused as a task's own key.
	if got[0] == "m" || got[1] == "m" {
		t.Errorf("an assigned key collided with the destination's own anchor: %#v", got)
	}
}

func TestAssignOrdinalsOrdersBySourceOrdinalAscending(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-3", Ordinal: float64Ptr(300)},
		{ID: "TASK-1", Ordinal: float64Ptr(100)},
		{ID: "TASK-2", Ordinal: float64Ptr(200)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	// got is parallel to tasks, not sorted by ordinal itself: got[1] is
	// TASK-1's own key (smallest ordinal, 100, assigned first), got[2] is
	// TASK-2's (200, assigned second), and got[0] is TASK-3's (300, the
	// input's own first element, but assigned LAST because its ordinal is
	// the greatest).
	if !(got[1] < got[2] && got[2] < got[0]) {
		t.Errorf("keys do not preserve the source ordinal's relative order: %#v", got)
	}
}

func TestAssignOrdinalsTiesOnTheSameOrdinalBreakByNaturalSourceIDOrderSimpleBeforeSubtask(t *testing.T) {
	tasks := []source.Task{
		// TASK-1.1 (a subtask of TASK-1) and TASK-1 (simple) share the same
		// main number, and the natural order puts the simple id first, the
		// same tie break identifiers.go's point 4 uses.
		{ID: "TASK-1.1", Ordinal: float64Ptr(500)},
		{ID: "TASK-1", Ordinal: float64Ptr(500)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	// got[1] is TASK-1's (simple) own key, got[0] is TASK-1.1's (subtask).
	if !(got[1] < got[0]) {
		t.Errorf("TASK-1 (simple) should sort before TASK-1.1 (subtask): %#v", got)
	}
}

func TestAssignOrdinalsTiesOnTheSameOrdinalBreakByMainNumberAscending(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-9", Ordinal: float64Ptr(500)},
		{ID: "TASK-2", Ordinal: float64Ptr(500)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	// got[1] is TASK-2's own key, got[0] is TASK-9's.
	if !(got[1] < got[0]) {
		t.Errorf("TASK-2 should sort before TASK-9 on a tied ordinal: %#v", got)
	}
}

func TestAssignOrdinalsATaskWithNoOrdinalDoesNotAppear(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-1", Ordinal: nil},
		{ID: "TASK-2", Ordinal: float64Ptr(100)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	if got[0] != "" {
		t.Errorf("TASK-1 has no ordinal, its position should be empty: %#v", got)
	}
	if got[1] == "" {
		t.Errorf("TASK-2 has an ordinal, its position should not be empty: %#v", got)
	}
}

func TestAssignOrdinalsATaskSkippedByPhase4bDoesNotAppear(t *testing.T) {
	// A caller filters out a task phase 4b skipped as already on the
	// destination before calling AssignOrdinals: this file has no notion
	// of "skipped" of its own, so the test only checks that a task simply
	// left out of the input never appears in the output, even though it
	// had a valid ordinal.
	tasks := []source.Task{
		{ID: "TASK-2", Ordinal: float64Ptr(100)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	if len(got) != 1 {
		t.Fatalf("len(got) = %d, want 1 (parallel to the single task passed in)", len(got))
	}
	if got[0] == "" {
		t.Errorf("TASK-2 has an ordinal, its position should not be empty: %#v", got)
	}
}

func TestAssignOrdinalsNeverCollidesWithAnExistingDestinationKey(t *testing.T) {
	board := destination.Board{Tasks: []destination.Task{
		{ID: "OLD-1", Ordinal: "a"},
		{ID: "OLD-2", Ordinal: "ai"},
		{ID: "OLD-3", Ordinal: "b"},
		{ID: "OLD-4", Ordinal: "zzi"},
		{ID: "OLD-5", Ordinal: "m"},
	}}
	existing := map[string]bool{"a": true, "ai": true, "b": true, "zzi": true, "m": true}

	tasks := []source.Task{
		{ID: "TASK-1", Ordinal: float64Ptr(1)},
		{ID: "TASK-2", Ordinal: float64Ptr(2)},
		{ID: "TASK-3", Ordinal: float64Ptr(3)},
	}

	got := AssignOrdinals(tasks, board)

	for i, key := range got {
		if existing[key] {
			t.Errorf("assigned key %q for %s collides with an existing destination key", key, tasks[i].ID)
		}
		if key <= "zzi" {
			t.Errorf("assigned key %q for %s is not strictly greater than the anchor %q", key, tasks[i].ID, "zzi")
		}
	}
}

// TestAssignOrdinalsGivesEachSharedIdCopyItsOwnKeyByPosition covers the bug
// found while implementing docs/especificacion.md, "Identificadores", point
// 1's id-reuse-after-archiving case (Backlog.md handing an archived task's
// number to the next one it creates, docs/decisiones.md's paragraph on it):
// two source.Task values that share the exact same literal ID (one
// archived, one not), each with its OWN distinct source ordinal, must each
// get their OWN correctly computed key at their OWN position in the result,
// never the other's. Before AssignOrdinals returned a slice, a map keyed by
// the shared literal ID would let the second one processed overwrite the
// first's key.
func TestAssignOrdinalsGivesEachSharedIdCopyItsOwnKeyByPosition(t *testing.T) {
	tasks := []source.Task{
		{ID: "TASK-4", Archived: false, Ordinal: float64Ptr(100)},
		{ID: "TASK-4", Archived: true, Ordinal: float64Ptr(50)},
	}
	board := destination.Board{}

	got := AssignOrdinals(tasks, board)

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2 (parallel to tasks)", len(got))
	}
	if got[0] == "" || got[1] == "" {
		t.Fatalf("both copies have their own ordinal, neither position should be empty: %#v", got)
	}
	if got[0] == got[1] {
		t.Fatalf("both copies got the SAME key, want two distinct ones: %#v", got)
	}
	// tasks[1] has the smaller source ordinal (50 < 100), so it must be
	// assigned first, the smaller key.
	if !(got[1] < got[0]) {
		t.Errorf("the archived copy (smaller source ordinal) should get the smaller key: got = %#v", got)
	}
}
