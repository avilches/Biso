package ops

import (
	"fmt"

	"biso/internal/board"
	"biso/internal/model"
)

// WhereResult is what `biso where` answers: the four data that identify the
// board in use, who is calling, the counts, and whatever candidate the
// resolution turned down (docs/spec/cmd/where.md).
//
// The four are separate data because each one changes on its own: renaming
// the board does not move it, moving the directory by hand does not rename
// it, and the identifier never changes at all.
type WhereResult struct {
	ID    string
	Board string
	// Path is always the board directory already resolved and expanded,
	// never the literal text the pointer carried.
	Path string
	// Source names the way that chose it, and with the pointer also the
	// directory that pointer came from.
	Source string
	// Me is the identity of whoever is calling, empty when none is set.
	Me string
	// Prefix is the board's task prefix, which the counts line needs to
	// name the highest identifier ever assigned.
	Prefix    string
	Counts    Counts
	Discarded []Discarded
}

// Counts are the two numbers of the counts row and the highest identifier
// the board has ever handed out, zero when it has handed out none.
type Counts struct {
	NotArchived         int
	Archived            int
	HighestEverAssigned int
}

// Discarded is a candidate the resolution turned down, with the reason in
// the free prose `biso where` prints under it.
type Discarded struct {
	ID     string
	Path   string
	Reason string
}

// Where resolves the board and reads what it takes to describe it.
//
// It does not dodge the unreadable database: saying which board is in use
// means opening it, so a board whose database does not open answers exit
// code 21 here like anywhere else (docs/spec/cmd/where.md).
func Where(env Env) (*WhereResult, error) {
	loc, facts, err := board.Resolve(board.Search{Dir: env.Dir, Machine: env.Machine})
	if err != nil {
		if err.Code == "no_board" {
			return nil, nothingSearchedError(facts)
		}
		return nil, err
	}

	b, openErr := board.Open(loc, env.Machine)
	if openErr != nil {
		return nil, openErr
	}
	defer b.Close()

	counts, countErr := b.Counts()
	if countErr != nil {
		return nil, countErr
	}
	result := &WhereResult{
		ID:     b.Location.ID,
		Board:  b.Config.ProjectName,
		Path:   b.Location.Dir,
		Source: b.Location.Source(),
		Me:     env.Me,
		Prefix: b.Config.TaskPrefix,
		Counts: Counts{
			NotArchived:         counts.NotArchived,
			Archived:            counts.Archived,
			HighestEverAssigned: counts.HighestEverAssigned,
		},
	}
	for _, d := range b.Location.Discarded {
		result.Discarded = append(result.Discarded, Discarded{
			ID: d.ID, Path: d.Dir, Reason: d.Reason,
		})
	}
	return result, nil
}

// nothingSearchedError is `biso where`'s own version of the no-board ending.
// Every other command prints the short one and points here; this one says
// what was looked at and where the walk upward stopped, which is the whole
// reason the command exists (docs/spec/cmd/where.md).
//
// The block is printed verbatim, with its own two columns, because the
// reason of a search is a sentence and does not fit in one column.
func nothingSearchedError(facts board.Searched) *model.Error {
	return &model.Error{
		ExitCode: 20,
		Code:     "no_board",
		Message:  "no board here, and none configured for this project",
		Detail: []string{
			"searched  this directory: not a board",
			fmt.Sprintf("          pointer:        not found between this directory and %s,", facts.Stop),
			"                          which is where the search stops",
		},
		Hints: []string{"`biso init` creates one"},
	}
}
