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
	if got["TASK-1"] != "i" {
		t.Errorf(`got["TASK-1"] = %q, want "i"`, got["TASK-1"])
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

	if got["TASK-1"] != want1 || got["TASK-2"] != want2 || got["TASK-3"] != want3 {
		t.Fatalf("got = %#v, want TASK-1=%q TASK-2=%q TASK-3=%q", got, want1, want2, want3)
	}
	if !(got["TASK-1"] < got["TASK-2"] && got["TASK-2"] < got["TASK-3"]) {
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

	if got["TASK-1"] != want1 {
		t.Errorf(`got["TASK-1"] = %q, want %q`, got["TASK-1"], want1)
	}
	if got["TASK-2"] != want2 {
		t.Errorf(`got["TASK-2"] = %q, want %q`, got["TASK-2"], want2)
	}
	// The anchor itself is never reused as a task's own key.
	if got["TASK-1"] == "m" || got["TASK-2"] == "m" {
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

	if !(got["TASK-1"] < got["TASK-2"] && got["TASK-2"] < got["TASK-3"]) {
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

	if !(got["TASK-1"] < got["TASK-1.1"]) {
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

	if !(got["TASK-2"] < got["TASK-9"]) {
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

	if _, ok := got["TASK-1"]; ok {
		t.Errorf("TASK-1 has no ordinal, it should not appear in the result: %#v", got)
	}
	if _, ok := got["TASK-2"]; !ok {
		t.Errorf("TASK-2 has an ordinal, it should appear in the result: %#v", got)
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

	if _, ok := got["TASK-1"]; ok {
		t.Errorf("TASK-1 was never passed in, it should not appear in the result: %#v", got)
	}
	if len(got) != 1 {
		t.Errorf("len(got) = %d, want 1", len(got))
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

	for id, key := range got {
		if existing[key] {
			t.Errorf("assigned key %q for %s collides with an existing destination key", key, id)
		}
		if key <= "zzi" {
			t.Errorf("assigned key %q for %s is not strictly greater than the anchor %q", key, id, "zzi")
		}
	}
}
