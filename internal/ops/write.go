package ops

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"biso/internal/board"
	"biso/internal/match"
	"biso/internal/model"
)

// This file is the engine every writing command shares: it takes the field
// flags of docs/spec/familias-de-flags.md, already classified into the nine
// fixed steps of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura, and
// applies them to a task in that order, whatever order they were written in.
//
// `biso new` and `biso set` are the first two callers; the six verbs of the
// cycle and `biso archive` are the same engine with a name and some defaults
// on top, which is what makes the specification's promise true that a flag
// means the same in every command that takes it.

// Step is one of the nine steps of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura. The
// numeric order is the order of that section, so sorting by it is the
// specified order and nothing else has to know what that order is.
type Step int

const (
	StepNone Step = iota
	StepClear
	StepReplace
	StepRemove
	StepAdd
	StepExt
	StepScalar
	StepCheckAC
	StepCommentDate
	StepComment
)

// Change is one value of one flag that writes a field: the flag's long name
// without its dashes, the step it belongs to, and the value. Key is filled
// only by the two flags whose value is a pair, --ext and --set-comment-date.
//
// It is a plain struct of strings and not the parser's own type because the
// dependency rule runs one way: internal/cli knows this package, and this
// package knows nothing about a command line.
type Change struct {
	Flag  string
	Step  Step
	Key   string
	Value string
}

// Warning is one line of docs/spec/salida-y-terminal.md#notas-y-avisos: its
// stable code, the text that follows "warning: " on stderr, its hints, and
// the fields the JSON object of data.warnings carries.
type Warning struct {
	Code    string
	Message string
	Hints   []string
	Fields  map[string]any
	// Detail are the lines printed under the message, each one already
	// carrying its own indentation, the way the unchecked criteria are
	// listed under the warning of
	// docs/spec/cmd/verbos-del-ciclo.md#biso-finish. They are text for
	// whoever is reading and never travel in the JSON object: what a
	// program needs about that warning is already in Fields.
	Detail []string
}

// TaskWrite is one task a write affected, with the three derived data of the
// status line of docs/spec/cmd/set.md#salida and the two lists of its JSON
// schema.
type TaskWrite struct {
	ID       string
	Status   string
	AcDone   int
	AcTotal  int
	Urgency  float64
	Changed  []string
	AcAdded  []int
	Archived bool
}

// WriteResult is what every command of kind task.write answers.
type WriteResult struct {
	Tasks []TaskWrite
	// Views are the cards --print asks for: the tasks the write affected,
	// as `biso get` would show them (docs/spec/cmd/flags-globales.md). It
	// is empty unless the call wrote that flag.
	Views    []TaskView
	Warnings []Warning
	Notes    []string
	DryRun   bool
	// Created marks the result of `biso new`, whose default output is the
	// identifier alone and not the status line
	// (docs/spec/cmd/new.md#salida).
	Created bool
	// Batch marks the result of the --from of `biso new`, whose preview
	// counts the lines of the file instead of saying "1 task", and
	// Previewed is that count (docs/spec/cmd/new.md#el-modo-lote).
	Batch     bool
	Previewed int
}

// writer is one writing invocation while it is being applied.
type writer struct {
	b       *board.Board
	env     Env
	now     time.Time
	changes []Change
	// manyTasks is what makes a selector that is not `all` a usage error:
	// the selector of one task does not have to mean the same in another
	// (docs/spec/familias-de-flags.md#selectores-de-criterios).
	manyTasks bool
	// newTask marks the engine running over a task that does not exist
	// yet, where every --clear-* does nothing and warns
	// (docs/spec/cmd/new.md).
	newTask bool
	// leading is what a verb writes into the comments of a task before the
	// --comment flags of the same call, whatever order the command line
	// had. `biso answer` is the one verb that has any
	// (docs/spec/cmd/verbos-del-ciclo.md#biso-answer).
	leading func(t *model.Task) error
	// claims are the conditional lease claims this write carries, which
	// internal/board checks inside the transaction that writes the rows.
	claims []board.LeaseClaim

	warnings []Warning
	notes    []string

	// board holds every task of the board, read once, for the cycles of
	// --add-deps and --parent and for the two terms of the urgency that
	// depend on the rest of the board.
	all     []*model.Task
	allRead bool

	// replaced remembers which --replace-* flag has already emptied its
	// list on the task being written, so that the second value of the same
	// flag accumulates instead of emptying again.
	replaced map[string]bool
	// acAdded are the keys the current task's --add-ac flags just created.
	acAdded []int
}

