package ops

import (
	"fmt"

	"biso/internal/model"
)

// This file is the writing half of the manual order
// (docs/spec/familias-de-flags.md#el-orden-manual): turning the one
// placement flag of a call into the keys the tasks it names will take.
//
// It runs once per call, before a single task is touched, for two reasons
// the specification gives. The neighbour is looked for against the board as
// it was before the write, like every other selector of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura. And
// the gap is computed one time, discounting the keys of the tasks this very
// call moves, so that a call that moves a block does not get in its own way
// (docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación).

// placementFlag answers whether a flag of a change places a task in the
// manual order. --clear-ordinal is not one of them: it takes the task out
// of the order, which needs no gap and no neighbour, and it is a --clear-*
// of step 1 like any other.
func placementFlag(flag string) bool {
	switch flag {
	case "ordinal", "above", "below":
		return true
	}
	return false
}

// placement answers the placement change of this call, if it wrote one. The
// parser has already refused a call with two of them, so there is at most
// one.
func (w *writer) placement() (Change, bool) {
	for _, c := range w.changes {
		if placementFlag(c.Flag) {
			return c, true
		}
	}
	return Change{}, false
}

// prepareOrdinal resolves the placement flag of this call into one key per
// task, in the order the references were written, and leaves them ready for
// the scalar step to hand out. `moving` are the tasks the call will write,
// empty for `biso new`, whose one task does not exist yet.
//
// Everything that can refuse the call happens here: the neighbour that does
// not exist, the text that matches several tasks, the task naming itself,
// and the neighbour with no key. All four answer before anything is
// written, which is what makes the promise of the case tables true that not
// one of the tasks of the call is touched.
func (w *writer) prepareOrdinal(moving []*model.Task) error {
	c, ok := w.placement()
	if !ok {
		return nil
	}
	all, err := w.tasks()
	if err != nil {
		return err
	}
	moved := make(map[string]bool, len(moving))
	for _, t := range moving {
		moved[t.ID] = true
	}

	prev, next, err := w.gapOf(c, all, moved)
	if err != nil {
		return err
	}

	// A call that moves several tasks drops them all into that one gap, in
	// the order the references were written: the first takes the midpoint,
	// and each of the rest the midpoint of what is left. That is what makes
	// `biso set A B --below C` leave C, A, B
	// (docs/spec/modelo-de-datos/orden-manual.md#varias-tareas-en-la-misma-llamada).
	count := len(moving)
	if w.newTask {
		count = 1
	}
	w.ordinalKeys = make([]string, 0, count)
	for i := 0; i < count; i++ {
		key := model.KeyBetween(prev, next)
		w.ordinalKeys = append(w.ordinalKeys, key)
		prev = key
	}
	return nil
}

// gapOf is the table of
// docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación: the
// key that has to end up above the tasks being placed and the one that has
// to end up below them, either of which may be empty when the gap reaches
// that end of the order.
//
// It looks at the whole board, archived and finished tasks included, because
// the key is global and where a task lands cannot depend on anybody's
// filters.
func (w *writer) gapOf(c Change, all []*model.Task, moved map[string]bool) (string, string, error) {
	switch c.Flag {
	case "ordinal":
		switch c.Value {
		case "first":
			return "", lowestKey(all, moved), nil
		case "last":
			return highestKey(all, moved), "", nil
		}
		// The parser closes this domain, so anything else here is a bug in
		// whoever built the change and not a case of the specification.
		return "", "", fmt.Errorf("--ordinal: %q is neither first nor last", c.Value)
	case "above", "below":
		neighbour, err := w.neighbourOf(c, all, moved)
		if err != nil {
			return "", "", err
		}
		if c.Flag == "above" {
			return keyBelow(all, moved, neighbour.Ordinal), neighbour.Ordinal, nil
		}
		return neighbour.Ordinal, keyAbove(all, moved, neighbour.Ordinal), nil
	}
	return "", "", fmt.Errorf("--%s does not place a task", c.Flag)
}

