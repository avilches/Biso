package convert

import (
	"fmt"
	"strings"
	"unicode"

	"backlog.md-migrate/internal/source"
)

// isLetterOrDigit reports whether r is a Unicode letter or a Unicode
// digit, the base of both the token alphabet (isTokenAlphabet) and the
// milestone/project slug algorithm (slugify).
func isLetterOrDigit(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

// isTokenAlphabet reports whether r is one of the characters
// docs/spec/valores-de-entrada.md, "El juego de caracteres de un token",
// allows in a labels or assignees value: a Unicode letter or digit, or one
// of the symbols '-', '_', '.', ':', '@'. The colon is included because a
// scoped label (docs/spec/valores-de-entrada.md, "Las etiquetas con
// ámbito") needs it.
func isTokenAlphabet(r rune) bool {
	if isLetterOrDigit(r) {
		return true
	}
	switch r {
	case '-', '_', '.', ':', '@':
		return true
	}
	return false
}

// validTokenAlphabet reports whether every rune of s is in the token
// alphabet. The empty string is vacuously valid, the same as an empty
// range never finding a violation.
func validTokenAlphabet(s string) bool {
	for _, r := range s {
		if !isTokenAlphabet(r) {
			return false
		}
	}
	return true
}

// collapseWhitespaceToHyphens trims Unicode whitespace from both ends of s
// and replaces every internal run of one or more whitespace characters
// (unicode.IsSpace: the ASCII space and tab, and every other character
// with Unicode's White_Space property) with a single '-' (U+002D). changed
// reports whether the result actually differs from s, which is false for a
// value that had no whitespace to begin with.
func collapseWhitespaceToHyphens(s string) (result string, changed bool) {
	trimmed := strings.TrimFunc(s, unicode.IsSpace)

	var b strings.Builder
	b.Grow(len(trimmed))
	inRun := false
	for _, r := range trimmed {
		if unicode.IsSpace(r) {
			if !inRun {
				b.WriteByte('-')
				inRun = true
			}
			continue
		}
		inRun = false
		b.WriteRune(r)
	}

	result = b.String()
	return result, result != s
}

// CleanTokenList applies docs/spec/valores-de-entrada.md, "El juego de
// caracteres de un token", to every value of a single labels or assignees
// list of one task, in this exact order:
//
//  1. Trim Unicode whitespace from both ends and collapse every internal
//     run of whitespace into a single hyphen (docs/decisiones.md, "Los
//     espacios de una etiqueta o un asignado se convierten en guiones").
//     A value this step actually changes raises a Finding; one with no
//     whitespace at all does not.
//  2. Drop a value that, after step 1, still has a character outside the
//     token alphabet, with a Finding.
//  3. Collapse duplicates: when two values of THIS SAME list end up equal
//     after steps 1 and 2, keep only the first one in the original order,
//     with no separate Finding, since the finding from step 1 or 2 that
//     produced the collision already explains it.
//
// field names the caller's list ("labels" or "assignees") only for the
// Finding messages; the rule is identical for both, per
// docs/decisiones.md, "Se aplica también a los asignados".
func CleanTokenList(file, field string, values []string) ([]string, []source.Finding) {
	var findings []source.Finding
	var cleaned []string
	seen := make(map[string]bool, len(values))

	for _, v := range values {
		collapsed, changed := collapseWhitespaceToHyphens(v)
		if changed {
			findings = append(findings, source.Finding{
				File:  file,
				Field: field,
				Message: fmt.Sprintf(
					"value %q converted to %q, biso's token alphabet has no space",
					v, collapsed,
				),
			})
		}

		if !validTokenAlphabet(collapsed) {
			findings = append(findings, source.Finding{
				File:  file,
				Field: field,
				Message: fmt.Sprintf(
					"value %q dropped, it has a character outside biso's token alphabet",
					collapsed,
				),
			})
			continue
		}

		if seen[collapsed] {
			continue
		}
		seen[collapsed] = true
		cleaned = append(cleaned, collapsed)
	}

	return cleaned, findings
}
