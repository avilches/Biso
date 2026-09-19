package model

import (
	"errors"
	"testing"
	"time"
)

// validTask is a task that Validate accepts, so that every case below
// changes exactly one thing and nothing else can explain the failure.
func validTask() *Task {
	task := &Task{
		Title:     "Normalize CRLF in the diff",
		Status:    "In Progress",
		Author:    "@avilches",
		Assignees: []string{"@claude"},
		Labels:    []string{"parser"},
		Ext:       map[string]string{"trello.card": "5f2a8c1e"},
	}
	task.AddCriterion("The exporter writes CRLF untouched")
	task.AddComment("@avilches", time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC), "A comment.")
	return task
}

func extensions() []string { return []string{"trello.card", "github.issue"} }

// assertSpecError checks that err is the *Error the specification fixes for
// a case, by exit code, code and field.
func assertSpecError(t *testing.T, err error, exitCode int, code, field string) *Error {
	t.Helper()

	var specErr *Error
	if !errors.As(err, &specErr) {
		t.Fatalf("error = %v, want a *Error of the specification", err)
	}
	if specErr.ExitCode != exitCode || specErr.Code != code {
		t.Fatalf("error = exit %d, code %q; want exit %d, code %q", specErr.ExitCode, specErr.Code, exitCode, code)
	}
	if specErr.Field != field {
		t.Fatalf("error field = %q, want %q", specErr.Field, field)
	}
	return specErr
}

func TestValidateAcceptsATaskWithEveryFieldFilled(t *testing.T) {
	if err := validTask().Validate(extensions()); err != nil {
		t.Fatalf("Validate on a well formed task: %v", err)
	}
}

// TestValidateRejectsAnEmptyTitle is the last row of the table of
// docs/spec/valores-de-entrada.md#el-valor-vacío: the title is the one
// field the model of docs/spec/modelo-de-datos/index.md marks as
// mandatory, and an empty one is exit code 2 with a literal message.
func TestValidateRejectsAnEmptyTitle(t *testing.T) {
	for _, title := range []string{"", "   ", "\t"} {
		task := validTask()
		task.Title = title

		err := task.Validate(extensions())
		specErr := assertSpecError(t, err, 2, "missing_title", "")
		if specErr.Message != "title cannot be empty" {
			t.Fatalf("message = %q, not the literal text of the specification", specErr.Message)
		}
	}
}

// TestValidateRejectsANewlineInEveryStringField covers
// docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string over
// the four fields it names, which is the distinction between a `string`
// and a `text` of docs/spec/modelo-de-datos/index.md made enforceable.
func TestValidateRejectsANewlineInEveryStringField(t *testing.T) {
	cases := []struct {
		name  string
		field string
		spoil func(*Task)
	}{
		{
			name:  "the title",
			field: StringFieldTitle,
			spoil: func(task *Task) { task.Title = "first line\nsecond line" },
		},
		{
			name:  "a carriage return in the title",
			field: StringFieldTitle,
			spoil: func(task *Task) { task.Title = "first line\rsecond line" },
		},
		{
			name:  "the author of the task",
			field: StringFieldAuthor,
			spoil: func(task *Task) { task.Author = "@avilches\n@claude" },
		},
		{
			name:  "the author of a comment",
			field: StringFieldAuthor,
			spoil: func(task *Task) { task.Comments[0].Author = "@avilches\n@claude" },
		},
		{
			name:  "the author of the open question",
			field: StringFieldAuthor,
			spoil: func(task *Task) {
				task.Question = &Question{Author: "@claude\n@sara", Body: "Which one?"}
			},
		},
		{
			name:  "the text of a criterion",
			field: StringFieldCriterionText,
			spoil: func(task *Task) { task.AcceptanceCriteria[0].Text = "first line\nsecond line" },
		},
		{
			name:  "the value of an extension field",
			field: StringFieldExt,
			spoil: func(task *Task) { task.Ext["trello.card"] = "5f2a8c1e\n5f2a8c1f" },
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			task := validTask()
			c.spoil(task)

			err := task.Validate(extensions())
			specErr := assertSpecError(t, err, 2, "malformed_string_value", c.field)
			wantHint := "a string field cannot contain a newline or a carriage return"
			if len(specErr.Hints) != 1 || specErr.Hints[0] != wantHint {
				t.Fatalf("hints = %q, not the literal text of the specification", specErr.Hints)
			}
		})
	}
}

// TestValidateExampleMessageOfTheSpecification pins the one message of
// that section the specification writes out in full.
func TestValidateExampleMessageOfTheSpecification(t *testing.T) {
	task := validTask()
	task.Title = "first line\nsecond line"

	err := task.Validate(extensions())
	specErr := assertSpecError(t, err, 2, "malformed_string_value", StringFieldTitle)
	want := `malformed title: "first line\nsecond line"`
	if specErr.Message != want {
		t.Fatalf("message = %q, want %q", specErr.Message, want)
	}
}

