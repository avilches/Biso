package ops

import (
	"fmt"
	"sort"
	"time"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the engine the two reading commands share, the way write.go
// is the engine the writing ones share: it reads the board once, computes
// the derived fields that no task carries by itself, and answers a view of
// a task that is the same for `biso ls`, for `biso get` and for the full
// card `--print` prints after a write.
//
// It is a set read in the sense of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar:
// a task it cannot decode is left out with a warning that names it, and
// never a reason to abort.

// TaskView is one task with its derived fields already computed: the ones
// docs/spec/modelo-de-datos/index.md#los-campos-derivados calls derived and
// that need the rest of the board to answer.
type TaskView struct {
	Task *model.Task

	Urgency float64
	// Breakdown is the formula term by term, and it is only filled when
	// the caller asked for it with --explain-urgency
	// (docs/spec/cmd/get.md).
	Breakdown *model.UrgencyBreakdown

	// Blocks are the tasks that depend on this one, Blocked says some
	// unfinished task blocks this one, and Waiting that it has an open
	// question.
	Blocks  []string
	Blocked bool
	Waiting bool

	LeaseExpired bool
}

// reader is one reading invocation while it is being answered.
type reader struct {
	b   *board.Board
	env Env
	now time.Time

	all  []*model.Task
	byID map[string]*model.Task

	warnings []Warning
	notes    []string
	// skipped are the tasks that could not be read, gathered whatever the
	// reason and wherever it was found, so that one single warning names
	// them all at the end: their identifiers also travel to the caller,
	// because data.skipped of docs/spec/cmd/ls.md#el-esquema-json carries
	// them.
	skipped []board.Skipped
}

func newReader(b *board.Board, env Env) *reader {
	return &reader{b: b, env: env, now: env.Now().UTC().Truncate(time.Second)}
}

func (r *reader) warn(w Warning) { r.warnings = append(r.warnings, w) }

func (r *reader) note(note string) { r.notes = append(r.notes, note) }

// load reads every task of the board once, archived ones included, and
// records the ones that could not be decoded.
//
// A row internal/board could not turn into a task is not the only way one
// ends up unreadable: a status, a type or a priority the board no longer
// configures is the other, and the two are folded into one list here,
// before any command looks at a filter
// (docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible). From this
// point on r.all and r.byID hold only tasks every reading command can show
// as if they were fine.
func (r *reader) load() error {
	if r.byID != nil {
		return nil
	}
	all, skipped, err := r.b.Tasks.All()
	if err != nil {
		return err
	}
	legible, badVocabulary := partitionByReadability(r.b.Config, all)
	r.all = legible
	r.byID = make(map[string]*model.Task, len(legible))
	for _, t := range legible {
		r.byID[t.ID] = t
	}
	r.skipped = append(r.skipped, mergeSkipped(skipped, badVocabulary)...)
	return nil
}

// readabilityError is the vocabulary half of
// docs/spec/garantias.md#qué-se-comprueba: a status, type or priority the
// board's configuration does not declare. The other half, a date that is
// not a date, an unknown list field or a column of the wrong type, is
// already decided while a row becomes a task (internal/board/rows.go),
// because that is the layer that reads the row and knows nothing about a
// board's configuration; this one is the layer that has both the task and
// the configuration.
//
// The comparison is exact and never the matching algorithm's: what a write
// stores is always the configured spelling
// (docs/spec/vocabularios.md#idéntica-al-leer-significa-también-lo-ya-guardado).
// status cannot be empty; type and priority can.
func readabilityError(cfg board.Config, t *model.Task) *model.Error {
	if err := vocabularyReadabilityError(t.ID, "status", t.Status, cfg.Statuses, false); err != nil {
		return err
	}
	if err := vocabularyReadabilityError(t.ID, "type", t.Type, cfg.Types, true); err != nil {
		return err
	}
	return vocabularyReadabilityError(t.ID, "priority", t.Priority, cfg.Priorities, true)
}

// readabilityErrorExcluding is readabilityError with status, type or
// priority left unchecked when this very call is about to write it: a
// value that is wrong right now is not a fault the call needs to answer
// for when the call itself is what replaces it. `biso start` and `biso
// finish` are the two callers, because both carry a precondition of their
// own (being archived, already finished, not ready to finish) that reads
// the task as it is before the write, and that precondition must never
// win over a fault the write does not touch
// (docs/spec/garantias.md#cómo-se-arregla-una-tarea-ilegible).
func readabilityErrorExcluding(cfg board.Config, t *model.Task, changes []Change) *model.Error {
	if !writesFlag(changes, "status") {
		if err := vocabularyReadabilityError(t.ID, "status", t.Status, cfg.Statuses, false); err != nil {
			return err
		}
	}
	if !writesFlag(changes, "type") && !writesFlag(changes, "clear-type") {
		if err := vocabularyReadabilityError(t.ID, "type", t.Type, cfg.Types, true); err != nil {
			return err
		}
	}
	if !writesFlag(changes, "priority") && !writesFlag(changes, "clear-priority") {
		if err := vocabularyReadabilityError(t.ID, "priority", t.Priority, cfg.Priorities, true); err != nil {
			return err
		}
	}
	return nil
}

func vocabularyReadabilityError(id, field, value string, configured []string, emptyOK bool) *model.Error {
	if value == "" && emptyOK {
		return nil
	}
	if containsString(configured, value) {
		return nil
	}
	return &model.Error{
		ExitCode: 3,
		Code:     "undecodable_task",
		Message: fmt.Sprintf("%s cannot be read: its %s is %q, which this board does not configure",
			id, field, value),
		Field: field,
		Given: value,
		Valid: append([]string(nil), configured...),
	}
}

// partitionByReadability splits all into the tasks readabilityError has
// nothing to say about and the ones it does, the latter turned into the
// same board.Skipped shape a row that failed to decode already comes in,
// so both kinds travel the rest of the way through one list.
func partitionByReadability(cfg board.Config, all []*model.Task) ([]*model.Task, []board.Skipped) {
	var legible []*model.Task
	var bad []board.Skipped
	for _, t := range all {
		if err := readabilityError(cfg, t); err != nil {
			bad = append(bad, board.Skipped{ID: t.ID, Reason: err})
			continue
		}
		legible = append(legible, t)
	}
	return legible, bad
}

// mergeSkipped answers base and extra as one list, in the ascending
// identifier order docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible
// names them in, whichever of the two kinds of unreadable task found them.
func mergeSkipped(base, extra []board.Skipped) []board.Skipped {
	if len(base) == 0 {
		return extra
	}
	if len(extra) == 0 {
		return base
	}
	all := make([]board.Skipped, 0, len(base)+len(extra))
	all = append(all, base...)
	all = append(all, extra...)
	sort.Slice(all, func(i, j int) bool {
		a, b := taskNumber(all[i].ID), taskNumber(all[j].ID)
		if a != b {
			return a < b
		}
		return all[i].ID < all[j].ID
	})
	return all
}

// skippedIDs are the identifiers of the tasks that were left out, in the
// order they were found.
func (r *reader) skippedIDs() []string {
	ids := make([]string, 0, len(r.skipped))
	for _, s := range r.skipped {
		ids = append(ids, s.ID)
	}
	return ids
}

// warnAboutSkipped emits the one warning of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// for every task this read left out. It is one warning naming all of them
// and not one per task, which is what that page's text shows.
func (r *reader) warnAboutSkipped() {
	if len(r.skipped) > 0 {
		r.warn(skippedWarning(r.skipped))
	}
}

// today is the UTC calendar day of this call, which is what the urgency
// compares against (docs/spec/modelo-de-datos/urgencia.md).
func (r *reader) today() time.Time {
	y, m, d := r.now.UTC().Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

// unfinished is what both urgency terms and the --blocked filter mean by a
// task that has not finished: neither in the terminal status nor archived,
// so archiving a task that never finished unblocks whoever depended on it
// in the same read (docs/spec/modelo-de-datos/urgencia.md and
// docs/spec/cmd/ls.md).
func (r *reader) unfinished(t *model.Task) bool {
	return !t.Archived && t.Status != r.b.Config.TerminalStatus
}

// blocks are the unfinished tasks that depend on this one, in identifier
// order, which is the order `All` answered them in.
func (r *reader) blocks(t *model.Task) []string {
	var out []string
	for _, other := range r.all {
		if other.ID == t.ID || !r.unfinished(other) {
			continue
		}
		for _, dep := range other.Dependencies {
			if dep == t.ID {
				out = append(out, other.ID)
				break
			}
		}
	}
	return out
}

// blocked says whether some unfinished task blocks this one.
func (r *reader) blocked(t *model.Task) bool {
	for _, dep := range t.Dependencies {
		if other, ok := r.byID[dep]; ok && r.unfinished(other) {
			return true
		}
	}
	return false
}

// view computes a task's derived fields. It answers the error of an
// undecodable task untouched, so that the caller decides between the two
// endings of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar:
// failing with exit code 3 on a targeted read, or skipping it on a set one.
func (r *reader) view(t *model.Task, explain bool) (TaskView, *model.Error) {
	// A task load() already filtered out never reaches here through r.all,
	// but one resolved straight off the store, such as `biso get` by a
	// well-formed identifier, does: this is the check that catches it,
	// docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible applied to
	// one task instead of the whole board.
	if err := readabilityError(r.b.Config, t); err != nil {
		return TaskView{}, err
	}
	blocks := r.blocks(t)
	breakdown, err := t.UrgencyBreakdown(model.UrgencyContext{
		Coefficients:   r.b.Config.Urgency,
		Priorities:     r.b.Config.Priorities,
		ActiveStatus:   r.b.Config.ActiveStatus,
		TerminalStatus: r.b.Config.TerminalStatus,
		Blocking:       len(blocks) > 0,
		Blocked:        r.blocked(t),
		Today:          r.today(),
	})
	if err != nil {
		return TaskView{}, err
	}
	v := TaskView{
		Task:    t,
		Urgency: breakdown.Total,
		Blocks:  blocks,
		Blocked: r.blocked(t),
		Waiting: t.Waiting(),
		LeaseExpired: !t.LeaseExpiresAt.IsZero() &&
			!t.LeaseExpiresAt.After(r.now),
	}
	if explain {
		v.Breakdown = breakdown
	}
	return v, nil
}

// undecodable turns a task the reader could not compute into the skip of a
// set read: it is left out of the answer and named in the warning, never a
// reason to fail (docs/spec/garantias.md).
func (r *reader) undecodable(t *model.Task, err *model.Error) {
	r.skipped = append(r.skipped, board.Skipped{ID: t.ID, Reason: err})
}
