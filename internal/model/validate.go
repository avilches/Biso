package model

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// The token alphabet of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token: Unicode
// letters and digits, plus these symbols.
const tokenSymbols = "-_.:@"

// ValidateLabel answers the error of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token when
// a label steps outside its alphabet, and nil when it does not. It is a
// problem of form, not of the board not recognizing the value, so the exit
// code is 2 and not the 3 of an unknown vocabulary value: neither labels
// nor assignees have a closed vocabulary when written.
//
// The alphabet admits the colon, so a label that carries one goes on
// through the rule of docs/spec/valores-de-entrada.md#las-etiquetas-con-ámbito
// as well: the two refusals share the `code` because they are the same kind
// of failure, and asking both here is what makes the rule hold wherever a
// task is written, from a flag or from a line of a batch.
func ValidateLabel(label string) *Error {
	if err := validateToken(label, tokenSymbols, "label", "malformed_label", "labels",
		"a label may contain letters, digits, and - _ . : @"); err != nil {
		return err
	}
	_, err := ParseLabel(label)
	return err
}

// ValidateLabelEntry is ValidateLabel for an entry of the `labels` list of
// the configuration, which is the one place where the key form `milestone::`
// is a declaration and not a malformed label
// (docs/spec/cmd/config.md#la-lista-labels). Everything else it refuses is
// the same, with the same `code` and the same exit code.
func ValidateLabelEntry(entry string) *Error {
	if err := validateToken(entry, tokenSymbols, "label", "malformed_label", "labels",
		"a label may contain letters, digits, and - _ . : @"); err != nil {
		return err
	}
	_, err := ParseLabelKeyOrLabel(entry)
	return err
}

// ValidateAssignee is ValidateLabel for an assignee, with the same
// alphabet and the same exit code.
func ValidateAssignee(assignee string) *Error {
	return validateToken(assignee, tokenSymbols, "assignee", "malformed_assignee", "assignees",
		"an assignee may contain letters, digits, and - _ . : @")
}

// validateToken is the shared body of the two above.
func validateToken(value, symbols, noun, code, field, hint string) *Error {
	if value != "" && allowedToken(value, symbols) {
		return nil
	}
	return &Error{
		ExitCode: 2,
		Code:     code,
		Message:  fmt.Sprintf("malformed %s: %q", noun, value),
		Hints:    []string{hint},
		Field:    field,
		Given:    value,
	}
}

// allowedToken answers whether every rune of value is a Unicode letter, a
// Unicode digit, or one of the symbols the alphabet adds.
func allowedToken(value, symbols string) bool {
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			continue
		}
		if strings.ContainsRune(symbols, r) {
			continue
		}
		return false
	}
	return true
}

// The three values the `field` key takes on a malformed_string_value error,
// per docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string.
// They name the field of the JSON envelope and not the flag that wrote it:
// the same criterion text arrives through --add-ac and through an import,
// and the reason it is rejected is the same one.
const (
	StringFieldTitle         = "title"
	StringFieldAuthor        = "author"
	StringFieldCriterionText = "criterion_text"
)

// stringFieldNouns is how each of those three names itself in the message.
// The specification writes only the title one literally, `error: malformed
// title: "first line\nsecond line"`, so the other two follow its shape
// with the field's own name in prose.
var stringFieldNouns = map[string]string{
	StringFieldTitle:         "title",
	StringFieldAuthor:        "author",
	StringFieldCriterionText: "criterion text",
}

// ValidateStringField answers the error of
// docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string when
// a one-line field carries a line break, and nil when it does not.
//
// The distinction it enforces is the one the type table of
// docs/spec/modelo-de-datos/index.md draws: a `string` is text of one line
// and a `text` is a block of prose. Only the three fields above are
// `string` and reachable from a value the caller writes, so only they pass
// through here; description, plan, notes, summary, a comment's body and a
// question's body are `text` and take line breaks without any restriction.
//
// Like a token outside its alphabet, it is a problem of form and not of an
// unknown value, so the exit code is 2.
func ValidateStringField(field, value string) *Error {
	if !strings.ContainsAny(value, "\r\n") {
		return nil
	}
	noun, ok := stringFieldNouns[field]
	if !ok {
		noun = field
	}
	return &Error{
		ExitCode: 2,
		Code:     "malformed_string_value",
		Message:  fmt.Sprintf("malformed %s: %q", noun, value),
		Hints:    []string{"a string field cannot contain a newline or a carriage return"},
		Field:    field,
		Given:    value,
	}
}

