package cli

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the translation of the two writing commands that exist
// today: the analyzed command line into their typed parameters, and their
// result into the literal text of docs/spec/cmd/new.md#salida and
// docs/spec/cmd/set.md#salida or into the task.write envelope of
// docs/spec/contrato-json.md.

func runNew(s Streams, p *Parsed, env ops.Env) int {
	if p.Has("from") {
		return runNewBatch(s, p, env)
	}
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
	case p.Has("print"):
		// --print replaces the default output with the whole card of
		// every task the call affected, and never adds it under the
		// line: the three data of that line are inside the card already
		// (docs/spec/cmd/flags-globales.md).
		printCards(s, result.Views)
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

// runNewBatch is `biso new --from`: the same command fed from NDJSON
// instead of from flags (docs/spec/cmd/new.md#el-modo-lote).
func runNewBatch(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	if len(p.Positionals) > 0 {
		// The table of docs/spec/cmd/new.md declares --from incompatible
		// with the title, which is a positional argument and therefore not
		// a row of the flag table the parser walks.
		return fail(s, asJSON, &model.Error{
			ExitCode: 2,
			Code:     "incompatible_flags",
			Message:  "--from and a title cannot be used together",
			Hints:    []string{"in a batch every field of every task travels in the file"},
		}, warningsOf(p))
	}
	if p.Has("print") {
		// A batch creates every task of the file from scratch in a board
		// that may have none, so there is no task that existed before to
		// print a card of, which is the same reason --print is a usage
		// error in `biso init --from` (docs/spec/cmd/new.md#el-modo-lote).
		return fail(s, asJSON, &model.Error{
			ExitCode: 2,
			Code:     "read_only_flag",
			Message:  "--print does not apply to a batch, which affects no task that existed before",
			Field:    "print",
		}, warningsOf(p))
	}
	from, _ := p.Value("from")
	content, err := readBatchSource(s, from)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.NewBatch(env, ops.BatchParams{
		Content: content, DryRun: p.Has("dry-run"),
	})
	if err != nil {
		return failWrite(s, p, asJSON, err, result)
	}
	printWriteWarnings(s, p, result)

	if asJSON {
		writeEnvelope(s, env, "task.write", writeData(result))
	} else {
		for _, t := range result.Tasks {
			fmt.Fprintln(s.Stdout, t.ID)
		}
	}
	if result.DryRun {
		fmt.Fprintf(s.Stderr, "%s would be created, nothing was written (--dry-run)\n",
			plural(result.Previewed, "task", "tasks"))
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	return 0
}

// readBatchSource reads what --from names: a file, or standard input when
// it is "-". It is not the @file of a text flag
// (docs/spec/valores-de-entrada.md#tres-formas-de-pasar-un-valor-largo):
// here the value is the path itself, so a file called @notes.ndjson is read
// by its own name.
func readBatchSource(s Streams, from string) (string, error) {
	if from == "-" {
		b, err := io.ReadAll(s.Stdin)
		if err != nil {
			return "", &model.Error{
				ExitCode: 8,
				Code:     "io_error",
				Message:  "--from: stdin cannot be read: " + err.Error(),
				Field:    "from",
				Given:    "-",
			}
		}
		return checkedUTF8(string(b), "-")
	}
	b, err := os.ReadFile(from)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", &model.Error{
				ExitCode: 4,
				Code:     "file_not_found",
				Message:  "--from: file not found: " + from,
				Field:    "from",
				Given:    from,
			}
		}
		return "", &model.Error{
			ExitCode: 8,
			Code:     "file_unreadable",
			Message:  "--from: file cannot be read: " + from,
			Field:    "from",
			Given:    from,
		}
	}
	return checkedUTF8(string(b), from)
}

// checkedUTF8 applies to a batch the rule every input of the program
// follows: the input is UTF-8 always, and a byte that is not is exit code 3
// pointing at it (docs/spec/salida-y-terminal.md#codificación-y-texto).
func checkedUTF8(content, given string) (string, error) {
	if offset := invalidUTF8At(content); offset >= 0 {
		return "", &model.Error{
			ExitCode: 3,
			Code:     "invalid_encoding",
			Message:  fmt.Sprintf("--from: invalid UTF-8 at byte %d", offset),
			Field:    "from",
			Given:    given,
		}
	}
	return content, nil
}

func runSet(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		return ops.Set(env, setParams(p))
	})
}

// The six verbs of the cycle. Each one is the same three lines: the
// analyzed call into its typed parameters, the call into internal/ops, and
// the shared printing of runWrite, because all seven writing commands over
// an existing task print the status line of docs/spec/cmd/set.md#salida.

func runStart(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		return ops.Start(env, ops.StartParams{
			Refs: p.Positionals, Mode: refMode(p), Reopen: p.Has("reopen"),
			Changes: changesOf(p), DryRun: p.Has("dry-run"), Print: p.Has("print"),
		})
	})
}

func runNote(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		params, err := oneRefParams(p, "note")
		if err != nil {
			return nil, err
		}
		return ops.Note(env, params)
	})
}

