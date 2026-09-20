package cli

import (
	"fmt"
	"strconv"
	"strings"

	"biso/internal/model"
	"biso/internal/ops"
)

// This file is the output of docs/spec/cmd/ls.md: the eight columns, the
// two shortened shapes (--ids and --count) and the task.list envelope.

func runList(s Streams, p *Parsed, env ops.Env) int {
	asJSON := p.Has("json")
	params, err := listParams(p)
	if err != nil {
		return fail(s, asJSON, err, warningsOf(p))
	}
	result, err := ops.List(env, params)
	if err != nil {
		return failWithCandidates(s, p, env, asJSON, err)
	}
	printWarnings(s, p)
	printOpsWarnings(s, result.Warnings)

	switch {
	case asJSON:
		writeEnvelope(s, env, "task.list", listData(result))
	case params.Count:
		fmt.Fprintln(s.Stdout, result.Matched)
	case params.IDs:
		for _, v := range result.Tasks {
			fmt.Fprintln(s.Stdout, v.Task.ID)
		}
	default:
		fmt.Fprint(s.Stdout, renderList(result.Tasks))
	}
	for _, note := range result.Notes {
		printNote(s, p, note)
	}
	return 0
}

// listParams turns the analyzed call into the typed parameters of the
// command.
func listParams(p *Parsed) (ops.ListParams, error) {
	if len(p.Positionals) > 0 {
		// `biso ls` takes no positional argument: a bare word is almost
		// always a filter whose flag was forgotten.
		return ops.ListParams{}, errUnexpectedArgument(p.Positionals[0])
	}
	params := ops.ListParams{
		Filters:   filtersOf(p),
		HasStatus: p.Has("status"),
		Mine:      p.Has("mine"),
		Reverse:   p.Has("reverse"),
		All:       p.Has("all"),
		IDs:       p.Has("ids"),
		Count:     p.Has("count"),
	}
	params.Archived = p.Has("archived")
	params.OnlyArchived = p.Has("only-archived")
	if v, ok := p.Value("sort"); ok {
		params.Sort = v
	}
	if v, ok := p.Value("due-before"); ok {
		if _, err := ops.ParseCalendarDay("due-before", v); err != nil {
			return ops.ListParams{}, err
		}
		params.DueBefore = &v
	}
	if v, ok := p.Value("limit"); ok {
		limit, err := parseLimit(v)
		if err != nil {
			return ops.ListParams{}, err
		}
		params.Limit, params.HasLimit = limit, true
	}
	return params, nil
}

// parseLimit reads the value of --limit. `biso ls` and `biso prime` take
// the flag with the same meaning, so a negative or malformed value answers
// the same error in both (docs/spec/cmd/prime.md#qué-hace-caso-a-caso).
func parseLimit(v string) (int, error) {
	limit, err := strconv.Atoi(v)
	if err != nil || limit < 0 {
		return 0, &model.Error{
			ExitCode: 2,
			Code:     "invalid_number",
			Message:  fmt.Sprintf("--limit: not a whole number of rows: %q", v),
			Hints:    []string{"a limit is zero or more"},
			Field:    "limit",
			Given:    v,
		}
	}
	return limit, nil
}

// filtersOf reads the filters of docs/spec/cmd/ls.md off the analyzed call.
// `biso export` takes the same ones, so they are read in one place: the two
// about the archive are not here, because each of the two commands has its
// own (docs/spec/cmd/export.md).
func filtersOf(p *Parsed) ops.Filters {
	f := ops.Filters{
		Status:     p.Values("status"),
		NotStatus:  p.Values("not-status"),
		AnyStatus:  p.Has("any-status"),
		Type:       p.Values("type"),
		Priority:   p.Values("priority"),
		Label:      p.Values("label"),
		LabelOr:    p.Values("label-or"),
		Assignee:   p.Values("assignee"),
		Unassigned: p.Has("unassigned"),
		Blocked:    either(p, "blocked", "not-blocked"),
		Waiting:    either(p, "waiting", "not-waiting"),
		Active:     either(p, "active", "not-active"),
		Overdue:    p.Has("overdue"),
		Unchecked:  p.Has("unchecked"),
	}
	if v, ok := p.Value("parent"); ok {
		f.Parent = &v
	}
	if v, ok := p.Value("search"); ok {
		f.Search = &v
	}
	return f
}

// either resolves a pair of opposite switches into the three values of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls: true, false, and null
// for "neither of the two was asked for".
func either(p *Parsed, yes, no string) *bool {
	switch {
	case p.Has(yes):
		t := true
		return &t
	case p.Has(no):
		f := false
		return &f
	}
	return nil
}

