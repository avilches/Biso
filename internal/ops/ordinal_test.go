package ops

import (
	"reflect"
	"strings"
	"testing"
)

// This file is the manual order as a caller sees it: the four flags of
// docs/spec/familias-de-flags.md#el-orden-manual over a real board, with
// the gap of each one, the refusals each one has, and the block of several
// tasks in a single call.
//
// What the keys themselves are is not asked here: that belongs to
// internal/model, which owns the midpoint function. What is asked here is
// where the tasks end up, because that is the whole point of the flags and
// it is what an implementation that computed the gap wrongly would get
// wrong while still producing perfectly valid keys.

// place is the four flags of the manual order, as changes.
func place(flag, value string) Change { return scalar(flag, value) }

// orderOf is the identifiers of the board, read in the manual order and
// with the tasks that have no key left out, which is what every test below
// compares.
func (h *harness) orderOf() []string {
	h.t.Helper()
	all, _, err := h.b.Tasks.All()
	if err != nil {
		h.t.Fatalf("reading the board: %v", err)
	}
	placed := make([]*taskKey, 0, len(all))
	for _, t := range all {
		if t.Ordinal != "" {
			placed = append(placed, &taskKey{t.ID, t.Ordinal})
		}
	}
	// The listing's own rule: by key, and the identifier breaking a tie.
	for i := 1; i < len(placed); i++ {
		for j := i; j > 0; j-- {
			if placed[j-1].key < placed[j].key {
				break
			}
			if placed[j-1].key == placed[j].key && placed[j-1].id <= placed[j].id {
				break
			}
			placed[j-1], placed[j] = placed[j], placed[j-1]
		}
	}
	ids := make([]string, 0, len(placed))
	for _, p := range placed {
		ids = append(ids, p.id)
	}
	return ids
}

type taskKey struct{ id, key string }

func (h *harness) assertOrder(want ...string) {
	h.t.Helper()
	got := h.orderOf()
	if !reflect.DeepEqual(got, want) {
		h.t.Errorf("the manual order is %v and not %v", got, want)
	}
}

// keyOf is the key a task ended up with, which only the tests about the gap
// itself look at.
func (h *harness) keyOf(id string) string {
	h.t.Helper()
	return h.load(id).Ordinal
}

// TestOrdinalFirstAndLastPlaceAtTheEnds is the first two rows of the table
// of docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación,
// including the row that says the first key of a board with none is the
// same whichever of the two was written.
func TestOrdinalFirstAndLastPlaceAtTheEnds(t *testing.T) {
	h := newHarness(t)
	first := h.create("The first one to be placed", place("ordinal", "first"))
	if h.keyOf(first) != "i" {
		t.Errorf("the first key of an empty board is %q, want i", h.keyOf(first))
	}

	last := h.create("Behind everything")
	h.set(last, place("ordinal", "last"))
	front := h.create("In front of everything")
	h.set(front, place("ordinal", "first"))
	h.assertOrder(front, first, last)

	// And doing it again moves the task again, which is the point of the
	// two flags being about the ends of the order and not about a value.
	h.set(first, place("ordinal", "last"))
	h.assertOrder(front, last, first)
}

// TestAboveAndBelowPlaceNextToTheNeighbour is the other two rows of that
// table: the gap of --above reaches down to the key just below the
// neighbour, and the gap of --below reaches up to the key just above it.
func TestAboveAndBelowPlaceNextToTheNeighbour(t *testing.T) {
	h := newHarness(t)
	a := h.create("A", place("ordinal", "last"))
	b := h.create("B", place("ordinal", "last"))
	c := h.create("C", place("ordinal", "last"))
	h.assertOrder(a, b, c)

	between := h.create("Between A and B")
	h.set(between, place("below", a))
	h.assertOrder(a, between, b, c)

	other := h.create("Also between A and B")
	h.set(other, place("above", b))
	h.assertOrder(a, between, other, b, c)
}

