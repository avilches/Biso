package convert

import (
	"fmt"
	"sort"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

// This is phase 4c of the conversion engine (docs/especificacion.md, "Orden
// manual", and docs/decisiones.md, "El orden manual se recalcula, no se
// copia"). It computes the final manual order key for every source task
// that is going to produce a line of output, that is, every source task
// AssignOrdinals is given: a task phase 4b skipped because it was already
// on the destination never reaches this file at all, and a task with no
// ordinal in the source never receives a key either. This file never reads
// or writes anything: it is a pure computation on top of source.Task and
// destination.Board, exactly like identifiers.go's naturalSourceIDLess,
// which it reuses unchanged for its own tie break.

// ordinalAlphabet is the 36 symbols a manual order key is built from, in
// order: "0123456789abcdefghijklmnopqrstuvwxyz". A symbol's value is its
// index in this string, 0 for '0' to 35 for 'z'
// (docs/spec/modelo-de-datos/orden-manual.md, "El algoritmo del punto
// medio").
const ordinalAlphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

// symbolValue returns c's position in ordinalAlphabet. It is only ever
// called on a byte that already came from a valid key (a symbol out of
// '0'-'9' or 'a'-'z'), which KeyBetween's own inputs and outputs always
// are, so an unrecognized byte is a bug in the caller, not a data problem
// to report: it panics rather than returning a wrong value silently.
func symbolValue(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 10
	default:
		panic(fmt.Sprintf("convert: %q is not a manual order key symbol", c))
	}
}

// KeyBetween implements clave_entre from
// docs/spec/modelo-de-datos/orden-manual.md, "El algoritmo del punto medio",
// literally, step by step, including its recursion. before is "anterior",
// the key that has to stay above (the lesser of the two), and after is
// "siguiente", the key that has to stay below (the greater of the two).
// Either one, or both, can be missing: since a real key is by definition a
// non-empty string, this file uses the empty string as "nada" (absent) for
// both parameters and for both of the recursion's own tails, exactly as the
// specification's step 1 defines "no anterior" to behave like the empty
// string for the rest of the algorithm.
//
// The result is always a valid key, strictly greater than before and
// strictly less than after, provided the precondition holds: after, when
// not empty, is strictly greater than before by the same code point
// comparison KeyBetween itself uses. Callers of this file (AssignOrdinals)
// never violate it, because they always call KeyBetween with after equal to
// "nada": every key it produces only ever has to stay above whatever came
// before it, never below anything.
func KeyBetween(before, after string) string {
	// Step 2: "si hay siguiente", take the common prefix of before and
	// after, reading before as '0' past its own end. The precondition
	// (after strictly greater than before) guarantees this common prefix
	// can never consume all of after, so afterTail below is never empty;
	// this is a defensive check for that invariant, not a case that can
	// happen, exactly as the specification allows.
	if after != "" {
		i := 0
		for i < len(after) {
			beforeSymbol := byte('0')
			if i < len(before) {
				beforeSymbol = before[i]
			}
			if beforeSymbol != after[i] {
				break
			}
			i++
		}
		if i == len(after) {
			panic("convert: KeyBetween precondition violated, after is not strictly greater than before")
		}
		if i > 0 {
			beforeTail := ""
			if i < len(before) {
				beforeTail = before[i:]
			}
			afterTail := after[i:]
			// The matched prefix is taken from after, not before: before
			// can be shorter than i (its own symbols were padded with '0'
			// past its end while comparing), and after[:i] holds the same
			// symbols the comparison above already found equal to
			// before's, real or padded, at every one of these i
			// positions.
			return after[:i] + KeyBetween(beforeTail, afterTail)
		}
	}

	// Step 3: the two keys (or their tails) now differ in their very first
	// symbol. a is before's first symbol's value, or 0 if before is empty;
	// b is after's first symbol's value, or 36 (one past the alphabet) if
	// after is empty.
	a := 0
	if before != "" {
		a = symbolValue(before[0])
	}
	b := 36
	if after != "" {
		b = symbolValue(after[0])
	}

	// Step 4: a value fits strictly between a and b.
	if b-a > 1 {
		return string(ordinalAlphabet[(a+b)/2])
	}

	// Step 5: nothing fits, but after has more than one symbol: reuse its
	// own first symbol, which is enough to stay strictly below the rest of
	// after and strictly above before (whose first symbol is a, one less
	// than after's, since nothing fit between them).
	if len(after) > 1 {
		return string(after[0])
	}

	// Step 6: nothing fits, and after is either absent or a single symbol:
	// the result keeps before's own first symbol (or '0' if before was
	// empty) and grows a new symbol out of before's own tail.
	beforeTail := ""
	if before != "" {
		beforeTail = before[1:]
	}
	return string(ordinalAlphabet[a]) + KeyBetween(beforeTail, "")
}

