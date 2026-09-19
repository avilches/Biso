package match

import "sort"

// Suggest returns the candidates closest to value, at most n of them, for the
// "did you mean" line of the five error messages that
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-mas-parecidas lists.
// Those five differ only in n.
//
// The metric is the Levenshtein distance between normalized forms, the same
// normalization Match uses, so that a candidate typed loosely is as close as
// one typed exactly. A candidate further than the threshold, which is half the
// length of the normalized value rounded up, is dropped; what survives is
// sorted by distance and returned in the board's own spelling.
//
// When nothing is close enough, or the vocabulary to suggest from is empty,
// the result is an empty list and never an error: an error message with no
// suggestion is better than n suggestions that resemble nothing, which would
// invite taking one without checking it.
//
// Ties are broken first by the alphabetical order of the normalized form,
// as the specification says, and then by the candidate's own spelling, so
// that two candidates that normalize the same (the board with both "bar-code"
// and "Bar Code") still come out in a fixed order instead of in whatever
// order the caller happened to assemble its list.
func Suggest(value string, candidates []string, n int) []string {
	if n <= 0 || len(candidates) == 0 {
		return nil
	}

	normalized := []rune(Normalize(value))
	threshold := (len(normalized) + 1) / 2

	type scored struct {
		value      string
		normalized string
		distance   int
	}
	kept := make([]scored, 0, len(candidates))
	for _, c := range candidates {
		cn := Normalize(c)
		d := levenshtein(normalized, []rune(cn))
		if d > threshold {
			continue
		}
		kept = append(kept, scored{value: c, normalized: cn, distance: d})
	}

	sort.SliceStable(kept, func(i, j int) bool {
		if kept[i].distance != kept[j].distance {
			return kept[i].distance < kept[j].distance
		}
		if kept[i].normalized != kept[j].normalized {
			return kept[i].normalized < kept[j].normalized
		}
		return kept[i].value < kept[j].value
	})

	if len(kept) > n {
		kept = kept[:n]
	}
	out := make([]string, len(kept))
	for i, k := range kept {
		out[i] = k.value
	}
	return out
}

// levenshtein is the minimum number of single-character insertions, deletions
// and substitutions that turn a into b, computed with the usual dynamic
// programming table of
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-mas-parecidas. It
// takes runes and not strings because the unit the specification counts is the
// character, and in UTF-8 a character is not a byte.
//
// Only one row of the table is kept at a time, which is the same result with
// memory proportional to the shorter of the two words instead of to their
// product.
func levenshtein(a, b []rune) int {
	if len(a) < len(b) {
		a, b = b, a
	}
	if len(b) == 0 {
		return len(a)
	}

	previous := make([]int, len(b)+1)
	current := make([]int, len(b)+1)
	for j := range previous {
		previous[j] = j
	}

	for i := 1; i <= len(a); i++ {
		current[0] = i
		for j := 1; j <= len(b); j++ {
			if a[i-1] == b[j-1] {
				current[j] = previous[j-1]
				continue
			}
			current[j] = 1 + min(previous[j], current[j-1], previous[j-1])
		}
		previous, current = current, previous
	}
	return previous[len(b)]
}