func newWriter(b *board.Board, env Env, changes []Change) *writer {
	ordered := make([]Change, len(changes))
	copy(ordered, changes)
	sort.SliceStable(ordered, func(i, j int) bool { return ordered[i].Step < ordered[j].Step })
	return &writer{
		b:        b,
		env:      env,
		now:      env.Now().UTC().Truncate(time.Second),
		changes:  ordered,
		replaced: map[string]bool{},
	}
}

func (w *writer) warn(warning Warning) { w.warnings = append(w.warnings, warning) }

// partial is what a failed write answers beside its error: the warnings it
// had already earned, which never disappear with the failure
// (docs/spec/salida-y-terminal.md#notas-y-avisos). The layer that prints
// them folds them into the same envelope as the error when the call asked
// for JSON (docs/spec/contrato-json.md#los-errores-en-json).
func (w *writer) partial() *WriteResult {
	return &WriteResult{Warnings: w.warnings}
}

func (w *writer) note(note string) { w.notes = append(w.notes, note) }

// ofStep answers the changes of one step, in the order of the command line.
func (w *writer) ofStep(step Step) []Change {
	var out []Change
	for _, c := range w.changes {
		if c.Step == step {
			out = append(out, c)
		}
	}
	return out
}

// tasks reads the whole board once. Two things need it and neither can be
// answered from the task alone: whether adding a dependency or a parent
// would close a cycle, and the two terms of the urgency that ask about other
// tasks (docs/spec/modelo-de-datos/urgencia.md).
func (w *writer) tasks() ([]*model.Task, error) {
	if w.allRead {
		return w.all, nil
	}
	all, skipped, err := w.b.Tasks.All()
	if err != nil {
		return nil, err
	}
	w.all, w.allRead = all, true
	if len(skipped) > 0 {
		w.warn(skippedWarning(skipped))
	}
	return w.all, nil
}

// skippedWarning is the one of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// for a set read that left a task out.
func skippedWarning(skipped []board.Skipped) Warning {
	ids := make([]string, 0, len(skipped))
	for _, s := range skipped {
		ids = append(ids, s.ID)
	}
	noun := "tasks could"
	if len(ids) == 1 {
		noun = "task could"
	}
	return Warning{
		Code: "task_skipped",
		Message: fmt.Sprintf("%d %s not be read and was skipped: %s",
			len(ids), noun, strings.Join(ids, ", ")),
		Fields: map[string]any{"count": len(ids), "tasks": ids},
	}
}

// apply writes every change onto the task, in the nine steps of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura.
func (w *writer) apply(t *model.Task) error {
	w.replaced = map[string]bool{}
	w.acAdded = nil

	// The two comment selectors are resolved here, against the list as it
	// was before the write, and not each one in its own step: they fall in
	// steps 3 and 8, so resolving them separately would let --rm-comment
	// delete the comment before --set-comment-date could ever conflict
	// with it (docs/spec/familias-de-flags.md#comentarios).
	rmComments, commentDates, err := w.commentSelectors(t)
	if err != nil {
		return err
	}

	for _, step := range []Step{StepClear, StepReplace, StepRemove, StepAdd, StepExt, StepScalar} {
		for _, c := range w.ofStep(step) {
			if c.Flag == "rm-comment" {
				continue
			}
			if err := w.applyOne(t, c); err != nil {
				return err
			}
		}
		if step == StepRemove {
			removeComments(t, rmComments)
		}
	}
	if err := w.applyChecks(t); err != nil {
		return err
	}
	w.applyCommentDates(t, commentDates)
	if err := w.applyComments(t); err != nil {
		return err
	}
	return nil
}

