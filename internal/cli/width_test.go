package cli

import "testing"

// The unit of the column algorithm of docs/spec/cmd/ls.md is the cell of a
// monospaced terminal, and these are the three widths that page fixes: zero
// for a combining mark, two for an East Asian ideograph or an emoji, and one
// for everything else.

func TestCellsCountsTheThreeWidthsOfTheSpecification(t *testing.T) {
	for _, c := range []struct {
		text  string
		cells int
		what  string
	}{
		{"", 0, "the empty string"},
		{"Crash on an empty repository", 28, "plain ASCII"},
		{"café", 4, "a precomposed accent, which is one character of one cell"},
		{"café", 4, "a combining accent, which adds nothing"},
		{"́", 0, "a combining mark on its own"},
		{"日本語", 6, "three ideographs"},
		{"ｆｕｌｌ", 8, "fullwidth Latin"},
		{"🙂", 2, "an emoji"},
		{"a日b", 4, "a mix"},
		{"‍", 0, "a zero width joiner, a format character"},
		{"👩‍🚀", 4, "an emoji joined to another one"},
		{"🇪🇸", 2, "a flag, which is two regional indicators of one cell each"},
	} {
		if got := cells(c.text); got != c.cells {
			t.Errorf("cells(%q) is %d and not %d (%s)", c.text, got, c.cells, c.what)
		}
	}
}

// TestTruncateNeverSplitsAGrapheme is the promise of
// docs/spec/cmd/ls.md#salida: the cut lands on a grapheme boundary, so the
// result can measure less than the cap, and never more.
func TestTruncateNeverSplitsAGrapheme(t *testing.T) {
	for _, c := range []struct {
		text string
		max  int
		want string
	}{
		{"abcdef", 10, "abcdef"},
		{"abcdef", 6, "abcdef"},
		{"abcdef", 3, "abc"},
		{"abcdef", 0, ""},
		// The accent goes with its letter or neither of them goes.
		{"aéx", 2, "aé"},
		// An ideograph measures two, so it does not fit in one cell.
		{"a日b", 2, "a"},
		{"a日b", 3, "a日"},
		// An emoji joined to another one is one grapheme of four cells.
		{"x👩‍🚀y", 3, "x"},
		{"x👩‍🚀y", 5, "x👩‍🚀"},
	} {
		got := truncateCells(c.text, c.max)
		if got != c.want {
			t.Errorf("truncateCells(%q, %d) is %q and not %q", c.text, c.max, got, c.want)
		}
		if cells(got) > c.max {
			t.Errorf("truncateCells(%q, %d) gave %q, which is wider than the cap",
				c.text, c.max, got)
		}
	}
}

// TestTitleIsCutAtAHundredCellsCountingTheDots is step 1 of the format of
// docs/spec/cmd/ls.md#salida: the three dots count towards the hundred, so
// what gets printed never measures more than a hundred cells.
func TestTitleIsCutAtAHundredCellsCountingTheDots(t *testing.T) {
	short := "A short title"
	if got := cutTitle(short); got != short {
		t.Errorf("a title of %d cells was changed: %q", cells(short), got)
	}

	long := ""
	for i := 0; i < 150; i++ {
		long += "a"
	}
	got := cutTitle(long)
	if cells(got) != 100 {
		t.Errorf("a long title was cut to %d cells and not to 100: %q", cells(got), got)
	}
	if got[len(got)-3:] != "..." {
		t.Errorf("a cut title does not end in three dots: %q", got)
	}

	// A title of exactly a hundred cells is not cut, because nothing is
	// being left out.
	exact := long[:100]
	if got := cutTitle(exact); got != exact {
		t.Errorf("a title of exactly 100 cells was cut: %q", got)
	}

	// An ideograph on the boundary: 97 cells cannot hold half of one, so
	// the cut lands at 96 and the whole string measures 99.
	wide := ""
	for i := 0; i < 60; i++ {
		wide += "日"
	}
	cut := cutTitle(wide)
	if cells(cut) != 99 {
		t.Errorf("a title of ideographs was cut to %d cells and not to 99: %q", cells(cut), cut)
	}
}