// TestTheEndsNeverFail is the promise of
// docs/spec/familias-de-flags.md#el-orden-manual that --above over the
// lowest key and --below over the highest are not errors: there is always a
// key below the lowest and one above the highest.
func TestTheEndsNeverFail(t *testing.T) {
	h := newHarness(t)
	low := h.create("The lowest", place("ordinal", "first"))
	high := h.create("The highest", place("ordinal", "last"))

	for i := 0; i < 8; i++ {
		under := h.create("Under the lowest")
		h.set(under, place("above", low))
		if h.keyOf(under) >= h.keyOf(low) {
			t.Fatalf("round %d: %q is not above the lowest %q",
				i, h.keyOf(under), h.keyOf(low))
		}
		low = under

		over := h.create("Over the highest")
		h.set(over, place("below", high))
		if h.keyOf(over) <= h.keyOf(high) {
			t.Fatalf("round %d: %q is not below the highest %q",
				i, h.keyOf(over), h.keyOf(high))
		}
		high = over
	}
	h.assertOrder(h.orderOf()...)
}

// TestABlockLandsInTheOrderItWasWritten is the example of
// docs/spec/familias-de-flags.md#el-orden-manual, both ways round: with
// --below the neighbour stays on top of the block, and with --above the
// block goes on top of the neighbour, and in both the references keep the
// order they were typed in.
func TestABlockLandsInTheOrderItWasWritten(t *testing.T) {
	h := newHarness(t)
	c := h.create("C", place("ordinal", "last"))
	tail := h.create("The one after C", place("ordinal", "last"))
	a := h.create("A")
	b := h.create("B")

	result, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{a, b}, Changes: []Change{place("below", c)},
	})
	if err != nil {
		t.Fatalf("biso set A B --below C: %v", err)
	}
	if len(result.Tasks) != 2 {
		t.Fatalf("the call wrote %d tasks and named two", len(result.Tasks))
	}
	h.assertOrder(c, a, b, tail)

	// The same call the other way round, over two tasks that are already
	// placed: what decides is the order of the references and not the keys
	// they happened to have.
	if _, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{b, a}, Changes: []Change{place("above", c)},
	}); err != nil {
		t.Fatalf("biso set B A --above C: %v", err)
	}
	h.assertOrder(b, a, c, tail)
}

// TestTheGapDiscountsTheTasksTheCallItselfMoves is the precision of
// docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación
// that is invisible in the result and decisive in the keys: moving a task
// just above the one it already sat on top of must not spend symbols on
// the key it is about to give up.
func TestTheGapDiscountsTheTasksTheCallItselfMoves(t *testing.T) {
	h := newHarness(t)
	a := h.create("A", place("ordinal", "last"))
	b := h.create("B", place("ordinal", "last"))

	// A is right above B. Asking for exactly that again has to leave A
	// with the key of a gap that reaches the bottom of the order, because
	// the only key under B is the one A is giving up. Counting that key
	// would squeeze the new one between the old one and B, spending a
	// symbol for a move that changed nothing.
	old := h.keyOf(a)
	h.set(a, place("above", b))
	h.assertOrder(a, b)
	if h.keyOf(a) > old {
		t.Errorf("moving A above B gave %q, which sits between the key it was "+
			"giving up (%q) and B (%q): its own key was counted in the gap",
			h.keyOf(a), old, h.keyOf(b))
	}

	// And a call that moves a whole block does not get in its own way
	// either: the three keys of the block come out of the gap the rest of
	// the board leaves, not out of a gap the block itself narrows.
	h = newHarness(t)
	top := h.create("The top", place("ordinal", "last"))
	one := h.create("One", place("ordinal", "last"))
	two := h.create("Two", place("ordinal", "last"))
	three := h.create("Three", place("ordinal", "last"))
	if _, err := SetOn(h.b, h.env, SetParams{
		Refs:    []string{one, two, three},
		Changes: []Change{place("below", top)},
	}); err != nil {
		t.Fatalf("biso set One Two Three --below The top: %v", err)
	}
	h.assertOrder(top, one, two, three)
	for _, id := range []string{one, two, three} {
		if h.keyOf(id) <= h.keyOf(top) {
			t.Errorf("%s ended at %q, which is not below the top %q",
				id, h.keyOf(id), h.keyOf(top))
		}
	}
}

