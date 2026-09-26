package ops

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	"biso/internal/model"
)

// This file holds what a write has to check and to settle once every field
// flag has been applied: the two cycles a reference can close, the lease of
// docs/spec/lease.md, the warnings of arriving at a terminal status, which
// fields really changed, and the three derived data of the status line.

// board is the board as this call will leave it: every task it reads, with
// the ones this very call is writing already carrying their new values, so
// that a cycle or an urgency is judged against the result and not against
// what was there a moment ago.
func (w *writer) boardAfter(pending []*model.Task) (map[string]*model.Task, error) {
	all, err := w.tasks()
	if err != nil {
		return nil, err
	}
	byID := make(map[string]*model.Task, len(all)+len(pending))
	for _, t := range all {
		byID[t.ID] = t
	}
	for _, t := range pending {
		if t.ID != "" {
			byID[t.ID] = t
		}
	}
	return byID, nil
}

// checkGraph answers the two cycles of docs/spec/cmd/new.md and
// docs/spec/cmd/set.md, and the one dependency that is the task itself.
//
// A task that does not exist yet cannot be in either cycle: it carries no
// identifier, so nothing can point at it. The checks run over it all the
// same and simply find nothing, which is what keeps one implementation for
// the two commands.
func (w *writer) checkGraph(t *model.Task, byID map[string]*model.Task) *model.Error {
	if t.ID != "" {
		for _, dep := range t.Dependencies {
			if dep == t.ID {
				return &model.Error{
					ExitCode: 2,
					Code:     "self_dependency",
					Message:  fmt.Sprintf("%s cannot depend on itself", t.ID),
					Field:    "dependencies",
					Given:    dep,
				}
			}
		}
		for _, dep := range t.Dependencies {
			if path := reaches(byID, dep, t.ID, func(x *model.Task) []string {
				return x.Dependencies
			}); path != nil {
				return &model.Error{
					ExitCode: 2,
					Code:     "dependency_cycle",
					Message: fmt.Sprintf("--add-deps would close a dependency cycle: %s",
						strings.Join(append([]string{t.ID}, path...), " -> ")),
					Field: "dependencies",
					Given: dep,
				}
			}
		}
	}
	if t.Parent == "" {
		return nil
	}
	if t.Parent == t.ID {
		return &model.Error{
			ExitCode: 2,
			Code:     "parent_cycle",
			Message:  fmt.Sprintf("%s cannot be its own parent", t.ID),
			Field:    "parent",
			Given:    t.Parent,
		}
	}
	if t.ID == "" {
		return nil
	}
	if path := reaches(byID, t.Parent, t.ID, func(x *model.Task) []string {
		if x.Parent == "" {
			return nil
		}
		return []string{x.Parent}
	}); path != nil {
		return &model.Error{
			ExitCode: 2,
			Code:     "parent_cycle",
			Message: fmt.Sprintf("--parent would close a parent cycle: %s",
				strings.Join(append([]string{t.ID}, path...), " -> ")),
			Field: "parent",
			Given: t.Parent,
		}
	}
	return nil
}

// reaches walks the graph edges gives it and answers the path from `from` to
// `target`, or nil when there is none.
func reaches(byID map[string]*model.Task, from, target string, edges func(*model.Task) []string) []string {
	seen := map[string]bool{}
	var walk func(id string) []string
	walk = func(id string) []string {
		if id == target {
			return []string{id}
		}
		if seen[id] {
			return nil
		}
		seen[id] = true
		t, ok := byID[id]
		if !ok {
			return nil
		}
		for _, next := range edges(t) {
			if path := walk(next); path != nil {
				return append([]string{id}, path...)
			}
		}
		return nil
	}
	return walk(from)
}

