package ops

import (
	"fmt"
	"strconv"
	"strings"

	"biso/internal/board"
	"biso/internal/match"
	"biso/internal/model"
)

// This file is docs/spec/referencias.md, whole: the grammar of a <ref>, the
// one text search the program has, and the three endings of resolving one.
//
// It lives here, as an internal function of internal/ops shared by every
// command that takes a <ref>, and not in a package of its own, because it
// needs to read the board for the text search and therefore cannot be a pure
// function of a leaf package the way internal/match is. That is section 3.5
// of docs/superpowers/specs/2026-09-10-arquitectura-implementacion-design.md.

// RefMode is the interpretation a call forces on a <ref> with --id or
// --match, and RefAuto is the grammar of
// docs/spec/referencias.md#la-gramática deciding on its own.
type RefMode int

const (
	RefAuto RefMode = iota
	// RefID forces an identifier. A value the grammar does not admit is
	// exit code 2 here, instead of becoming a text query.
	RefID
	// RefText forces a text query, which is how a task whose title is a
	// number is looked up by that title.
	RefText
)

// AmbiguousRef is the exit code 5 of
// docs/spec/referencias.md#la-búsqueda-por-texto: a text that matches more
// than one task. It carries the candidates because that ending prints them
// on stdout, in the format of `biso ls`, next to the error line on stderr.
//
// It wraps the ordinary *model.Error rather than replacing it, so that every
// layer that only knows how to print an error keeps working unchanged: the
// error is reached with errors.As, and only a caller that wants to print the
// candidates has to know this type exists.
type AmbiguousRef struct {
	Err        *model.Error
	Candidates []*model.Task
	// Listing is the candidates as `biso ls` would print them: the same
	// order, the same limit of thirty and the same truncation warning. The
	// command that resolved the reference fills it in, because the order
	// of that listing needs the board's configuration and this function
	// does not have it.
	Listing *ListResult
}

func (a *AmbiguousRef) Error() string { return a.Err.Error() }

// Unwrap is what makes errors.As find the *model.Error inside.
func (a *AmbiguousRef) Unwrap() error { return a.Err }

// Resolved is one reference that resolved, with the note the text search
// owes its caller when a text, and not an identifier, is what found the task
// (docs/spec/referencias.md#la-búsqueda-por-texto).
//
// A task the text search could not decode is not reported here. A set read
// never aborts on one and never hides it
// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar),
// and every command that resolves a reference reads the whole board anyway,
// so the one warning that names the skipped tasks is emitted there, once,
// instead of once per reference resolved.
type Resolved struct {
	Task *model.Task
	Note string
}

// resolveRef resolves one reference to one task, exactly as
// docs/spec/referencias.md says and with no variant per command.
func resolveRef(b *board.Board, ref string, mode RefMode) (*Resolved, error) {
	return resolveRefWith(b, nil, ref, mode)
}

// resolveRefWith is resolveRef over a board whose tasks the caller has
// already read. Every command that resolves a reference reads the whole
// board anyway, for the derived fields or for the cycles, so handing that
// list in keeps the text search from running the same five queries twice.
// A nil list means the search reads them itself.
func resolveRefWith(b *board.Board, all []*model.Task, ref string, mode RefMode) (*Resolved, error) {
	num, wellFormed := parseTaskRef(b.Config.TaskPrefix, ref)
	switch {
	case mode == RefID && !wellFormed:
		return nil, malformedIDError(ref)
	case mode == RefText:
		return searchForRef(b, all, ref)
	case wellFormed:
		task, err := b.Tasks.Load(fmt.Sprintf("%s-%d", b.Config.TaskPrefix, num))
		if err != nil {
			return nil, err
		}
		return &Resolved{Task: task}, nil
	}
	return searchForRef(b, all, ref)
}

// parseTaskRef reads the three shapes of an identifier of
// docs/spec/referencias.md#la-gramática: "MYP-11" with the prefix written in
// any case, "11", and "#11". It answers the number and whether the value is
// an identifier at all; anything else is a text query.
func parseTaskRef(prefix, ref string) (int, bool) {
	rest := strings.TrimPrefix(ref, "#")
	if len(rest) != len(ref) {
		return positiveNumber(rest)
	}
	if lower := strings.ToLower(ref); strings.HasPrefix(lower, strings.ToLower(prefix)+"-") {
		return positiveNumber(ref[len(prefix)+1:])
	}
	return positiveNumber(ref)
}