// applyOne writes one change. Every flag of docs/spec/familias-de-flags.md
// answers here and nowhere else, so that adding a command never adds a
// second meaning for a flag that already has one.
func (w *writer) applyOne(t *model.Task, c Change) error {
	if w.newTask && c.Step == StepClear {
		// On a task that does not exist yet there is nothing to empty, so
		// every --clear-* does nothing and says so (docs/spec/cmd/new.md).
		w.warn(Warning{
			Code:    "clear_on_new_task",
			Message: "--" + c.Flag + " has no effect on a new task",
			Fields:  map[string]any{"field": c.Flag},
		})
		return nil
	}
	if field, op, ok := listFlag(c.Flag); ok {
		return w.applyList(t, c, field, op)
	}
	switch c.Flag {
	case "clear-desc":
		t.Description = ""
	case "clear-plan":
		t.Plan = ""
	case "clear-notes":
		t.Notes = ""
	case "clear-summary":
		t.Summary = ""
	case "clear-acs":
		t.AcceptanceCriteria = nil
	case "clear-ext":
		t.Ext = nil
	case "clear-type":
		t.Type = ""
	case "clear-priority":
		t.Priority = ""
	case "clear-parent":
		t.Parent = ""
	case "clear-due":
		t.Due = time.Time{}
	case "clear-ordinal":
		t.Ordinal = nil
	case "clear-author":
		t.Author = ""
	case "append-desc":
		appendProse(&t.Description, c.Value)
	case "append-plan":
		appendProse(&t.Plan, c.Value)
	case "append-note":
		appendProse(&t.Notes, c.Value)
	case "append-summary":
		appendProse(&t.Summary, c.Value)
	case "add-ac":
		w.acAdded = append(w.acAdded, t.AddCriterion(c.Value).Key)
	case "rm-ac":
		return w.removeCriteria(t, c)
	case "ext":
		if t.Ext == nil {
			t.Ext = map[string]string{}
		}
		t.Ext[c.Key] = c.Value
	case "rm-ext":
		if _, ok := t.Ext[c.Value]; !ok {
			w.warnNotPresent(t, c)
			return nil
		}
		delete(t.Ext, c.Value)
	case "title":
		t.Title = c.Value
	case "status":
		status, err := match.Match(match.Status, c.Value, w.b.Config.Statuses)
		if err != nil {
			return err
		}
		t.Status = status
	case "type":
		value, err := match.Match(match.Type, c.Value, w.b.Config.Types)
		if err != nil {
			return err
		}
		t.Type = value
	case "priority":
		value, err := match.Match(match.Priority, c.Value, w.b.Config.Priorities)
		if err != nil {
			return err
		}
		t.Priority = value
	case "parent":
		resolved, err := resolveRef(w.b, c.Value, RefAuto)
		if err != nil {
			return err
		}
		if resolved.Note != "" {
			w.note(resolved.Note)
		}
		t.Parent = resolved.Task.ID
	case "due":
		due, err := parseDueDate(c.Value)
		if err != nil {
			return err
		}
		t.Due = due
		if due.Before(w.today()) {
			w.warn(Warning{
				Code:    "due_in_past",
				Message: fmt.Sprintf("--due %s is in the past", c.Value),
				Fields:  map[string]any{"value": c.Value},
			})
		}
	case "ordinal":
		n, err := strconv.Atoi(c.Value)
		if err != nil {
			return &model.Error{
				ExitCode: 2,
				Code:     "invalid_number",
				Message:  fmt.Sprintf("--ordinal: not a whole number: %q", c.Value),
				Field:    "ordinal",
				Given:    c.Value,
			}
		}
		t.Ordinal = &n
	case "author":
		t.Author = c.Value
	default:
		return fmt.Errorf("no field is written by --%s", c.Flag)
	}
	return nil
}

