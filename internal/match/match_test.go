package match

import (
	"errors"
	"testing"

	"biso/internal/model"
)

// defaultStatuses is the vocabulary of a board created with the default
// statuses, the one every example of docs/spec/vocabularios.md uses.
var defaultStatuses = []string{"To Do", "In Progress", "Done"}

func asBisoError(t *testing.T, err error) *model.Error {
	t.Helper()
	var e *model.Error
	if !errors.As(err, &e) {
		t.Fatalf("error is %T, want *model.Error: %v", err, err)
	}
	return e
}

// TestMatchTable walks the table of
// docs/spec/vocabularios.md#el-algoritmo-de-coincidencia row by row.
func TestMatchTable(t *testing.T) {
	cases := []struct {
		name string
		in   string
		// want is the configured value the row resolves to, and wantCode the
		// error code of the row that fails. Exactly one of the two is set: a
		// row that fails says which error it is, not just that there was one.
		want     string
		wantCode string
	}{
		{name: "exact match wins by step a", in: "To Do", want: "To Do"},
		{name: "lowercase", in: "todo", want: "To Do"},
		{name: "uppercase", in: "TODO", want: "To Do"},
		{name: "hyphen is removed", in: "To-Do", want: "To Do"},
		{name: "underscore is removed", in: "TO_DO", want: "To Do"},
		{name: "every space is removed", in: "to  do", want: "To Do"},
		{name: "an embedded tab is removed", in: "To\tDo", want: "To Do"},
		{name: "a trailing dot is not removed", in: "To Do.", wantCode: "unknown_status"},
		{name: "a dot in the middle is not removed", in: "To.Do", wantCode: "unknown_status"},
		{name: "another status, spelled loosely", in: "in progress", want: "In Progress"},
		{name: "another status, with a hyphen", in: "In-Progress", want: "In Progress"},
		{name: "the terminal status", in: "DONE", want: "Done"},
		{name: "a status the board does not have", in: "Pending", wantCode: "unknown_status"},
		{name: "the empty value", in: "", wantCode: "unknown_status"},
		{name: "only separators", in: "-_", wantCode: "unknown_status"},
		{name: "no prefix matching", in: "To", wantCode: "unknown_status"},
		{name: "no substring matching", in: "oD", wantCode: "unknown_status"},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Match(Status, c.in, defaultStatuses)
			if c.wantCode != "" {
				if err == nil {
					t.Fatalf("Match(Status, %q) = %q, want the error %s", c.in, got, c.wantCode)
				}
				e := asBisoError(t, err)
				if e.Code != c.wantCode {
					t.Fatalf("Match(Status, %q) failed with %q, want %q", c.in, e.Code, c.wantCode)
				}
				if e.ExitCode != 3 {
					t.Fatalf("Match(Status, %q) has exit code %d, want 3", c.in, e.ExitCode)
				}
				if got != "" {
					t.Fatalf("Match(Status, %q) = %q with an error, want the empty string", c.in, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("Match(Status, %q) failed: %v", c.in, err)
			}
			if got != c.want {
				t.Fatalf("Match(Status, %q) = %q, want %q", c.in, got, c.want)
			}
		})
	}
}

// TestMatchFoldsCaseAndNotOnlyLowercases is the Greek board of
// docs/spec/vocabularios.md#el-algoritmo-de-coincidencia: a status that ends
// in a final sigma, typed in uppercase. Lowercasing leaves the two spellings
// of sigma apart and the board would reject its own status.
func TestMatchFoldsCaseAndNotOnlyLowercases(t *testing.T) {
	configured := []string{"Δοκιμές", "Σε εξέλιξη", "Έτοιμο"}
	for _, in := range []string{"ΔΟΚΙΜΕΣ", "δοκιμες", "Δοκιμές"} {
		got, err := Match(Status, in, configured)
		if err != nil {
			t.Fatalf("Match(Status, %q) failed: %v", in, err)
		}
		if got != "Δοκιμές" {
			t.Fatalf("Match(Status, %q) = %q, want Δοκιμές", in, got)
		}
	}
}

// TestMatchOnAConfiguredValueListedTwice is the board that repeats a value:
// two copies of the same spelling are one value, so they resolve instead of
// colliding.
func TestMatchOnAConfiguredValueListedTwice(t *testing.T) {
	configured := []string{"To Do", "In Progress", "To Do", "Done"}
	got, err := Match(Status, "todo", configured)
	if err != nil {
		t.Fatalf("Match(Status, %q) failed: %v", "todo", err)
	}
	if got != "To Do" {
		t.Fatalf("Match(Status, %q) = %q, want To Do", "todo", got)
	}
}

