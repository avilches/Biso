package convert

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"backlog.md-migrate/internal/destination"
	"backlog.md-migrate/internal/source"
)

// This is the final assembly step of the conversion engine. It does not run
// any conversion of its own: every value a Line carries was already computed
// by phase 4a (Task, convert.go), phase 4b (Identifiers, identifiers.go),
// phase 4c (AssignOrdinals, ordinal.go), or read verbatim from the
// source.Task no earlier phase touches (References, Documentation,
// ModifiedFiles, Archived). Assemble's only job is to run those phases in
// the right order and join their outputs into what "biso new --from" expects
// (docs/especificacion.md, "La salida", and docs/spec/cmd/new.md, "El modo
// lote").

// AcceptanceCriterion is one element of a Line's acceptanceCriteria, in the
// object shape docs/spec/cmd/new.md, "El modo lote", accepts: key, text,
// checked, in that order. Checked has no "omitempty": the batch format's own
// example line writes "checked":false explicitly, and biso needs to see
// that false to know the criterion is unchecked rather than treat it as
// unset.
type AcceptanceCriterion struct {
	Key     int    `json:"key"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

// CommentLine is one element of a Line's comments. Author and CreatedAt are
// omitted when the source comment had none: docs/especificacion.md,
// "Comentarios", says a comment without an author, or without a created
// date, is valid and not a Finding. Body carries no "omitempty": a comment
// that reached this point always has one (source.Task never keeps a comment
// block without a body).
type CommentLine struct {
	Author    string `json:"author,omitempty"`
	CreatedAt string `json:"createdAt,omitempty"`
	Body      string `json:"body"`
}

// Line is one task ready to be serialized as a single line of NDJSON, with
// its fields declared in the exact fixed order docs/especificacion.md, "La
// salida", requires: encoding/json serializes a struct's fields in
// declaration order, so this declaration order alone fixes the output's key
// order, with no hand-written encoder needed.
//
// Every field but ID and Title carries "omitempty": docs/especificacion.md
// says a field that is absent or empty (an empty string, an empty list, or
// false) is omitted from the line entirely, never written as null or [].
type Line struct {
	ID                 string                `json:"id"`
	Title              string                `json:"title"`
	Status             string                `json:"status,omitempty"`
	Type               string                `json:"type,omitempty"`
	Priority           string                `json:"priority,omitempty"`
	Assignees          []string              `json:"assignees,omitempty"`
	Labels             []string              `json:"labels,omitempty"`
	Dependencies       []string              `json:"dependencies,omitempty"`
	Parent             string                `json:"parent,omitempty"`
	Ordinal            string                `json:"ordinal,omitempty"`
	Due                string                `json:"due,omitempty"`
	Documentation      []string              `json:"documentation,omitempty"`
	References         []string              `json:"references,omitempty"`
	ModifiedFiles      []string              `json:"modifiedFiles,omitempty"`
	Archived           bool                  `json:"archived,omitempty"`
	CreatedAt          string                `json:"createdAt,omitempty"`
	UpdatedAt          string                `json:"updatedAt,omitempty"`
	Description        string                `json:"description,omitempty"`
	Plan               string                `json:"plan,omitempty"`
	Notes              string                `json:"notes,omitempty"`
	Summary            string                `json:"summary,omitempty"`
	AcceptanceCriteria []AcceptanceCriterion `json:"acceptanceCriteria,omitempty"`
	Comments           []CommentLine         `json:"comments,omitempty"`
}

// Assemble runs the whole conversion engine over a full source board and a
// full destination board and returns the final, ordered list of Lines ready
// to serialize as NDJSON, plus every Finding phases 4a and 4b raise along
// the way (converting each task, then resolving identifiers), in that
// order.
//
// sourceBoard.Findings (raised while reading the board, phase 2) is
// deliberately NOT part of the returned findings: Assemble only knows about
// the findings the conversion engine itself raises. A caller combines the
// two, source board findings first, since those happened first in the
// pipeline.
//
// A non-nil error is returned instead, with lines and findings both nil,
// only when Identifiers fails its point-1 validation
// (docs/especificacion.md, "Identificadores", point 1: a malformed id, a
// shared-prefix violation, or a duplicate id), the one failure the
// specification turns into exit code 3 rather than a Finding. A caller must
// check this error before looking at anything else Assemble returns.
func Assemble(sourceBoard source.Board, destBoard destination.Board) ([]Line, []source.Finding, error) {
	var findings []source.Finding

	milestoneSlugs, f := MilestoneSlugs(sourceBoard.Milestones)
	findings = append(findings, f...)

	batch := make([]TaskInput, len(sourceBoard.Tasks))
	for i, t := range sourceBoard.Tasks {
		result, taskFindings := Task(t, milestoneSlugs, destBoard.Config)
		findings = append(findings, taskFindings...)
		batch[i] = TaskInput{Task: t, Result: result}
	}

	identified, identifierFindings, err := Identifiers(batch, destBoard)
	if err != nil {
		return nil, nil, err
	}
	findings = append(findings, identifierFindings...)

	// AssignOrdinals needs exactly the tasks that are going to produce a
	// line of output (its own doc comment: a skipped task must never reach
	// it), which is precisely what identified already enumerates, one
	// Identified per producing task. Looked up by BatchIndex, not SourceID:
	// docs/especificacion.md, "Identificadores", point 1's id-reuse case
	// means two Identified values can share the exact same SourceID, so a
	// map keyed by SourceID (as this used to be) would silently collapse
	// two distinct tasks into one, handing AssignOrdinals, and every line
	// built below, the wrong source.Task for one of them.
	//
	// AssignOrdinals now returns a slice parallel to producing rather than a
	// map keyed by source id, for the exact same reason: two tasks sharing a
	// reused id could otherwise overwrite each other's assigned key.
	// producing[i] is built from identified[i], so ordinals[i] is always
	// identified[i]'s own key, read back below by that same index.
	producing := make([]source.Task, len(identified))
	for i, id := range identified {
		producing[i] = batch[id.BatchIndex].Task
	}
	ordinals := AssignOrdinals(producing, destBoard)

	type entry struct {
		line   Line
		parsed parsedSourceID
	}
	entries := make([]entry, len(identified))
	for i, id := range identified {
		result := batch[id.BatchIndex].Result
		sourceTask := batch[id.BatchIndex].Task

		parsed, ok := parseSourceID(id.SourceID)
		if !ok {
			// Identifiers only ever produces an Identified whose SourceID
			// already passed validateSourceShape's point-1 check, so this
			// can never actually happen; a failure here would be a bug in
			// an earlier phase, not a new data problem to report.
			panic(fmt.Sprintf("convert: Identifiers returned SourceID %q, which does not parse as a source id", id.SourceID))
		}

		entries[i] = entry{
			line:   buildLine(id, result, ordinals[i], sourceTask),
			parsed: parsed,
		}
	}

	// docs/especificacion.md, "La salida", requires ascending order by the
	// source id's own number. naturalSourceIDLess is the same natural order
	// identifiers.go and ordinal.go already use, reused here unchanged so
	// all three phases agree on what that order means.
	sort.SliceStable(entries, func(i, j int) bool {
		return naturalSourceIDLess(entries[i].parsed, entries[j].parsed)
	})

	lines := make([]Line, len(entries))
	for i, e := range entries {
		lines[i] = e.line
	}

	return lines, findings, nil
}

// buildLine combines one Identified with its matching Result, its assigned
// ordinal key (empty when it has none), and its original source.Task into
// the Line that gets serialized. The three-way split follows
// docs/especificacion.md, "El mapeo de campos", plus the fields no earlier
// phase touches at all (References, Documentation, ModifiedFiles, Archived),
// which are read straight from sourceTask.
func buildLine(id Identified, result Result, ordinal string, sourceTask source.Task) Line {
	return Line{
		ID:                 id.ID,
		Title:              id.Title,
		Status:             result.Status,
		Type:               result.Type,
		Priority:           result.Priority,
		Assignees:          result.Assignees,
		Labels:             id.Labels,
		Dependencies:       id.Dependencies,
		Parent:             id.Parent,
		Ordinal:            ordinal,
		Due:                result.Due,
		Documentation:      sourceTask.Documentation,
		References:         sourceTask.References,
		ModifiedFiles:      sourceTask.ModifiedFiles,
		Archived:           sourceTask.Archived,
		CreatedAt:          result.CreatedAt,
		UpdatedAt:          result.UpdatedAt,
		Description:        id.Description,
		Plan:               id.Plan,
		Notes:              id.Notes,
		Summary:            id.Summary,
		AcceptanceCriteria: convertCriteria(id.AcceptanceCriteria),
		Comments:           convertComments(id.Comments),
	}
}

// convertCriteria maps source.Checkbox to the acceptanceCriteria object
// shape (key/text/checked), or nil for an empty list so the "omitempty" tag
// omits the key entirely rather than writing "[]".
func convertCriteria(criteria []source.Checkbox) []AcceptanceCriterion {
	if len(criteria) == 0 {
		return nil
	}
	out := make([]AcceptanceCriterion, len(criteria))
	for i, c := range criteria {
		out[i] = AcceptanceCriterion{Key: c.Number, Text: c.Text, Checked: c.Checked}
	}
	return out
}

// convertComments maps Comment to the comments object shape
// (author/createdAt/body, no key), or nil for an empty list for the same
// "omitempty" reason as convertCriteria.
func convertComments(comments []Comment) []CommentLine {
	if len(comments) == 0 {
		return nil
	}
	out := make([]CommentLine, len(comments))
	for i, c := range comments {
		out[i] = CommentLine{Author: c.Author, CreatedAt: c.CreatedAt, Body: c.Body}
	}
	return out
}

// EncodeNDJSON writes one compact JSON object per line to w, in lines'
// order, disabling HTML escaping so a title, description, or comment body
// containing '<', '>', or '&' is not silently rewritten into a Unicode
// escape sequence: this output is NDJSON for "biso new --from"
// (docs/especificacion.md, "La salida"), never HTML. It lives in this
// package, next to Line itself, rather than in the cli package that calls
// it, so that a test exercising the exact byte-for-byte shape of a line
// (field order, omitted fields, escaping) can call the very same encoder
// the real import command uses instead of a second one built just for the
// test.
func EncodeNDJSON(w io.Writer, lines []Line) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, line := range lines {
		if err := enc.Encode(line); err != nil {
			return err
		}
	}
	return nil
}
