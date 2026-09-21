package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// These tests run the rule of docs/spec/cmd/get.md#salida for a list in the
// card through the compiled program: the values a task holds are the ones
// --json gives, exact and never escaped, and the card writes each of them
// with the escape of the input read backwards.

// refsBoard is a board with one task, MYP-1, holding the references that are
// given the way the input reads them (a backslash and a comma escape).
func refsBoard(t *testing.T, addRefs string) *machine {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "A task", "--add-refs", addRefs).assertCode(t, 0)
	return m
}

// refsLine is the line of the card that starts with "refs", without its
// terminator.
func refsLine(t *testing.T, card string) string {
	t.Helper()
	var found []string
	for _, line := range strings.Split(card, "\n") {
		if strings.HasPrefix(line, "refs ") {
			found = append(found, line)
		}
	}
	if len(found) != 1 {
		t.Fatalf("the card has %d refs lines, want 1:\n%s", len(found), card)
	}
	return found[0]
}

func TestTheCardEscapesTheCommaOfAReference(t *testing.T) {
	m := refsBoard(t, `a\,b,c.md`)

	got := m.run(t, "get", "MYP-1", "--section", "meta").assertCode(t, 0)

	assertEqual(t, refsLine(t, got.stdout), `refs       a\,b, c.md`, "the refs line of the card")
}

func TestTheCardEscapesTheBackslashOfAReference(t *testing.T) {
	// Two references, `C:\dir\` and `notes/b.md`, given with the escape.
	m := refsBoard(t, `C:\\dir\\,notes/b.md`)

	got := m.run(t, "get", "MYP-1").assertCode(t, 0)

	assertEqual(t, refsLine(t, got.stdout), `refs       C:\\dir\\, notes/b.md`, "the refs line of the card")
}

func TestTheCardTellsTwoReferencesFromOneWithACommaAndASpace(t *testing.T) {
	// The references `a\` and `b` on one board, the single reference `a, b`
	// on another: they must not print the same line.
	two := refsBoard(t, `a\\,b`)
	one := refsBoard(t, `a\, b`)

	twoLine := refsLine(t, two.run(t, "get", "MYP-1").assertCode(t, 0).stdout)
	oneLine := refsLine(t, one.run(t, "get", "MYP-1").assertCode(t, 0).stdout)

	assertEqual(t, twoLine, `refs       a\\, b`, "the line of the references a\\ and b")
	assertEqual(t, oneLine, `refs       a\, b`, "the line of the reference a, b")
}

func TestTheJsonOfATaskKeepsTheReferencesExactAndUnescaped(t *testing.T) {
	m := refsBoard(t, `a\,b,C:\\dir\\,p\q,c.md`)

	got := m.run(t, "get", "MYP-1", "--json").assertCode(t, 0)

	var envelope struct {
		Data struct {
			Task struct {
				References []string `json:"references"`
			} `json:"task"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(got.stdout), &envelope); err != nil {
		t.Fatalf("the output is not JSON: %v\n%s", err, got.stdout)
	}
	want := []string{"a,b", `C:\dir\`, `p\q`, "c.md"}
	if strings.Join(envelope.Data.Task.References, "\x00") != strings.Join(want, "\x00") {
		t.Errorf("references in --json = %q, want %q", envelope.Data.Task.References, want)
	}
	// The card of the same task escapes them, the JSON does not.
	card := m.run(t, "get", "MYP-1", "--section", "meta").assertCode(t, 0)
	assertEqual(t, refsLine(t, card.stdout), `refs       a\,b, C:\\dir\\, p\\q, c.md`, "the refs line of the card")
}

// TestTheValueOfTheCardIsTheOneThatTheFlagTakes is the reason for the rule:
// what the card shows for a reference, copied into --rm-refs, removes it.
func TestTheValueOfTheCardIsTheOneThatTheFlagTakes(t *testing.T) {
	m := refsBoard(t, `a\,b,C:\\dir\\,c.md`)
	line := refsLine(t, m.run(t, "get", "MYP-1").assertCode(t, 0).stdout)
	if line != `refs       a\,b, C:\\dir\\, c.md` {
		t.Fatalf("the refs line is %q", line)
	}

	// The whole line, after its label, is a valid value of the flag once the
	// separating spaces are dropped: the same escape reads both ways.
	value := strings.ReplaceAll(strings.TrimPrefix(line, "refs       "), ", ", ",")
	got := m.run(t, "set", "MYP-1", "--rm-refs", value).assertCode(t, 0)

	if strings.Contains(got.stderr, "not present") {
		t.Errorf("a reference copied from the card was not found: %q", got.stderr)
	}
	after := m.run(t, "get", "MYP-1").assertCode(t, 0)
	assertEqual(t, refsLine(t, after.stdout), "refs       -", "the refs line once all of them are removed")
}

func TestTheOtherListsOfTheCardAreUnchanged(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@avilches"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	m.run(t, "new", "Blocker").assertCode(t, 0)
	m.run(t, "new", "Blocked", "--add-deps", "MYP-1",
		"--add-labels", "one,two", "--add-assignees", "@claude,@sara").assertCode(t, 0)

	got := m.run(t, "get", "MYP-2", "--section", "meta").assertCode(t, 0)

	for _, want := range []string{"@claude, @sara", "one, two", "MYP-1"} {
		if !strings.Contains(got.stdout, want) {
			t.Errorf("the card lacks %q:\n%s", want, got.stdout)
		}
	}
	assertEqual(t, refsLine(t, got.stdout), "refs       -", "the refs line of a task without references")
}
