package match

import (
	"strings"
	"testing"
)

// exampleLabels is the label set of the example in
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-mas-parecidas.
var exampleLabels = []string{
	"api", "backend", "bug", "docs", "frontend",
	"infra", "parser", "security", "ui", "urgent",
}

func TestLevenshtein(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"", "abc", 3},
		{"abc", "", 3},
		{"abc", "abc", 0},
		{"fronted", "frontend", 1},
		{"kitten", "sitting", 3},
		{"flaw", "lawn", 2},
		{"revision", "revisin", 1},
		{"todo", "done", 3},
	}

	for _, c := range cases {
		t.Run(c.a+"/"+c.b, func(t *testing.T) {
			if got := levenshtein([]rune(c.a), []rune(c.b)); got != c.want {
				t.Fatalf("levenshtein(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
			}
			if got := levenshtein([]rune(c.b), []rune(c.a)); got != c.want {
				t.Fatalf("levenshtein(%q, %q) = %d, want %d, the metric is symmetric", c.b, c.a, got, c.want)
			}
		})
	}
}

// TestLevenshteinCountsCharactersNotBytes guards the one mistake this metric
// invites in Go: measuring len(string) instead of the number of runes.
func TestLevenshteinCountsCharactersNotBytes(t *testing.T) {
	if got := levenshtein([]rune("días"), []rune("dias")); got != 1 {
		t.Fatalf("levenshtein of two four-character words = %d, want 1", got)
	}
}

// TestSuggestTable walks the table of the example in
// docs/spec/vocabularios.md#el-algoritmo-de-sugerencias-mas-parecidas.
func TestSuggestTable(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{"a typo one edit away", "fronted", []string{"frontend"}},
		{"nothing close enough", "xyz", nil},
		{"an exact candidate", "backend", []string{"backend"}},
		{"a candidate spelled loosely", "Front-End", []string{"frontend"}},
		{"nothing to suggest from an empty input", "", nil},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Suggest(c.in, exampleLabels, 5)
			if !equal(got, c.want) {
				t.Fatalf("Suggest(%q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestSuggestOnAnEmptyVocabulary is the board with no labels at all: the
// result is an empty list, never an error and never an invented value.
func TestSuggestOnAnEmptyVocabulary(t *testing.T) {
	if got := Suggest("frontend", nil, 5); len(got) != 0 {
		t.Fatalf("Suggest on a board with no labels = %q, want nothing", got)
	}
}

// TestSuggestOrdersByDistanceNotAlphabetically is the explicit promise of the
// specification: the closest first, whatever its initial letter.
func TestSuggestOrdersByDistanceNotAlphabetically(t *testing.T) {
	candidates := []string{"zzbug", "bugs", "bug"}
	want := []string{"bug", "bugs", "zzbug"}
	if got := Suggest("bug", candidates, 5); !equal(got, want) {
		t.Fatalf("Suggest = %q, want %q", got, want)
	}
}

// TestSuggestBreaksTiesAlphabetically covers the tie: two candidates at the
// same distance come out in alphabetical order of their normalized form.
func TestSuggestBreaksTiesAlphabetically(t *testing.T) {
	candidates := []string{"Baz", "bar", "BAT"}
	want := []string{"bar", "BAT", "Baz"}
	if got := Suggest("ba", candidates, 5); !equal(got, want) {
		t.Fatalf("Suggest = %q, want %q, ties in alphabetical order of the normalized form", got, want)
	}
}

// TestSuggestBreaksATieOfNormalizedForms is the second tie-breaker, the one
// the specification did not have to name until two candidates normalized the
// same: the original spelling decides, so the order never depends on the
// order the caller happened to build its list in.
func TestSuggestBreaksATieOfNormalizedForms(t *testing.T) {
	candidates := []string{"bar-code", "Bar Code", "barcode"}
	want := []string{"Bar Code", "bar-code", "barcode"}
	if got := Suggest("barcod", candidates, 5); !equal(got, want) {
		t.Fatalf("Suggest = %q, want %q", got, want)
	}
}

// TestSuggestThreshold checks step 3 literally: the threshold is half the
// length of the normalized input, rounded up, and a candidate exactly at the
// threshold is kept while the next one out is dropped.
func TestSuggestThreshold(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		candidate string
		want      bool
	}{
		{"five characters, distance 3 is exactly the threshold", "abcde", "ab", true},
		{"five characters, distance 4 is over the threshold", "abcde", "a", false},
		{"three characters, distance 2 is exactly the threshold", "bug", "bugle", true},
		{"three characters, distance 3 is over the threshold", "bug", "bugles", false},
		{"four characters, distance 2 is exactly the threshold", "docs", "do", true},
		{"four characters, distance 3 is over the threshold", "docs", "d", false},
		{"empty input admits only an empty candidate", "", "_-", true},
		{"empty input rejects anything else", "", "a", false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := Suggest(c.in, []string{c.candidate}, 5)
			if c.want && len(got) != 1 {
				t.Fatalf("Suggest(%q, %q) = %q, want it kept", c.in, c.candidate, got)
			}
			if !c.want && len(got) != 0 {
				t.Fatalf("Suggest(%q, %q) = %q, want it dropped", c.in, c.candidate, got)
			}
		})
	}
}

// TestSuggestCaps checks step 5: at most N, and N is the only thing that
// changes between the five places of the specification that promise a list.
func TestSuggestCaps(t *testing.T) {
	candidates := []string{"aa", "ab", "ac", "ad", "ae", "af", "ag"}
	for _, n := range []int{1, 3, 5} {
		got := Suggest("aa", candidates, n)
		if len(got) != n {
			t.Fatalf("Suggest with N = %d returned %d values: %q", n, len(got), got)
		}
	}
	if got := Suggest("aa", candidates, 100); len(got) != len(candidates) {
		t.Fatalf("Suggest with N above the number of candidates returned %q", got)
	}
	if got := Suggest("aa", candidates, 0); len(got) != 0 {
		t.Fatalf("Suggest with N = 0 returned %q", got)
	}
}

// TestSuggestReturnsTheConfiguredSpelling checks that what comes back is the
// candidate as the board spells it, not the normalized form the comparison
// used: the hint tells the reader what to type.
func TestSuggestReturnsTheConfiguredSpelling(t *testing.T) {
	got := Suggest("inprogres", []string{"In Progress"}, 3)
	if len(got) != 1 || got[0] != "In Progress" {
		t.Fatalf("Suggest = %q, want the candidate as it is spelled", got)
	}
}

// TestSuggestFitsTheAssigneeExample reproduces the second example of
// docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué, where the
// candidates carry the @ that every person's name carries.
func TestSuggestFitsTheAssigneeExample(t *testing.T) {
	people := []string{"@claude", "@sara", "@avilches"}
	got := Suggest("@clude", people, 5)
	if len(got) != 1 || got[0] != "@claude" {
		t.Fatalf("Suggest = %q, want only @claude", got)
	}
}

// TestSuggestDoesNotTouchItsArguments guards the purity of the package.
func TestSuggestDoesNotTouchItsArguments(t *testing.T) {
	candidates := []string{"urgent", "ui", "api"}
	before := strings.Join(candidates, ",")
	Suggest("ur", candidates, 5)
	if after := strings.Join(candidates, ","); after != before {
		t.Fatalf("Suggest reordered its candidates: %q", candidates)
	}
}

func equal(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}
