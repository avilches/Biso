package model

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"unicode"
)

// The two token alphabets of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token. Both
// take Unicode letters and digits; they differ only in which symbols they
// add. An extension key leaves out @ and :, which have no documented use
// there, and =, which --ext <key>=<value> already spends on separating the
// two halves.
const (
	tokenSymbols        = "-_.:@"
	extensionKeySymbols = "-_."
)

// ValidateLabel answers the error of
// docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token when
// a label steps outside its alphabet, and nil when it does not. It is a
// problem of form, not of the board not recognizing the value, so the exit
// code is 2 and not the 3 of an unknown vocabulary value: neither labels
// nor assignees have a closed vocabulary when written.
func ValidateLabel(label string) *Error {
	return validateToken(label, tokenSymbols, "label", "malformed_label", "labels",
		"a label may contain letters, digits, and - _ . : @")
}

// ValidateAssignee is ValidateLabel for an assignee, with the same
// alphabet and the same exit code.
func ValidateAssignee(assignee string) *Error {
	return validateToken(assignee, tokenSymbols, "assignee", "malformed_assignee", "assignees",
		"an assignee may contain letters, digits, and - _ . : @")
}

// ValidateExtensionKeySyntax answers the same kind of error for the key of
// an extension field, whose alphabet is the narrower one
// (docs/spec/modelo-de-datos/campos-externos.md). It says nothing about
// whether the board declares that key: that is ValidateExtensionKey, and
// the two are separate because they are different failures with different
// exit codes.
func ValidateExtensionKeySyntax(key string) *Error {
	return validateToken(key, extensionKeySymbols, "extension key", "malformed_extension_key", "ext",
		"an extension key may contain letters, digits, and - _ .")
}

// validateToken is the shared body of the three above.
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

// The four values the `field` key takes on a malformed_string_value error,
// per docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string.
// They name the field of the JSON envelope and not the flag that wrote it:
// the same criterion text arrives through --add-ac and through an import,
// and the reason it is rejected is the same one.
const (
	StringFieldTitle         = "title"
	StringFieldAuthor        = "author"
	StringFieldCriterionText = "criterion_text"
	StringFieldExt           = "ext"
)

// stringFieldNouns is how each of those four names itself in the message.
// The specification writes only the title one literally, `error: malformed
// title: "first line\nsecond line"`, so the other three follow its shape
// with the field's own name in prose.
var stringFieldNouns = map[string]string{
	StringFieldTitle:         "title",
	StringFieldAuthor:        "author",
	StringFieldCriterionText: "criterion text",
	StringFieldExt:           "extension value",
}

// ValidateStringField answers the error of
// docs/spec/valores-de-entrada.md#el-salto-de-línea-en-un-campo-string when
// a one-line field carries a line break, and nil when it does not.
//
// The distinction it enforces is the one the type table of
// docs/spec/modelo-de-datos/index.md draws: a `string` is text of one line
// and a `text` is a block of prose. Only the four fields above are
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

// ValidateExtensionKey answers the error of
// docs/spec/modelo-de-datos/campos-externos.md when a key is not one of
// the ones the board declares in its `extensions` list, and nil when it
// is. Writing an undeclared key is exit code 3.
//
// The second line the specification prints under the message,
//
//	error: unknown extension key: "jira.key"
//	       declared keys on this board: trello.card, github.issue
//
// is the rendering of Valid, the same shape as the unknown-status error of
// docs/spec/vocabularios.md, so it is not a hint and does not live here:
// this builds the error, and the layer that writes to stderr writes both
// lines from it.
func ValidateExtensionKey(key string, declared []string) *Error {
	for _, d := range declared {
		if d == key {
			return nil
		}
	}
	valid := make([]string, len(declared))
	copy(valid, declared)
	return &Error{
		ExitCode: 3,
		Code:     "unknown_extension_key",
		Message:  fmt.Sprintf("unknown extension key: %q", key),
		Field:    "ext",
		Given:    key,
		Valid:    valid,
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
// declaredExtensions is the board's `extensions` list, the one closed
// vocabulary a task carries inside itself
// (docs/spec/modelo-de-datos/campos-externos.md). The vocabularies a board
// configures for status, type and priority are not checked here: they are
// matched by internal/match against a configuration the model does not
// see.
//
// Two kinds of failure come out of it. The rules of the specification
// answer a *Error with its exit code and its code, ready to print. The
// keys of a criterion answer a plain error instead: the program assigns
// them, so a criterion with key 0 or two criteria sharing a key are a bug
// in whoever built the task, which is why the specification gives them no
// code of their own. Both matter here for the same reason: without this
// check the second one reached SQLite and came back as a raw constraint
// failure naming a table.
func (t *Task) Validate(declaredExtensions []string) error {
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
	for _, key := range SortedExtKeys(t.Ext) {
		if err := ValidateExtensionKeySyntax(key); err != nil {
			return err
		}
		if err := ValidateExtensionKey(key, declaredExtensions); err != nil {
			return err
		}
		if err := ValidateStringField(StringFieldExt, t.Ext[key]); err != nil {
			return err
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

// SortedExtKeys answers the keys of an extension map in a fixed order, so
// that a task with two bad keys always fails on the same one.
func SortedExtKeys(ext map[string]string) []string {
	keys := make([]string, 0, len(ext))
	for k := range ext {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
