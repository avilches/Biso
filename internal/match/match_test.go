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
		name    string
		in      string
		want    string
		wantErr bool
	}{
		{"exact match wins by step a", "To Do", "To Do", false},
		{"lowercase", "todo", "To Do", false},
		{"uppercase", "TODO", "To Do", false},
		{"hyphen is removed", "To-Do", "To Do", false},
		{"underscore is removed", "TO_DO", "To Do", false},
		{"every space is removed", "to  do", "To Do", false},
		{"an embedded tab is removed", "To\tDo", "To Do", false},
		{"a trailing dot is not removed", "To Do.", "", true},
		{"a dot in the middle is not removed", "To.Do", "", true},
		{"another status, spelled loosely", "in progress", "In Progress", false},
		{"another status, with a hyphen", "In-Progress", "In Progress", false},
		{"the terminal status", "DONE", "Done", false},
		{"a status the board does not have", "Pending", "", true},
		{"the empty value", "", "", true},
		{"only separators", "-_", "", true},
		{"no prefix matching", "To", "", true},
		{"no substring matching", "oD", "", true},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got, err := Match(Status, c.in, defaultStatuses)
			if c.wantErr {
				if err == nil {
					t.Fatalf("Match(Status, %q) = %q, want an error", c.in, got)
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
	if e.Hint != "" {
		t.Errorf("Hint = %q, want it empty", e.Hint)
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
	if e.Hint != wantHint {
		t.Errorf("Hint = %q, want %q", e.Hint, wantHint)
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

// TestTheSameTextIsWorthTheSameInBothDirections is the contract table of
// docs/spec/vocabularios.md#el-mismo-texto-vale-lo-mismo-en-los-dos-sentidos.
// Writing (biso set -s <v>) and filtering (biso ls -s <v>) call the very same
// function, so the test asserts it twice per row and compares the two
// outcomes to each other as well as to the expected one.
func TestTheSameTextIsWorthTheSameInBothDirections(t *testing.T) {
	cases := []struct {
		in       string
		resolves string
		fails    bool
	}{
		{in: "To Do", resolves: "To Do"},
		{in: "todo", resolves: "To Do"},
		{in: "TO_DO", resolves: "To Do"},
		{in: "In-Progress", resolves: "In Progress"},
		{in: "Pending", fails: true},
		{in: "", fails: true},
	}

	for _, c := range cases {
		t.Run(c.in, func(t *testing.T) {
			writeValue, writeErr := Match(Status, c.in, defaultStatuses)
			filterValue, filterErr := Match(Status, c.in, defaultStatuses)

			if writeValue != filterValue {
				t.Fatalf("writing gave %q and filtering gave %q", writeValue, filterValue)
			}
			if (writeErr == nil) != (filterErr == nil) {
				t.Fatalf("writing gave %v and filtering gave %v", writeErr, filterErr)
			}

			if c.fails {
				if writeErr == nil {
					t.Fatalf("Match(Status, %q) succeeded, want exit code 3 in both directions", c.in)
				}
				w := asBisoError(t, writeErr)
				f := asBisoError(t, filterErr)
				if w.ExitCode != 3 || f.ExitCode != 3 {
					t.Fatalf("exit codes %d and %d, want 3 in both directions", w.ExitCode, f.ExitCode)
				}
				if w.Message != f.Message {
					t.Fatalf("messages differ: %q and %q", w.Message, f.Message)
				}
				return
			}
			if writeErr != nil {
				t.Fatalf("Match(Status, %q) failed: %v", c.in, writeErr)
			}
			if writeValue != c.resolves {
				t.Fatalf("Match(Status, %q) = %q, want %q", c.in, writeValue, c.resolves)
			}
		})
	}
}