// TestValidateAcceptsANewlineInEveryTextField is the other half of the
// same rule: the fields typed `text` take line breaks with no restriction,
// so a check that rejected them everywhere would be just as wrong.
func TestValidateAcceptsANewlineInEveryTextField(t *testing.T) {
	task := validTask()
	task.Description = "first line\nsecond line"
	task.Plan = "1. one\n2. two"
	task.Notes = "a note\nand another"
	task.Summary = "what happened\nin the end"
	task.Comments[0].Body = "a comment\nwith two lines"
	task.Question = &Question{Author: "@claude", Body: "a question\nwith two lines"}

	if err := task.Validate(extensions()); err != nil {
		t.Fatalf("Validate rejected a line break in a text field: %v", err)
	}
}

// TestValidateRejectsALabelAndAnAssigneeOutsideTheirAlphabet is
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token
// applied to the two fields of its table that are not an extension key.
func TestValidateRejectsALabelAndAnAssigneeOutsideTheirAlphabet(t *testing.T) {
	withLabel := validTask()
	withLabel.Labels = []string{"parser", "urgent!"}
	specErr := assertSpecError(t, withLabel.Validate(extensions()), 2, "malformed_label", "labels")
	if specErr.Message != `malformed label: "urgent!"` {
		t.Fatalf("message = %q, not the literal text of the specification", specErr.Message)
	}

	withAssignee := validTask()
	withAssignee.Assignees = []string{"sara smith"}
	specErr = assertSpecError(t, withAssignee.Validate(extensions()), 2, "malformed_assignee", "assignees")
	if specErr.Message != `malformed assignee: "sara smith"` {
		t.Fatalf("message = %q, not the literal text of the specification", specErr.Message)
	}
}

// TestValidateRejectsAnUndeclaredOrMalformedExtensionKey keeps the two
// extension-key errors covered now that they are reached through Validate.
func TestValidateRejectsAnUndeclaredOrMalformedExtensionKey(t *testing.T) {
	undeclared := validTask()
	undeclared.Ext = map[string]string{"jira.key": "PROJ-1"}
	assertSpecError(t, undeclared.Validate(extensions()), 3, "unknown_extension_key", "ext")

	malformed := validTask()
	malformed.Ext = map[string]string{"trello card": "5f2a8c1e"}
	assertSpecError(t, malformed.Validate(extensions()), 2, "malformed_extension_key", "ext")
}

// TestValidateRejectsANegativeOrdinal is the `int (>= 0)` of the field
// table of docs/spec/modelo-de-datos/index.md. Zero is a value and not an
// absence, which is why Ordinal is a pointer, so it has to pass.
func TestValidateRejectsANegativeOrdinal(t *testing.T) {
	negative := -1
	task := validTask()
	task.Ordinal = &negative
	specErr := assertSpecError(t, task.Validate(extensions()), 2, "invalid_number", "ordinal")
	if specErr.Message != "ordinal cannot be negative: -1" {
		t.Fatalf("message = %q", specErr.Message)
	}

	zero := 0
	task.Ordinal = &zero
	if err := task.Validate(extensions()); err != nil {
		t.Fatalf("Validate rejected ordinal 0, which the field table admits: %v", err)
	}
}

// TestValidateRejectsACriterionKeyThatBreaksItsContract covers the two
// halves of the key rule of docs/spec/modelo-de-datos/criterios.md: a key
// is a positive integer, and no two criteria of a task share one. The
// program assigns them, so neither is an error of the specification with a
// code of its own; what matters is that it is the program that says so and
// not a constraint of SQLite naming a table.
func TestValidateRejectsACriterionKeyThatBreaksItsContract(t *testing.T) {
	zeroKey := validTask()
	zeroKey.AcceptanceCriteria = []Criterion{{Key: 0, Text: "No key at all"}}
	err := zeroKey.Validate(extensions())
	if err == nil {
		t.Fatalf("Validate accepted a criterion with key 0")
	}
	var specErr *Error
	if errors.As(err, &specErr) {
		t.Fatalf("a criterion key is a program invariant, not a case of the specification: %+v", specErr)
	}

	duplicate := validTask()
	duplicate.AcceptanceCriteria = []Criterion{
		{Key: 1, Text: "The first one"},
		{Key: 1, Text: "The same key again"},
	}
	if err := duplicate.Validate(extensions()); err == nil {
		t.Fatalf("Validate accepted two criteria sharing a key")
	}
}
