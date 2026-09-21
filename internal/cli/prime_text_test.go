package cli

import (
	"sort"
	"strings"
	"testing"
)

// The FIELD FLAGS grid of the startup message is literal text, transcribed
// from docs/spec/cmd/prime.md#la-salida-literal, and it is the one place of
// the message where a list of flag names is written by hand: the block
// `--full` adds is generated from the table of fields.go and cannot drift,
// and the grid cannot be generated from that table because its eleven
// lines are a layout, with their own grouping and their own order inside
// each line, which no rule in the table carries.
//
// So the grid is tied to the table by this test instead: the flags it names
// are exactly the flags the table has, and a flag added to one and not to
// the other turns it red on the spot. What the test does not check is the
// layout, which is what the golden file of the message already compares
// character for character.

// commentAuthorFlag is the one entry of the table the grid leaves out. It
// signs a comment instead of writing a field, so the message shows it in
// the COMMANDS block, on the line of `biso comment`, where the person who
// is about to write one will see it.
const commentAuthorFlag = "--comment-author"

func TestTheFieldFlagsGridNamesEveryFlagOfTheTable(t *testing.T) {
	inGrid := flagsOfTheGrid(t)

	var missing []string
	for _, f := range fieldFlags() {
		name := "--" + f.Name
		if name == commentAuthorFlag {
			if inGrid[name] {
				t.Errorf("%s is in the grid, so this exception is stale", name)
			}
			if !strings.Contains(primeCommands, " ["+name+" @who]") {
				t.Errorf("%s is in neither the grid nor the COMMANDS block", name)
			}
			continue
		}
		if !inGrid[name] {
			missing = append(missing, name)
		}
		delete(inGrid, name)
	}
	if len(missing) > 0 {
		t.Errorf("the FIELD FLAGS grid does not name %s", strings.Join(missing, ", "))
	}
	if left := sorted(inGrid); len(left) > 0 {
		t.Errorf("the FIELD FLAGS grid names %s, which the table of fields.go does not have",
			strings.Join(left, ", "))
	}
}

// flagsOfTheGrid is the set of flag names the grid writes.
func flagsOfTheGrid(t *testing.T) map[string]bool {
	t.Helper()
	found := map[string]bool{}
	for _, token := range tokensOfTheGrid(t) {
		found[token] = true
	}
	return found
}

// tokensOfTheGrid is every flag name of the grid, in the order it writes
// them. The heading is left out, and so is any token that is not a flag.
func tokensOfTheGrid(t *testing.T) []string {
	t.Helper()
	lines := strings.Split(strings.TrimSuffix(primeFieldFlags, "\n"), "\n")
	if len(lines) < 2 {
		t.Fatalf("the FIELD FLAGS grid has %d lines", len(lines))
	}
	var tokens []string
	for _, line := range lines[1:] {
		for _, token := range strings.Fields(line) {
			if strings.HasPrefix(token, "-") {
				tokens = append(tokens, token)
			}
		}
	}
	return tokens
}

func sorted(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}
