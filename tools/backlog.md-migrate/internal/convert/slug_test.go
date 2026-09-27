package convert

import "testing"

// TestSlugifyExamples reproduces the two exact examples
// docs/especificacion.md, "Milestone y proyecto", gives.
func TestSlugifyExamples(t *testing.T) {
	cases := map[string]string{
		"Puesta en uso":  "puesta-en-uso",
		"Implementación": "implementacion",
	}
	for title, want := range cases {
		if got := slugify(title); got != want {
			t.Errorf("slugify(%q) = %q, want %q", title, got, want)
		}
	}
}

// TestSlugifyOfATitleWithNoLetterOrDigitIsEmpty covers
// docs/especificacion.md, "Milestone y proyecto"'s example: a title with
// no letter or digit at all (only dots) slugifies to nothing.
func TestSlugifyOfATitleWithNoLetterOrDigitIsEmpty(t *testing.T) {
	if got := slugify("..."); got != "" {
		t.Errorf("slugify(\"...\") = %q, want empty", got)
	}
}
