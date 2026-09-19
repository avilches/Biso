// Package match implements the vocabulary rules of docs/spec/vocabularios.md:
// the matching algorithm that resolves a typed value against the values a
// board has configured, and the nearest-suggestions algorithm that five
// error messages of the specification promise.
//
// Everything here is a pure function: no I/O, no clock, no globals that
// change. The package depends only on internal/model, to build the errors of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md.
package match

import (
	"sort"
	"strings"
	"unicode"
)

// Normalize is step 1 of docs/spec/vocabularios.md#el-algoritmo-de-coincidencia:
// lowercase according to Unicode, drop the diacritics, and drop every
// separator character. The result is the form in which two values are
// compared, both by Match and by Suggest.
//
// The separators removed are exactly the four the specification lists: space
// (U+0020), tab (U+0009), hyphen-minus (U+002D) and low line (U+005F). No
// other whitespace is a separator, so a no-break space or a newline survives
// normalization and makes the value fail to match, which is what a value
// carrying one deserves.
//
// Dropping the diacritics means canonical decomposition followed by removing
// the combining marks. A combining mark that is already there is removed
// directly; a precomposed letter is folded through the table of fold.go,
// which covers the Latin, Greek and Cyrillic blocks. A letter outside those
// blocks keeps its diacritic, and a letter that is not a diacritic variant of
// another one is never folded: "ß", "ø" and "æ" survive, because none of them
// is a letter with an accent on top.
func Normalize(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range strings.ToLower(s) {
		if r == ' ' || r == '\t' || r == '-' || r == '_' {
			continue
		}
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(fold(r))
	}
	return b.String()
}

// fold returns the base letter of a precomposed letter with a diacritic, or
// the rune itself when it carries none.
func fold(r rune) rune {
	if r < foldFrom[0] || r > foldFrom[len(foldFrom)-1] {
		return r
	}
	i := sort.Search(len(foldFrom), func(i int) bool { return foldFrom[i] >= r })
	if i < len(foldFrom) && foldFrom[i] == r {
		return foldTo[i]
	}
	return r
}