// TestRepeatingAPlacementConverges is the precision of
// docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación
// about asking twice for a place a task already has: the discount takes the
// task's own key out of the gap, so the answer does not depend on where the
// task was, and asking again lands on the very same key. Every one of the
// four placements settles after at most one rewrite.
func TestRepeatingAPlacementConverges(t *testing.T) {
	for _, c := range []struct {
		name  string
		place func(neighbour string) Change
	}{
		{"--ordinal first", func(string) Change { return place("ordinal", "first") }},
		{"--ordinal last", func(string) Change { return place("ordinal", "last") }},
		{"--above", func(neighbour string) Change { return place("above", neighbour) }},
		{"--below", func(neighbour string) Change { return place("below", neighbour) }},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			anchor := h.create("The anchor", place("ordinal", "last"))
			mover := h.create("The one being placed", place("ordinal", "last"))

			h.set(mover, c.place(anchor))
			settled := h.keyOf(mover)
			if settled == "" {
				t.Fatal("the placement wrote no key")
			}

			for i := 0; i < 3; i++ {
				again := h.set(mover, c.place(anchor))
				if h.keyOf(mover) != settled {
					t.Fatalf("call %d moved the key again, from %q to %q",
						i+2, settled, h.keyOf(mover))
				}
				if len(again.Tasks[0].Changed) != 0 {
					t.Errorf("call %d reports the change %v over a task it did not move",
						i+2, again.Tasks[0].Changed)
				}
			}
		})
	}

	// The one rewrite the first call may cost is real and worth pinning:
	// a task that is already the first of the board is moved all the same,
	// because the gap that opens when its own key is discounted reaches the
	// next task and its midpoint is not where the task was.
	h := newHarness(t)
	first := h.create("The first one", place("ordinal", "last"))
	h.create("The second one", place("ordinal", "last"))
	before := h.keyOf(first)

	moved := h.set(first, place("ordinal", "first"))

	if h.keyOf(first) == before {
		t.Errorf("--ordinal first left the key at %q, and the discount widens the gap", before)
	}
	if len(moved.Tasks[0].Changed) != 1 || moved.Tasks[0].Changed[0] != "ordinal" {
		t.Errorf("the call changed %v, want only the ordinal", moved.Tasks[0].Changed)
	}
}

// TestATaskCannotBeItsOwnNeighbour is the self_ordinal_neighbour of
// docs/spec/familias-de-flags.md#el-orden-manual, with its literal message,
// over the task itself and over another task of the same call.
func TestATaskCannotBeItsOwnNeighbour(t *testing.T) {
	h := newHarness(t)
	a := h.create("A", place("ordinal", "last"))
	b := h.create("B", place("ordinal", "last"))
	before := h.keyOf(a)

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{a}, Changes: []Change{place("above", a)},
	})
	assertSpec(t, err, 2, "self_ordinal_neighbour")
	if got := specError(t, err).Message; got != "--above: "+a+" cannot be its own neighbour" {
		t.Errorf("message is %q", got)
	}

	// A task named among the references of the same call is just as much
	// its own neighbour, which is what makes the block rule safe.
	_, err = SetOn(h.b, h.env, SetParams{
		Refs: []string{a, b}, Changes: []Change{place("below", b)},
	})
	assertSpec(t, err, 2, "self_ordinal_neighbour")
	if got := specError(t, err).Message; got != "--below: "+b+" cannot be its own neighbour" {
		t.Errorf("message is %q", got)
	}
	if h.keyOf(a) != before {
		t.Error("the refused call wrote a key all the same")
	}
}

// TestANeighbourWithoutAKeyIsAPreconditionAndNotAUsageError is
// neighbour_without_ordinal, the one code of the manual order that exits 6,
// with the two hints the specification writes out whole.
func TestANeighbourWithoutAKeyIsAPreconditionAndNotAUsageError(t *testing.T) {
	h := newHarness(t)
	loose := h.create("A task with no place")
	mover := h.create("The one being moved", place("ordinal", "last"))
	before := h.keyOf(mover)

	for _, c := range []struct{ flag, where string }{
		{"above", "above"},
		{"below", "below"},
	} {
		_, err := SetOn(h.b, h.env, SetParams{
			Refs: []string{mover}, Changes: []Change{place(c.flag, loose)},
		})
		assertSpec(t, err, 6, "neighbour_without_ordinal")
		e := specError(t, err)
		if e.Message != "--"+c.flag+": "+loose+" has no ordinal" {
			t.Errorf("message is %q", e.Message)
		}
		want := []string{
			"a task without one has no place in the manual order, " +
				"so there is nothing to write " + c.where,
			"`biso set " + loose + " --ordinal last` gives it one, and then --" +
				c.flag + " " + loose + " works",
		}
		if !reflect.DeepEqual(e.Hints, want) {
			t.Errorf("hints are %q, want %q", e.Hints, want)
		}
		if h.keyOf(mover) != before {
			t.Fatal("the refused call moved the task all the same")
		}
	}

	// And the remedy the hint proposes really works, which is the whole
	// reason it is in the message.
	h.set(loose, place("ordinal", "last"))
	h.set(mover, place("above", loose))
	h.assertOrder(mover, loose)
}

