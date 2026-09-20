package cli

import (
	"fmt"
	"strconv"
	"strings"

	"biso/internal/ops"
)

// This file is the translation of the two writing commands that exist
// today: the analyzed command line into their typed parameters, and their
// result into the literal text of docs/spec/cmd/new.md#salida and
// docs/spec/cmd/set.md#salida or into the task.write envelope of
// docs/spec/contrato-json.md.

func runNew(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	params, err := newParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.New(env, params)
	if err != nil {
		return failWrite(s, p, asJSON, err, result)
	}
	printWriteWarnings(s, p, result)

	switch {
	case asJSON:
		writeEnvelope(s, env, "task.write", writeData(result))
	default:
		// The default output of `biso new` is one line per task created,
		// with the identifier and nothing else: it is the one writing
		// command that does not print the status line, because the three
		// derived data of that line say nothing about a task that has just
		// been born (docs/spec/cmd/new.md#salida).
		for _, t := range result.Tasks {
			fmt.Fprintln(s.Stdout, t.ID)
		}
	}
	if result.DryRun {
		// A preview writes nothing, so there is no identifier to print and
		// the whole answer is this line on stderr
		// (docs/spec/cmd/new.md#--dry-run-sobre-una-sola-tarea).
		fmt.Fprintf(s.Stderr, "1 task would be created, nothing was written (--dry-run)\n")
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	return 0
}

func runSet(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	params, err := setParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.Set(env, params)
	if err != nil {
		return failWrite(s, p, asJSON, err, result)
	}
	printWriteWarnings(s, p, result)

	switch {
	case asJSON:
		writeEnvelope(s, env, "task.write", writeData(result))
	case p.Has("quiet"):
		// --quiet reduces stdout to the identifiers a write affected, one
		// per line (docs/spec/cmd/flags-globales.md).
		for _, t := range result.Tasks {
			fmt.Fprintln(s.Stdout, t.ID)
		}
	default:
		if result.DryRun {
			fmt.Fprintln(s.Stdout, dryRunHeader(len(result.Tasks)))
		}
		for _, t := range result.Tasks {
			fmt.Fprintln(s.Stdout, statusLine(t))
		}
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	return 0
}

// dryRunHeader is the line that marks the status lines under it as
// hypothetical, in the singular when exactly one task would be affected
// (docs/spec/cmd/set.md#salida).
func dryRunHeader(n int) string {
	if n == 1 {
		return "1 task would be affected (--dry-run)"
	}
	return fmt.Sprintf("%d tasks would be affected (--dry-run)", n)
}

// statusLine is the line of docs/spec/cmd/set.md#salida, with its four
// rules: the progress of the criteria only when the task has any, the keys
// of the criteria the call just created only when it created some, and the
// word `archived` only when the task ends up archived.
func statusLine(t ops.TaskWrite) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s  %s", t.ID, t.Status)
	if t.AcTotal > 0 {
		fmt.Fprintf(&b, "  ac %d/%d", t.AcDone, t.AcTotal)
	}
	fmt.Fprintf(&b, "  urgency %s", formatUrgency(t.Urgency))
	if len(t.AcAdded) > 0 {
		keys := make([]string, 0, len(t.AcAdded))
		for _, key := range t.AcAdded {
			keys = append(keys, "#"+strconv.Itoa(key))
		}
		fmt.Fprintf(&b, "  added ac %s", strings.Join(keys, ", "))
	}
	if t.Archived {
		b.WriteString("  archived")
	}
	return b.String()
}

// formatUrgency writes the one derived number of the line with the single
// decimal digit of docs/spec/contrato-json.md#números-fechas-y-ausencias, so
// that a whole number is printed as 19.0 and never as 19.
func formatUrgency(u float64) string { return strconv.FormatFloat(u, 'f', 1, 64) }

// newParams turns the analyzed call into the parameters of `biso new`.
func newParams(p *Parsed) (ops.NewParams, error) {
	if len(p.Positionals) > 1 {
		return ops.NewParams{}, errUnexpectedArgument(p.Positionals[1])
	}
	params := ops.NewParams{
		Start:   p.Has("start"),
		Changes: changesOf(p),
		DryRun:  p.Has("dry-run"),
	}
	if len(p.Positionals) == 1 {
		params.Title, params.HasTitle = p.Positionals[0], true
	}
	return params, nil
}

// setParams turns the analyzed call into the parameters of `biso set`.
func setParams(p *Parsed) (ops.SetParams, error) {
	params := ops.SetParams{
		Refs:    p.Positionals,
		Changes: changesOf(p),
		DryRun:  p.Has("dry-run"),
	}
	switch {
	case p.Has("id"):
		params.Mode = ops.RefID
	case p.Has("match"):
		params.Mode = ops.RefText
	}
	return params, nil
}

// changesOf turns the values the parser classified into the changes the
// engine of internal/ops applies. The steps line up one for one, because
// both tables are the same section of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura.
func changesOf(p *Parsed) []ops.Change {
	var out []ops.Change
	for _, c := range p.Changes() {
		out = append(out, ops.Change{
			Flag:  c.Flag.Name,
			Step:  ops.Step(c.Category),
			Key:   c.Key,
			Value: c.Value,
		})
	}
	// A --replace-* given an empty value leaves the field empty, like its
	// --clear-* sibling, and produces no occurrence of its own
	// (docs/spec/valores-de-entrada.md#el-valor-vacío).
	for _, name := range p.EmptiedFlags() {
		out = append(out, ops.Change{Flag: name, Step: ops.Step(Replace)})
	}
	// --comment-author writes no field of its own, so it is not a change,
	// but the engine needs it to know who signs the comments of this call.
	if author, ok := p.Value("comment-author"); ok {
		out = append(out, ops.Change{Flag: "comment-author", Value: author})
	}
	return out
}

// printWriteWarnings prints the warnings of the command line and then the
// ones the write itself produced, in that order, which is the order they
// happened in.
func printWriteWarnings(s Streams, p *Parsed, result *ops.WriteResult) {
	printWarnings(s, p)
	if result == nil {
		return
	}
	for _, w := range result.Warnings {
		fmt.Fprint(s.Stderr, prefixed("warning: ", w.Message))
		for _, hint := range w.Hints {
			fmt.Fprint(s.Stderr, prefixed("hint: ", hint))
		}
	}
}

// failWrite is fail with the warnings a failed write had already produced,
// which never disappear with the error
// (docs/spec/salida-y-terminal.md#notas-y-avisos).
func failWrite(s Streams, p *Parsed, asJSON bool, err error, result *ops.WriteResult) int {
	warnings := warningsOf(p)
	if result != nil {
		for _, w := range result.Warnings {
			warnings = append(warnings, Warning{
				Code: w.Code, Message: w.Message, Hints: w.Hints, Fields: w.Fields,
			})
		}
	}
	return fail(s, asJSON, err, warnings)
}

// The task.write envelope of docs/spec/cmd/set.md#el-esquema-json, shared by
// every command that writes a task so that whoever consumes the output does
// not have to tell which verb produced it.
type writeEnvelopeData struct {
	Tasks    []writeEnvelopeTask `json:"tasks"`
	Warnings []map[string]any    `json:"warnings"`
}

type writeEnvelopeTask struct {
	ID      string   `json:"id"`
	Status  string   `json:"status"`
	AcDone  int      `json:"acDone"`
	AcTotal int      `json:"acTotal"`
	Urgency urgency  `json:"urgency"`
	Changed []string `json:"changed"`
	AcAdded []int    `json:"acAdded"`
}

// urgency is a float with exactly one digit after the point, which is what
// docs/spec/contrato-json.md#números-fechas-y-ausencias fixes for this
// field and what the default marshalling of a whole number would lose.
type urgency float64

func (u urgency) MarshalJSON() ([]byte, error) {
	return []byte(formatUrgency(float64(u))), nil
}

func writeData(r *ops.WriteResult) writeEnvelopeData {
	data := writeEnvelopeData{
		Tasks:    []writeEnvelopeTask{},
		Warnings: []map[string]any{},
	}
	for _, t := range r.Tasks {
		changed, added := t.Changed, t.AcAdded
		if changed == nil {
			changed = []string{}
		}
		if added == nil {
			added = []int{}
		}
		data.Tasks = append(data.Tasks, writeEnvelopeTask{
			ID:      t.ID,
			Status:  t.Status,
			AcDone:  t.AcDone,
			AcTotal: t.AcTotal,
			Urgency: urgency(t.Urgency),
			Changed: changed,
			AcAdded: added,
		})
	}
	for _, w := range r.Warnings {
		object := map[string]any{"code": w.Code}
		for key, value := range w.Fields {
			object[key] = value
		}
		data.Warnings = append(data.Warnings, object)
	}
	return data
}
