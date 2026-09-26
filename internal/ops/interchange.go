package ops

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"slices"
	"sort"
	"strings"
	"time"

	"biso/internal/model"
)

// This file is the one definition of the interchange format there is: the
// shape `biso export` writes and the shape `biso new --from` reads back.
//
// The two directions share one struct and therefore one list of keys. That
// is what makes the symmetry guarantee of docs/spec/cmd/export.md mean
// something: a second list of keys written by hand somewhere else could
// drift away from this one without any test noticing, and then exporting a
// board and importing it would stop reproducing it field for field.
//
// The only keys the two directions do not share are `definitionOfDone`,
// `documentation` and `modifiedFiles`, which `new --from` accepts and
// `export` never writes because none is a field of the model
// (docs/spec/cmd/new.md#el-modo-lote). They live in wireInput, which embeds
// wireTask, so even that difference is written once.

// wireTask is one task as a line of the interchange format. Every key is
// always written: a scalar with no value goes as null and a list or a map
// with nothing in it as [] or {}, never null, which is the rule that makes
// re-importing the output reproduce the exact task
// (docs/spec/cmd/export.md).
//
// The field order is the order the keys come out in, because that is the
// order encoding/json writes a struct in.
type wireTask struct {
	ID                 string          `json:"id"`
	Title              string          `json:"title"`
	Status             *string         `json:"status"`
	Type               *string         `json:"type"`
	Priority           *string         `json:"priority"`
	Parent             *string         `json:"parent"`
	Assignees          []string        `json:"assignees"`
	Author             *string         `json:"author"`
	Labels             []string        `json:"labels"`
	Dependencies       []string        `json:"dependencies"`
	References         []string        `json:"references"`
	Due                *string         `json:"due"`
	Ordinal            *string         `json:"ordinal"`
	Description        *string         `json:"description"`
	Plan               *string         `json:"plan"`
	Notes              *string         `json:"notes"`
	Summary            *string         `json:"summary"`
	AcceptanceCriteria []wireCriterion `json:"acceptanceCriteria"`
	Comments           []wireComment   `json:"comments"`
	Question           *wireQuestion   `json:"question"`
	CreatedAt          *string         `json:"createdAt"`
	UpdatedAt          *string         `json:"updatedAt"`
	LeaseExpiresAt     *string         `json:"leaseExpiresAt"`
	LeaseHolder        *string         `json:"leaseHolder"`
	Archived           bool            `json:"archived"`
}

// wireInput is a line of a batch: everything above plus the three keys that
// only the reading direction knows, all of them from a foreign batch.
type wireInput struct {
	wireTask
	DefinitionOfDone []wireCriterion `json:"definitionOfDone"`
	Documentation    []string        `json:"documentation"`
	ModifiedFiles    []string        `json:"modifiedFiles"`
}