// DestinationAnchor is the ancla del destino
// (docs/especificacion.md, "Orden manual"): the greatest manual order key
// among every task already on the destination board, compared by Unicode
// code point exactly as KeyBetween itself compares (plain Go string
// comparison, since a key's alphabet is ASCII), or the empty string
// ("nada", no anchor at all) when no task on the destination has a manual
// order key.
//
// destination.Task.Ordinal is already the empty string for a task with no
// manual order at all (see destination.go), so this needs no separate
// notion of "missing" beyond the same empty-string convention KeyBetween
// uses for "nada".
func DestinationAnchor(board destination.Board) string {
	anchor := ""
	for _, t := range board.Tasks {
		if t.Ordinal == "" {
			continue
		}
		if t.Ordinal > anchor {
			anchor = t.Ordinal
		}
	}
	return anchor
}

// AssignOrdinals computes the final manual order key for every task in
// tasks that has a source ordinal, following
// docs/especificacion.md, "Orden manual", and docs/decisiones.md, "El
// orden manual se recalcula, no se copia":
//
//  1. The anchor is DestinationAnchor(board).
//  2. Only the tasks whose source.Task.Ordinal is not nil receive a key. A
//     task with no ordinal at all is simply absent from tasks, or absent
//     from the result if it is present with Ordinal == nil.
//  3. Those tasks are ordered by Ordinal ascending; a tie is broken between
//     the tied tasks by naturalSourceIDLess on their own source id, the
//     exact same tie break identifiers.go (phase 4b) uses to reassign a
//     colliding id's number, so the two phases agree on what "the source
//     id's own natural order" means and a batch converts to the same
//     result every time it runs.
//  4. Keys are assigned in that order as if the whole batch were placed
//     with "--ordinal last", one task after another, behind everything
//     already on the destination: the first task takes
//     KeyBetween(anchor, ""), and every following task takes
//     KeyBetween(the key just assigned, "").
//
// tasks must already be the tasks that are going to produce a line of
// output: this file does not know about phase 4b's own notion of "already
// on the destination, skipped", so a caller filters those out before
// calling this function. Passing a skipped task in by mistake would give it
// a key it should never have.
//
// The result maps a source task's own id (source.Task.ID, exactly as read,
// the same join key Identified.SourceID uses) to its assigned key, with one
// entry only for a task that received one. None of the keys this function
// returns can ever collide with a key already on the destination: each one
// is built by KeyBetween(previous, "") from either the anchor or the key
// assigned right before it, and KeyBetween's own guarantee is that its
// result is always strictly greater than its first argument, so every
// assigned key is strictly greater than the anchor (which is itself the
// greatest key already on the destination, or nothing at all) and strictly
// greater than every key assigned before it in the same call. That chain of
// strict inequalities places every assigned key in the empty stretch of the
// order that starts right after the destination's own greatest key,
// regardless of what the destination's keys actually are.
func AssignOrdinals(tasks []source.Task, board destination.Board) map[string]string {
	type entry struct {
		task    source.Task
		parsed  parsedSourceID
		ordinal float64
	}

	var withOrdinal []entry
	for _, t := range tasks {
		if t.Ordinal == nil {
			continue
		}
		parsed, ok := parseSourceID(t.ID)
		if !ok {
			// Every task that reaches this function already produces a
			// line of output, which means it already went through phase
			// 4b's validateSourceShape, which requires this exact shape
			// for every task in the batch, skipped or not. A task whose id
			// does not parse here would mean a caller passed in a task
			// that never went through that validation, which is a bug in
			// the caller, not a data problem this file can report.
			panic(fmt.Sprintf("convert: AssignOrdinals given task %q with an id that does not parse as a source id", t.ID))
		}
		withOrdinal = append(withOrdinal, entry{task: t, parsed: parsed, ordinal: *t.Ordinal})
	}

	sort.SliceStable(withOrdinal, func(i, j int) bool {
		if withOrdinal[i].ordinal != withOrdinal[j].ordinal {
			return withOrdinal[i].ordinal < withOrdinal[j].ordinal
		}
		return naturalSourceIDLess(withOrdinal[i].parsed, withOrdinal[j].parsed)
	})

	result := make(map[string]string, len(withOrdinal))
	anchor := DestinationAnchor(board)
	for _, e := range withOrdinal {
		key := KeyBetween(anchor, "")
		result[e.task.ID] = key
		anchor = key
	}
	return result
}