// settleLease is docs/spec/lease.md applied to one write, in the order that
// page fixes: the invariant of "El vaciado" wins over the renewal of "La
// renovación", and the warning of somebody else's lease is emitted either
// way.
func (w *writer) settleLease(t *model.Task) {
	alive := !t.LeaseExpiresAt.IsZero() && t.LeaseExpiresAt.After(w.now)
	somebodyElses := t.LeaseHolder != "" && t.LeaseHolder != w.env.Me
	if alive && somebodyElses {
		w.warn(Warning{
			Code: "lease_held",
			Message: fmt.Sprintf("%s's lease is held by %s until %s",
				t.ID, t.LeaseHolder, t.LeaseExpiresAt.UTC().Format(model.InstantLayout)),
			Fields: map[string]any{
				"task":   t.ID,
				"holder": t.LeaseHolder,
				"until":  t.LeaseExpiresAt.UTC().Format(model.InstantLayout),
			},
		})
	}

	if !w.leaseInvariantHolds(t) {
		// The fields only have a value on a task that is at once active,
		// assigned and not archived, and losing any of the three empties
		// them in this very write, whoever made it.
		t.LeaseExpiresAt, t.LeaseHolder = time.Time{}, ""
		return
	}
	if t.LeaseHolder != "" && t.LeaseHolder == w.env.Me {
		// Any write by the holder over their own task is a heartbeat,
		// including one that changed no field at all.
		t.LeaseExpiresAt = w.now.Add(time.Duration(w.b.Config.LeaseMinutes) * time.Minute)
	}
}

func (w *writer) leaseInvariantHolds(t *model.Task) bool {
	return t.Status == w.b.Config.ActiveStatus && len(t.Assignees) > 0 && !t.Archived
}

// claimLease takes the lease for whoever is calling, which is what `biso
// start` does and what `biso new --start` does at creation. It is the only
// way leaseHolder is ever set outside an import.
func (w *writer) claimLease(t *model.Task) {
	if w.env.Me == "" || !w.leaseInvariantHolds(t) {
		return
	}
	t.LeaseHolder = w.env.Me
	t.LeaseExpiresAt = w.now.Add(time.Duration(w.b.Config.LeaseMinutes) * time.Minute)
}

// warnAboutTerminal is the three warnings of arriving at the terminal
// status. They speak of arriving, so a write over a task that was already
// there does not repeat them. `biso finish` is the one command that does not
// go through here: it asks the same three questions of the task as this
// write leaves it, whether it arrived now or was already closed
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
func (w *writer) warnAboutTerminal(t *model.Task, previousStatus string) {
	terminal := w.b.Config.TerminalStatus
	if t.Status != terminal || previousStatus == terminal {
		return
	}
	for _, warning := range []*Warning{
		acUncheckedWarning(t, terminal),
		noSummaryWarning(t),
		openQuestionOnTerminalWarning(t, terminal),
	} {
		if warning != nil {
			w.warn(*warning)
		}
	}
}

// The three warnings above, one function each, so that `biso finish` can ask
// for the same three in its own order, turn them into the error 6 of
// --strict or drop them all with --no-checks, with no second implementation
// of any of them.
//
// status is the one the message names, which is the status the write leaves
// the task in. It is always the board's terminal one, in both callers,
// because these are the warnings of arriving there: a `biso finish --status` that
// names another status closes nothing and asks none of them
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).

func acUncheckedWarning(t *model.Task, status string) *Warning {
	unchecked := t.AcTotal() - t.AcDone()
	if unchecked == 0 {
		return nil
	}
	warning := &Warning{
		Code: "terminal_ac_unchecked",
		Message: fmt.Sprintf("%s moved to %s with %d of %d acceptance criteria unchecked",
			t.ID, status, unchecked, t.AcTotal()),
		Fields: map[string]any{
			"task": t.ID, "unchecked": unchecked, "total": t.AcTotal(),
		},
	}
	for _, c := range t.AcceptanceCriteria {
		if !c.Checked {
			warning.Detail = append(warning.Detail, fmt.Sprintf("  #%d %s", c.Key, c.Text))
		}
	}
	return warning
}

func noSummaryWarning(t *model.Task) *Warning {
	if t.Summary != "" {
		return nil
	}
	return &Warning{
		Code:    "terminal_no_summary",
		Message: fmt.Sprintf("%s finished without a final summary", t.ID),
		Fields:  map[string]any{"task": t.ID},
	}
}

func openQuestionOnTerminalWarning(t *model.Task, status string) *Warning {
	if t.Question == nil {
		return nil
	}
	return &Warning{
		Code: "open_question_on_terminal",
		Message: fmt.Sprintf("%s moved to %s with an open question, asked by %s",
			t.ID, status, t.Question.Author),
		Fields: map[string]any{"task": t.ID, "author": t.Question.Author},
	}
}