// renderList is the two steps of docs/spec/cmd/ls.md#salida: every title is
// cut to a hundred cells first, and only then is each column padded to the
// widest value it holds among the rows this call prints.
//
// The width of a column is therefore a property of the call and not of the
// board, which is why the example of that page pads column 5 to twenty
// eight cells and not to a fixed number.
func renderList(views []ops.TaskView) string {
	rows := make([][]string, 0, len(views))
	for _, v := range views {
		rows = append(rows, listRow(v))
	}
	widths := columnWidths(rows)

	var b strings.Builder
	for _, row := range rows {
		b.WriteString(renderRow(row, widths))
		b.WriteString("\n")
	}
	return b.String()
}

// columnWidths is step 2 of that algorithm: each column padded to the
// widest value it holds among the rows given. `biso prime` computes them
// over the rows of its four blocks together, which is why this is a
// function of a set of rows and not of one listing
// (docs/spec/cmd/prime.md#la-salida-literal).
func columnWidths(rows [][]string) []int {
	widths := make([]int, columns)
	for _, row := range rows {
		// Column 8 is never padded, because it is the last one and there
		// is nothing after it to line up.
		for i := 0; i < columns-1; i++ {
			if w := cells(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

// renderRow writes one row padded to those widths, with no line break of
// its own.
func renderRow(row []string, widths []int) string {
	var b strings.Builder
	for i, cell := range row {
		if i > 0 {
			b.WriteString(columnGap)
		}
		b.WriteString(cell)
		if i < columns-1 {
			b.WriteString(strings.Repeat(" ", widths[i]-cells(cell)))
		}
	}
	return b.String()
}

// columns is how many the listing has, and columnGap the two literal spaces
// that always separate two of them, padded or not.
const (
	columns   = 8
	columnGap = "  "
	// dash is what an empty cell prints.
	dash = "-"
)

// listRow is one line of the table: the eight columns of
// docs/spec/cmd/ls.md#salida, in order.
func listRow(v ops.TaskView) []string {
	t := v.Task
	return []string{
		t.ID,
		t.Status,
		orDash(t.Type),
		orDash(t.Priority),
		cutTitle(t.Title),
		criteriaCell(t),
		assigneeCell(t),
		orDash(dueCell(t)),
	}
}

func orDash(s string) string {
	if s == "" {
		return dash
	}
	return s
}

// criteriaCell is column 6, and it is a dash and not "ac 0/0" when the task
// has no acceptance criteria: a progress of nothing out of nothing would
// read as a task that has some and has done none.
func criteriaCell(t *model.Task) string {
	if t.AcTotal() == 0 {
		return dash
	}
	return fmt.Sprintf("ac %d/%d", t.AcDone(), t.AcTotal())
}

// assigneeCell is column 7: the first person the task is assigned to, and
// how many more there are.
func assigneeCell(t *model.Task) string {
	if len(t.Assignees) == 0 {
		return dash
	}
	if len(t.Assignees) == 1 {
		return t.Assignees[0]
	}
	return fmt.Sprintf("%s+%d", t.Assignees[0], len(t.Assignees)-1)
}

func dueCell(t *model.Task) string {
	if t.Due.IsZero() {
		return ""
	}
	return t.Due.UTC().Format(model.DateLayout)
}

// listData is the task.list envelope of docs/spec/cmd/ls.md#el-esquema-json.
//
// data.filters is the parameters marshalled as they are: the tags of
// ops.Filters are the table of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls, so there is no
// translation here that could disagree with the filter that really ran.
func listData(r *ops.ListResult) map[string]any {
	tasks := make([]map[string]any, 0, len(r.Tasks))
	for _, v := range r.Tasks {
		tasks = append(tasks, taskObject(v))
	}
	warnings := make([]map[string]any, 0, len(r.Warnings))
	for _, w := range r.Warnings {
		object := map[string]any{"code": w.Code}
		for key, value := range w.Fields {
			object[key] = value
		}
		warnings = append(warnings, object)
	}
	return map[string]any{
		"tasks":     tasks,
		"shown":     r.Shown,
		"matched":   r.Matched,
		"hidden":    r.Hidden,
		"truncated": r.Truncated,
		"skipped":   list(r.Skipped),
		"sort":      r.Sort,
		"filters":   r.Filters,
		"warnings":  warnings,
	}
}

// printOpsWarnings prints the warnings a command's logic produced, after
// the ones the command line produced, which is the order they happened in.
func printOpsWarnings(s Streams, warnings []ops.Warning) {
	for _, w := range warnings {
		fmt.Fprint(s.Stderr, prefixed("warning: ", w.Message))
		for _, hint := range w.Hints {
			fmt.Fprint(s.Stderr, prefixed("hint: ", hint))
		}
	}
}
