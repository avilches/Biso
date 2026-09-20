package cli

import (
	"strings"
	"testing"
)

// The catalog of commandCatalog and the two help blocks say the same thing
// twice: the catalog is data, with a name and a one-line summary, and the
// blocks are a layout, with their columns lined up by hand. Neither can be
// generated from the other, because a layout is not derivable and a
// listing is not parseable, so what keeps them from drifting is this file.

// summaryColumn is where the summary of a command starts in both blocks,
// counted in bytes from the start of the line. It is a single column and
// not a run of spaces because the layout lines them up: the longest line,
// `comment <ref> TEXT`, leaves exactly one space in front of its summary,
// so there is no gap to look for, only a place.
const summaryColumn = 21

// TestTheCatalogAndTheHelpBlocksAgree walks the two blocks and the catalog
// against each other, in both directions: a command in one and not in the
// other fails here, and so does a summary that was reworded in one place
// only.
func TestTheCatalogAndTheHelpBlocksAgree(t *testing.T) {
	printed := map[string]string{}
	for _, line := range append(commandLinesOf(topLevelHelp), commandLinesOf(adminBlock)...) {
		if len(line) <= summaryColumn || line[summaryColumn] == ' ' {
			t.Errorf("a command line of the help does not line its summary up: %q", line)
			continue
		}
		name := strings.Fields(line)[0]
		summary := line[summaryColumn:]
		if first, seen := printed[name]; seen {
			t.Errorf("%s is listed twice, as %q and as %q", name, first, summary)
		}
		printed[name] = summary
	}

	for _, c := range commandCatalog {
		summary, listed := printed[c.Name]
		if !listed {
			t.Errorf("%s is in the catalog and in neither help block", c.Name)
			continue
		}
		if summary != c.Summary {
			t.Errorf("%s reads %q in the help and %q in the catalog", c.Name, summary, c.Summary)
		}
		delete(printed, c.Name)
	}
	for name := range printed {
		t.Errorf("%s is printed by the help and is not in the catalog", name)
	}
}

// TestEveryCommandTheParserTakesIsInTheCatalog is the other direction of
// the same promise: a command biso answers to and `biso help` does not know
// about would be a command nobody can find.
func TestEveryCommandTheParserTakesIsInTheCatalog(t *testing.T) {
	for _, c := range Commands() {
		if !inCatalog(c.Name) {
			t.Errorf("biso takes %s and the catalog does not list it", c.Name)
		}
	}
}

// commandLinesOf is the lines of a help block that list a command. In the
// top-level help those are the ones before `Global options:`, which is
// where the flags start and the commands end; in the administration block
// they are all of them but the heading.
func commandLinesOf(block string) []string {
	var lines []string
	for _, line := range strings.Split(block, "\n") {
		if strings.HasPrefix(line, "Global options:") {
			break
		}
		if !strings.HasPrefix(line, "  ") || strings.HasPrefix(strings.TrimSpace(line), "-") {
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
