package main

import (
	"path/filepath"
	"testing"
)

// These are the golden tests of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro: the three
// endings of naming a task that is not there, each one compared character
// for character with the block transcribed from that page.

// boardOfTheThreeMessages is the board those three examples describe: fifty
// two tasks, and a counter left at ninety. MYP-53 is therefore an
// identifier the board handed out at some point and no longer has, and
// MYP-999 one it never handed out.
func boardOfTheThreeMessages(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	dir := filepath.Join(m.home, ".biso", "boards", "my-project-3f9a2b1c")
	m.buildBoard(t, dir, "3f9a2b1c", "My project", "MYP", 52, 0, 90)
	m.write(t, filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	return m
}

func TestTheThreeMessagesOfNotFindingATask(t *testing.T) {
	for _, c := range []struct {
		argv    []string
		code    int
		fixture string
	}{
		{[]string{"get", "MYP-1.1", "--id"}, 2, "ref-malformed-id.txt"},
		{[]string{"get", "MYP-999"}, 4, "ref-never-allocated.txt"},
		{[]string{"get", "MYP-53"}, 4, "ref-not-found.txt"},
	} {
		m := boardOfTheThreeMessages(t)

		got := m.run(t, c.argv...).assertCode(t, c.code)

		assertEqual(t, "$ biso "+shellArgs(c.argv)+"\n"+got.stderr, fixture(t, c.fixture),
			"the refusal of biso "+shellArgs(c.argv))
		assertEqual(t, got.stdout, "", "the standard output of a failed biso get")
	}
}

// TestAnIdentifierWithoutTheFlagIsATextQuery is the other half of the first
// of those three: the same string, written without --id, is what the
// grammar of docs/spec/referencias.md#la-gramática calls anything else, so
// it is a text query and its ending is the one of a text that matches
// nothing.
func TestAnIdentifierWithoutTheFlagIsATextQuery(t *testing.T) {
	m := boardOfTheThreeMessages(t)

	got := m.run(t, "get", "MYP-1.1").assertCode(t, 4)

	assertEqual(t, got.stderr, "error: no task matches \"MYP-1.1\"\n",
		"the refusal of a text that matches nothing")
}
