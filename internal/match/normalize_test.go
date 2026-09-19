package match

import "testing"

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

func TestNormalizeIsIdempotent(t *testing.T) {
	for _, in := range []string{"To Do", "Revisión", "TO_DO", "Δοκιμή", "設計"} {
		once := Normalize(in)
		if twice := Normalize(once); twice != once {
			t.Fatalf("Normalize(Normalize(%q)) = %q, want %q", in, twice, once)
		}
	}
}
