package convert

import "strings"

// slugify implements the milestone/project slug algorithm of
// docs/especificacion.md, "Milestone y proyecto": Unicode case folding,
// then NFD decomposition with combining marks
// stripped, then every run of one or more characters that is NOT a Unicode
// letter or digit collapsed into a single hyphen, with no leading or
// trailing hyphen. "Puesta en uso" slugifies to "puesta-en-uso", and
// "Implementación" to "implementacion" (the diacritic is dropped, the
// letter is not).
//
// slugify("...") is "", because a title with no letter or digit at all has
// nothing to build a slug from. slugify itself never raises a Finding for
// this: the empty result is exactly what its caller (MilestoneSlugs, or
// ScopedLabels for a project value) checks for, to fall back to something
// else with a Finding of its own.
func slugify(title string) string {
	normalized := stripCombiningMarks(foldCase(title))

	var parts []string
	var current strings.Builder
	flush := func() {
		if current.Len() > 0 {
			parts = append(parts, current.String())
			current.Reset()
		}
	}
	for _, r := range normalized {
		if isLetterOrDigit(r) {
			current.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()

	return strings.Join(parts, "-")
}
