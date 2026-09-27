package convert

import (
	"fmt"
	"strings"

	"backlog.md-migrate/internal/source"
)

// normalizeVocabulary implements normalizar(x) from
// docs/spec/vocabularios.md, "El algoritmo de coincidencia": Unicode case
// folding, then NFD decomposition with combining marks stripped, then
// removing every space (U+0020), tab (U+0009), hyphen ('-', U+002D), or
// underscore ('_', U+005F). No other whitespace is a separator here: a
// non-breaking space or a line break survives, exactly as the
// specification requires.
//
// This is used only by MatchVocabulary, for status/type/priority. The
// milestone/project slug (slugify) and the scoped label key comparison
// (scopedLabelKey's caller) each strip something different and have their
// own function: reusing this one for them would be wrong, not just
// redundant, because it does not collapse runs of removed characters into
// a hyphen the way a slug does.
func normalizeVocabulary(x string) string {
	folded := stripCombiningMarks(foldCase(x))
	var b strings.Builder
	b.Grow(len(folded))
	for _, r := range folded {
		switch r {
		case ' ', '\t', '-', '_':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// MatchVocabulary implements coincidir(v, configured) from
// docs/spec/vocabularios.md, "El algoritmo de coincidencia":
//
//  1. An exact string match (no folding at all) wins outright.
//  2. Otherwise, v and every configured value are each run through
//     normalizeVocabulary; a configured value that repeats letter for
//     letter is counted once, not once per repetition.
//  3. Exactly one normalized match returns it with ok=true.
//  4. Zero matches, or two or more matches that are themselves distinct
//     configured values, return ok=false: this tool treats that the same
//     way in both cases, as a finding to report rather than biso's own
//     error 3 (see MatchField).
func MatchVocabulary(v string, configured []string) (match string, ok bool) {
	for _, c := range configured {
		if c == v {
			return c, true
		}
	}

	normalizedV := normalizeVocabulary(v)
	seenExact := make(map[string]bool, len(configured))
	var distinct []string
	for _, c := range configured {
		if seenExact[c] {
			continue
		}
		if normalizeVocabulary(c) != normalizedV {
			continue
		}
		seenExact[c] = true
		distinct = append(distinct, c)
	}

	if len(distinct) == 1 {
		return distinct[0], true
	}
	return "", false
}

// MatchField matches a single status, type, or priority value against the
// destination's configured vocabulary for that field, reusing
// MatchVocabulary for all three rather than duplicating the algorithm,
// since docs/spec/vocabularios.md, "El algoritmo de coincidencia", defines
// coincidir(v, configured) once for all three fields.
//
// value is the source task's raw value for the field; an empty value (the
// task simply has no status/type/priority) is not attempted and raises no
// finding, since there is nothing to match. A non-empty value that
// MatchVocabulary cannot resolve unambiguously returns "" together with a
// Finding naming the original value and the destination's configured
// values: this tool never aborts on that, it omits the field and reports it
// instead of biso's own error 3.
func MatchField(file, field, value string, configured []string) (string, []source.Finding) {
	if value == "" {
		return "", nil
	}
	if match, ok := MatchVocabulary(value, configured); ok {
		return match, nil
	}
	return "", []source.Finding{{
		File:  file,
		Field: field,
		Message: fmt.Sprintf(
			"unmatched %s %q, configured values on the destination: %s",
			field, value, strings.Join(configured, ", "),
		),
	}}
}
