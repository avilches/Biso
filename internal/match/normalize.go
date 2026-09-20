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
// case-fold according to Unicode, drop the diacritics, and drop every
// separator character. The result is the form in which two values are
// compared, both by Match and by Suggest.
//
// Case folding, and not simple lowercasing, is what the specification asks
// for, and the difference is not academic: Greek final sigma is a different
// code point from the sigma in the middle of a word, so lowercasing the
// uppercase spelling of a status that ends in one produces a string that is
// not equal to the status typed in lowercase, and the board rejects a value
// that is its own status. Folding puts every spelling of a letter on the same
// representative and the two compare equal.
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
	for _, r := range Simplify(s) {
		if r == ' ' || r == '\t' || r == '-' || r == '_' {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// Simplify is that same step without the separators: it folds the case and
// drops the diacritics, and leaves every other character where it was.
//
// It exists because two derivations of the specification ask for exactly
// this half and would be wrong with the whole one. The board folder's slug
// (docs/spec/resolucion-del-tablero.md#cómo-se-deriva-el-nombre-de-la-carpeta)
// collapses each run of characters that are neither an ASCII letter nor an
// ASCII digit into one hyphen, so "Mi Proyecto" has to still carry its space
// when that step runs, or it would give "miproyecto" instead of
// "mi-proyecto". The task prefix
// (docs/spec/modelo-de-datos/identificadores.md#identificador-de-tarea)
// keeps only the ASCII letters afterwards, which reaches the same answer
// either way, and asking both of them for the same function is what keeps
// "Peña" giving "pena" and "PENA" in one place.
func Simplify(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(fold(foldCase(r)))
	}
	return b.String()
}

// foldCase returns the representative that every case variant of r shares:
// the lowercase of its uppercase. Going through the uppercase is what makes
// the mapping a case folding and not a lowercasing, because the variants that
// lowercasing leaves alone all have the same uppercase and come back from it
// as the same letter: Greek final sigma and Greek sigma both become sigma,
// the micro sign becomes mu, the Kelvin sign becomes k and the long s becomes
// s. A letter with no case, such as a Han ideograph, is its own uppercase and
// its own lowercase, so it comes back untouched, and the sharp s survives too
// because it has no single-rune uppercase to travel through.
func foldCase(r rune) rune {
	if r < 0x80 {
		if 'A' <= r && r <= 'Z' {
			return r + ('a' - 'A')
		}
		return r
	}
	return unicode.ToLower(unicode.ToUpper(r))
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

// FoldCase folds the case of every character and touches nothing else: no
// diacritic is dropped and no separator is removed.
//
// It is the comparison that docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma
// fixes for a label and for a person, and it is deliberately weaker than the
// two above. A label and a person are stored letter for letter, so "Parser"
// and "parser" really are two labels, but a reading filter is not allowed to
// tell them apart. It stops at the case because a label is a short token and
// not prose: it does not share the accent rule of the text selectors, and
// dropping the separators would make "in-progress" and "inprogress" the same
// label, which the vocabulary rules never promised for a field the board does
// not configure.
func FoldCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		b.WriteRune(foldCase(r))
	}
	return b.String()
}
