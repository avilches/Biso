package ops

import (
	"fmt"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is `biso archive` (docs/spec/cmd/archive.md), the eighth and
// last command built on the write engine of write.go. It is `biso set` with
// one field of its own, the one no flag writes, and the two idempotent
// notes of asking for a state a task is already in.
//
// There is no delete command, on purpose. Archiving keeps the task, keeps
// its identifier reserved and keeps its history; the refusal of `biso
// delete` lives in internal/cli, where a command name is read.

// ArchiveParams is one `biso archive` call.
type ArchiveParams struct {
	Refs []string
	Mode RefMode
	// Unarchive puts the tasks back on the board instead of taking them
	// off it, with the status each one had.
	Unarchive bool
	Changes   []Change
	DryRun    bool
	Print     bool
}

// Archive takes one or more tasks off the board, or puts them back.
func Archive(env Env, p ArchiveParams) (*WriteResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return ArchiveOn(b, env, p)
}

// ArchiveOn is Archive over a board that is already open.
func ArchiveOn(b *board.Board, env Env, p ArchiveParams) (*WriteResult, error) {
	return writeOn(b, env, SetParams{
		Refs: p.Refs, Mode: p.Mode, Changes: p.Changes,
		DryRun: p.DryRun, Print: p.Print,
		// The field this verb writes has no flag, so a call with no field
		// flag at all still has something to do.
		AllowNoChanges: true,
	}, verb{
		name: "archive",
		// The two notes about being archived are this verb's own: the
		// generic ones of the loop would say "MYP-11 is archived" right
		// after being asked to archive it, and "MYP-11 unchanged" where
		// the answer is that it was already off the board.
		ownStateNotes: true,
		after: func(w *writer, t, before *model.Task, byID map[string]*model.Task) error {
			if p.Unarchive {
				if !before.Archived {
					// Idempotent, exactly like asking to archive one
					// that is already archived
					// (docs/spec/cmd/archive.md).
					w.note(t.ID + " was not archived")
				}
				// The status the task had is the status it comes back
				// with: nothing here touches it. The lease stays empty,
				// because archiving emptied it and nothing claims one
				// here; whoever wants the task takes it with `biso
				// start`.
				t.Archived = false
				return nil
			}
			if before.Archived {
				w.note(t.ID + " was already archived")
				return nil
			}
			t.Archived = true
			w.warnAboutDependents(t, byID)
			return nil
		},
	})
}

// warnAboutDependents is the warning of taking off the board a task that
// others are still waiting on. It never refuses: archiving unblocks them in
// the same read, and whether that is what the caller meant is the caller's
// business (docs/spec/salida-y-terminal.md#notas-y-avisos).
//
// One warning per dependent, because the message names one and the JSON
// object carries one `dependent`.
func (w *writer) warnAboutDependents(t *model.Task, byID map[string]*model.Task) {
	for _, id := range sortedIDs(byID) {
		other := byID[id]
		if other.ID == t.ID || !w.unfinished(other) {
			continue
		}
		for _, dep := range other.Dependencies {
			if dep != t.ID {
				continue
			}
			w.warn(Warning{
				Code: "dependency_of_unfinished",
				Message: fmt.Sprintf("%s is a dependency of %s, which is not finished",
					t.ID, other.ID),
				Fields: map[string]any{"task": t.ID, "dependent": other.ID},
			})
			break
		}
	}
}
