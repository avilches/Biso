package main

import (
	"strings"
	"testing"
)

// The help of `biso set` and `biso new` has to say which way a dependency
// points, because writing it the wrong way round is valid and nothing in the
// program can detect it (docs/decisiones/detalles.md, "La ayuda enseña la
// dirección de una dependencia").
func TestTheHelpOfSetAndNewSaysWhichWayADependencyPoints(t *testing.T) {
	m := newMachine(t)

	for _, c := range []struct {
		command string
		phrases []string
	}{
		{"set", []string{
			"MYP-4 blocks MYP-10:",
			"biso set MYP-10 --add-deps MYP-4",
			"written on the task that waits, never on the one that blocks",
			"--parent <ref>",
			"tasks that must be done before this one, so each blocks it",
			"--add-refs <text>",
		}},
		{"new", []string{
			"means MYP-4 goes first and blocks it",
			"biso set MYP-10 --add-deps <new id>",
			"--parent <ref>",
			"tasks that must be done first, so each blocks the new task",
			"--add-refs <text>",
		}},
	} {
		// Compared with the whitespace collapsed, so that a phrase the
		// help wraps onto two lines still counts as one phrase.
		got := strings.Join(strings.Fields(m.run(t, c.command, "--help").assertCode(t, 0).stdout), " ")
		for _, phrase := range c.phrases {
			if !strings.Contains(got, phrase) {
				t.Errorf("biso %s --help does not say %q:\n%s", c.command, phrase, got)
			}
		}
	}

	set := m.run(t, "set", "--help").assertCode(t, 0).stdout
	if strings.Contains(set, "no rule to learn beyond the name") {
		t.Errorf("biso set --help still claims that the names are the whole rule:\n%s", set)
	}
}
