package model

import (
	"fmt"
	"strings"
	"time"
)

// DateLayout is how a `date` field that names a calendar day is written:
// YYYY-MM-DD, per docs/spec/contrato-json.md#números-fechas-y-ausencias. It
// is the layout of `due`, the only field of that shape.
const DateLayout = "2006-01-02"

// InstantLayout is how a `date` field that names an instant is written:
// ISO 8601 in UTC, ending in Z, with second precision, per
// docs/spec/contrato-json.md#números-fechas-y-ausencias.
const InstantLayout = "2006-01-02T15:04:05Z"

// Task is one task of a board, the logical model of
// docs/spec/modelo-de-datos/index.md. It knows nothing about SQLite and
// nothing about JSON: translating it to rows is internal/board's job and
// translating it to the interchange format is internal/ops's.
//
// Two conventions run through the whole type, and both come from the
// specification rather than from Go:
//
//   - **An empty string is the absence of a value**, for every string and
//     text field. docs/spec/valores-de-entrada.md#el-valor-vacío makes the
//     empty string never a storable value: on a scalar it is an error, and
//     the way to leave a field with no value is --clear-<field>. So "" and
//     "no value" can never be told apart by anything the caller can do, and
//     the model does not carry the distinction either. The JSON contract
//     writes them as null.
//   - **A zero time.Time is the absence of a date**, for the same reason:
//     no date field can hold the zero instant.
//
// Ordinal is the one field that needs a pointer, because 0 is a value a
// caller can legitimately give it (docs/spec/modelo-de-datos/index.md says
// int >= 0), so it cannot double as the absence of one.
//
// The derived fields of docs/spec/modelo-de-datos/index.md#los-campos-derivados
// are not fields here: none of them is stored, so they are methods
// (AcDone, AcTotal, CommentCount, Waiting, Urgency) or, for the ones that
// need the rest of the board, arguments of the call that computes them.
type Task struct {
	// The automatic fields.
	ID             string    // "MYP-11", allocated once and never reused
	CreatedAt      time.Time // UTC, second precision
	UpdatedAt      time.Time // UTC, second precision
	Archived       bool
	LeaseExpiresAt time.Time // UTC, zero when there is no lease
	LeaseHolder    string

	// The fields the caller sets.
	Title         string
	Status        string
	Type          string
	Priority      string
	Parent        string
	Assignees     []string
	Author        string
	Labels        []string
	Dependencies  []string
	References    []string
	ModifiedFiles []string
	Due           time.Time // a calendar day at UTC midnight, zero when unset
	Ordinal       *int
	Ext           map[string]string
	Description   string
	Plan          string
	Notes         string
	Summary       string

	AcceptanceCriteria []Criterion
	Comments           []Comment
	Question           *Question

	// NextCriterionKey and NextCommentKey are the two per-task counters of
	// docs/spec/modelo-de-datos/criterios.md and
	// docs/spec/modelo-de-datos/comentarios.md. They only grow, so removing
	// an element never frees its key. They are stored with the task, and
	// therefore exported: internal/board has to read them back. Zero means
	// "no element created yet", so the first key of an empty list is 1
	// whether the counter was written or not.
	NextCriterionKey int
	NextCommentKey   int
}

// Criterion is one acceptance criterion, the only checklist a task has
// (docs/spec/modelo-de-datos/criterios.md). It has no date and no author of
// its own.
type Criterion struct {
	Key     int // positive, assigned by the program, never reassigned
	Text    string
	Checked bool
}

// Comment is one comment of a task (docs/spec/modelo-de-datos/comentarios.md).
// Its body and its author are never edited once written: the only things
// the program does to an existing comment are removing it whole and
// correcting its date.
type Comment struct {
	Key       int // positive, assigned by the program, never reassigned
	Author    string
	CreatedAt time.Time // UTC, second precision
	Body      string
}

// Question is a task's open question
// (docs/spec/modelo-de-datos/pregunta-abierta.md). There is at most one, so
// it carries no key: a single value is not an element of a list that needs
// addressing. A task with one is waiting for a person to answer, whatever
// its status.
type Question struct {
	Author  string
	AskedAt time.Time // UTC, second precision
	Body    string
}

