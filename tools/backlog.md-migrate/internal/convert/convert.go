// Package convert computes the "local" pieces of turning a source.Task
// into what "biso new --from" expects: the ones a single task can compute
// on its own, needing at most the destination's configured vocabulary and
// the source board's milestone titles, but never another task in the same
// batch and never a task that already exists on the destination.
//
// This file implements the per-task field conversion of the conversion
// engine (docs/especificacion.md, "El mapeo de campos"). It covers status,
// type and priority vocabulary matching, the labels/assignees token
// alphabet, date conversion to UTC, the milestone:: and project:: scoped
// labels, and folding Definition of Done into acceptance criteria. It
// deliberately does NOT touch identifiers, parent, dependencies, ordinal,
// title, description, plan, notes, summary, or a comment's body: rewriting
// mentions and identifiers needs to see every task in the batch and the
// tasks already on the destination, which identifiers.go handles, built on
// top of this file's output.
package convert

import (
	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

// Result holds the fields of a single task that this phase converts. Every
// other field of the source.Task it was built from (identifiers, Parent,
// Dependencies, Ordinal, Title, Description, Plan, Notes, Summary, and a
// comment's Author/Body) is untouched by this phase and does not appear
// here; a later phase combines this Result with those fields and with the
// backlog.id:: label before writing a line of NDJSON.
type Result struct {
	// Status, Type, and Priority are the destination's own configured
	// spelling once MatchField found an unambiguous match, or the empty
	// string when nothing matched (the field is then omitted from the
	// batch line, and a Finding already explains why).
	Status   string
	Type     string
	Priority string

	// Labels is the task's origin labels, cleaned through
	// CleanTokenList, with any label colliding with a derived
	// milestone:: or project:: key removed, followed by milestone::
	// and then project:: when the task has one.
	Labels []string
	// Assignees is the task's origin assignees, cleaned through
	// CleanTokenList the same way as Labels.
	Assignees []string

	// Due is task.DueDate copied verbatim: docs/especificacion.md, "El
	// mapeo de campos", says due_date -> due is "YYYY-MM-DD, tal cual",
	// so it is never passed through ConvertDate.
	Due string
	// CreatedAt and UpdatedAt are task.CreatedDate and task.UpdatedDate
	// converted to biso's UTC instant shape by ConvertDate. UpdatedAt
	// is CreatedAt's already-converted value when the task had no
	// updated_date at all (docs/especificacion.md, "Fechas").
	CreatedAt string
	UpdatedAt string

	// Comments mirrors task.Comments in the same order, with CreatedAt
	// converted by ConvertDate (or left empty when the source comment
	// had none) and Author/Body copied verbatim.
	Comments []Comment

	// AcceptanceCriteria is the task's own acceptance criteria followed
	// by its Definition of Done items merged in by
	// MergeDefinitionOfDone.
	AcceptanceCriteria []source.Checkbox
}

// Comment is one converted comment: Author and Body copied verbatim from
// source.Comment, CreatedAt converted to biso's UTC instant shape.
type Comment struct {
	Author    string
	CreatedAt string
	Body      string
}

// Task converts the pieces of a single source.Task that this phase owns
// into a Result, plus every Finding raised while doing so.
//
// milestoneSlugs is the map MilestoneSlugs already computed once for the
// whole batch (every milestone id the source board knows about, mapped to
// its slug); Task looks up t.Milestone in it rather than recomputing a
// slug itself, and falls back to using the id when t.Milestone names a
// milestone that map does not have an entry for (the milestone file the id
// pointed at was never found while reading the board).
//
// config is the destination's configuration; only Statuses, Types, and
// Priorities are used here. config.TaskPrefix belongs to a later phase's
// identifier rewriting.
//
// Task never looks at any other source.Task and never looks at what
// already exists on the destination board: identifiers, Parent,
// Dependencies rewriting, Ordinal placement, and the backlog.id:: label
// all need that wider view and are a later phase's job.
func Task(t source.Task, milestoneSlugs map[string]string, config destination.Config) (Result, []source.Finding) {
	var findings []source.Finding

	status, f := MatchField(t.File, "status", t.Status, config.Statuses)
	findings = append(findings, f...)
	typ, f := MatchField(t.File, "type", t.Type, config.Types)
	findings = append(findings, f...)
	priority, f := MatchField(t.File, "priority", t.Priority, config.Priorities)
	findings = append(findings, f...)

	labels, f := CleanTokenList(t.File, "labels", t.Labels)
	findings = append(findings, f...)
	assignees, f := CleanTokenList(t.File, "assignees", t.Assignees)
	findings = append(findings, f...)

	labels, f = ScopedLabels(t.File, labels, t.Milestone, t.Project, milestoneSlugs)
	findings = append(findings, f...)

	createdAt := ConvertDate(t.CreatedDate)
	updatedAt := createdAt
	if t.UpdatedDate != "" {
		updatedAt = ConvertDate(t.UpdatedDate)
	}

	var comments []Comment
	for _, c := range t.Comments {
		comments = append(comments, Comment{
			Author:    c.Author,
			CreatedAt: ConvertDate(c.CreatedAt),
			Body:      c.Body,
		})
	}

	criteria := MergeDefinitionOfDone(t.AcceptanceCriteria, t.DefinitionOfDone)

	return Result{
		Status:             status,
		Type:               typ,
		Priority:           priority,
		Labels:             labels,
		Assignees:          assignees,
		Due:                t.DueDate,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
		Comments:           comments,
		AcceptanceCriteria: criteria,
	}, findings
}
