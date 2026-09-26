package ops

import (
	"biso/internal/board"
	"biso/internal/model"
)

// SetParams is one `biso set` call, already read off the command line.
type SetParams struct {
	// Refs are the references the call named, in the order they were
	// written, at least one (docs/spec/cmd/set.md).
	Refs []string
	// Mode is the interpretation --id or --match forced on them.
	Mode RefMode
	// Changes are the field flags, in any order: the engine sorts them
	// into the nine steps of
	// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura.
	Changes []Change
	DryRun  bool
	// Print asks for the card of every task the call affected
	// (docs/spec/cmd/flags-globales.md).
	Print bool
	// Warnings are the ones the caller had already earned before the
	// write began, which the six verbs of the cycle use for the empty
	// positional text of `biso note` and `biso comment`.
	Warnings []Warning
	// AllowNoChanges lets a call through that writes no field. `biso set`
	// never does, because a call of its own with no field flag has nothing
	// to say; a verb of the cycle does, because what it writes may be the
	// question, or nothing at all with a warning.
	AllowNoChanges bool
}

// Set changes any field of one or more tasks, all or nothing
// (docs/spec/cmd/set.md).
func Set(env Env, p SetParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return SetOn(b, env, p)
}

// verb is what the six verbs of the cycle add on top of `biso set`: three
// moments of the one write loop below, and nothing else. None of them opens
// a transaction, resolves a reference or writes a field flag, because all of
// that is already here; what a verb contributes is its refusals, its
// defaults and its own effects
// (docs/spec/cmd/verbos-del-ciclo.md).
type verb struct {
	// name is the command the messages of the shared loop name, empty for
	// `biso set` itself.
	name string
	// before runs over the task exactly as it was read, ahead of every
	// field flag. It is where a verb refuses: `biso start` over an
	// archived task, `biso ask` over one that already has a question.
	before func(w *writer, t *model.Task) error
	// after runs once the nine steps of
	// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura
	// have been applied, and before anything derived from the result is
	// computed: the status this verb sets, the identity it assigns, the
	// question it fills or empties, and the checks `biso finish` reads off
	// the task as this write leaves it.
	after func(w *writer, t, before *model.Task, byID map[string]*model.Task) error
	// settled runs after docs/spec/lease.md has been applied, which is
	// where `biso start` claims: the warning of somebody else's live lease
	// has to be emitted before the lease changes hands.
	settled func(w *writer, t, before *model.Task)
	// terminalWarnings says whether the warnings of arriving at the
	// terminal status are this verb's business. `biso finish` emits its
	// own, because it reads them off the result whether the task was
	// already closed or not, and because --strict turns them into an
	// error and --no-checks silences them.
	ownTerminalWarnings bool
	// ownStateNotes says whether the two notes the loop writes about the
	// task's own state, "unchanged" and "is archived", are this verb's
	// business instead. `biso archive` is the one verb they belong to,
	// because it is the command that asks for that state
	// (docs/spec/cmd/archive.md).
	ownStateNotes bool
	// scope is which half of the board a text reference of this call
	// looks at, and only `biso archive --unarchive` moves it
	// (docs/spec/cmd/archive.md#la-referencia-de---unarchive).
	scope RefScope
}

// SetOn is Set over a board that is already open.
func SetOn(b *board.Board, env Env, p SetParams) (*WriteResult, error) {
	return writeOn(b, env, p, verb{})
}