// ListField names one of the five list<string> fields of
// docs/spec/modelo-de-datos/index.md. It is a closed vocabulary: ListField
// and SetListField reject any other name instead of answering an empty
// list, per the rule of the project's CLAUDE.md that a value that does not
// exist is an error whether it is being written or read.
type ListField string

// The five list<string> fields. The names are the ones the JSON contract
// uses, so the same constant serves the model, the storage and the wire.
const (
	FieldAssignees     ListField = "assignees"
	FieldLabels        ListField = "labels"
	FieldDependencies  ListField = "dependencies"
	FieldReferences    ListField = "references"
	FieldModifiedFiles ListField = "modifiedFiles"
)

// listFields is the closed vocabulary itself, in the order of the table of
// docs/spec/modelo-de-datos/index.md.
var listFields = []ListField{
	FieldAssignees,
	FieldLabels,
	FieldDependencies,
	FieldReferences,
	FieldModifiedFiles,
}

// ListFields answers the five list<string> field names, in the order of the
// specification's table. The caller gets a copy: the vocabulary is closed
// and nobody outside this package extends it.
func ListFields() []ListField {
	out := make([]ListField, len(listFields))
	copy(out, listFields)
	return out
}

// ErrUnknownListField is what ListField and SetListField answer for a name
// that is not one of the five. It is not a *Error of the specification:
// no command lets a caller name a list field freely, so reaching it means
// the program asked for a field that does not exist, which is a bug and
// not a case of docs/spec/codigos-de-salida.md.
type ErrUnknownListField ListField

// Error implements the standard library's error interface.
func (e ErrUnknownListField) Error() string {
	return fmt.Sprintf("unknown list field %q", string(e))
}

// ListField answers the values of one of the five list fields.
func (t *Task) ListField(f ListField) ([]string, error) {
	switch f {
	case FieldAssignees:
		return t.Assignees, nil
	case FieldLabels:
		return t.Labels, nil
	case FieldDependencies:
		return t.Dependencies, nil
	case FieldReferences:
		return t.References, nil
	case FieldModifiedFiles:
		return t.ModifiedFiles, nil
	}
	return nil, ErrUnknownListField(f)
}

// SetListField replaces the values of one of the five list fields, in the
// order given. Lists are never sorted on their own
// (docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura).
func (t *Task) SetListField(f ListField, values []string) error {
	switch f {
	case FieldAssignees:
		t.Assignees = values
	case FieldLabels:
		t.Labels = values
	case FieldDependencies:
		t.Dependencies = values
	case FieldReferences:
		t.References = values
	case FieldModifiedFiles:
		t.ModifiedFiles = values
	default:
		return ErrUnknownListField(f)
	}
	return nil
}

// AddCriterion appends a criterion with the next key of the task's own
// counter and answers it. The returned pointer points into the task's
// slice, so writing through it changes the task.
func (t *Task) AddCriterion(text string) *Criterion {
	if t.NextCriterionKey < 1 {
		t.NextCriterionKey = 1
	}
	t.AcceptanceCriteria = append(t.AcceptanceCriteria, Criterion{
		Key:  t.NextCriterionKey,
		Text: text,
	})
	t.NextCriterionKey++
	return &t.AcceptanceCriteria[len(t.AcceptanceCriteria)-1]
}

// Criterion answers the criterion with that key, or nil. A criterion is
// always addressed by its key and never by its position.
func (t *Task) Criterion(key int) *Criterion {
	for i := range t.AcceptanceCriteria {
		if t.AcceptanceCriteria[i].Key == key {
			return &t.AcceptanceCriteria[i]
		}
	}
	return nil
}

// CriterionOrError is Criterion with the error of
// docs/spec/familias-de-flags.md#selectores-de-criterios when there is no
// such key: exit code 4, code criterion_not_found, and a message that
// lists the keys the task does have, so the caller does not have to ask
// again to find out.
func (t *Task) CriterionOrError(key int) (*Criterion, *Error) {
	if c := t.Criterion(key); c != nil {
		return c, nil
	}
	return nil, &Error{
		ExitCode: 4,
		Code:     "criterion_not_found",
		Message: fmt.Sprintf(
			"no acceptance criterion #%d on %s (keys: %s)",
			key, t.ID, joinKeys(t.CriterionKeys()),
		),
	}
}