func runComment(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		params, err := oneRefParams(p, "comment")
		if err != nil {
			return nil, err
		}
		return ops.Comment(env, params)
	})
}

func runAsk(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		params, err := oneRefParams(p, "ask")
		if err != nil {
			return nil, err
		}
		return ops.Ask(env, params)
	})
}

func runAnswer(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		params, err := oneRefParams(p, "answer")
		if err != nil {
			return nil, err
		}
		return ops.Answer(env, params)
	})
}

func runFinish(s Streams, p *Parsed, env ops.Env) int {
	return runWrite(s, p, env, func() (*ops.WriteResult, error) {
		return ops.Finish(env, ops.FinishParams{
			Refs: p.Positionals, Mode: refMode(p),
			Strict: p.Has("strict"), NoChecks: p.Has("no-checks"),
			Changes: changesOf(p), DryRun: p.Has("dry-run"), Print: p.Has("print"),
		})
	})
}

// runWrite is the printing every writing command over an existing task
// shares: the warnings, then the status line or whatever --json, --quiet and
// --print ask for instead, then the notes.
func runWrite(s Streams, p *Parsed, env ops.Env, run func() (*ops.WriteResult, error)) int {
	asJSON := p.Has("json")
	result, err := run()
	if err != nil {
		return failWriteWithCandidates(s, p, env, asJSON, err, result)
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
	case p.Has("print"):
		if result.DryRun {
			fmt.Fprintln(s.Stdout, dryRunHeader(len(result.Tasks)))
		}
		printCards(s, result.Views)
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

// oneRefParams is the shape of the four verbs that take exactly one
// reference and read the rest of the positionals as text
// (docs/spec/cmd/verbos-del-ciclo.md).
func oneRefParams(p *Parsed, command string) (ops.NoteParams, error) {
	if len(p.Positionals) == 0 {
		return ops.NoteParams{}, &model.Error{
			ExitCode: 2,
			Code:     "missing_ref",
			Message:  "biso " + command + " needs one task reference",
			Hints:    []string{"biso " + command + " MYP-11 \"...\""},
		}
	}
	texts := make([]ops.Text, 0, len(p.Texts))
	for _, t := range p.Texts {
		texts = append(texts, ops.Text{Typed: t.Typed, Value: t.Value})
	}
	return ops.NoteParams{
		Ref: p.Positionals[0], Mode: refMode(p), Texts: texts,
		Changes: changesOf(p), DryRun: p.Has("dry-run"), Print: p.Has("print"),
	}, nil
}

// refMode is --id and --match, the two flags that force how a positional
// reference is read (docs/spec/referencias.md#la-gramática).
func refMode(p *Parsed) ops.RefMode {
	switch {
	case p.Has("id"):
		return ops.RefID
	case p.Has("match"):
		return ops.RefText
	}
	return ops.RefAuto
}

// printCards writes the whole card of every task a write affected, which
// is what --print asks for. The cards are separated by a blank line, so
// that a call over several tasks does not run two of them together.
func printCards(s Streams, views []ops.TaskView) {
	for i, v := range views {
		if i > 0 {
			fmt.Fprintln(s.Stdout)
		}
		fmt.Fprint(s.Stdout, renderCard(&ops.GetResult{
			Task: v, Sections: ops.Sections, WholeCard: true,
		}))
	}
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
		Print:   p.Has("print"),
	}
	if len(p.Positionals) == 1 {
		params.Title, params.HasTitle = p.Positionals[0], true
	}
	return params, nil
}

// setParams turns the analyzed call into the parameters of `biso set`.
func setParams(p *Parsed) ops.SetParams {
	return ops.SetParams{
		Refs:    p.Positionals,
		Mode:    refMode(p),
		Changes: changesOf(p),
		DryRun:  p.Has("dry-run"),
		Print:   p.Has("print"),
	}
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
		printWarning(s, cliWarning(w))
	}
}

// cliWarning is one warning of internal/ops as this layer prints it and
// folds it into an envelope.
func cliWarning(w ops.Warning) Warning {
	return Warning{
		Code: w.Code, Message: w.Message, Hints: w.Hints,
		Fields: w.Fields, Detail: w.Detail,
	}
}

// failWriteWithCandidates is failWrite for a command that resolved a
// reference: the candidates of an ambiguous one are printed the way
// `biso ls` prints a listing (docs/spec/referencias.md).
func failWriteWithCandidates(s Streams, p *Parsed, env ops.Env, asJSON bool,
	err error, result *ops.WriteResult) int {
	var ambiguous *ops.AmbiguousRef
	if errors.As(err, &ambiguous) {
		return failWithCandidates(s, p, env, asJSON, err)
	}
	return failWrite(s, p, asJSON, err, result)
}

// failWrite is fail with the warnings a failed write had already produced,
// which never disappear with the error
// (docs/spec/salida-y-terminal.md#notas-y-avisos).
func failWrite(s Streams, p *Parsed, asJSON bool, err error, result *ops.WriteResult) int {
	warnings := warningsOf(p)
	if result != nil {
		for _, w := range result.Warnings {
			warnings = append(warnings, cliWarning(w))
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