// TestANeighbourThatDoesNotExistIsTheOrdinaryRefusal is the sentence of
// that section saying the reference is resolved like any other: a task that
// is not there is exit code 4, and a text that matches several is 5.
func TestANeighbourThatDoesNotExistIsTheOrdinaryRefusal(t *testing.T) {
	h := newHarness(t)
	mover := h.create("The one being moved")
	h.create("Normalize CRLF in the parser", place("ordinal", "last"))
	h.create("Normalize CRLF in the diff", place("ordinal", "last"))

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{mover}, Changes: []Change{place("above", "MYP-900")},
	})
	assertSpec(t, err, 4, "never_allocated")

	_, err = SetOn(h.b, h.env, SetParams{
		Refs: []string{mover}, Changes: []Change{place("below", "Normalize CRLF")},
	})
	assertSpec(t, err, 5, "ambiguous_reference")
}

// TestClearOrdinalTakesTheTaskOutOfTheOrder is the fifth flag, and the one
// that needs no gap: it empties the field, and the task drops to the block
// of the ones with no place.
func TestClearOrdinalTakesTheTaskOutOfTheOrder(t *testing.T) {
	h := newHarness(t)
	a := h.create("A", place("ordinal", "last"))
	b := h.create("B", place("ordinal", "last"))
	h.assertOrder(a, b)

	h.set(a, clear("clear-ordinal"))
	h.assertOrder(b)
	if h.keyOf(a) != "" {
		t.Errorf("--clear-ordinal left the key %q", h.keyOf(a))
	}
	// Which is exactly what makes it stop being a legal neighbour.
	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{b}, Changes: []Change{place("above", a)},
	})
	assertSpec(t, err, 6, "neighbour_without_ordinal")
}

// TestATaskCanBeBornPlaced is the bullet of
// docs/spec/cmd/new.md#parámetros-propios: the four flags mean in `biso new`
// what they mean in `biso set`, and a neighbour without a key refuses the
// creation itself.
func TestATaskCanBeBornPlaced(t *testing.T) {
	h := newHarness(t)
	anchor := h.create("The anchor", place("ordinal", "last"))
	under := h.create("Born under the anchor", place("below", anchor))
	over := h.create("Born over the anchor", place("above", anchor))
	h.assertOrder(over, anchor, under)

	loose := h.create("With no place of its own")
	_, err := NewOn(h.b, h.env, NewParams{
		Title: "Born next to a task with no place", HasTitle: true,
		Changes: []Change{place("below", loose)},
	})
	assertSpec(t, err, 6, "neighbour_without_ordinal")
	all, _, readErr := h.b.Tasks.All()
	if readErr != nil {
		t.Fatal(readErr)
	}
	if len(all) != 4 {
		t.Errorf("the board holds %d tasks and four were created, "+
			"so the refused creation wrote one", len(all))
	}
}