// TestMatchQuotesWhatItWasGiven pins the boundary of this package: a value of
// nothing but spaces is the empty value, and turning it into one belongs to
// the layer that reads it (docs/spec/valores-de-entrada.md#el-valor-vacío).
// Match never receives one, and if it did it would report it as it arrived
// rather than pretend it had been converted.
func TestMatchQuotesWhatItWasGiven(t *testing.T) {
	_, err := Match(Status, "   ", defaultStatuses)
	e := asBisoError(t, err)
	if e.Code != "unknown_status" {
		t.Errorf("Code = %q, want unknown_status", e.Code)
	}
	if e.Message != `unknown status: "   "` {
		t.Errorf("Message = %q, want it to quote the value as received", e.Message)
	}
	if e.Given != "   " {
		t.Errorf("Given = %q, want the value as received", e.Given)
	}
}

// TestMatchUnknownError checks the shape of the error of step d, the one
// docs/spec/contrato-json.md#los-errores-en-json shows field by field.
func TestMatchUnknownError(t *testing.T) {
	_, err := Match(Status, "Pending", defaultStatuses)
	e := asBisoError(t, err)

	if e.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", e.ExitCode)
	}
	if e.Code != "unknown_status" {
		t.Errorf("Code = %q, want unknown_status", e.Code)
	}
	if e.Message != `unknown status: "Pending"` {
		t.Errorf("Message = %q, want %q", e.Message, `unknown status: "Pending"`)
	}
	if len(e.Hints) != 0 {
		t.Errorf("Hints = %q, want none", e.Hints)
	}
	if e.Field != "status" {
		t.Errorf("Field = %q, want status", e.Field)
	}
	if e.Given != "Pending" {
		t.Errorf("Given = %q, want Pending", e.Given)
	}
	if got, want := e.Valid, defaultStatuses; len(got) != len(want) {
		t.Fatalf("Valid = %q, want %q", got, want)
	}
	for i := range e.Valid {
		if e.Valid[i] != defaultStatuses[i] {
			t.Fatalf("Valid = %q, want %q in the configured order", e.Valid, defaultStatuses)
		}
	}
}

// TestMatchUnknownMessagePerField checks that the three closed vocabularies
// name themselves in their own message and their own code.
func TestMatchUnknownMessagePerField(t *testing.T) {
	cases := []struct {
		field       Field
		configured  []string
		in          string
		wantCode    string
		wantMessage string
	}{
		{Status, defaultStatuses, "Pending", "unknown_status", `unknown status: "Pending"`},
		{Type, []string{"feature", "bug", "chore"}, "epic", "unknown_type", `unknown type: "epic"`},
		{Priority, []string{"low", "medium", "high"}, "", "unknown_priority", `unknown priority: ""`},
	}

	for _, c := range cases {
		t.Run(string(c.field), func(t *testing.T) {
			_, err := Match(c.field, c.in, c.configured)
			e := asBisoError(t, err)
			if e.Code != c.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, c.wantCode)
			}
			if e.Message != c.wantMessage {
				t.Errorf("Message = %q, want %q", e.Message, c.wantMessage)
			}
			if e.Field != string(c.field) {
				t.Errorf("Field = %q, want %q", e.Field, c.field)
			}
		})
	}
}

// TestMatchDoesNotAliasTheConfiguredSlice guards the purity of the package:
// the caller's slice must survive whatever the caller does to the error.
func TestMatchDoesNotAliasTheConfiguredSlice(t *testing.T) {
	configured := []string{"To Do", "In Progress", "Done"}
	_, err := Match(Status, "Pending", configured)
	e := asBisoError(t, err)
	e.Valid[0] = "mutated"
	if configured[0] != "To Do" {
		t.Fatalf("the error aliased the caller's slice: %q", configured)
	}
}

// TestMatchAmbiguous covers step e: two configured values that normalize the
// same, so no normalized comparison can pick one of them.
func TestMatchAmbiguous(t *testing.T) {
	configured := []string{"To Do", "In Progress", "To-Do", "Done"}

	_, err := Match(Status, "todo", configured)
	e := asBisoError(t, err)

	if e.ExitCode != 3 {
		t.Errorf("ExitCode = %d, want 3", e.ExitCode)
	}
	if e.Code != "ambiguous_vocabulary" {
		t.Errorf("Code = %q, want ambiguous_vocabulary", e.Code)
	}
	want := `ambiguous status: "todo" matches 2 configured values: To Do, To-Do`
	if e.Message != want {
		t.Errorf("Message = %q, want %q", e.Message, want)
	}
	wantHint := "type one of them exactly, or rename one so the two no longer normalize the same"
	if len(e.Hints) != 1 || e.Hints[0] != wantHint {
		t.Errorf("Hints = %q, want exactly [%q]", e.Hints, wantHint)
	}
	if e.Field != "status" || e.Given != "todo" {
		t.Errorf("Field = %q and Given = %q, want status and todo", e.Field, e.Given)
	}
	if len(e.Valid) != 2 || e.Valid[0] != "To Do" || e.Valid[1] != "To-Do" {
		t.Errorf("Valid = %q, want only the two that collide, in configured order", e.Valid)
	}
}

