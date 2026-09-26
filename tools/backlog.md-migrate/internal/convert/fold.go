package convert

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// foldCase returns s with every rune replaced by the smallest rune in its
// Unicode simple case folding orbit (the cycle unicode.SimpleFold walks).
// This is Unicode case folding, not strings.ToLower: a Greek final sigma
// 'ς' (U+03C2), a capital sigma 'Σ' (U+03A3), and a lowercase sigma 'σ'
// (U+03C3) all fold to the same representative, while strings.ToLower
// alone leaves 'ς' distinct from 'σ' (the lowercase form of 'Σ'). This is
// exactly the behavior docs/spec/vocabularios.md, "Qué es el paso 1",
// requires of the vocabulary matching algorithm's first step; the slug
// algorithm and the scoped label key comparison in this package reuse it
// for the same reason (both are specified as "case folding", not
// lowercasing).
func foldCase(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		b.WriteRune(foldRune(r))
	}
	return b.String()
}

// foldRune returns a stable representative for every rune in r's Unicode
// simple case folding orbit (the cycle unicode.SimpleFold walks): the
// lowercase member of the orbit when the orbit has one, and otherwise the
// numerically smallest member. Preferring the lowercase member is what
// makes foldCase actually produce lowercase text for an ordinary cased
// letter ('P' folds to 'p', not the other way around, even though 'P' is
// the numerically smaller code point); falling back to the smallest member
// for an orbit with no lowercase member at all (a symbol, a digit, or any
// other rune docs/spec/vocabularios.md's case folding leaves untouched)
// still gives every rune in that orbit the same representative, which is
// all normalizeVocabulary and scopedLabelKey's comparison need.
func foldRune(r rune) rune {
	best := r
	bestIsLower := unicode.IsLower(r)
	for f := unicode.SimpleFold(r); f != r; f = unicode.SimpleFold(f) {
		fIsLower := unicode.IsLower(f)
		switch {
		case fIsLower && !bestIsLower:
			best, bestIsLower = f, true
		case fIsLower == bestIsLower && f < best:
			best = f
		}
	}
	return best
}

// stripCombiningMarks decomposes s to Unicode's canonical (NFD) form and
// removes every combining mark (Unicode category Mn), which is how both
// docs/spec/vocabularios.md's normalizar() and this package's slug
// algorithm strip diacritics: 'ñ' becomes 'n' and 'ü' becomes 'u', but 'ß',
// 'ø', and 'æ' are left untouched because NFD has no canonical
// decomposition for them, they are not a letter with a mark on top.
func stripCombiningMarks(s string) string {
	decomposed := norm.NFD.String(s)
	var b strings.Builder
	b.Grow(len(decomposed))
	for _, r := range decomposed {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