// RemoveCriterion removes the criterion with that key and answers whether
// there was one. The keys of the others do not move, and the counter does
// not go back: a task can perfectly well have criteria #1 and #3.
func (t *Task) RemoveCriterion(key int) bool {
	for i := range t.AcceptanceCriteria {
		if t.AcceptanceCriteria[i].Key == key {
			t.AcceptanceCriteria = append(t.AcceptanceCriteria[:i], t.AcceptanceCriteria[i+1:]...)
			return true
		}
	}
	return false
}

// CriterionKeys answers the keys of the criteria the task has, in the
// order the criteria are in.
func (t *Task) CriterionKeys() []int {
	keys := make([]int, 0, len(t.AcceptanceCriteria))
	for _, c := range t.AcceptanceCriteria {
		keys = append(keys, c.Key)
	}
	return keys
}

// AcTotal is the derived field of the same name: the number of criteria
// present, never the highest key
// (docs/spec/modelo-de-datos/criterios.md).
func (t *Task) AcTotal() int { return len(t.AcceptanceCriteria) }

// AcDone is the derived field of the same name: how many of those are
// checked.
func (t *Task) AcDone() int {
	done := 0
	for _, c := range t.AcceptanceCriteria {
		if c.Checked {
			done++
		}
	}
	return done
}

// AddComment appends a comment with the next key of the task's own
// counter and answers it. Comments are kept in creation order and never in
// order of createdAt (docs/spec/modelo-de-datos/comentarios.md).
func (t *Task) AddComment(author string, at time.Time, body string) *Comment {
	if t.NextCommentKey < 1 {
		t.NextCommentKey = 1
	}
	t.Comments = append(t.Comments, Comment{
		Key:       t.NextCommentKey,
		Author:    author,
		CreatedAt: at,
		Body:      body,
	})
	t.NextCommentKey++
	return &t.Comments[len(t.Comments)-1]
}

// Comment answers the comment with that key, or nil.
func (t *Task) Comment(key int) *Comment {
	for i := range t.Comments {
		if t.Comments[i].Key == key {
			return &t.Comments[i]
		}
	}
	return nil
}

// CommentOrError is Comment with the error of
// docs/spec/familias-de-flags.md#comentarios when there is no such key.
func (t *Task) CommentOrError(key int) (*Comment, *Error) {
	if c := t.Comment(key); c != nil {
		return c, nil
	}
	return nil, &Error{
		ExitCode: 4,
		Code:     "comment_not_found",
		Message: fmt.Sprintf(
			"no comment #%d on %s (keys: %s)",
			key, t.ID, joinKeys(t.CommentKeys()),
		),
	}
}

// RemoveComment removes the comment with that key and answers whether
// there was one. Like a criterion's, the other keys do not move.
func (t *Task) RemoveComment(key int) bool {
	for i := range t.Comments {
		if t.Comments[i].Key == key {
			t.Comments = append(t.Comments[:i], t.Comments[i+1:]...)
			return true
		}
	}
	return false
}

// CommentKeys answers the keys of the comments the task has, in order.
func (t *Task) CommentKeys() []int {
	keys := make([]int, 0, len(t.Comments))
	for _, c := range t.Comments {
		keys = append(keys, c.Key)
	}
	return keys
}

// CommentCount is the derived field of the same name.
func (t *Task) CommentCount() int { return len(t.Comments) }

// Waiting is the derived field of the same name: a task with an open
// question is waiting for a person to answer
// (docs/spec/modelo-de-datos/pregunta-abierta.md).
func (t *Task) Waiting() bool { return t.Question != nil }

// joinKeys writes a key list the way the two not-found messages print it:
// "1, 3".
func joinKeys(keys []int) string {
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%d", k))
	}
	return strings.Join(parts, ", ")
}
