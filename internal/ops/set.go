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

// SetOn is Set over a board that is already open.
func SetOn(b *board.Board, env Env, p SetParams) (*WriteResult, error) {
	if len(p.Refs) == 0 {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "missing_ref",
			Message:  "biso set needs at least one task reference",
			Hints:    []string{"biso set MYP-11 --priority high"},
		}
	}
	if len(p.Changes) == 0 {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "nothing_to_change",
			Message:  "nothing to change",
			Hints:    []string{"`biso get " + p.Refs[0] + "` shows what the task has now"},
		}
	}

	w := newWriter(b, env, p.Changes)
	w.manyTasks = len(p.Refs) > 1

	tasks, err := w.resolveAll(p.Refs, p.Mode)
	if err != nil {
		return w.partial(), err
	}

	byID, err := w.boardAfter(tasks)
	if err != nil {
		return w.partial(), err
	}

	result := &WriteResult{DryRun: p.DryRun}
	for _, t := range tasks {
		before := cloneTask(t)
		if err := w.apply(t); err != nil {
			return w.partial(), err
		}
		if err := w.checkGraph(t, byID); err != nil {
			return w.partial(), err
		}
		changed := changedFields(before, t)
		if len(changed) > 0 {
			t.UpdatedAt = w.now
		} else {
			// The note speaks of the fields of the task, and none of them
			// changed; the lease is settled below and renews all the same
			// (docs/spec/lease.md#la-renovación).
			w.note(t.ID + " unchanged")
		}
		w.warnAboutTerminal(t, before.Status)
		w.settleLease(t)
		if t.Archived {
			w.note(t.ID + " is archived")
		}
		summary, err := w.summarize(t, changed, byID)
		if err != nil {
			return w.partial(), err
		}
		result.Tasks = append(result.Tasks, summary)
	}

	if !p.DryRun {
		// One transaction for every task the call named, which is what
		// makes the all-or-nothing of
		// docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables
		// observable and not just intended.
		if err := b.Tasks.SaveAll(tasks); err != nil {
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
func (w *writer) resolveAll(refs []string, mode RefMode) ([]*model.Task, error) {
	var tasks []*model.Task
	seen := map[string]bool{}
	for _, ref := range refs {
		all, readErr := w.tasks()
		if readErr != nil {
			return nil, readErr
		}
		resolved, err := resolveRefWith(w.b, all, ref, mode)
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
