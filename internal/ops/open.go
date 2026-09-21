package ops

import (
	"biso/internal/board"
	"biso/internal/model"
)

// openBoard resolves the board of a call and opens it.
//
// Section 3.5 of
// docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md
// gives every command of this package the signature
// `func(b *board.Board, p XParams)`, and that is the shape of the function
// that holds each command's logic here. What that section does not say is
// who opens the board, and it cannot be internal/cli: the dependency rule
// keeps internal/board two layers below it, so naming that type up there is
// exactly what the layering test forbids. So each command has a thin
// exported entry point that opens the board, runs the logic and closes it,
// and the logic itself takes the open board as its first argument and can be
// driven by a test without any of this.
func openBoard(env Env) (*board.Board, error) {
	loc, _, err := board.Resolve(board.Search{Dir: env.Dir, Machine: env.Machine})
	if err != nil {
		return nil, err
	}
	return board.Open(loc, env.Machine)
}

// cloneTask is a task copied deeply enough that writing on one of the two
// never reaches the other. It is what lets a write compare what it left
// behind with what it found, which is the `changed` key of the JSON schema
// of docs/spec/cmd/set.md.
func cloneTask(t *model.Task) *model.Task {
	clone := *t
	clone.Assignees = append([]string(nil), t.Assignees...)
	clone.Labels = append([]string(nil), t.Labels...)
	clone.Dependencies = append([]string(nil), t.Dependencies...)
	clone.References = append([]string(nil), t.References...)
	clone.AcceptanceCriteria = append([]model.Criterion(nil), t.AcceptanceCriteria...)
	clone.Comments = append([]model.Comment(nil), t.Comments...)
	if t.Ordinal != nil {
		ordinal := *t.Ordinal
		clone.Ordinal = &ordinal
	}
	if t.Ext != nil {
		clone.Ext = make(map[string]string, len(t.Ext))
		for k, v := range t.Ext {
			clone.Ext[k] = v
		}
	}
	if t.Question != nil {
		question := *t.Question
		clone.Question = &question
	}
	return &clone
}