// unfinishedSubtasks is the warning of closing a parent whose
// children are not closed. An archived subtask that never reached the
// terminal status is still unfinished, and it is marked so that a caller
// does not mistake it for a live one
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
func (w *writer) unfinishedSubtasks(t *model.Task, byID map[string]*model.Task) *Warning {
	var listed, plain []string
	for _, id := range sortedIDs(byID) {
		child := byID[id]
		if child.Parent != t.ID || child.Status == w.b.Config.TerminalStatus {
			continue
		}
		plain = append(plain, child.ID)
		if child.Archived {
			listed = append(listed, child.ID+" (archived)")
			continue
		}
		listed = append(listed, child.ID)
	}
	if len(listed) == 0 {
		return nil
	}
	return &Warning{
		Code: "unfinished_subtasks",
		Message: fmt.Sprintf("%s has unfinished subtasks: %s",
			t.ID, strings.Join(listed, ", ")),
		Fields: map[string]any{"task": t.ID, "subtasks": plain},
	}
}

// sortedIDs is the identifiers of a board, in ascending numeric order,
// which is the order every list of tasks of a message is written in: the
// order of the number and not of the text, so MYP-9 comes before MYP-10.
func sortedIDs(byID map[string]*model.Task) []string {
	ids := make([]string, 0, len(byID))
	for id := range byID {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return taskNumber(ids[i]) < taskNumber(ids[j]) })
	return ids
}

// warnAboutOpenQuestionOnStart is the warning of taking a task that is
// waiting for somebody's answer. It warns and never refuses, for the same
// reason unresolved dependencies do
// (docs/spec/cmd/verbos-del-ciclo.md#biso-start).
func (w *writer) warnAboutOpenQuestionOnStart(t *model.Task) {
	if t.Question == nil {
		return
	}
	w.warn(Warning{
		Code: "open_question_on_start",
		Message: fmt.Sprintf("%s has an open question, asked by %s",
			t.ID, t.Question.Author),
		Fields: map[string]any{"task": t.ID, "author": t.Question.Author},
	})
}

// warnAboutUnresolvedDependencies is the warning of starting a blocked task,
// which `biso new --start` owes for the same reason `biso start` does: the
// two have to leave the same task behind.
func (w *writer) warnAboutUnresolvedDependencies(t *model.Task, byID map[string]*model.Task) {
	var unresolved []string
	for _, dep := range t.Dependencies {
		other, ok := byID[dep]
		if !ok || other.Archived || other.Status == w.b.Config.TerminalStatus {
			continue
		}
		unresolved = append(unresolved, fmt.Sprintf("%s (%s)", other.ID, other.Status))
	}
	if len(unresolved) == 0 {
		return
	}
	w.warn(Warning{
		Code: "unresolved_dependencies",
		Message: fmt.Sprintf("%s has unresolved dependencies: %s",
			t.ID, strings.Join(unresolved, ", ")),
		Fields: map[string]any{"task": t.ID, "dependencies": unresolved},
	})
}

// changedFields is the `changed` key of the JSON schema of
// docs/spec/cmd/set.md: which fields really changed, which is not the same
// as which flags were written. The two lease fields are not among them,
// because a renewal on its own is not a change to the task
// (docs/spec/modelo-de-datos/index.md and docs/spec/lease.md#la-renovación).
func changedFields(before, after *model.Task) []string {
	fields := []struct {
		name string
		of   func(*model.Task) any
	}{
		{"title", func(t *model.Task) any { return t.Title }},
		{"status", func(t *model.Task) any { return t.Status }},
		{"type", func(t *model.Task) any { return t.Type }},
		{"priority", func(t *model.Task) any { return t.Priority }},
		{"parent", func(t *model.Task) any { return t.Parent }},
		{"assignees", func(t *model.Task) any { return t.Assignees }},
		{"author", func(t *model.Task) any { return t.Author }},
		{"labels", func(t *model.Task) any { return t.Labels }},
		{"dependencies", func(t *model.Task) any { return t.Dependencies }},
		{"references", func(t *model.Task) any { return t.References }},
		{"due", func(t *model.Task) any { return t.Due }},
		{"ordinal", func(t *model.Task) any { return t.Ordinal }},
		{"description", func(t *model.Task) any { return t.Description }},
		{"plan", func(t *model.Task) any { return t.Plan }},
		{"notes", func(t *model.Task) any { return t.Notes }},
		{"summary", func(t *model.Task) any { return t.Summary }},
		{"acceptanceCriteria", func(t *model.Task) any { return t.AcceptanceCriteria }},
		{"comments", func(t *model.Task) any { return t.Comments }},
		{"question", func(t *model.Task) any { return t.Question }},
		{"archived", func(t *model.Task) any { return t.Archived }},
	}
	changed := []string{}
	for _, f := range fields {
		if !sameValue(f.of(before), f.of(after)) {
			changed = append(changed, f.name)
		}
	}
	return changed
}