// TestTheGapIsTheWholeBoard is the precision that the key is global: an
// archived task and a finished one count for the gap exactly as any other,
// because where a task lands cannot depend on anybody's filters. Those two
// are precisely the ones a listing hides by default, so they are the ones a
// gap computed over what `biso ls` shows would lose in silence.
//
// Each case puts the hidden task at the end the flag reaches for, which is
// the only arrangement where leaving it out changes the answer: with the
// hidden task in the middle, the visible ones already close the gap.
func TestTheGapIsTheWholeBoard(t *testing.T) {
	for _, c := range []struct {
		name string
		hide func(h *harness, id string)
	}{
		{"an archived task", func(h *harness, id string) {
			if _, err := ArchiveOn(h.b, h.env, ArchiveParams{Refs: []string{id}}); err != nil {
				h.t.Fatalf("biso archive: %v", err)
			}
			if !h.load(id).Archived {
				h.t.Fatal("the task meant to be archived is not archived")
			}
		}},
		{"a finished task", func(h *harness, id string) {
			if _, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}}); err != nil {
				h.t.Fatalf("biso finish: %v", err)
			}
			if got := h.load(id).Status; got != h.b.Config.TerminalStatus {
				h.t.Fatalf("the task meant to be finished is in %q", got)
			}
		}},
	} {
		t.Run(c.name+" at the top", func(t *testing.T) {
			h := newHarness(t)
			visible := h.create("On the board", place("ordinal", "last"))
			hidden := h.create("Hidden, and holding the highest key", place("ordinal", "last"))
			c.hide(h, hidden)

			newcomer := h.create("Placed last of all", place("ordinal", "last"))

			if h.keyOf(newcomer) <= h.keyOf(hidden) {
				t.Errorf("--ordinal last gave %q, which is not past the hidden %q: "+
					"the gap left it out", h.keyOf(newcomer), h.keyOf(hidden))
			}
			h.assertOrder(visible, hidden, newcomer)
		})

		t.Run(c.name+" at the bottom", func(t *testing.T) {
			h := newHarness(t)
			visible := h.create("On the board", place("ordinal", "last"))
			hidden := h.create("Hidden, and holding the lowest key", place("ordinal", "first"))
			c.hide(h, hidden)

			newcomer := h.create("Placed first of all", place("ordinal", "first"))

			if h.keyOf(newcomer) >= h.keyOf(hidden) {
				t.Errorf("--ordinal first gave %q, which is not under the hidden %q: "+
					"the gap left it out", h.keyOf(newcomer), h.keyOf(hidden))
			}
			h.assertOrder(newcomer, hidden, visible)
		})
	}
}

// TestTwoTasksMaySharAKeyAndTheGapStaysUsable is the last precision of the
// gap table: the neighbours are looked for with a strict comparison, so a
// tie never leaves a gap with nothing in it.
func TestTwoTasksMaySharAKeyAndTheGapStaysUsable(t *testing.T) {
	h := newHarness(t)
	// Two tasks with the same key can only arrive through a batch, so they
	// are written straight onto the board here.
	first := h.create("The first of the tie")
	second := h.create("The second of the tie")
	for _, id := range []string{first, second} {
		task := h.load(id)
		task.Ordinal = "m"
		if err := h.b.Tasks.Save(task); err != nil {
			t.Fatalf("writing a key by hand: %v", err)
		}
	}

	over := h.create("Over both of them")
	h.set(over, place("above", first))
	under := h.create("Under both of them")
	h.set(under, place("below", second))

	if h.keyOf(over) >= "m" || h.keyOf(under) <= "m" {
		t.Fatalf("the tie was treated as a gap: %q and %q around m",
			h.keyOf(over), h.keyOf(under))
	}
	// The listing breaks the tie by identifier, which is its rule for any
	// other tie (docs/spec/cmd/ls.md#la-regla-de-orden-completa).
	h.assertOrder(over, first, second, under)
}

// TestAPlacementCountsAsAChangedField is what the status line and the
// `changed` key of the envelope answer: a call that only moves a task has
// changed the ordinal and nothing else.
func TestAPlacementCountsAsAChangedField(t *testing.T) {
	h := newHarness(t)
	anchor := h.create("The anchor", place("ordinal", "last"))
	mover := h.create("The mover")

	result := h.set(mover, place("below", anchor))
	if !reflect.DeepEqual(result.Tasks[0].Changed, []string{"ordinal"}) {
		t.Errorf("changed = %v, want [ordinal]", result.Tasks[0].Changed)
	}
	if len(result.Notes) != 0 {
		t.Errorf("a call that moved a task noted %v", result.Notes)
	}
}

// TestADryRunPlacementWritesNothing is --dry-run over the one flag family
// whose value the program computes: the preview runs the whole resolution,
// refusals included, and the board comes out untouched.
func TestADryRunPlacementWritesNothing(t *testing.T) {
	h := newHarness(t)
	anchor := h.create("The anchor", place("ordinal", "last"))
	mover := h.create("The mover")

	if _, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{mover}, Changes: []Change{place("below", anchor)}, DryRun: true,
	}); err != nil {
		t.Fatalf("the preview failed: %v", err)
	}
	if h.keyOf(mover) != "" {
		t.Errorf("--dry-run wrote the key %q", h.keyOf(mover))
	}

	loose := h.create("With no place")
	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{mover}, Changes: []Change{place("above", loose)}, DryRun: true,
	})
	assertSpec(t, err, 6, "neighbour_without_ordinal")
}