// today is the UTC calendar day of this call, which is what the urgency and
// the past-due warning compare against
// (docs/spec/modelo-de-datos/urgencia.md).
func (w *writer) today() time.Time {
	y, m, d := w.now.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// listOp is which of the four things a flag of
// docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma does.
type listOp int

const (
	opAdd listOp = iota
	opRemove
	opClear
	opReplace
)

// listSuffixes maps the tail of a list flag to the field it writes. The four
// prefixes and these five suffixes are the whole family: there is no field
// that breaks the shape, which is what `biso set --help` promises.
var listSuffixes = map[string]model.ListField{
	"labels":    model.FieldLabels,
	"assignees": model.FieldAssignees,
	"refs":      model.FieldReferences,
	"deps":      model.FieldDependencies,
	"files":     model.FieldModifiedFiles,
}

func listFlag(flag string) (model.ListField, listOp, bool) {
	for prefix, op := range map[string]listOp{
		"add-": opAdd, "rm-": opRemove, "clear-": opClear, "replace-": opReplace,
	} {
		suffix, ok := strings.CutPrefix(flag, prefix)
		if !ok {
			continue
		}
		if field, ok := listSuffixes[suffix]; ok {
			return field, op, true
		}
	}
	return "", 0, false
}

func (w *writer) applyList(t *model.Task, c Change, field model.ListField, op listOp) error {
	values, err := t.ListField(field)
	if err != nil {
		return err
	}
	switch op {
	case opClear:
		return t.SetListField(field, nil)
	case opReplace:
		if !w.replaced[c.Flag] {
			w.replaced[c.Flag] = true
			if len(values) > 0 {
				w.warn(Warning{
					Code: "overwrite",
					Message: fmt.Sprintf("--%s replaced %d existing %s",
						c.Flag, len(values), pluralItems(len(values), string(field))),
					Fields: map[string]any{
						"task": t.ID, "field": string(field), "count": len(values),
					},
				})
			}
			values = nil
			if err := t.SetListField(field, nil); err != nil {
				return err
			}
		}
		if c.Value == "" {
			// --replace-labels "" is emptying, and emptying is explicit
			// (docs/spec/valores-de-entrada.md#el-valor-vacío).
			return nil
		}
		return w.addValue(t, c, field, values)
	case opRemove:
		kept := make([]string, 0, len(values))
		found := false
		for _, v := range values {
			if v == c.Value {
				found = true
				continue
			}
			kept = append(kept, v)
		}
		if !found {
			w.warnNotPresent(t, c)
			return nil
		}
		return t.SetListField(field, kept)
	}
	return w.addValue(t, c, field, values)
}

// addValue adds one value to a list field, with the two rules every list of
// this family shares: a dependency is validated as it is written and stored
// as the identifier it resolves to, and a value the list already has is kept
// once and warns (docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma).
func (w *writer) addValue(t *model.Task, c Change, field model.ListField, values []string) error {
	value := c.Value
	if field == model.FieldDependencies {
		resolved, err := resolveRef(w.b, value, RefAuto)
		if err != nil {
			return err
		}
		if resolved.Note != "" {
			w.note(resolved.Note)
		}
		value = resolved.Task.ID
	}
	for _, v := range values {
		if v == value {
			w.warn(Warning{
				Code:    "value_already_present",
				Message: fmt.Sprintf("--%s: %q already present, kept once", c.Flag, value),
				Fields:  map[string]any{"flag": "--" + c.Flag, "value": value, "task": t.ID},
			})
			return nil
		}
	}
	return t.SetListField(field, append(values, value))
}

// warnNotPresent is the tolerant half of every flag that removes: taking out
// a value the task does not have, or a key the map does not have, warns and
// never fails, so that removing never forces a read first.
func (w *writer) warnNotPresent(t *model.Task, c Change) {
	w.warn(Warning{
		Code:    "value_not_present",
		Message: fmt.Sprintf("--%s: %q not present, nothing removed", c.Flag, c.Value),
		Fields:  map[string]any{"flag": "--" + c.Flag, "value": c.Value, "task": t.ID},
	})
}

// pluralItems writes the noun of the overwrite warning the way
// docs/spec/salida-y-terminal.md#notas-y-avisos prints it: the field's own
// name, in the singular when it replaced exactly one.
func pluralItems(count int, field string) string {
	if count != 1 {
		return field
	}
	switch field {
	case "dependencies":
		return "dependency"
	}
	return strings.TrimSuffix(field, "s")
}

// appendProse is the one rule every prose field shares: adding to an empty
// field is the same as setting it, and adding over existing content leaves a
// blank line in between, one paragraph per repetition of the flag
// (docs/spec/familias-de-flags.md#campos-de-prosa).
func appendProse(field *string, value string) {
	if *field == "" {
		*field = value
		return
	}
	*field += "\n\n" + value
}

// parseDueDate reads the one calendar day of the model. It is a day and not
// an instant, so its layout is YYYY-MM-DD and the message says so.
func parseDueDate(value string) (time.Time, *model.Error) {
	return ParseCalendarDay("due", value)
}

// ParseCalendarDay reads a YYYY-MM-DD written after a flag, and answers the
// same message for every flag that takes one: --due when a task is written
// and --due-before when a listing is filtered
// (docs/spec/cmd/ls.md#códigos-de-salida puts a malformed date of that
// filter among the usage errors, exactly like --due).
func ParseCalendarDay(flag, value string) (time.Time, *model.Error) {
	day, err := time.ParseInLocation(model.DateLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, &model.Error{
			ExitCode: 2,
			Code:     "invalid_date",
			Message:  fmt.Sprintf("--%s: invalid date: %q", flag, value),
			Hints:    []string{"a due date is written YYYY-MM-DD"},
			Field:    flag,
			Given:    value,
		}
	}
	return day, nil
}

// removeCriteria is --rm-ac, at step 3. Its selector is resolved against the
// list as it stands when the step runs, which is after --clear-acs and
// before any --add-ac, so a criterion the same call creates is never a
// target of it.
func (w *writer) removeCriteria(t *model.Task, c Change) error {
	sel, err := w.selectorOf(c, criteriaList)
	if err != nil {
		return err
	}
	keys, warning, selErr := sel.resolve(c.Flag, t.ID, criteriaList, criteriaElements(t))
	if selErr != nil {
		return selErr
	}
	if warning != nil {
		w.warn(*warning)
	}
	for _, key := range keys {
		t.RemoveCriterion(key)
	}
	return nil
}

// applyChecks is step 7, --check-ac and --uncheck-ac. The two are resolved
// together, over the list as it stands after the additions of step 4, which
// is what makes `--clear-acs --add-ac "A" --check-ac all` check the
// criterion the same call just created
// (docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura).
func (w *writer) applyChecks(t *model.Task) error {
	checked, err := w.resolveCriteriaFlag(t, "check-ac")
	if err != nil {
		return err
	}
	unchecked, err := w.resolveCriteriaFlag(t, "uncheck-ac")
	if err != nil {
		return err
	}
	// The overlap is judged over the sets already resolved and not over the
	// text of the selectors, so --check-ac all --uncheck-ac 3 is the same
	// error as --check-ac 3 --uncheck-ac 3.
	for _, key := range unchecked {
		if containsKey(checked, key) {
			return &model.Error{
				ExitCode: 2,
				Code:     "criterion_selector_overlap",
				Message: fmt.Sprintf(
					"--check-ac and --uncheck-ac both select acceptance criterion #%d of %s",
					key, t.ID),
			}
		}
	}
	for _, key := range checked {
		if c := t.Criterion(key); c != nil {
			c.Checked = true
		}
	}
	for _, key := range unchecked {
		if c := t.Criterion(key); c != nil {
			c.Checked = false
		}
	}
	return nil
}

func (w *writer) resolveCriteriaFlag(t *model.Task, flag string) ([]int, error) {
	var keys []int
	for _, c := range w.ofStep(StepCheckAC) {
		if c.Flag != flag {
			continue
		}
		sel, err := w.selectorOf(c, criteriaList)
		if err != nil {
			return nil, err
		}
		found, warning, selErr := sel.resolve(flag, t.ID, criteriaList, criteriaElements(t))
		if selErr != nil {
			return nil, selErr
		}
		if warning != nil {
			w.warn(*warning)
		}
		for _, key := range found {
			if !containsKey(keys, key) {
				keys = append(keys, key)
			}
		}
	}
	return keys, nil
}

// commentSelectors resolves --rm-comment and --set-comment-date against the
// comments the task had before the write, and answers the conflict between
// them before either has been applied.
func (w *writer) commentSelectors(t *model.Task) ([]int, map[int]time.Time, error) {
	elements := commentElements(t)
	var removed []int
	for _, c := range w.ofStep(StepRemove) {
		if c.Flag != "rm-comment" {
			continue
		}
		sel, err := w.selectorOf(c, commentsList)
		if err != nil {
			return nil, nil, err
		}
		keys, warning, selErr := sel.resolve(c.Flag, t.ID, commentsList, elements)
		if selErr != nil {
			return nil, nil, selErr
		}
		if warning != nil {
			w.warn(*warning)
		}
		for _, key := range keys {
			if !containsKey(removed, key) {
				removed = append(removed, key)
			}
		}
	}

	dates := map[int]time.Time{}
	for _, c := range w.ofStep(StepCommentDate) {
		instant, err := parseInstantValue(c.Value)
		if err != nil {
			return nil, nil, err
		}
		sel, selErr := w.selectorOf(Change{Flag: c.Flag, Value: c.Key}, commentsList)
		if selErr != nil {
			return nil, nil, selErr
		}
		keys, warning, resolveErr := sel.resolve(c.Flag, t.ID, commentsList, elements)
		if resolveErr != nil {
			return nil, nil, resolveErr
		}
		if warning != nil {
			w.warn(*warning)
		}
		for _, key := range keys {
			dates[key] = instant
		}
	}
	for key := range dates {
		if containsKey(removed, key) {
			return nil, nil, &model.Error{
				ExitCode: 2,
				Code:     "comment_selector_overlap",
				Message: fmt.Sprintf(
					"--rm-comment and --set-comment-date both select comment #%d of %s",
					key, t.ID),
			}
		}
	}
	return removed, dates, nil
}

// selectorOf reads one selector value and enforces the rule that the
// selector of one task does not have to mean the same in another: with more
// than one reference, only `all` is admitted.
func (w *writer) selectorOf(c Change, kind listKind) (selector, *model.Error) {
	sel, err := parseSelector(c.Flag, c.Value)
	if err != nil {
		return selector{}, err
	}
	if w.manyTasks && !sel.isAll() {
		return selector{}, &model.Error{
			ExitCode: 2,
			Code:     "key_selector_with_many_tasks",
			Message: fmt.Sprintf(
				"--%s: with several tasks the selector has to be all, and this one is %q",
				c.Flag, c.Value),
			Hints: []string{fmt.Sprintf(
				"the keys of the %s of one task do not name the same thing in another",
				kind.plural)},
			Field: c.Flag,
			Given: c.Value,
		}
	}
	return sel, nil
}

func removeComments(t *model.Task, keys []int) {
	for _, key := range keys {
		t.RemoveComment(key)
	}
}

func (w *writer) applyCommentDates(t *model.Task, dates map[int]time.Time) {
	for key, instant := range dates {
		if c := t.Comment(key); c != nil {
			c.CreatedAt = instant
		}
	}
}

// applyComments is step 9, the last one: a comment the call adds is never a
// target of --rm-comment or --set-comment-date, because those two resolved
// their selectors against the list of before.
func (w *writer) applyComments(t *model.Task) error {
	if w.leading != nil {
		// The two comments of `biso answer` go in front of every
		// --comment of the same call, even one written before the
		// answer on the command line
		// (docs/spec/cmd/verbos-del-ciclo.md#biso-answer).
		if err := w.leading(t); err != nil {
			return err
		}
	}
	added := w.ofStep(StepComment)
	if len(added) == 0 {
		return nil
	}
	author := w.commentAuthor()
	if author == "" {
		return &model.Error{
			ExitCode: 2,
			Code:     "missing_identity",
			Message:  "--comment-author is required, no identity is configured",
			Field:    "comment-author",
		}
	}
	for _, c := range added {
		t.AddComment(author, w.now, c.Value)
	}
	return nil
}

// commentAuthor is --comment-author when it was written, and the caller's
// identity otherwise (docs/spec/invocacion.md#variables-de-entorno).
func (w *writer) commentAuthor() string {
	for _, c := range w.changes {
		if c.Flag == "comment-author" {
			return c.Value
		}
	}
	return w.env.Me
}

// parseInstantValue reads the UTC instant --set-comment-date takes, which is
// a full instant and not a bare day, because a comment's createdAt is an
// instant (docs/spec/familias-de-flags.md#comentarios).
func parseInstantValue(value string) (time.Time, *model.Error) {
	instant, err := time.ParseInLocation(model.InstantLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, &model.Error{
			ExitCode: 2,
			Code:     "invalid_date",
			Message:  fmt.Sprintf("--set-comment-date: invalid instant: %q", value),
			Hints:    []string{"an instant is written YYYY-MM-DDTHH:MM:SSZ, in UTC"},
			Field:    "set-comment-date",
			Given:    value,
		}
	}
	return instant, nil
}

// containsKey answers whether a resolved set of keys already holds one.
func containsKey(keys []int, key int) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