// wireCriterion is one acceptance criterion on the wire. Reading it takes
// the two forms of docs/spec/cmd/new.md#el-modo-lote, a bare string or an
// object, and writing it always produces the object.
type wireCriterion struct {
	Key     *int   `json:"key"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// MarshalJSON writes the object form, with the key as a number and never as
// null: `export` only ever writes criteria that already have one.
func (c wireCriterion) MarshalJSON() ([]byte, error) {
	key := 0
	if c.Key != nil {
		key = *c.Key
	}
	return json.Marshal(struct {
		Key     int    `json:"key"`
		Text    string `json:"text"`
		Checked bool   `json:"checked"`
	}{key, c.Text, c.Checked})
}

// UnmarshalJSON reads either form. A bare string is a criterion with no key
// of its own, which the import then takes from the task's counter.
func (c *wireCriterion) UnmarshalJSON(data []byte) error {
	if trimmed := strings.TrimSpace(string(data)); strings.HasPrefix(trimmed, "\"") {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		c.Key, c.Text, c.Checked = nil, text, false
		return nil
	}
	if err := objectKeys(data, []string{"key", "text", "checked"}); err != nil {
		return err
	}
	type plain wireCriterion
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*c = wireCriterion(value)
	return nil
}

// wireComment is one comment on the wire. Its createdAt is optional and its
// key follows the same rule as a criterion's
// (docs/spec/cmd/new.md#el-modo-lote).
type wireComment struct {
	Key       *int    `json:"key"`
	Author    *string `json:"author"`
	CreatedAt *string `json:"createdAt"`
	Body      string  `json:"body"`
}

func (c *wireComment) UnmarshalJSON(data []byte) error {
	if err := objectKeys(data, []string{"key", "author", "createdAt", "body"}); err != nil {
		return err
	}
	type plain wireComment
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*c = wireComment(value)
	return nil
}

// wireQuestion is the open question on the wire, whose askedAt is optional
// for the same reason a comment's createdAt is.
type wireQuestion struct {
	Author  *string `json:"author"`
	AskedAt *string `json:"askedAt"`
	Body    string  `json:"body"`
}

func (q *wireQuestion) UnmarshalJSON(data []byte) error {
	if err := objectKeys(data, []string{"author", "askedAt", "body"}); err != nil {
		return err
	}
	type plain wireQuestion
	var value plain
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*q = wireQuestion(value)
	return nil
}

// objectKeys rejects any key the shape does not declare. A key that is not
// in the format is a failure of validation and never something read past in
// silence, and the message names it (docs/spec/cmd/new.md#el-modo-lote).
//
// The keys are looked at in alphabetical order and not in the order they
// were written, so that a line with two of them always fails on the same
// one and the message of a batch is the same twice.
//
// It answers a typed error and not a bare one because `unknown_key` is a
// `code` of exit status 2 in the table of
// docs/spec/contrato-json.md#los-identificadores-de-error: a key the format
// does not have is a file written wrong, not a value this board cannot
// interpret, and wrapping it as `invalid_line` would file it under 3.
func objectKeys(data []byte, allowed []string) error {
	raw, err := rawObject(data)
	if err != nil {
		return err
	}
	for _, key := range sortedRawKeys(raw) {
		if !containsString(allowed, key) {
			return &model.Error{
				ExitCode: 2,
				Code:     "unknown_key",
				Message:  fmt.Sprintf("unknown key: %q", key),
				Field:    key,
			}
		}
	}
	return nil
}

func rawObject(data []byte) (map[string]json.RawMessage, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}

func sortedRawKeys(raw map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// interchangeKeys are the keys a line of a batch may carry, taken from the
// struct tags of wireInput itself. Deriving them instead of repeating them
// is what keeps the check and the format from drifting apart.
var interchangeKeys = keysOf(reflect.TypeOf(wireInput{}))

// listKeys are the keys whose empty value is [] or {} and never null
// (docs/spec/cmd/new.md#el-modo-lote), so an explicit null in one of them
// is a failure of validation and not the absence of the key.
//
// They are deduced from the shape itself, like interchangeKeys, and for the
// same reason: a list written by hand here would be a second list of fields
// that a new one could be left out of, and leaving it out would not fail
// anything, it would start accepting null in silence.
//
// The shape it reads is wireTask and not wireInput, which is exactly the
// rule: the lists of the model are the ones whose empty value is [] or {},
// and definitionOfDone, documentation and modifiedFiles are not among them,
// so null there means the key was not written (docs/spec/cmd/new.md#el-modo-lote).
var listKeys = listKeysOf(reflect.TypeOf(wireTask{}))

func keysOf(t reflect.Type) []string {
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			keys = append(keys, keysOf(f.Type)...)
			continue
		}
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if tag != "" && tag != "-" {
			keys = append(keys, tag)
		}
	}
	return keys
}

// listKeysOf is keysOf narrowed to the fields whose value is a list or a
// map, which is what makes a null in them a failure.
func listKeysOf(t reflect.Type) []string {
	var keys []string
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if f.Anonymous {
			keys = append(keys, listKeysOf(f.Type)...)
			continue
		}
		switch f.Type.Kind() {
		case reflect.Slice, reflect.Map:
		default:
			continue
		}
		tag := strings.Split(f.Tag.Get("json"), ",")[0]
		if tag != "" && tag != "-" {
			keys = append(keys, tag)
		}
	}
	return keys
}

// encodeTask writes one task as a line of the interchange format.
func encodeTask(t *model.Task) ([]byte, error) {
	w := wireTask{
		ID:                 t.ID,
		Title:              t.Title,
		Status:             orNil(t.Status),
		Type:               orNil(t.Type),
		Priority:           orNil(t.Priority),
		Parent:             orNil(t.Parent),
		Assignees:          listOrEmpty(t.Assignees),
		Author:             orNil(t.Author),
		Labels:             listOrEmpty(t.Labels),
		Dependencies:       listOrEmpty(t.Dependencies),
		References:         listOrEmpty(t.References),
		Due:                dayOrNil(t.Due),
		Ordinal:            orNil(t.Ordinal),
		Description:        orNil(t.Description),
		Plan:               orNil(t.Plan),
		Notes:              orNil(t.Notes),
		Summary:            orNil(t.Summary),
		AcceptanceCriteria: []wireCriterion{},
		Comments:           []wireComment{},
		CreatedAt:          instantOrNil(t.CreatedAt),
		UpdatedAt:          instantOrNil(t.UpdatedAt),
		LeaseExpiresAt:     instantOrNil(t.LeaseExpiresAt),
		LeaseHolder:        orNil(t.LeaseHolder),
		Archived:           t.Archived,
	}
	for _, c := range t.AcceptanceCriteria {
		key := c.Key
		w.AcceptanceCriteria = append(w.AcceptanceCriteria,
			wireCriterion{Key: &key, Text: c.Text, Checked: c.Checked})
	}
	for _, c := range t.Comments {
		key := c.Key
		w.Comments = append(w.Comments, wireComment{
			Key:       &key,
			Author:    orNil(c.Author),
			CreatedAt: instantOrNil(c.CreatedAt),
			Body:      c.Body,
		})
	}
	if t.Question != nil {
		w.Question = &wireQuestion{
			Author:  orNil(t.Question.Author),
			AskedAt: instantOrNil(t.Question.AskedAt),
			Body:    t.Question.Body,
		}
	}
	return json.Marshal(w)
}

// decoded is one line of a batch, already read but not yet judged against
// the board: the task it describes, which of the optional keys it really
// carried, and the criteria that came from definitionOfDone.
type decoded struct {
	task *model.Task
	// hasID, hasLease and the rest say the key was written, which a zero
	// value cannot: an explicit null in an optional scalar is the same as
	// not writing the key (docs/spec/cmd/new.md#el-modo-lote), but a key
	// that was written with a value has rules of its own.
	hasID             bool
	hasLeaseExpiresAt bool
	hasLeaseHolder    bool
	// dodKeys are the keys the conversion of definitionOfDone created, in
	// the order it created them. They are the one thing `biso new` ever
	// announces about a key it assigned (docs/spec/cmd/new.md#el-modo-lote).
	dodKeys []int
	// docCount is how many elements of documentation the line carried, not
	// empty, and were merged into references, which the warning announces.
	docCount int
	// fileCount is the same for the elements of modifiedFiles.
	fileCount int
	// emptyDropped is what the line dropped for being empty or only spaces
	// (docs/spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote), one
	// entry per list that lost something, in the fixed order of
	// elementLists and not in the order the keys were written in.
	emptyDropped []emptyDrop
	// rawStatus, rawType and rawPriority are the values as the line wrote
	// them, before the board's vocabulary has judged them.
	rawStatus, rawType, rawPriority string
	hasStatus                       bool
}

// emptyDrop is the count of elements one list of a line lost for being empty.
type emptyDrop struct {
	field string
	count int
}

// The lists of a line whose elements are text, in the order the warning of
// the elements dropped from each of them is emitted in, which is the one of
// docs/spec/cmd/new.md#el-modo-lote whatever order the keys were written in.
// The two lists of criteria are apart because an element of them may be an
// object, and a null in the place of one reads differently.
var (
	textLists     = []string{"assignees", "labels", "dependencies", "references"}
	criteriaLists = []string{"acceptanceCriteria", "definitionOfDone"}
	pointerLists  = []string{"documentation", "modifiedFiles"}
	// commentLists holds only "comments", apart from the other three because
	// its element takes neither of their shapes: a comment is always an
	// object, never a bare string, so anything else in its place, null
	// included, fails the line
	// (docs/spec/cmd/new.md#el-modo-lote).
	commentLists = []string{"comments"}
)

// checkElements is the failure of a value that is not text where a list of
// text has one, which for a null is what makes it different from an empty
// string: an empty string is dropped and a null is a line that cannot be read
// (docs/spec/cmd/new.md#el-modo-lote). The message names the list and the
// position, like the one for a number that encoding/json writes itself.
//
// It reads the raw elements and not the decoded ones because a null decodes
// as an empty string, which is exactly the distinction to keep.
func checkElements(raw map[string]json.RawMessage) error {
	for _, key := range slices.Concat(textLists, pointerLists) {
		for i, element := range rawElements(raw[key]) {
			if kind := jsonKind(element); kind == "null" {
				return fmt.Errorf("%s.%d: expected text, got null", key, i)
			}
		}
	}
	for _, key := range criteriaLists {
		for i, element := range rawElements(raw[key]) {
			switch kind := jsonKind(element); kind {
			case "string":
			case "object":
				fields, err := rawObject(element)
				if err != nil {
					return err
				}
				if text, ok := fields["text"]; ok {
					if kind := jsonKind(text); kind != "string" {
						return fmt.Errorf("%s.%d.text: expected text, got %s", key, i, kind)
					}
				}
			default:
				return fmt.Errorf("%s.%d: expected text or an object, got %s", key, i, kind)
			}
		}
	}
	for _, key := range commentLists {
		for i, element := range rawElements(raw[key]) {
			if kind := jsonKind(element); kind != "object" {
				return fmt.Errorf("%s.%d: expected an object, got %s", key, i, kind)
			}
		}
	}
	return nil
}

// rawElements is the elements of a value if it is a list, and nothing
// otherwise: a value that is not a list is the failure encoding/json reports
// with the name of the key.
func rawElements(value json.RawMessage) []json.RawMessage {
	var elements []json.RawMessage
	if jsonKind(value) != "array" || json.Unmarshal(value, &elements) != nil {
		return nil
	}
	return elements
}

// jsonKind names the type of a raw JSON value the way encoding/json does in
// its own messages.
func jsonKind(value json.RawMessage) string {
	trimmed := strings.TrimSpace(string(value))
	switch {
	case trimmed == "":
		return "null"
	case trimmed == "null":
		return "null"
	case trimmed[0] == '"':
		return "string"
	case trimmed[0] == '{':
		return "object"
	case trimmed[0] == '[':
		return "array"
	case trimmed == "true" || trimmed == "false":
		return "bool"
	}
	return "number"
}

// withoutEmpty drops the elements that are empty or only spaces, and answers
// what is left and how many went. What is left is stored as it came, without
// trimming (docs/spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote).
func withoutEmpty(values []string) ([]string, int) {
	dropped := 0
	for _, v := range values {
		if isEmpty(v) {
			dropped++
		}
	}
	if dropped == 0 {
		return values, 0
	}
	kept := make([]string, 0, len(values)-dropped)
	for _, v := range values {
		if !isEmpty(v) {
			kept = append(kept, v)
		}
	}
	return kept, dropped
}

// withoutEmptyCriteria is withoutEmpty for criteria, whose element is empty
// when its text is, whether it came as a string or as an object, and an object
// with no text has none. It is dropped before anything else of the element is
// looked at, so its key reserves nothing and its `checked` marks nothing.
func withoutEmptyCriteria(criteria []wireCriterion) ([]wireCriterion, int) {
	dropped := 0
	for _, c := range criteria {
		if isEmpty(c.Text) {
			dropped++
		}
	}
	if dropped == 0 {
		return criteria, 0
	}
	kept := make([]wireCriterion, 0, len(criteria)-dropped)
	for _, c := range criteria {
		if !isEmpty(c.Text) {
			kept = append(kept, c)
		}
	}
	return kept, dropped
}

// isEmpty is the definition of an empty value of
// docs/spec/valores-de-entrada.md#el-valor-vacío, the one the flags use.
func isEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// mergeIntoReferences appends each value to the references of t unless they
// already hold it, and answers how many values there were, which is what the
// warning of the batch counts, repeated ones included.
func mergeIntoReferences(t *model.Task, values []string) int {
	for _, v := range values {
		if !slices.Contains(t.References, v) {
			t.References = append(t.References, v)
		}
	}
	return len(values)
}

// decodeTask reads one line of the interchange format into a task.
//
// It is the counterpart of encodeTask and it reads the same struct, so a
// key that one writes is a key the other accepts. `now` is the instant a
// comment or a question with no date of its own is stamped with.
func decodeTask(line []byte, now time.Time) (*decoded, error) {
	if err := objectKeys(line, interchangeKeys); err != nil {
		return nil, err
	}
	raw, err := rawObject(line)
	if err != nil {
		return nil, err
	}
	for _, key := range listKeys {
		if value, ok := raw[key]; ok && string(value) == "null" {
			return nil, fmt.Errorf(
				"%s is null; an empty list or map is written [] or {}, never null", key)
		}
	}

	if err := checkElements(raw); err != nil {
		return nil, err
	}

	var in wireInput
	if err := json.Unmarshal(line, &in); err != nil {
		return nil, unwrapJSONError(err)
	}

	// An element that is empty or only spaces is dropped before anything
	// else of the line is done with it: it is not a malformed label, nor a
	// dependency that does not exist, nor a criterion that takes a key
	// (docs/spec/valores-de-entrada.md#un-elemento-vacío-en-un-lote). The
	// lists are visited in the order the warnings come out in.
	var emptyDropped []emptyDrop
	note := func(field string, dropped int) {
		if dropped > 0 {
			emptyDropped = append(emptyDropped, emptyDrop{field, dropped})
		}
	}
	var dropped int
	in.Assignees, dropped = withoutEmpty(in.Assignees)
	note("assignees", dropped)
	in.Labels, dropped = withoutEmpty(in.Labels)
	note("labels", dropped)
	in.Dependencies, dropped = withoutEmpty(in.Dependencies)
	note("dependencies", dropped)
	in.References, dropped = withoutEmpty(in.References)
	note("references", dropped)
	in.AcceptanceCriteria, dropped = withoutEmptyCriteria(in.AcceptanceCriteria)
	note("acceptanceCriteria", dropped)
	in.DefinitionOfDone, dropped = withoutEmptyCriteria(in.DefinitionOfDone)
	note("definitionOfDone", dropped)
	in.Documentation, dropped = withoutEmpty(in.Documentation)
	note("documentation", dropped)
	in.ModifiedFiles, dropped = withoutEmpty(in.ModifiedFiles)
	note("modifiedFiles", dropped)

	t := &model.Task{
		Title:        in.Title,
		Type:         value(in.Type),
		Priority:     value(in.Priority),
		Parent:       value(in.Parent),
		Assignees:    in.Assignees,
		Author:       value(in.Author),
		Labels:       in.Labels,
		Dependencies: in.Dependencies,
		References:   in.References,
		Ordinal:      value(in.Ordinal),
		Description:  value(in.Description),
		Plan:         value(in.Plan),
		Notes:        value(in.Notes),
		Summary:      value(in.Summary),
		LeaseHolder:  value(in.LeaseHolder),
		Archived:     in.Archived,
	}
	d := &decoded{
		task:              t,
		hasID:             in.ID != "",
		hasLeaseExpiresAt: in.LeaseExpiresAt != nil,
		hasLeaseHolder:    in.LeaseHolder != nil,
		rawStatus:         value(in.Status),
		rawType:           value(in.Type),
		rawPriority:       value(in.Priority),
		hasStatus:         in.Status != nil,
		emptyDropped:      emptyDropped,
	}
	t.ID = in.ID

	// The ordinal key is judged here and not by the model's validation,
	// because this is the last place where a key written as "" can still be
	// told from a key that was not written at all: inside a task the empty
	// string already means "no key". A line that writes one is writing a
	// value that is not a key, and the rule of
	// docs/spec/cmd/new.md#el-modo-lote is that a key of the wrong shape is
	// malformed_ordinal (the empty key of
	// docs/spec/modelo-de-datos/orden-manual.md#qué-es-una-clave-de-orden).
	// An explicit null, like a missing key, is a task with no place, which
	// is why only a value that really came through is asked.
	if in.Ordinal != nil {
		if e := model.ValidateOrdinal(*in.Ordinal); e != nil {
			return nil, e
		}
	}

	for _, f := range []struct {
		key   string
		value *string
		into  *time.Time
	}{
		{"createdAt", in.CreatedAt, &t.CreatedAt},
		{"updatedAt", in.UpdatedAt, &t.UpdatedAt},
		{"leaseExpiresAt", in.LeaseExpiresAt, &t.LeaseExpiresAt},
	} {
		if f.value == nil {
			continue
		}
		instant, err := readInstant(f.key, *f.value)
		if err != nil {
			return nil, err
		}
		*f.into = instant
	}
	if in.Due != nil {
		day, err := time.ParseInLocation(model.DateLayout, *in.Due, time.UTC)
		if err != nil {
			return nil, fmt.Errorf("due: invalid date: %q; a due date is written YYYY-MM-DD", *in.Due)
		}
		t.Due = day
	}

	if err := readCriteria(t, in.AcceptanceCriteria); err != nil {
		return nil, err
	}
	// definitionOfDone is converted after acceptanceCriteria and takes the
	// next free key of the counter, discarding whatever key it carried:
	// the two lists had counters of their own, so the same key in both is
	// no contradiction (docs/spec/cmd/new.md#el-modo-lote).
	for _, c := range in.DefinitionOfDone {
		added := t.AddCriterion(c.Text)
		added.Checked = c.Checked
		d.dodKeys = append(d.dodKeys, added.Key)
	}
	// documentation and then modifiedFiles are merged into references, after
	// the ones the line already had and in the order they came, whatever
	// order the keys were written in. A value that references already holds
	// is not added again. That is a rule of this merge only: a references list
	// that already carries a repeated value keeps it repeated.
	d.docCount = mergeIntoReferences(t, in.Documentation)
	d.fileCount = mergeIntoReferences(t, in.ModifiedFiles)
	if err := readComments(t, in.Comments, now); err != nil {
		return nil, err
	}
	if in.Question != nil {
		q := &model.Question{Author: value(in.Question.Author), Body: in.Question.Body, AskedAt: now}
		if in.Question.AskedAt != nil {
			instant, err := readInstant("question.askedAt", *in.Question.AskedAt)
			if err != nil {
				return nil, err
			}
			q.AskedAt = instant
		}
		t.Question = q
	}
	return d, nil
}

// readCriteria fills the acceptance criteria, settling every key: the one
// the line gave, or the next free one of the task's counter, which ends up
// above the highest key imported (docs/spec/cmd/new.md#el-modo-lote).
func readCriteria(t *model.Task, criteria []wireCriterion) error {
	highest := 0
	for _, c := range criteria {
		if c.Key != nil && *c.Key > highest {
			highest = *c.Key
		}
	}
	t.NextCriterionKey = highest + 1
	seen := map[int]bool{}
	for _, c := range criteria {
		if c.Key == nil {
			t.AcceptanceCriteria = append(t.AcceptanceCriteria,
				model.Criterion{Key: t.NextCriterionKey, Text: c.Text, Checked: c.Checked})
			t.NextCriterionKey++
			continue
		}
		if *c.Key < 1 {
			return fmt.Errorf("acceptanceCriteria: key %d; a key is a positive whole number", *c.Key)
		}
		if seen[*c.Key] {
			return fmt.Errorf("acceptanceCriteria: key %d appears twice in the same task", *c.Key)
		}
		seen[*c.Key] = true
		t.AcceptanceCriteria = append(t.AcceptanceCriteria,
			model.Criterion{Key: *c.Key, Text: c.Text, Checked: c.Checked})
	}
	return nil
}

// readComments is readCriteria for the comments, with two differences the
// specification gives them: a comment with no createdAt is stamped with the
// instant of the import, and an empty or blank body fails the line instead
// of being dropped, because a comment with nothing in it is not a hole that
// can be skipped but a comment missing the one thing that makes it one
// (docs/spec/cmd/new.md#el-modo-lote).
func readComments(t *model.Task, comments []wireComment, now time.Time) error {
	highest := 0
	for _, c := range comments {
		if c.Key != nil && *c.Key > highest {
			highest = *c.Key
		}
	}
	t.NextCommentKey = highest + 1
	seen := map[int]bool{}
	for i, c := range comments {
		if isEmpty(c.Body) {
			return fmt.Errorf("comments.%d: comment body cannot be empty", i)
		}
		at := now
		if c.CreatedAt != nil {
			instant, err := readInstant("comments.createdAt", *c.CreatedAt)
			if err != nil {
				return err
			}
			at = instant
		}
		comment := model.Comment{Author: value(c.Author), CreatedAt: at, Body: c.Body}
		if c.Key == nil {
			comment.Key = t.NextCommentKey
			t.NextCommentKey++
		} else {
			if *c.Key < 1 {
				return fmt.Errorf("comments: key %d; a key is a positive whole number", *c.Key)
			}
			if seen[*c.Key] {
				return fmt.Errorf("comments: key %d appears twice in the same task", *c.Key)
			}
			seen[*c.Key] = true
			comment.Key = *c.Key
		}
		t.Comments = append(t.Comments, comment)
	}
	return nil
}

func readInstant(key, value string) (time.Time, error) {
	instant, err := time.ParseInLocation(model.InstantLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, fmt.Errorf(
			"%s: invalid instant: %q; an instant is written YYYY-MM-DDTHH:MM:SSZ, in UTC", key, value)
	}
	return instant, nil
}

// unwrapJSONError turns what encoding/json says about a value of the wrong
// type into the sentence a line of a batch prints, which names the key and
// not the Go type behind it.
func unwrapJSONError(err error) error {
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return fmt.Errorf("%s: expected %s, got %s", typeErr.Field, goTypeName(typeErr), typeErr.Value)
	}
	return err
}

// goTypeName writes the expected type the way the format talks about it and
// not the way Go does: a list and not a slice, text and not string.
func goTypeName(e *json.UnmarshalTypeError) string {
	switch e.Type.Kind() {
	case reflect.Slice:
		return "a list"
	case reflect.Map:
		return "an object"
	case reflect.String:
		return "text"
	case reflect.Bool:
		return "true or false"
	case reflect.Int, reflect.Int64, reflect.Float64:
		return "a number"
	}
	return e.Type.String()
}

func orNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func value(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func listOrEmpty(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}

func instantOrNil(at time.Time) *string {
	if at.IsZero() {
		return nil
	}
	s := at.UTC().Format(model.InstantLayout)
	return &s
}

func dayOrNil(at time.Time) *string {
	if at.IsZero() {
		return nil
	}
	s := at.UTC().Format(model.DateLayout)
	return &s
}