// neighbourOf resolves the reference of --above or --below, exactly as any
// other reference is resolved (docs/spec/referencias.md), and then asks the
// two things the specification adds about this one: it is not a task this
// same call is moving, and it has a key.
func (w *writer) neighbourOf(c Change, all []*model.Task, moved map[string]bool) (*model.Task, error) {
	resolved, err := resolveRefWith(w.b, all, c.Value, RefAuto)
	if err != nil {
		// The candidates of an ambiguous reference are printed the way
		// `biso ls` prints a listing, here as everywhere else.
		return nil, withCandidates(w.b, w.env, all, err)
	}
	if resolved.Note != "" {
		w.note(resolved.Note)
	}
	t := resolved.Task
	if moved[t.ID] {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "self_ordinal_neighbour",
			Message:  fmt.Sprintf("--%s: %s cannot be its own neighbour", c.Flag, t.ID),
			Field:    "ordinal",
			Given:    c.Value,
		}
	}
	if t.Ordinal == "" {
		// Exit code 6 and not 2: the call is written correctly and what
		// does not admit it is the state of the board, so the same call
		// works as soon as the neighbour has a key. The message carries the
		// whole remedy because it is two calls and not one
		// (docs/spec/familias-de-flags.md#el-orden-manual).
		return nil, &model.Error{
			ExitCode: 6,
			Code:     "neighbour_without_ordinal",
			Message:  fmt.Sprintf("--%s: %s has no ordinal", c.Flag, t.ID),
			Hints: []string{
				fmt.Sprintf(
					"a task without one has no place in the manual order, "+
						"so there is nothing to write %s", c.Flag),
				fmt.Sprintf(
					"`biso set %s --ordinal last` gives it one, and then --%s %s works",
					t.ID, c.Flag, t.ID),
			},
		}
	}
	return t, nil
}

// lowestKey is the smallest key of the board, and highestKey the largest,
// both leaving out the tasks this call moves. Empty means the board has no
// key at all once those are discounted, and then the gap reaches both ends.
func lowestKey(all []*model.Task, moved map[string]bool) string {
	lowest := ""
	for _, t := range all {
		if t.Ordinal == "" || moved[t.ID] {
			continue
		}
		if lowest == "" || t.Ordinal < lowest {
			lowest = t.Ordinal
		}
	}
	return lowest
}

func highestKey(all []*model.Task, moved map[string]bool) string {
	highest := ""
	for _, t := range all {
		if t.Ordinal == "" || moved[t.ID] {
			continue
		}
		if t.Ordinal > highest {
			highest = t.Ordinal
		}
	}
	return highest
}

// keyBelow is the largest key strictly smaller than one, and keyAbove the
// smallest key strictly larger. The comparison is strict on purpose: two
// tasks may share a key, and a strict one leaves the new task above or below
// both of them instead of inside a gap that does not exist
// (docs/spec/modelo-de-datos/orden-manual.md#el-hueco-de-cada-colocación).
func keyBelow(all []*model.Task, moved map[string]bool, key string) string {
	below := ""
	for _, t := range all {
		if t.Ordinal == "" || moved[t.ID] || t.Ordinal >= key {
			continue
		}
		if t.Ordinal > below {
			below = t.Ordinal
		}
	}
	return below
}

func keyAbove(all []*model.Task, moved map[string]bool, key string) string {
	above := ""
	for _, t := range all {
		if t.Ordinal == "" || moved[t.ID] || t.Ordinal <= key {
			continue
		}
		if above == "" || t.Ordinal < above {
			above = t.Ordinal
		}
	}
	return above
}

// takeOrdinalKey hands out the key of the next task of the call, which is
// what the scalar step writes. The tasks are applied in the order their
// references were written, so taking them in order is the rule of the block.
func (w *writer) takeOrdinalKey() string {
	if w.ordinalNext >= len(w.ordinalKeys) {
		// Every task of the call got a key in prepareOrdinal, so running
		// out means the two loops disagree, which is a bug and not a case
		// of the specification. Answering the last key keeps the write
		// coherent instead of leaving a task with none.
		if len(w.ordinalKeys) == 0 {
			return ""
		}
		return w.ordinalKeys[len(w.ordinalKeys)-1]
	}
	key := w.ordinalKeys[w.ordinalNext]
	w.ordinalNext++
	return key
}
