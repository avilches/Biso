package cli

import "testing"

// TestUseColor walks the table of
// docs/spec/salida-y-terminal.md#interactividad-terminal-y-color row by row,
// in the order it is written, because the first row that applies decides.
func TestUseColor(t *testing.T) {
	for _, c := range []struct {
		name       string
		when       ColorWhen
		noColor    bool
		isTerminal bool
		want       bool
	}{
		{"--color always", ColorAlways, false, false, true},
		{"--color always beats NO_COLOR", ColorAlways, true, false, true},
		{"--color never", ColorNever, false, true, false},
		{"--color never beats a terminal", ColorNever, true, true, false},
		{"NO_COLOR with no --color at all", "", true, true, false},
		{"--color auto on a terminal", ColorAuto, false, true, true},
		{"--color auto redirected", ColorAuto, false, false, false},
		{"--color auto beats NO_COLOR, because writing the flag is the choice",
			ColorAuto, true, true, true},
		{"nothing said, on a terminal", "", false, true, true},
		{"nothing said, redirected", "", false, false, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			if got := UseColor(c.when, c.noColor, c.isTerminal); got != c.want {
				t.Errorf("UseColor(%q, noColor=%v, terminal=%v) = %v, want %v",
					c.when, c.noColor, c.isTerminal, got, c.want)
			}
		})
	}
}

// TestTheOutputDoesNotDependOnTheTerminal is the promise of that same page:
// the output of a command is identical byte for byte with a terminal and
// without one, save the color codes, and nothing biso prints today carries
// any.
func TestTheOutputDoesNotDependOnTheTerminal(t *testing.T) {
	m := newMachine(t).withIDs("3f9a2b1c")
	m.run("init", "My project").assertCode(t, 0)

	withTerminal := m.runWithTerminal(true, "where")
	without := m.runWithTerminal(false, "where")
	assertEqual(t, withTerminal.stdout, without.stdout,
		"the output of biso where with and without a terminal")
}
