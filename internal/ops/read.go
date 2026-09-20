package ops

import (
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
func (r *reader) load() error {
	if r.byID != nil {
		return nil
	}
	all, skipped, err := r.b.Tasks.All()
	if err != nil {
		return err
	}
	r.all = all
	r.byID = make(map[string]*model.Task, len(all))
	for _, t := range all {
		r.byID[t.ID] = t
	}
	r.skipped = append(r.skipped, skipped...)
	return nil
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
