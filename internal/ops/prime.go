package ops

import (
	"biso/internal/board"
	"biso/internal/model"
)

// This file is the reading half of docs/spec/cmd/prime.md: the board's own
// vocabulary and the four blocks the summary is made of, each one already
// ordered and cut by --limit. Nothing here knows what the message looks
// like: the literal text, its size budget and the cascade that trims it
// when the summary does not fit are internal/cli's job.

// DefaultPrimeLimit is how many rows ASSIGNED TO YOU and NEXT UP share
// when the call does not say (docs/spec/cmd/prime.md).
const DefaultPrimeLimit = 5

// PrimeParams is one `biso prime` call.
type PrimeParams struct {
	Limit    int
	HasLimit bool
}

// PrimeBoard is what the `BOARD` block and the `board` key of the envelope
// say about the board itself, as opposed to about its tasks.
type PrimeBoard struct {
	Name           string
	Statuses       []string
	InitialStatus  string
	ActiveStatus   string
	TerminalStatus string
	Types          []string
	Priorities     []string
	// CountByStatus counts the not archived tasks of each status, and has
	// one entry per configured status even when it is zero.
	CountByStatus map[string]int
}

// PrimeResult is what `biso prime` answers.
type PrimeResult struct {
	Board PrimeBoard
	// Me is the configured identity, empty when there is none, which is
	// also what makes AssignedToYou always empty.
	Me string

	InProgress    []TaskView
	NeedsAnswer   []TaskView
	AssignedToYou []TaskView
	NextUp        []TaskView

	// HiddenCount is how many tasks --limit left out of AssignedToYou and
	// NextUp together.
	HiddenCount int

	// Empty says the board holds no task at all, archived ones included,
	// which is the one case that replaces the four blocks
	// (docs/spec/cmd/prime.md#tablero-vacío).
	Empty bool

	Skipped []string
}

// Prime reads everything the startup message says (docs/spec/cmd/prime.md).
func Prime(env Env, p PrimeParams) (*PrimeResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return PrimeOn(b, env, p)
}

// PrimeOn is Prime over a board that is already open.
//
// It emits no warning of its own, not even for a task it could not read:
// `biso prime` writes nothing on stderr at all, and the two facts that
// would travel there (the missing identity and the unreadable task) are
// lines of the message instead, which is the exception
// docs/spec/cmd/prime.md declares against
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar.
func PrimeOn(b *board.Board, env Env, p PrimeParams) (*PrimeResult, error) {
	r := newReader(b, env)
	if err := r.load(); err != nil {
		return nil, err
	}

	result := &PrimeResult{
		Me: env.Me,
		Board: PrimeBoard{
			Name:           b.Config.ProjectName,
			Statuses:       append([]string(nil), b.Config.Statuses...),
			InitialStatus:  b.Config.InitialStatus,
			ActiveStatus:   b.Config.ActiveStatus,
			TerminalStatus: b.Config.TerminalStatus,
			Types:          append([]string(nil), b.Config.Types...),
			Priorities:     append([]string(nil), b.Config.Priorities...),
			CountByStatus:  map[string]int{},
		},
	}
	for _, status := range b.Config.Statuses {
		// Every configured status has its number, zero included: the line
		// is a fact about the board's vocabulary and not a list of the
		// statuses that happen to be in use.
		result.Board.CountByStatus[status] = 0
	}
	result.Empty = len(r.all) == 0 && len(r.skipped) == 0

	for _, t := range r.all {
		if t.Archived {
			continue
		}
		v, viewErr := r.view(t, false)
		if viewErr != nil {
			// The same skip as any other set read: the task is left out
			// and named, and never a reason to fail. It is left out of
			// the counts line too, which would otherwise claim to
			// describe a task the message could not look at.
			r.undecodable(t, viewErr)
			continue
		}
		if _, configured := result.Board.CountByStatus[t.Status]; configured {
			result.Board.CountByStatus[t.Status]++
		}
		if t.Status == b.Config.TerminalStatus {
			// The four blocks leave out the finished tasks, and the
			// counts line above has already taken note of this one.
			continue
		}
		result.place(b, env, v)
	}

	// The four blocks are ordered by the rule of order of
	// docs/spec/cmd/ls.md, the same one `biso ls` uses with no --sort.
	for _, block := range []*[]TaskView{
		&result.InProgress, &result.NeedsAnswer,
		&result.AssignedToYou, &result.NextUp,
	} {
		r.sortTasks(*block, ListParams{})
	}
	result.applyLimit(p)
	result.Skipped = r.skippedIDs()
	return result, nil
}

// place puts one task in the first of the four blocks that accepts it,
// which is the precedence of docs/spec/cmd/prime.md#la-salida-literal. No
// task is ever in two of them.
func (r *PrimeResult) place(b *board.Board, env Env, v TaskView) {
	switch {
	case v.Waiting:
		r.NeedsAnswer = append(r.NeedsAnswer, v)
	case v.Task.Status == b.Config.ActiveStatus:
		r.InProgress = append(r.InProgress, v)
	case env.Me != "" && assignedTo(v.Task, env.Me):
		r.AssignedToYou = append(r.AssignedToYou, v)
	default:
		r.NextUp = append(r.NextUp, v)
	}
}

// assignedTo folds the case, the same way every other reading comparison
// over a person does
// (docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma).
func assignedTo(t *model.Task, me string) bool {
	return containsFold(t.Assignees, me)
}

// applyLimit shares --limit between ASSIGNED TO YOU and NEXT UP, in that
// order of preference, and records how many tasks the cut left out of the
// two together (docs/spec/cmd/prime.md#la-salida-literal).
func (r *PrimeResult) applyLimit(p PrimeParams) {
	limit := DefaultPrimeLimit
	if p.HasLimit {
		limit = p.Limit
	}
	total := len(r.AssignedToYou) + len(r.NextUp)

	assigned := limit
	if assigned > len(r.AssignedToYou) {
		assigned = len(r.AssignedToYou)
	}
	next := limit - assigned
	if next > len(r.NextUp) {
		next = len(r.NextUp)
	}
	r.AssignedToYou = r.AssignedToYou[:assigned]
	r.NextUp = r.NextUp[:next]
	r.HiddenCount = total - assigned - next
}
