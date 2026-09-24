package cli

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the output of docs/spec/cmd/get.md: the card of one task,
// the explanation of its urgency, the task.get envelope, and the candidates
// an ambiguous reference prints, which every command that resolves one
// shares (docs/spec/referencias.md).

func runGet(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	params, err := getParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.Get(env, params)
	if err != nil {
		var extra []Warning
		if result != nil {
			for _, w := range result.Warnings {
				extra = append(extra, cliWarning(w))
			}
		}
		return failWithCandidates(s, p, env, asJSON, err, extra)
	}
	printWarnings(s, p)
	printOpsWarnings(s, result.Warnings)

	if asJSON {
		writeEnvelope(s, env, "task.get", map[string]any{"task": getObject(result)})
	} else {
		fmt.Fprint(s.Stdout, renderCard(result))
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	return 0
}

// getParams turns the analyzed call into the typed parameters of the
// command.
func getParams(p *Parsed) (ops.GetParams, error) {
	if len(p.Positionals) == 0 {
		return ops.GetParams{}, &model.Error{
			ExitCode: 2,
			Code:     "missing_ref",
			Message:  "biso get needs a task reference",
			Hints:    []string{"biso get MYP-11"},
		}
	}
	if len(p.Positionals) > 1 {
		return ops.GetParams{}, errUnexpectedArgument(p.Positionals[1])
	}
	params := ops.GetParams{
		Ref:            p.Positionals[0],
		Sections:       p.Values("section"),
		ExplainUrgency: p.Has("explain-urgency"),
		Closure:        p.Has("closure"),
	}
	switch {
	case p.Has("id"):
		params.Mode = ops.RefID
	case p.Has("match"):
		params.Mode = ops.RefText
	}
	return params, nil
}

// renderCard is the output block of docs/spec/cmd/get.md.
//
// Without --section every one of the eight sections is printed, the empty
// ones marked; with it, an empty section is left out, and when nothing is
// left the output is empty and the exit code is still 0.
func renderCard(r *ops.GetResult) string {
	t := r.Task.Task
	printed := make([]string, 0, len(r.Sections))
	for _, section := range r.Sections {
		if r.WholeCard || !ops.EmptySection(t, section) {
			printed = append(printed, section)
		}
	}
	if len(printed) == 0 && r.Task.Breakdown == nil && r.Task.Closure == nil {
		return ""
	}

	var b strings.Builder
	fmt.Fprintf(&b, "%s%s%s\n", t.ID, columnGap, t.Title)
	for _, section := range printed {
		if section == ops.SectionMeta {
			b.WriteString(metaBlock(r.Task))
			continue
		}
		b.WriteString("\n")
		fmt.Fprintf(&b, "## %s\n", sectionHeadings[section])
		b.WriteString(sectionBody(t, section))
	}
	if r.Task.Breakdown != nil {
		b.WriteString("\n")
		b.WriteString(renderUrgency(r.Task))
	}
	if r.Task.Closure != nil {
		b.WriteString("\n")
		b.WriteString(renderClosure(r.Task))
	}
	return b.String()
}

// sectionHeadings are the headings the card prints over each section. They
// are a presentation format and not a storage one
// (docs/spec/cmd/get.md#salida).
var sectionHeadings = map[string]string{
	ops.SectionDesc:     "Description",
	ops.SectionAC:       "Acceptance Criteria",
	ops.SectionPlan:     "Implementation Plan",
	ops.SectionNotes:    "Implementation Notes",
	ops.SectionSummary:  "Final Summary",
	ops.SectionComments: "Comments",
	ops.SectionQuestion: "Open Question",
}

// emptyMark is what an empty section prints in the whole card, which is
// only ever seen without --section.
const emptyMark = "(empty)"

func sectionBody(t *model.Task, section string) string {
	switch section {
	case ops.SectionDesc:
		return prose(t.Description)
	case ops.SectionPlan:
		return prose(t.Plan)
	case ops.SectionNotes:
		return prose(t.Notes)
	case ops.SectionSummary:
		return prose(t.Summary)
	case ops.SectionAC:
		if len(t.AcceptanceCriteria) == 0 {
			return emptyMark + "\n"
		}
		var b strings.Builder
		for _, c := range t.AcceptanceCriteria {
			mark := " "
			if c.Checked {
				mark = "x"
			}
			fmt.Fprintf(&b, "- [%s] #%d %s\n", mark, c.Key, c.Text)
		}
		return b.String()
	case ops.SectionComments:
		if len(t.Comments) == 0 {
			return emptyMark + "\n"
		}
		var b strings.Builder
		for i, c := range t.Comments {
			if i > 0 {
				b.WriteString("\n")
			}
			// The key is part of the line because a comment is an element
			// of a list that a selector can point at
			// (docs/spec/modelo-de-datos/comentarios.md).
			fmt.Fprintf(&b, "#%d%s%s, %s\n", c.Key, columnGap, c.Author, minute(c.CreatedAt))
			b.WriteString(prose(c.Body))
		}
		return b.String()
	case ops.SectionQuestion:
		if t.Question == nil {
			return emptyMark + "\n"
		}
		// No key: the open question is a single value and not an element
		// of a list (docs/spec/modelo-de-datos/pregunta-abierta.md).
		return fmt.Sprintf("%s, %s\n", t.Question.Author, minute(t.Question.AskedAt)) +
			prose(t.Question.Body)
	}
	return ""
}

func prose(text string) string {
	if text == "" {
		return emptyMark + "\n"
	}
	if strings.HasSuffix(text, "\n") {
		return text
	}
	return text + "\n"
}

// The layout of the block of metadata: a label of eleven cells, a value of
// twenty one, and then the second pair. The last rows carry one field
// and pad nothing after it.
const (
	metaLabelWidth = 11
	metaValueWidth = 21
)

// metaBlock is the `meta` section: the fields of the task in two columns,
// with the lease line only when the task has a lease
// (docs/spec/cmd/get.md#salida).
func metaBlock(v ops.TaskView) string {
	t := v.Task
	var b strings.Builder
	pair := func(leftLabel, leftValue, rightLabel, rightValue string) {
		b.WriteString(pad(leftLabel, metaLabelWidth))
		b.WriteString(pad(leftValue, metaValueWidth))
		b.WriteString(pad(rightLabel, metaLabelWidth))
		b.WriteString(strings.TrimRight(rightValue, " "))
		b.WriteString("\n")
	}
	single := func(label, value string) {
		b.WriteString(pad(label, metaLabelWidth))
		b.WriteString(value)
		b.WriteString("\n")
	}

	pair("status", t.Status, "type", orDash(t.Type))
	pair("priority", orDash(t.Priority), "urgency", formatUrgency(v.Urgency))
	pair("assignees", joined(t.Assignees), "author", orDash(t.Author))
	pair("labels", joined(t.Labels), "parent", orDash(t.Parent))
	pair("due", orDash(dueCell(t)), "ordinal", ordinalCell(t))
	pair("created", minute(t.CreatedAt), "updated", minute(t.UpdatedAt))
	pair("depends", joined(t.Dependencies), "blocks", joined(v.Blocks))
	// blocked by/unblocks are the transitive closure's two counts, always
	// printed even at zero, because zero is a fact about the task and not
	// the absence of one (docs/spec/cmd/get.md#salida).
	single("blocked by", fmt.Sprintf("%d total", v.BlockedByCount))
	single("unblocks", fmt.Sprintf("%d total", v.UnblocksCount))
	if !t.LeaseExpiresAt.IsZero() {
		// The two fields of a lease appear and disappear together, so one
		// line carries both. A task without one does not print the line,
		// instead of printing it with two dashes
		// (docs/spec/lease.md#el-vaciado).
		pair("lease", minute(t.LeaseExpiresAt), "holder", orDash(t.LeaseHolder))
	}
	single("refs", joined(t.References))
	return b.String()
}

// pad writes a value in a field of that many cells, and lets it overflow
// with a two-space gap when it is wider, which keeps the next field from
// being glued to it.
func pad(value string, width int) string {
	if w := cells(value); w < width {
		return value + strings.Repeat(" ", width-w)
	}
	return value + columnGap
}

// joined writes a list of values on one line, per
// docs/spec/cmd/get.md#salida: each value is escaped with the rule of the
// input read backwards (escapeListValue), and the values are separated by a
// bare comma and a space. Only a reference can carry a comma or a backslash,
// so the other lists come out unchanged, but the rule is the same for all of
// them and does not have to know which list it is writing.
func joined(values []string) string {
	if len(values) == 0 {
		return dash
	}
	escaped := make([]string, len(values))
	for i, v := range values {
		escaped[i] = escapeListValue(v)
	}
	return strings.Join(escaped, ", ")
}

// ordinalCell is the one value of the card that is not printed as it is
// stored. The key cannot be typed and says nothing a reader can use, so
// what the card prints is whether the task has a place decided by hand,
// which is what lets it be named as the neighbour of an --above or a
// --below; the raw key is in --json (docs/spec/cmd/get.md#salida).
func ordinalCell(t *model.Task) string {
	if t.Ordinal == "" {
		return dash
	}
	return "manual"
}

// minute is how the card writes an instant: the calendar day and the time
// of day to the minute, which is the precision a person reads. The second
// is in --json, where a program reads it.
func minute(at time.Time) string {
	if at.IsZero() {
		return dash
	}
	return at.UTC().Format("2006-01-02 15:04")
}

// renderUrgency is the block of --explain-urgency: one line per term of the
// formula, with the coefficient, the factor and the product, and the sum
// under a rule (docs/spec/cmd/get.md#salida).
func renderUrgency(v ops.TaskView) string {
	b := v.Breakdown
	var out strings.Builder
	fmt.Fprintf(&out, "urgency %s\n", formatUrgency(v.Urgency))
	if b.Terminal {
		// A terminal task's urgency is zero with no term computed, so
		// there is nothing to break down
		// (docs/spec/modelo-de-datos/urgencia.md).
		out.WriteString("  terminal status, urgency is zero by definition\n")
		return out.String()
	}
	for _, line := range []struct {
		label string
		term  model.UrgencyTerm
	}{
		{"priority " + priorityLabel(b.PriorityName), b.Priority},
		{activeLabel(b.ActiveReason), b.Active},
		{"blocking", b.Blocking},
		{"blocked", b.Blocked},
		{"due", b.Due},
		{"has criteria", b.Criteria},
		{fmt.Sprintf("age %d days", b.AgeDays), b.Age},
	} {
		out.WriteString(urgencyLine(line.label, line.term))
	}
	out.WriteString(strings.Repeat(" ", totalWidth-ruleWidth) +
		strings.Repeat("-", ruleWidth) + "\n")
	fmt.Fprintf(&out, "%*s\n", totalWidth, strconv.FormatFloat(noNegativeZero(b.Sum), 'f', 2, 64))
	return out.String()
}

// renderClosure is the block of --closure: the full transitive closure of
// `dependencies` in both directions, with the same two counts the block of
// metadata already carries and, this time, the complete list of
// identifiers each one reaches (docs/spec/cmd/get.md#--closure).
func renderClosure(v ops.TaskView) string {
	c := v.Closure
	var b strings.Builder
	b.WriteString("closure\n")
	b.WriteString(closureLine("blocked by", v.BlockedByCount, c.BlockedBy))
	b.WriteString(closureLine("unblocks", v.UnblocksCount, c.Unblocks))
	return b.String()
}

// closureLine is one line of that block: the label padded to the eleven
// cells of the block of metadata, the count as `N total`, two literal
// spaces, and the list of identifiers, or the usual dash when it is empty
// (docs/spec/cmd/get.md#--closure).
func closureLine(label string, count int, ids []string) string {
	list := dash
	if len(ids) > 0 {
		list = strings.Join(ids, ", ")
	}
	return "  " + pad(label, metaLabelWidth) + fmt.Sprintf("%d total", count) + "  " + list + "\n"
}

// The widths of that block: the coefficient ends at coefficientWidth cells,
// the whole line at totalWidth, and the rule over the sum is ruleWidth
// dashes long.
const (
	coefficientWidth = 24
	totalWidth       = 40
	ruleWidth        = 7
)

func urgencyLine(label string, t model.UrgencyTerm) string {
	left := "  " + label
	coefficient := strconv.FormatFloat(t.Coefficient, 'f', 1, 64)
	value := strconv.FormatFloat(noNegativeZero(t.Value), 'f', 2, 64)
	gap := coefficientWidth - cells(left) - len(coefficient)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + coefficient +
		" * " + strconv.FormatFloat(t.Factor, 'f', 2, 64) +
		" =" + fmt.Sprintf("%*s", totalWidth-coefficientWidth-len(" * 0.00 ="), value) + "\n"
}

// noNegativeZero turns a minus zero into a zero. A coefficient that is
// negative, such as the one of `blocked`, times a factor of zero gives a
// float that prints as "-0.00", and a term that contributes nothing is
// nothing whichever sign the multiplication happened to keep.
func noNegativeZero(f float64) float64 {
	if f == 0 {
		return 0
	}
	return f
}

// priorityLabel names the priority the weight came from, and says so when
// the task has none: that case has a weight of its own and is not the same
// as a priority the board no longer configures, which makes the task
// unreadable (docs/spec/modelo-de-datos/urgencia.md).
func priorityLabel(priority string) string {
	if priority == "" {
		return "(none)"
	}
	return priority
}

// activeLabel is the label of the `active` term, which says which of the
// two reasons applies when the term contributes nothing
// (docs/spec/cmd/get.md#salida).
func activeLabel(reason string) string {
	switch reason {
	case model.NotActive:
		return "not active"
	case model.ActiveButWaiting:
		return "active, waiting"
	}
	return "active"
}

// failWithCandidates is fail for a command that resolved a reference: an
// ambiguous one prints its candidates the way `biso ls` would print them,
// which is the ending docs/spec/referencias.md#la-búsqueda-por-texto fixes
// for a text that matches more than one task.
//
// With --json, `biso get` is the one command that answers a data envelope
// there, `task.candidates`, with the exit code 5 all the same: the table of
// docs/spec/contrato-json.md#el-sobre gives that kind to `get` and to no
// other command, so the rest answer the ordinary error envelope.
// extra is the warnings a call had already produced before the error, such
// as `biso get`'s own r.warnings when the reference it resolved by text
// named no task because the only one that matched could not be read
// (docs/spec/garantias.md#qué-hace-cada-comando): the tasks a set read left
// out are named whether or not the reference resolves.
func failWithCandidates(s Streams, p *Parsed, env ops.Env, asJSON bool, err error, extra []Warning) int {
	var ambiguous *ops.AmbiguousRef
	if !errors.As(err, &ambiguous) || ambiguous.Listing == nil {
		return fail(s, asJSON, err, append(warningsOf(p), extra...))
	}
	listing := ambiguous.Listing

	if asJSON && p.Command == "get" {
		printWarnings(s, p)
		for _, w := range extra {
			printWarning(s, w)
		}
		printOpsWarnings(s, listing.Warnings)
		tasks := make([]map[string]any, 0, len(listing.Tasks))
		for _, v := range listing.Tasks {
			tasks = append(tasks, taskObject(v))
		}
		writeEnvelope(s, env, "task.candidates", map[string]any{"tasks": tasks})
		return ambiguous.Err.ExitCode
	}

	warnings := append(warningsOf(p), extra...)
	for _, w := range listing.Warnings {
		warnings = append(warnings, Warning{
			Code: w.Code, Message: w.Message, Hints: w.Hints, Fields: w.Fields,
		})
	}
	code := fail(s, asJSON, err, warnings)
	if !asJSON {
		// The candidates are data, so they go on stdout, under the error
		// line that stderr already carries.
		fmt.Fprint(s.Stdout, renderList(listing.Tasks))
	}
	return code
}
