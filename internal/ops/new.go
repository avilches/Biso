package ops

import (
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// NewParams is one `biso new` call, already read off the command line.
type NewParams struct {
	// Title is the positional argument, and HasTitle says whether the call
	// wrote one at all, which a call with none did not.
	Title    string
	HasTitle bool
	// Start creates the task already in the active status, assigned to the
	// caller and with the lease claimed, exactly as `biso start` over it
	// would leave it (docs/spec/cmd/new.md).
	Start   bool
	Changes []Change
	DryRun  bool
}

// New creates one task and answers its identifier (docs/spec/cmd/new.md).
//
// The batch of --from is not this function: it is the same engine fed from
// NDJSON instead of from flags, and it lives with `biso export`, whose format
// it has to stay symmetrical with.
func New(env Env, p NewParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return NewOn(b, env, p)
}

// NewOn is New over a board that is already open.
func NewOn(b *board.Board, env Env, p NewParams) (*WriteResult, error) {
	if strings.TrimSpace(p.Title) == "" {
		// A call with no title at all and one with a title of nothing but
		// spaces are the same mistake, and the specification gives them
		// one message (docs/spec/valores-de-entrada.md#el-valor-vacío).
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "missing_title",
			Message:  "title cannot be empty",
		}
	}

	w := newWriter(b, env, p.Changes)
	w.newTask = true

	task := &model.Task{
		Title:  p.Title,
		Status: b.Config.InitialStatus,
		// The author is the caller's identity unless --author says
		// otherwise, and a call with no identity configured creates the
		// task without one, with no warning
		// (docs/spec/modelo-de-datos/autor.md).
		Author:    env.Me,
		CreatedAt: w.now,
		UpdatedAt: w.now,
	}
	if err := w.apply(task); err != nil {
		return w.partial(), err
	}
	if p.Start {
		w.start(task)
	}

	byID, err := w.boardAfter(nil)
	if err != nil {
		return w.partial(), err
	}
	if err := w.checkGraph(task, byID); err != nil {
		return w.partial(), err
	}
	if p.Start {
		w.warnAboutUnresolvedDependencies(task, byID)
	}
	// A task created straight into the terminal status has arrived at it,
	// so it earns the same warnings any other arrival does.
	w.warnAboutTerminal(task, "")

	result := &WriteResult{DryRun: p.DryRun, Created: true}
	if p.DryRun {
		// Nothing was written, so no identifier was spent and there is
		// none to name the task by: the answer is the count and the
		// warnings the real call would have produced
		// (docs/spec/cmd/new.md#--dry-run-sobre-una-sola-tarea).
		if err := task.Validate(b.Config.Extensions); err != nil {
			return w.partial(), err
		}
		result.Warnings, result.Notes = w.warnings, w.notes
		return result, nil
	}

	if err := b.Tasks.Create(task); err != nil {
		return w.partial(), err
	}
	summary, err := w.summarize(task, changedFields(&model.Task{}, task), byID)
	if err != nil {
		return w.partial(), err
	}
	result.Tasks = append(result.Tasks, summary)
	result.Warnings, result.Notes = w.warnings, w.notes
	return result, nil
}

// start is what --start adds on top of the creation, and it has to leave the
// same task `biso new "X"` followed by `biso start` would leave, or the
// shortcut would not be one (docs/spec/cmd/new.md).
func (w *writer) start(t *model.Task) {
	t.Status = w.b.Config.ActiveStatus
	if len(t.Assignees) == 0 {
		if w.env.Me == "" {
			// No identity to attribute it to, so the task is left
			// unassigned and, by the invariant of docs/spec/lease.md,
			// without a lease either.
			w.note("no identity configured, task left unassigned")
		} else {
			t.Assignees = append(t.Assignees, w.env.Me)
		}
	}
	// Whoever writes is who takes the lease, not whoever is in assignees:
	// `biso new "X" --start -a @sara` leaves the task assigned to @sara
	// with the lease of the caller.
	w.claimLease(t)
}
