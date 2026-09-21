package model

import "strings"

// This file is docs/spec/modelo-de-datos/orden-manual.md: what an ordinal
// key is, and the midpoint function that finds one inside a gap.
//
// It is the only place in the program that builds a key. Every flag that
// places a task, in every command that writes, ends up calling KeyBetween,
// which is what makes the promises of that page checkable in one spot:
// between any two keys there is another, there is always one below the
// lowest and one above the highest, and none of them ever ends in 0.

// ordinalSymbols is the alphabet of a key, in the order its code points
// already have. A symbol's position in this string is its value as a digit
// of the base 36 fraction a key is, which is why comparing two keys by
// their code points compares the two fractions.
const ordinalSymbols = "0123456789abcdefghijklmnopqrstuvwxyz"

// ordinalBase is how many symbols the alphabet has, and therefore the value
// that sits just past the last one: it is what stands for "no key above" in
// the midpoint function.
const ordinalBase = len(ordinalSymbols)

// OrdinalHint is the parenthetical that every message about a malformed key
// carries, so that the rule is stated where it is broken.
const OrdinalHint = "an ordinal key is made of 0-9 and a-z, and never ends in 0"

// ValidOrdinal answers whether a value is a key: not empty, every symbol out
// of 0-9a-z, and never ending in 0. The empty string is how a task with no
// key is carried through the model, so it is not a key and answers false.
func ValidOrdinal(key string) bool {
	if key == "" || key[len(key)-1] == '0' {
		return false
	}
	for i := 0; i < len(key); i++ {
		if ordinalValue(key[i]) < 0 {
			return false
		}
	}
	return true
}

// ValidateOrdinal answers the error of docs/spec/cmd/new.md#el-modo-lote for
// a key that does not keep its form, and nil when it does. The batch of
// `biso new --from` is the one place where a key arrives written, so it is
// the one place this can fail: every other key is built by KeyBetween.
func ValidateOrdinal(key string) *Error {
	if ValidOrdinal(key) {
		return nil
	}
	return &Error{
		ExitCode: 2,
		Code:     "malformed_ordinal",
		Message:  "malformed ordinal: \"" + key + "\" (" + OrdinalHint + ")",
		Field:    "ordinal",
		Given:    key,
	}
}

// KeyBetween is `clave_entre` of
// docs/spec/modelo-de-datos/orden-manual.md#el-algoritmo-del-punto-medio: the
// key of a task that has to end up strictly after prev and strictly before
// next. Either one may be empty, which means the gap reaches that end of the
// order, and with both empty the answer is the first key of a board that has
// none.
//
// The caller owes it two things, and every caller inside the program keeps
// them: prev and next are keys or empty, and prev sorts before next. What
// comes back is always a key, always inside the gap, and the recursion
// always ends, because each step consumes a symbol of one of the two.
func KeyBetween(prev, next string) string {
	// The common prefix is taken symbol by symbol, with a 0 wherever prev
	// has already run out, which is the same padding that makes "m" and
	// "m0" the same fraction. What is left behind it is the same problem
	// one symbol shorter.
	if next != "" {
		if n := commonOrdinalPrefix(prev, next); n > 0 && n < len(next) {
			return next[:n] + KeyBetween(ordinalTail(prev, n), next[n:])
		}
	}

	// From here the two differ in their first symbol, so only those two
	// matter: whatever else prev carries is below the value of its first
	// symbol, and whatever else next carries is above the value of its own.
	a := firstOrdinalValue(prev, 0)
	b := firstOrdinalValue(next, ordinalBase)
	if b-a > 1 {
		// There is room for a symbol between the two, and the middle one is
		// what keeps the keys growing slowly.
		return string(ordinalSymbols[(a+b)/2])
	}
	if next != "" && len(next) > 1 {
		// The two symbols are consecutive, so nothing fits between them,
		// but next carries more behind its first symbol: that first symbol
		// alone is above prev and below next. It is never a 0, because a
		// next beginning with 0 would have shared a prefix with prev.
		return next[:1]
	}
	// Nothing fits and next is one symbol or is not there at all, so the key
	// grows: it keeps the first symbol of prev and looks for room behind it,
	// where the gap now reaches the top of the alphabet.
	return string(ordinalSymbols[a]) + KeyBetween(ordinalTail(prev, 1), "")
}

// commonOrdinalPrefix is how many symbols of next prev matches, reading a 0
// wherever prev has already run out.
func commonOrdinalPrefix(prev, next string) int {
	n := 0
	for n < len(next) {
		symbol := byte('0')
		if n < len(prev) {
			symbol = prev[n]
		}
		if symbol != next[n] {
			break
		}
		n++
	}
	return n
}

// ordinalTail is what is left of a key once its first n symbols are gone,
// which is the empty string when it had no more than that.
func ordinalTail(key string, n int) string {
	if len(key) <= n {
		return ""
	}
	return key[n:]
}

// ordinalValue is the value of one symbol, or -1 when the character is not
// one of the alphabet.
func ordinalValue(symbol byte) int {
	return strings.IndexByte(ordinalSymbols, symbol)
}

// firstOrdinalValue is the value of the first symbol of a key, and absent
// when the key is empty. A character outside the alphabet reads as 0, which
// only a key stored by something other than this program can be: it keeps
// the arithmetic below inside the alphabet instead of indexing past it.
func firstOrdinalValue(key string, absent int) int {
	if key == "" {
		return absent
	}
	if value := ordinalValue(key[0]); value >= 0 {
		return value
	}
	return 0
}