func positiveNumber(s string) (int, bool) {
	// strconv.Atoi admits a sign and this grammar does not: "MYP--3" and
	// "+3" are not identifiers.
	for _, r := range s {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 1 {
		return 0, false
	}
	return n, true
}

// malformedIDError is the first of the three messages of
// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro. The other
// two are born in internal/board, which is the layer that knows the highest
// identifier the board ever handed out.
func malformedIDError(ref string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "malformed_id",
		Message:  fmt.Sprintf("malformed task id: %q", ref),
		Hints: []string{
			"ids look like MYP-11 or 11. A subtask is an ordinary task with --parent MYP-1",
		},
		Field: "ref",
		Given: ref,
	}
}

// searchForRef is the text search used to resolve a reference, which is the
// same search --search uses with the extra rules of
// docs/spec/referencias.md#la-búsqueda-por-texto laid on top: it looks only
// at tasks that are not archived, a hit in the title beats a hit anywhere
// else, and the three counts end differently.
func searchForRef(b *board.Board, all []*model.Task, query string) (*Resolved, error) {
	if all == nil {
		var err error
		if all, _, err = b.Tasks.All(); err != nil {
			return nil, err
		}
	}
	var titles, anywhere []*model.Task
	for _, t := range all {
		if t.Archived {
			continue
		}
		if TextMatches(t.Title, query) {
			titles = append(titles, t)
			anywhere = append(anywhere, t)
			continue
		}
		if TaskMatchesText(t, query) {
			anywhere = append(anywhere, t)
		}
	}

	// A hit in the title wins over a hit anywhere else: if the text is in
	// the title of one single task, that is the answer however many other
	// tasks carry it in their body. With two titles carrying it, those two
	// are the ambiguity, and the bodies never enlarge it.
	candidates := anywhere
	if len(titles) > 0 {
		candidates = titles
	}

	switch len(candidates) {
	case 1:
		return &Resolved{
			Task: candidates[0],
			Note: fmt.Sprintf("%q matched %s", query, candidates[0].ID),
		}, nil
	case 0:
		return nil, &model.Error{
			ExitCode: 4,
			Code:     "not_found",
			Message:  fmt.Sprintf("no task matches %q", query),
			Field:    "ref",
			Given:    query,
		}
	}
	return nil, &AmbiguousRef{
		Err: &model.Error{
			ExitCode: 5,
			Code:     "ambiguous_reference",
			Message:  fmt.Sprintf("%q matches %d tasks", query, len(candidates)),
			Field:    "ref",
			Given:    query,
		},
		Candidates: candidates,
	}
}

// TaskMatchesText is the one search scope the whole program has
// (docs/spec/referencias.md#la-búsqueda-por-texto): the title, the
// description, the plan, the notes, the final summary, the text of the
// acceptance criteria, the body of the comments, the body of the open
// question and the labels. It does not look at the identifiers, the
// references, the documentation or the extension fields.
//
// It is exported because `biso ls --search` and `biso export --search` use
// this very function and not a second implementation of the same list.
func TaskMatchesText(t *model.Task, query string) bool {
	for _, field := range []string{t.Title, t.Description, t.Plan, t.Notes, t.Summary} {
		if TextMatches(field, query) {
			return true
		}
	}
	for _, c := range t.AcceptanceCriteria {
		if TextMatches(c.Text, query) {
			return true
		}
	}
	for _, c := range t.Comments {
		if TextMatches(c.Body, query) {
			return true
		}
	}
	if t.Question != nil && TextMatches(t.Question.Body, query) {
		return true
	}
	for _, label := range t.Labels {
		if TextMatches(label, query) {
			return true
		}
	}
	return false
}

// TextMatches is the comparison every text search of the specification uses:
// a substring, with the case folded and the diacritics dropped on both sides
// alike (docs/spec/familias-de-flags.md#selectores-de-criterios fixes the two
// steps, and docs/spec/referencias.md#la-búsqueda-por-texto asks for the same
// two words, "sin distinguir mayúsculas ni acentos").
//
// It folds through match.Simplify and not through match.Normalize: the
// vocabulary algorithm also drops spaces, hyphens and underscores, which is
// right for a status typed as "To-Do" and wrong for a fragment of prose,
// where a space is a character of the text like any other.
func TextMatches(haystack, needle string) bool {
	return strings.Contains(match.Simplify(haystack), match.Simplify(needle))
}
