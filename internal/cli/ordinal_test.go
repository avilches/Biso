package cli

import (
	"strings"
	"testing"
)

// This file is the command line of the manual order: what the parser
// refuses before the board is even opened, and what the two readings print
// (docs/spec/familias-de-flags.md#el-orden-manual).

// TestOrdinalTakesOnlyFirstAndLast is the closed domain of that flag. It is
// the program's own domain and not the board's, so an unknown value is exit
// code 2 with invalid_ordinal_value and never the 3 of a value the board
// does not have.
func TestOrdinalTakesOnlyFirstAndLast(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)

	for _, value := range []string{"3", "3000", "m8", "First", "top"} {
		out := m.run("set", "MYP-1", "--ordinal", value).assertCode(t, 2)
		want := `error: --ordinal: unknown value: "` + value + `"`
		if !strings.Contains(out.stderr, want) {
			t.Errorf("--ordinal %s said:\n%s", value, out.stderr)
		}
		if !strings.Contains(out.stderr,
			"hint: --ordinal takes first or last; "+
				"to place a task next to another one, use --above or --below") {
			t.Errorf("--ordinal %s printed no hint:\n%s", value, out.stderr)
		}
	}

	// The empty string is the exception of the exception of
	// docs/spec/valores-de-entrada.md#el-valor-vacío: it is not the
	// empty_scalar_value of every other scalar, because this domain is the
	// program's.
	out := m.run("set", "MYP-1", "--ordinal", "", "--json").assertCode(t, 2)
	if !strings.Contains(out.stderr, `"code": "invalid_ordinal_value"`) {
		t.Errorf(`--ordinal "" answered:\n%s`, out.stderr)
	}
	if !strings.Contains(out.stderr, `"valid"`) {
		t.Errorf("the envelope carries no valid list:\n%s", out.stderr)
	}

	for _, value := range []string{"first", "last"} {
		m.run("set", "MYP-1", "--ordinal", value).assertCode(t, 0)
	}
}

// TestTheFlagsOfTheManualOrderAreIncompatibleWithOneAnother is the sentence
// of that section saying all four write the same field, so no two of them
// can share a call. The message is the ordinary one of a pair, and it does
// not depend on the order they were typed in.
func TestTheFlagsOfTheManualOrderAreIncompatibleWithOneAnother(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)
	m.run("new", "Another task").assertCode(t, 0)

	for _, c := range []struct {
		one, other []string
		message    string
	}{
		{[]string{"--above", "MYP-2"}, []string{"--below", "MYP-2"},
			"--above and --below cannot be used together"},
		{[]string{"--ordinal", "last"}, []string{"--above", "MYP-2"},
			"--ordinal and --above cannot be used together"},
		{[]string{"--ordinal", "last"}, []string{"--below", "MYP-2"},
			"--ordinal and --below cannot be used together"},
		{[]string{"--clear-ordinal"}, []string{"--ordinal", "last"},
			"--clear-ordinal and --ordinal cannot be used together"},
		{[]string{"--clear-ordinal"}, []string{"--above", "MYP-2"},
			"--clear-ordinal and --above cannot be used together"},
		{[]string{"--clear-ordinal"}, []string{"--below", "MYP-2"},
			"--clear-ordinal and --below cannot be used together"},
	} {
		argv := append([]string{"set", "MYP-1"}, append(c.one, c.other...)...)
		out := m.run(argv...).assertCode(t, 2)
		if !strings.Contains(out.stderr, "error: "+c.message) {
			t.Errorf("%v said:\n%s", argv, out.stderr)
		}
		// The other way round says exactly the same thing, because a
		// message names two flags in the order of the table and never in
		// the order of the call.
		reversed := append([]string{"set", "MYP-1"}, append(c.other, c.one...)...)
		other := m.run(reversed...).assertCode(t, 2)
		assertEqual(t, other.stderr, out.stderr,
			"the message of "+c.message+", written the other way round")
	}

	// And the pair of a scalar with its own --clear-<field> is still not a
	// conflict anywhere else, which is what keeps this rule about the
	// manual order and not about scalars in general.
	m.run("set", "MYP-1", "--clear-due", "--due", "2026-09-20").assertCode(t, 0)
}