// writeOn is the one write loop of the program: `biso set` with no verb on
// top, and each of the six verbs of the cycle with its own.
func writeOn(b *board.Board, env Env, p SetParams, v verb) (*WriteResult, error) {
	command := "set"
	if v.name != "" {
		command = v.name
	}
	if len(p.Refs) == 0 {
		// The hint of `biso set` is the one docs/spec/cmd/set.md prints,
		// which names a flag because a call of that command with no flag
		// has nothing to do either; a verb of the cycle needs no flag, so
		// its hint is the call alone.
		hint := "biso " + command + " MYP-11"
		if v.name == "" {
			hint = "biso set MYP-11 --priority high"
		}
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "missing_ref",
			Message:  "biso " + command + " needs at least one task reference",
			Hints:    []string{hint},
		}
	}
	if len(p.Changes) == 0 && !p.AllowNoChanges {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "nothing_to_change",
			Message:  "nothing to change",
			Hints:    []string{"`biso get " + p.Refs[0] + "` shows what the task has now"},
		}
	}

	w := newWriter(b, env, p.Changes)
	w.manyTasks = len(p.Refs) > 1
	w.warnings = append(w.warnings, p.Warnings...)
	if err := w.prepareLabels(); err != nil {
		return w.partial(), err
	}

	tasks, err := w.resolveAll(p.Refs, p.Mode, v.scope)
	if err != nil {
		// The targets are not known yet, so none of them can be excluded,
		// but a reference that fails to resolve is still worth naming the
		// rest of the board's unreadable tasks for
		// (docs/spec/garantias.md#el-primer-caso-una-tarea-ilegible).
		w.warnAboutSkippedExcept(nil)
		return w.partial(), err
	}
	w.warnAboutSkippedExcept(tasks)

	// The manual order is resolved here, before the loop below writes
	// anything: the neighbour is looked for against the board as it was,
	// and the gap is computed once for every task the call moves
	// (docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura).
	if err := w.prepareOrdinal(tasks); err != nil {
		return w.partial(), err
	}

	byID, err := w.boardAfter(tasks)
	if err != nil {
		return w.partial(), err
	}

	result := &WriteResult{DryRun: p.DryRun}
	// Only `biso set`, `biso start` and `biso finish` can ever repair a
	// task's vocabulary, because only they can write status, type or
	// priority (docs/spec/cmd/set.md#comportamiento-caso-a-caso): `biso
	// note`, `biso comment`, `biso ask`, `biso answer` and `biso archive`
	// never repair, whatever they write, so their illegibility is settled
	// before anything else, including a verb's own precondition. A
	// question such as "does it have an open question" reads a status or a
	// priority the board no longer configures, which is exactly what makes
	// the task unreadable in the first place.
	//
	// `biso start` and `biso finish` are not that simple, because they
	// carry their own precondition too (`is archived`, `is already
	// Done`, `is not ready to finish`), and that precondition reads
	// fields of the task as it already is, before this call's status
	// change is applied. Refusing on it first, ahead of the readability
	// check, would let an unrelated fault, say a bad priority, hide
	// behind "is archived" and answer with the wrong exit code. So the
	// two are checked early as well, but excluding the fields this very
	// call is about to write: that is what keeps
	// docs/spec/garantias.md#cómo-se-arregla-una-tarea-ilegible's
	// exception working when the fault the call excludes is the one
	// that made the task illegible, and still catches every fault it
	// does not touch before the verb's own precondition runs.
	for _, t := range tasks {
		switch v.name {
		case "":
			// Nothing here to preempt: the final loop below judges the
			// task as `biso set` leaves it.
		case "start", "finish":
			if err := readabilityErrorExcluding(b.Config, t, w.changes); err != nil {
				return w.partial(), err
			}
		default:
			if err := readabilityError(b.Config, t); err != nil {
				return w.partial(), err
			}
		}
		before := cloneTask(t)
		if v.before != nil {
			if err := v.before(w, t); err != nil {
				return w.partial(), err
			}
		}
		if err := w.apply(t); err != nil {
			return w.partial(), err
		}
		if v.after != nil {
			if err := v.after(w, t, before, byID); err != nil {
				return w.partial(), err
			}
		}
		if err := w.checkGraph(t, byID); err != nil {
			return w.partial(), err
		}
		changed := changedFields(before, t)
		if len(changed) > 0 {
			t.UpdatedAt = w.now
		} else if !v.ownStateNotes {
			// The note speaks of the fields of the task, and none of them
			// changed; the lease is settled below and renews all the same
			// (docs/spec/lease.md#la-renovación).
			w.note(t.ID + " unchanged")
		}
		if !v.ownTerminalWarnings {
			w.warnAboutTerminal(t, before.Status)
		}
		w.settleLease(t)
		if v.settled != nil {
			v.settled(w, t, before)
		}
		if t.Archived && !v.ownStateNotes {
			w.note(t.ID + " is archived")
		}
		summary, err := w.summarize(t, changed, byID)
		if err != nil {
			return w.partial(), err
		}
		result.Tasks = append(result.Tasks, summary)
	}

	// The model's own validation, which the store runs again before it
	// writes, asked here too and in the same order, because a preview that
	// answered 0 where the real write answers 3 would be a preview that
	// lies: --dry-run promises the outcome of the call and not the outcome
	// of everything except the last check
	// (docs/spec/cmd/flags-globales.md).
	//
	// The readability check right after it judges the task as this write
	// would leave it, not as it was: a call that clears the fault, such as
	// `biso set MYP-11 --priority medium`, is applied, and one that does
	// not is error 3 with nothing written, which is the one exception of
	// docs/spec/garantias.md#cómo-se-arregla-una-tarea-ilegible.
	for _, t := range tasks {
		if err := t.Validate(); err != nil {
			return w.partial(), err
		}
		if err := readabilityError(b.Config, t); err != nil {
			return w.partial(), err
		}
	}

	if !p.DryRun {
		// One transaction for every task the call named, which is what
		// makes the all-or-nothing of
		// docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables
		// observable and not just intended.
		if err := b.Tasks.SaveAll(tasks, w.claims...); err != nil {
			return w.partial(), err
		}
	}
	if p.Print {
		if result.Views, err = w.views(tasks, byID); err != nil {
			return w.partial(), err
		}
	}
	result.Warnings, result.Notes = w.warnings, w.notes
	return result, nil
}

// resolveAll turns every reference of the call into the task it names,
// before anything is written: a reference that does not exist stops the call
// with its own exit code and not one task is touched.
//
// The same task named twice is one task: the second mention says nothing the
// first did not, and applying the same changes to it twice would append the
// same comment or the same criterion two times over.
func (w *writer) resolveAll(refs []string, mode RefMode, scope RefScope) ([]*model.Task, error) {
	var tasks []*model.Task
	seen := map[string]bool{}
	for _, ref := range refs {
		all, readErr := w.tasks()
		if readErr != nil {
			return nil, readErr
		}
		resolved, err := resolveRefIn(w.b, all, ref, mode, scope)
		if err != nil {
			// The candidates of an ambiguous reference are printed the
			// way `biso ls` prints a listing, in every command that
			// resolves one (docs/spec/referencias.md).
			return nil, withCandidates(w.b, w.env, all, err)
		}
		if resolved.Note != "" {
			w.note(resolved.Note)
		}
		if seen[resolved.Task.ID] {
			continue
		}
		seen[resolved.Task.ID] = true
		tasks = append(tasks, resolved.Task)
	}
	return tasks, nil
}