// Validate answers the first thing about the task that a board must not
// store, or nil when there is nothing.
//
// It lives on the model and not on the layer that writes rows because
// every one of these rules is a rule of the specification about the task
// itself, not about SQLite: the same task validated the same way whether
// it arrives from a flag, from an import or from a test.
//
// The vocabularies a board configures for status, type and priority are
// not checked here: they are matched by internal/match against a
// configuration the model does not see.
//
// Two kinds of failure come out of it. The rules of the specification
// answer a *Error with its exit code and its code, ready to print. The
// keys of a criterion answer a plain error instead: the program assigns
// them, so a criterion with key 0 or two criteria sharing a key are a bug
// in whoever built the task, which is why the specification gives them no
// code of their own. Both matter here for the same reason: without this
// check the second one reached SQLite and came back as a raw constraint
// failure naming a table.
func (t *Task) Validate() error {
	if strings.TrimSpace(t.Title) == "" {
		return &Error{
			ExitCode: 2,
			Code:     "missing_title",
			Message:  "title cannot be empty",
		}
	}
	if err := ValidateStringField(StringFieldTitle, t.Title); err != nil {
		return err
	}
	if t.Author != "" {
		if err := ValidateStringField(StringFieldAuthor, t.Author); err != nil {
			return err
		}
	}
	for _, assignee := range t.Assignees {
		if err := ValidateAssignee(assignee); err != nil {
			return err
		}
	}
	for _, label := range t.Labels {
		if err := ValidateLabel(label); err != nil {
			return err
		}
	}
	// docs/spec/modelo-de-datos/index.md types `ordinal` as int >= 0, and a
	// negative one is a number the field does not admit, which is the
	// invalid_number of docs/spec/contrato-json.md#los-identificadores-de-error.
	if t.Ordinal != nil && *t.Ordinal < 0 {
		return &Error{
			ExitCode: 2,
			Code:     "invalid_number",
			Message:  fmt.Sprintf("ordinal cannot be negative: %d", *t.Ordinal),
			Field:    "ordinal",
			Given:    strconv.Itoa(*t.Ordinal),
		}
	}
	if err := t.validateCriteria(); err != nil {
		return err
	}
	return t.validatePeopleOfTheListsWithStructure()
}

// validateCriteria checks the part of
// docs/spec/modelo-de-datos/criterios.md that the task can break on its
// own: a key is a positive integer, no two criteria share one, and the
// text is a one-line field.
func (t *Task) validateCriteria() error {
	seen := make(map[int]bool, len(t.AcceptanceCriteria))
	for _, c := range t.AcceptanceCriteria {
		if c.Key < 1 {
			return fmt.Errorf(
				"acceptance criterion with key %d: a criterion's key is a positive integer assigned by the program",
				c.Key,
			)
		}
		if seen[c.Key] {
			return fmt.Errorf(
				"two acceptance criteria share the key %d: a criterion's key is unique within its task and is never reassigned",
				c.Key,
			)
		}
		seen[c.Key] = true
		if err := ValidateStringField(StringFieldCriterionText, c.Text); err != nil {
			return err
		}
	}
	return nil
}

// validatePeopleOfTheListsWithStructure checks the author of every comment
// and of the open question, which are `string` fields like the task's own
// author (docs/spec/modelo-de-datos/comentarios.md and
// docs/spec/modelo-de-datos/pregunta-abierta.md). Their bodies are `text`
// and are deliberately not checked.
//
// The keys of the comments are not checked here, in contrast with the
// criteria: a comment is never addressed by anything the caller writes
// beyond its key, and the same check on both lists would be the same rule
// twice; the schema's primary key is what keeps them apart.
func (t *Task) validatePeopleOfTheListsWithStructure() error {
	for _, c := range t.Comments {
		if c.Author == "" {
			continue
		}
		if err := ValidateStringField(StringFieldAuthor, c.Author); err != nil {
			return err
		}
	}
	if t.Question != nil && t.Question.Author != "" {
		if err := ValidateStringField(StringFieldAuthor, t.Question.Author); err != nil {
			return err
		}
	}
	return nil
}