// TestTheCardPrintsManualAndNeverTheKey is the paragraph of
// docs/spec/cmd/get.md#salida: the key cannot be typed and says nothing a
// reader can use, so the card says whether there is one and --json carries
// the key itself.
func TestTheCardPrintsManualAndNeverTheKey(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "A task").assertCode(t, 0)

	before := m.run("get", "MYP-1").assertCode(t, 0)
	if !strings.Contains(before.stdout, "ordinal    -") {
		t.Errorf("a task with no key printed:\n%s", before.stdout)
	}

	m.run("set", "MYP-1", "--ordinal", "first").assertCode(t, 0)
	after := m.run("get", "MYP-1").assertCode(t, 0)
	if !strings.Contains(after.stdout, "ordinal    manual") {
		t.Errorf("a placed task printed:\n%s", after.stdout)
	}
	if strings.Contains(after.stdout, "ordinal    i") {
		t.Error("the card printed the raw key")
	}

	// The envelope of the card carries the key itself, as a string.
	envelope := m.run("get", "MYP-1", "--json").assertCode(t, 0)
	if !strings.Contains(envelope.stdout, `"ordinal": "i"`) {
		t.Errorf("the envelope of biso get does not carry the key:\n%s", envelope.stdout)
	}
}

// TestTheEnvelopesCarryTheKeyAsAStringOrNull is the rule of
// docs/spec/contrato-json.md#números-fechas-y-ausencias for this one field:
// a string or null, never a number, in every envelope that carries a task.
func TestTheEnvelopesCarryTheKeyAsAStringOrNull(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "Placed").assertCode(t, 0)
	m.run("new", "Not placed").assertCode(t, 0)
	m.run("set", "MYP-1", "--ordinal", "last").assertCode(t, 0)

	listing := m.run("ls", "--json").assertCode(t, 0)
	for _, want := range []string{`"ordinal": "i"`, `"ordinal": null`} {
		if !strings.Contains(listing.stdout, want) {
			t.Errorf("the listing envelope has no %s:\n%s", want, listing.stdout)
		}
	}

	exported := m.run("export").assertCode(t, 0)
	for _, want := range []string{`"ordinal":"i"`, `"ordinal":null`} {
		if !strings.Contains(exported.stdout, want) {
			t.Errorf("the export has no %s:\n%s", want, exported.stdout)
		}
	}
}

// TestTheFlagsOfTheManualOrderWorkInEveryWritingCommand is the promise
// `biso set --help` makes about every field flag, asked of the four that
// place a task: `biso new` creates a task already placed, and a verb of the
// cycle moves one while it does its own work.
func TestTheFlagsOfTheManualOrderWorkInEveryWritingCommand(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "The anchor", "--ordinal", "last").assertCode(t, 0)
	m.run("new", "Born below the anchor", "--below", "MYP-1").assertCode(t, 0)
	m.run("new", "Born above the anchor", "--above", "MYP-1").assertCode(t, 0)
	m.run("start", "MYP-2", "--ordinal", "first").assertCode(t, 0)
	m.run("note", "MYP-3", "A note", "--ordinal", "last").assertCode(t, 0)

	got := m.run("ls", "--sort", "ordinal", "--ids", "--any-status").assertCode(t, 0)
	assertEqual(t, got.stdout, "MYP-2\nMYP-1\nMYP-3\n",
		"the manual order after placing from new, start and note")
}

// TestANeighbourWithoutAKeyIsSixOnTheCommandLine is the exit code the shell
// sees for the one refusal of this family that is about the board and not
// about the call, with the whole remedy printed under it.
func TestANeighbourWithoutAKeyIsSixOnTheCommandLine(t *testing.T) {
	m := newMachine(t).board()
	m.run("new", "With no place").assertCode(t, 0)
	m.run("new", "The one being moved").assertCode(t, 0)

	out := m.run("set", "MYP-2", "--above", "MYP-1").assertCode(t, 6)
	for _, want := range []string{
		"error: --above: MYP-1 has no ordinal",
		"hint: a task without one has no place in the manual order, " +
			"so there is nothing to write above",
		"hint: `biso set MYP-1 --ordinal last` gives it one, " +
			"and then --above MYP-1 works",
	} {
		if !strings.Contains(out.stderr, want) {
			t.Errorf("the refusal does not say %q:\n%s", want, out.stderr)
		}
	}

	// The remedy the hint gives works, which is why it is written out.
	m.run("set", "MYP-1", "--ordinal", "last").assertCode(t, 0)
	m.run("set", "MYP-2", "--above", "MYP-1").assertCode(t, 0)
}