// TestAMalformedKeyInABatchIsItsOwnFailure is the bullet of
// docs/spec/cmd/new.md#el-modo-lote, which is the one place a key arrives
// written: the wrong shape is malformed_ordinal with the line in front, and
// a value that is not text at all is invalid_line.
func TestAMalformedKeyInABatchIsItsOwnFailure(t *testing.T) {
	h := newHarness(t)
	_, err := NewBatchOn(h.b, h.env, BatchParams{Content: strings.Join([]string{
		`{"title":"A well formed one","ordinal":"m8"}`,
		`{"title":"A number where a key goes","ordinal":"3000"}`,
		`{"title":"A key that ends in zero","ordinal":"m0"}`,
		`{"title":"The empty string, which is not a key either","ordinal":""}`,
		``,
	}, "\n")})
	e := specError(t, err)
	if e.ExitCode != 7 || e.Code != "batch_invalid" {
		t.Fatalf("the batch answered %d/%s", e.ExitCode, e.Code)
	}
	want := []string{
		`  line 2: malformed ordinal: "3000" ` +
			`(an ordinal key is made of 0-9 and a-z, and never ends in 0)`,
		`  line 3: malformed ordinal: "m0" ` +
			`(an ordinal key is made of 0-9 and a-z, and never ends in 0)`,
		`  line 4: malformed ordinal: "" ` +
			`(an ordinal key is made of 0-9 and a-z, and never ends in 0)`,
	}
	if !reflect.DeepEqual(e.Detail, want) {
		t.Errorf("the report is %q, want %q", e.Detail, want)
	}
	for _, d := range e.Details {
		if d.Code != "malformed_ordinal" || d.ExitCode != 2 {
			t.Errorf("a failure of the report is %d/%s", d.ExitCode, d.Code)
		}
	}

	// An explicit null, on the other hand, is a task with no place, exactly
	// like a line that does not write the key at all
	// (docs/spec/cmd/new.md#el-modo-lote). It is the one value of that field
	// the empty string does not share, which is why the two are told apart
	// while the line is being read and not afterwards.
	if _, err := NewBatchOn(h.b, h.env, BatchParams{Content: strings.Join([]string{
		`{"id":"MYP-1","title":"An explicit null","ordinal":null}`,
		`{"id":"MYP-2","title":"No key written at all"}`,
		``,
	}, "\n")}); err != nil {
		t.Fatalf("a null ordinal was refused: %v", err)
	}
	for _, id := range []string{"MYP-1", "MYP-2"} {
		if got := h.keyOf(id); got != "" {
			t.Errorf("%s came out with the key %q, want none", id, got)
		}
	}

	// A value of the wrong type is not a malformed key: it is a line the
	// format cannot read, like any other key given a number.
	_, err = NewBatchOn(h.b, h.env, BatchParams{
		Content: `{"title":"A real number","ordinal":3000}` + "\n",
	})
	e = specError(t, err)
	if len(e.Details) != 1 || e.Details[0].Code != "invalid_line" {
		t.Fatalf("a numeric ordinal answered %v", e.Details)
	}
	if !strings.Contains(e.Detail[0], "ordinal: expected text, got number") {
		t.Errorf("the report says %q", e.Detail[0])
	}
}

// TestABatchKeepsTheKeyExactly is the half of the symmetry with
// `biso export` that lives here: a key of a line is stored as it came, with
// nothing recomputed, and two lines may perfectly well share one.
func TestABatchKeepsTheKeyExactly(t *testing.T) {
	h := newHarness(t)
	if _, err := NewBatchOn(h.b, h.env, BatchParams{Content: strings.Join([]string{
		`{"id":"MYP-1","title":"The lowest","ordinal":"00i"}`,
		`{"id":"MYP-2","title":"A long one","ordinal":"zzzzzzzz1"}`,
		`{"id":"MYP-3","title":"The same key as the next","ordinal":"m"}`,
		`{"id":"MYP-4","title":"The same key as the one before","ordinal":"m"}`,
		`{"id":"MYP-5","title":"No key at all"}`,
		``,
	}, "\n")}); err != nil {
		t.Fatalf("the batch failed: %v", err)
	}
	for id, want := range map[string]string{
		"MYP-1": "00i", "MYP-2": "zzzzzzzz1", "MYP-3": "m", "MYP-4": "m", "MYP-5": "",
	} {
		if got := h.keyOf(id); got != want {
			t.Errorf("%s kept the key %q, want %q", id, got, want)
		}
	}
	h.assertOrder("MYP-1", "MYP-3", "MYP-4", "MYP-2")
}
