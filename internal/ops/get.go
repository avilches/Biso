package ops

import (
	"fmt"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is docs/spec/cmd/get.md: one task, whole or by sections, with
// the urgency explained if the call asked for it.

// The eight sections of a task's card, in the fixed order they are always
// printed in (docs/spec/cmd/get.md). It is a closed vocabulary: a name that
// is not one of these is a usage error that lists the eight.
const (
	SectionMeta     = "meta"
	SectionDesc     = "desc"
	SectionAC       = "ac"
	SectionPlan     = "plan"
	SectionNotes    = "notes"
	SectionSummary  = "summary"
	SectionComments = "comments"
	SectionQuestion = "question"
)

// Sections are the eight, in the order of the card. Several --section come
// out in this order and never in the order they were asked for, so
// `--section plan,ac` and `--section ac,plan` print the same thing.
var Sections = []string{
	SectionMeta, SectionDesc, SectionAC, SectionPlan,
	SectionNotes, SectionSummary, SectionComments, SectionQuestion,
}

// GetParams is one `biso get` call, already read off the command line.
type GetParams struct {
	Ref string
	// Mode is the interpretation --id or --match forced on the reference.
	Mode RefMode
	// Sections are the ones the call asked for, in any order, and empty
	// when it asked for none, which is the whole card.
	Sections       []string
	ExplainUrgency bool
}

// GetResult is what `biso get` answers.
type GetResult struct {
	Task TaskView
	// Sections are the ones to print, in the fixed order of the card, and
	// every one of them when the call named none.
	Sections []string
	// WholeCard says no --section was written, which is what decides
	// whether an empty section is printed as `(empty)` or left out
	// (docs/spec/cmd/get.md).
	WholeCard bool

	Warnings []Warning
	Notes    []string
}

// Get shows one task (docs/spec/cmd/get.md).
func Get(env Env, p GetParams) (*GetResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return GetOn(b, env, p)
}

// GetOn is Get over a board that is already open.
func GetOn(b *board.Board, env Env, p GetParams) (*GetResult, error) {
	if strings.TrimSpace(p.Ref) == "" {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "missing_ref",
			Message:  "biso get needs a task reference",
			Hints:    []string{"biso get MYP-11"},
		}
	}

	for _, section := range p.Sections {
		if !containsString(Sections, section) {
			return nil, UnknownSection(section)
		}
	}

	r := newReader(b, env)
	if err := r.load(); err != nil {
		return nil, err
	}

	resolved, err := resolveRefWith(b, r.all, p.Ref, p.Mode)
	if err != nil {
		return nil, withCandidates(b, env, r.all, err)
	}
	if resolved.Note != "" {
		r.note(resolved.Note)
	}
	if resolved.Task.Archived {
		r.note(resolved.Task.ID + " is archived")
	}

	// A targeted read of a task that cannot be decoded is exit code 3 with
	// the reason, and not the skip of a set read: there is nothing else to
	// answer with when the caller asked for that one task
	// (docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar).
	view, viewErr := r.view(resolved.Task, p.ExplainUrgency)
	if viewErr != nil {
		return nil, viewErr
	}

	result := &GetResult{
		Task:      view,
		Sections:  sectionsOf(p.Sections),
		WholeCard: len(p.Sections) == 0,
		Notes:     r.notes,
	}
	// The tasks the text search could not read are named here and not
	// inside the resolution, once for the whole call
	// (docs/spec/garantias.md).
	r.warnAboutSkipped()
	result.Warnings = r.warnings
	return result, nil
}

// sectionsOf puts the sections the call asked for into the fixed order of
// the card, and answers all eight when it asked for none.
func sectionsOf(asked []string) []string {
	if len(asked) == 0 {
		return append([]string(nil), Sections...)
	}
	var out []string
	for _, name := range Sections {
		if containsString(asked, name) {
			out = append(out, name)
		}
	}
	return out
}

// EmptySection answers whether a section of this task has nothing in it,
// which is what decides between printing `(empty)` and leaving it out.
// `meta` is never empty: a task always has an identifier and a status.
func EmptySection(t *model.Task, section string) bool {
	switch section {
	case SectionMeta:
		return false
	case SectionDesc:
		return t.Description == ""
	case SectionAC:
		return len(t.AcceptanceCriteria) == 0
	case SectionPlan:
		return t.Plan == ""
	case SectionNotes:
		return t.Notes == ""
	case SectionSummary:
		return t.Summary == ""
	case SectionComments:
		return len(t.Comments) == 0
	case SectionQuestion:
		return t.Question == nil
	}
	return true
}

// UnknownSection is the error of a --section nobody has: exit code 2 with
// the eight valid names, per the table of docs/spec/cmd/get.md.
func UnknownSection(given string) *model.Error {
	return &model.Error{
		ExitCode: 2,
		Code:     "unknown_section",
		Message:  fmt.Sprintf("--section: unknown value: %q", given),
		Field:    "section",
		Given:    given,
		Valid:    append([]string(nil), Sections...),
	}
}
