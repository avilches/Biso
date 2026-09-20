package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// This file is the precedence of the row limit of
// docs/spec/invocacion.md#variables-de-entorno, seen from outside the
// program: --limit over BISO_LIMIT, BISO_LIMIT over the `default_limit` key
// of the machine's configuration, and thirty when none of the three says
// anything.

// boardOfThirtyFive is a board with more tasks than the built-in default, so
// that every rung of the precedence cuts at a different place and the four
// answers are told apart by counting rows.
func boardOfThirtyFive(t *testing.T) *machine {
	t.Helper()
	m := newMachine(t)
	dir := filepath.Join(m.home, ".biso", "boards", "my-project-3f9a2b1c")
	m.buildBoard(t, dir, "3f9a2b1c", "My project", "MYP", 35, 0, 35)
	m.write(t, filepath.Join(m.dir, ".biso.json"),
		"{ \"version\": 1, \"id\": \"3f9a2b1c\" }\n")
	return m
}

// machineConfig writes ~/.biso/config.json with one key in it.
func (m *machine) machineConfig(t *testing.T, content string) {
	t.Helper()
	m.write(t, filepath.Join(m.home, ".biso", "config.json"), content+"\n")
}

func rows(out string) int {
	if out == "" {
		return 0
	}
	return len(strings.Split(strings.TrimRight(out, "\n"), "\n"))
}

func TestTheRowLimitReadsTheFourRungsOfItsPrecedence(t *testing.T) {
	for _, c := range []struct {
		name   string
		config string
		env    string
		argv   []string
		want   int
	}{
		{"the flag wins over everything",
			`{ "default_limit": 3 }`, "2", []string{"ls", "--limit", "1"}, 1},
		{"the variable wins over the key",
			`{ "default_limit": 3 }`, "2", []string{"ls"}, 2},
		{"the key answers when the variable is not set",
			`{ "default_limit": 3 }`, "", []string{"ls"}, 3},
		{"thirty when nothing says otherwise",
			"", "", []string{"ls"}, 30},
	} {
		t.Run(c.name, func(t *testing.T) {
			m := boardOfThirtyFive(t)
			if c.config != "" {
				m.machineConfig(t, c.config)
			}
			if c.env != "" {
				m.env["BISO_LIMIT"] = c.env
			}

			got := m.run(t, c.argv...).assertCode(t, 0)

			if n := rows(got.stdout); n != c.want {
				t.Errorf("the listing has %d rows and not %d:\n%s", n, c.want, got.stdout)
			}
			if !strings.Contains(got.stderr, "more tasks match; showing") {
				t.Errorf("the warning of the cut is not there:\n%s", got.stderr)
			}
		})
	}
}

// TestAllLiftsEveryRungOfTheLimit is the other end of the same table:
// --all is the one that asks for everything, whatever the variable or the
// key say (docs/spec/cmd/ls.md).
func TestAllLiftsEveryRungOfTheLimit(t *testing.T) {
	m := boardOfThirtyFive(t)
	m.machineConfig(t, `{ "default_limit": 3 }`)
	m.env["BISO_LIMIT"] = "2"

	got := m.run(t, "ls", "--all").assertCode(t, 0)

	if n := rows(got.stdout); n != 35 {
		t.Errorf("the listing has %d rows and not the thirty five of the board", n)
	}
}

// TestAVariableThatIsNotAWholeNumberOfRowsIsRefused is the same judgement
// the `default_limit` key already gets: a value outside its domain is an
// error and never something read half way
// (docs/spec/invocacion.md#variables-de-entorno).
func TestAVariableThatIsNotAWholeNumberOfRowsIsRefused(t *testing.T) {
	for _, value := range []string{"lots", "-1", "3.5"} {
		m := boardOfThirtyFive(t)
		m.env["BISO_LIMIT"] = value

		got := m.run(t, "ls").assertCode(t, 2)

		want := "error: BISO_LIMIT: not a whole number of rows: \"" + value + "\"\n" +
			"hint: a limit is zero or more\n"
		assertEqual(t, got.stderr, want, "the refusal of BISO_LIMIT="+value)
	}
}

// TestTheLimitOfPrimeIsItsOwn is the boundary of the two rungs above:
// BISO_LIMIT and `default_limit` are spelled out as the default of
// `biso ls`, and `biso prime` keeps its own five
// (docs/spec/cmd/prime.md).
func TestTheLimitOfPrimeIsItsOwn(t *testing.T) {
	m := boardOfThirtyFive(t)
	m.machineConfig(t, `{ "default_limit": 3 }`)
	m.env["BISO_LIMIT"] = "2"

	got := m.run(t, "prime").assertCode(t, 0)

	if !strings.Contains(got.stdout, "30 more not shown") {
		t.Errorf("prime shared its five rows differently:\n%s", got.stdout)
	}
}

// TestATextPositionalAndAFlagCannotBothReadStdin is the rule of one
// standard input per invocation reaching the positional argument that a
// verb reads as long text, and reaching it before anything is read: the
// refusal is the whole of what the call writes, with no warning in front
// of it about a value that was never going to be stored
// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo).
func TestATextPositionalAndAFlagCannotBothReadStdin(t *testing.T) {
	m := boardOfThirtyFive(t)

	got := runWithStdin(t, m, "", "note", "MYP-1", "-", "--append-plan", "-")

	if got.code != 2 {
		t.Fatalf("the call answered %d and not 2:\n%s", got.code, got.stderr)
	}
	assertEqual(t, got.stderr,
		"error: - can be given only once per invocation; "+
			"text and --append-plan both read stdin\n",
		"the refusal of two arguments reading standard input")
}