// sameValue compares two field values the way the caller sees them: an empty
// list and a list that was never there are the same field with no value, so
// they are equal (docs/spec/contrato-json.md#números-fechas-y-ausencias).
func sameValue(a, b any) bool {
	if isEmptySlice(a) && isEmptySlice(b) {
		return true
	}
	if isEmptyMap(a) && isEmptyMap(b) {
		return true
	}
	return reflect.DeepEqual(a, b)
}

func isEmptySlice(v any) bool {
	value := reflect.ValueOf(v)
	return value.Kind() == reflect.Slice && value.Len() == 0
}

func isEmptyMap(v any) bool {
	value := reflect.ValueOf(v)
	return value.Kind() == reflect.Map && value.Len() == 0
}

// summarize is the status line of docs/spec/cmd/set.md#salida as data: the
// resulting status, the progress of the criteria and the urgency
// recalculated, which are the three things the caller could not know without
// reading the task.
func (w *writer) summarize(t *model.Task, changed []string, byID map[string]*model.Task) (TaskWrite, error) {
	urgency, err := t.Urgency(model.UrgencyContext{
		Coefficients:   w.b.Config.Urgency,
		Priorities:     w.b.Config.Priorities,
		ActiveStatus:   w.b.Config.ActiveStatus,
		TerminalStatus: w.b.Config.TerminalStatus,
		Blocking:       w.blocking(t, byID),
		Blocked:        w.blocked(t, byID),
		Today:          w.today(),
	})
	if err != nil {
		return TaskWrite{}, err
	}
	added := w.acAdded
	if added == nil {
		added = []int{}
	}
	return TaskWrite{
		ID:       t.ID,
		Status:   t.Status,
		AcDone:   t.AcDone(),
		AcTotal:  t.AcTotal(),
		Urgency:  urgency,
		Changed:  changed,
		AcAdded:  added,
		Archived: t.Archived,
	}, nil
}

// unfinished is what both urgency terms mean by "sin terminar": neither in
// the terminal status nor archived, so archiving a task that never finished
// unblocks whoever depended on it in the same read
// (docs/spec/modelo-de-datos/urgencia.md).
func (w *writer) unfinished(t *model.Task) bool {
	return !t.Archived && t.Status != w.b.Config.TerminalStatus
}

func (w *writer) blocking(t *model.Task, byID map[string]*model.Task) bool {
	if t.ID == "" {
		return false
	}
	for _, other := range byID {
		if other.ID == t.ID || !w.unfinished(other) {
			continue
		}
		for _, dep := range other.Dependencies {
			if dep == t.ID {
				return true
			}
		}
	}
	return false
}

func (w *writer) blocked(t *model.Task, byID map[string]*model.Task) bool {
	for _, dep := range t.Dependencies {
		if other, ok := byID[dep]; ok && w.unfinished(other) {
			return true
		}
	}
	return false
}

// views is the cards of the tasks a write affected, which is what --print
// prints (docs/spec/cmd/flags-globales.md). They are built over the board
// as this call leaves it, so the card shows what was just written and not
// what was there a moment ago.
func (w *writer) views(tasks []*model.Task, byID map[string]*model.Task) ([]TaskView, error) {
	r := newReader(w.b, w.env)
	r.byID = byID
	r.all = make([]*model.Task, 0, len(byID))
	for _, t := range byID {
		r.all = append(r.all, t)
	}
	sort.Slice(r.all, func(i, j int) bool {
		return taskNumber(r.all[i].ID) < taskNumber(r.all[j].ID)
	})

	out := make([]TaskView, 0, len(tasks))
	for _, t := range tasks {
		v, err := r.view(t, false)
		if err != nil {
			return nil, err
		}
		// The card --print writes is the whole card of `biso get`
		// (docs/spec/cmd/flags-globales.md), which always carries `blocked
		// by`/`unblocks`.
		v = r.withClosure(v, false)
		out = append(out, v)
	}
	return out, nil
}
