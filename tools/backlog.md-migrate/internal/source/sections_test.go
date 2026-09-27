package source

import "testing"

// TestFindUnclosedSectionsDoesNotJustCompareTotals exercises the two shapes
// that a naive "count BEGINs vs count ENDs" check would get wrong, per the
// reasoning in findUnclosedSections' doc comment. Neither shape is produced
// by the real Backlog.md CLI; both are constructed directly as a body
// string, since there is no task-level fixture that isolates the ordering
// of markers this precisely.
func TestFindUnclosedSectionsDoesNotJustCompareTotals(t *testing.T) {
	t.Run("a closed pair earlier does not cover a later unrelated BEGIN", func(t *testing.T) {
		body := "<!-- SECTION:NOTES:BEGIN -->\nclosed\n<!-- SECTION:NOTES:END -->\n" +
			"some other text\n" +
			"<!-- SECTION:NOTES:BEGIN -->\nnever closed\n"
		got := findUnclosedSections(body)
		if len(got) != 1 || got[0] != "NOTES" {
			t.Fatalf("findUnclosedSections = %v, want [NOTES]: two BEGINs and one END for the same name must still be unclosed", got)
		}
	})

	t.Run("a stray END before any BEGIN does not close it", func(t *testing.T) {
		// Total counts are equal (one BEGIN, one END), which is exactly
		// the case a naive comparison of totals would wrongly call
		// closed.
		body := "<!-- SECTION:NOTES:END -->\n" +
			"some other text\n" +
			"<!-- SECTION:NOTES:BEGIN -->\nnever closed\n"
		got := findUnclosedSections(body)
		if len(got) != 1 || got[0] != "NOTES" {
			t.Fatalf("findUnclosedSections = %v, want [NOTES]: a stray END before the BEGIN must not close it", got)
		}
	})

	t.Run("a normal closed pair reports nothing", func(t *testing.T) {
		body := "<!-- SECTION:NOTES:BEGIN -->\nclosed\n<!-- SECTION:NOTES:END -->\n"
		got := findUnclosedSections(body)
		if len(got) != 0 {
			t.Fatalf("findUnclosedSections = %v, want none", got)
		}
	})

	t.Run("names are reported in the order they first appear", func(t *testing.T) {
		body := "<!-- SECTION:PLAN:BEGIN -->\nnever closed\n" +
			"<!-- SECTION:NOTES:BEGIN -->\nalso never closed\n"
		got := findUnclosedSections(body)
		want := []string{"PLAN", "NOTES"}
		if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
			t.Fatalf("findUnclosedSections = %v, want %v", got, want)
		}
	})
}
