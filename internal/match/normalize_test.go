package match

import (
	"strings"
	"testing"
)

func TestNormalize(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"already normal", "todo", "todo"},
		{"uppercase", "TODO", "todo"},
		{"mixed case with a space", "To Do", "todo"},
		{"several spaces", "to  do", "todo"},
		{"hyphen", "To-Do", "todo"},
		{"underscore", "TO_DO", "todo"},
		{"embedded tab", "To\tDo", "todo"},
		{"leading and trailing spaces", "  To Do  ", "todo"},
		{"a dot survives", "To Do.", "todo."},
		{"a dot in the middle survives", "To.Do", "to.do"},
		{"precomposed diacritics", "Revisión Técnica", "revisiontecnica"},
		{"decomposed diacritics", "Revisión", "revision"},
		{"german umlaut", "Prüfung", "prufung"},
		{"cedilla", "Façade", "facade"},
		{"greek tonos", "Δοκιμή", "δοκιμη"},
		{"greek final sigma", "Δοκιμές", "δοκιμεσ"},
		{"greek uppercase folds onto the final sigma", "ΔΟΚΙΜΕΣ", "δοκιμεσ"},
		{"the micro sign folds onto mu", "µ", "μ"},
		{"the kelvin sign folds onto k", "K", "k"},
		{"the long s folds onto s", "ſ", "s"},
		{"a sign decomposes like a letter", ";", ";"},
		{"cyrillic breve", "Йод", "иод"},
		{"empty", "", ""},
		{"only separators", "-_ \t", ""},
		{"eszett is not a diacritic", "Straße", "straße"},
		{"slashed o is not a diacritic", "Ø", "ø"},
		{"non latin script is untouched", "設計", "設計"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := Normalize(c.in); got != c.want {
				t.Fatalf("Normalize(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestNormalizeIsCaseFoldingAndNotLowercasing is the difference step 1 of
// docs/spec/vocabularios.md#el-algoritmo-de-coincidencia insists on: these
// pairs are the same word in two spellings, and simple lowercasing leaves
// them as two different strings.
func TestNormalizeIsCaseFoldingAndNotLowercasing(t *testing.T) {
	pairs := [][2]string{
		{"ΔΟΚΙΜΕΣ", "Δοκιμές"},
		{"Σ", "ς"},
		{"Μ", "µ"},
		{"ſ", "s"},
	}
	for _, p := range pairs {
		if got, want := Normalize(p[0]), Normalize(p[1]); got != want {
			t.Errorf("Normalize(%q) = %q and Normalize(%q) = %q, want the same form", p[0], got, p[1], want)
		}
		if strings.ToLower(p[0]) == strings.ToLower(p[1]) {
			t.Errorf("%q and %q no longer tell case folding from lowercasing apart", p[0], p[1])
		}
	}
}

func TestNormalizeIsIdempotent(t *testing.T) {
	for _, in := range []string{"To Do", "Revisión", "TO_DO", "Δοκιμή", "ΔΟΚΙΜΕΣ", "設計"} {
		once := Normalize(in)
		if twice := Normalize(once); twice != once {
			t.Fatalf("Normalize(Normalize(%q)) = %q, want %q", in, twice, once)
		}
	}
}