// TestMatchAmbiguousCountsMoreThanTwo checks that the message says how many
// collide when they are more than two, without pretending there are only two.
func TestMatchAmbiguousCountsMoreThanTwo(t *testing.T) {
	configured := []string{"To Do", "TO_DO", "to-do", "Done"}
	_, err := Match(Status, "todo", configured)
	e := asBisoError(t, err)
	want := `ambiguous status: "todo" matches 3 configured values: To Do, TO_DO, to-do`
	if e.Message != want {
		t.Errorf("Message = %q, want %q", e.Message, want)
	}
}

// TestMatchExactBeatsAmbiguity is step a winning over step e: even on a board
// whose two statuses normalize the same, typing one of them exactly resolves.
func TestMatchExactBeatsAmbiguity(t *testing.T) {
	configured := []string{"To Do", "To-Do", "Done"}
	for _, in := range []string{"To Do", "To-Do"} {
		got, err := Match(Status, in, configured)
		if err != nil {
			t.Fatalf("Match(Status, %q) failed: %v", in, err)
		}
		if got != in {
			t.Fatalf("Match(Status, %q) = %q, want the value itself", in, got)
		}
	}
}

// TestMatchOnAnEmptyVocabulary is the degenerate board: nothing configured,
// so nothing can match and the error carries no valid values.
func TestMatchOnAnEmptyVocabulary(t *testing.T) {
	_, err := Match(Type, "bug", nil)
	e := asBisoError(t, err)
	if e.Code != "unknown_type" {
		t.Errorf("Code = %q, want unknown_type", e.Code)
	}
	if len(e.Valid) != 0 {
		t.Errorf("Valid = %q, want it empty", e.Valid)
	}
}

// outcome is one cell of the contract table of
// docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos,
// transcribed from the document: either the configured value that cell
// resolves to, or the error code it fails with.
type outcome struct {
	resolves string
	code     string
	message  string
}

// TestTheSameTextIsWorthTheSameInBothDirections walks the contract table of
// docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos
// row by row. Each row transcribes its two columns separately, the one for
// writing (biso set -s <v>) and the one for filtering (biso ls -s <v>), and
// the test checks the implementation against each column on its own. The two
// columns are data copied from the document, not one computed from the other,
// so a row whose two halves stopped agreeing, or that stopped agreeing with
// the document, fails here.
//
// What this test cannot yet prove is the other half of the promise: that the
// command that writes and the command that filters both reach this function
// instead of validating on their own. Today there is only one implementation
// and no command at all, so the equivalence is true by construction; the day
// biso set and biso ls exist it stops being so, and a test that calls the two
// commands with the same text has to assert it end to end. That belongs to
// TASK-13, which is where those two commands are built.
func TestTheSameTextIsWorthTheSameInBothDirections(t *testing.T) {
	const unknownPending = `unknown status: "Pending"`
	const unknownEmpty = `unknown status: ""`

	cases := []struct {
		in     string
		write  outcome
		filter outcome
	}{
		{in: "To Do", write: outcome{resolves: "To Do"}, filter: outcome{resolves: "To Do"}},
		{in: "todo", write: outcome{resolves: "To Do"}, filter: outcome{resolves: "To Do"}},
		{in: "TO_DO", write: outcome{resolves: "To Do"}, filter: outcome{resolves: "To Do"}},
		{in: "In-Progress", write: outcome{resolves: "In Progress"}, filter: outcome{resolves: "In Progress"}},
		{
			in:     "Pending",
			write:  outcome{code: "unknown_status", message: unknownPending},
			filter: outcome{code: "unknown_status", message: unknownPending},
		},
		{
			in:     "",
			write:  outcome{code: "unknown_status", message: unknownEmpty},
			filter: outcome{code: "unknown_status", message: unknownEmpty},
		},
	}

	check := func(t *testing.T, direction, in string, want outcome) {
		t.Helper()
		got, err := Match(Status, in, defaultStatuses)
		if want.code == "" {
			if err != nil {
				t.Fatalf("%s: Match(Status, %q) failed: %v", direction, in, err)
			}
			if got != want.resolves {
				t.Fatalf("%s: Match(Status, %q) = %q, want %q", direction, in, got, want.resolves)
			}
			return
		}
		if err == nil {
			t.Fatalf("%s: Match(Status, %q) = %q, want the error %s", direction, in, got, want.code)
		}
		e := asBisoError(t, err)
		if e.ExitCode != 3 {
			t.Errorf("%s: exit code %d, want 3", direction, e.ExitCode)
		}
		if e.Code != want.code {
			t.Errorf("%s: Code = %q, want %q", direction, e.Code, want.code)
		}
		if e.Message != want.message {
			t.Errorf("%s: Message = %q, want %q", direction, e.Message, want.message)
		}
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			check(t, "writing", c.in, c.write)
			check(t, "filtering", c.in, c.filter)
		})
	}
}
