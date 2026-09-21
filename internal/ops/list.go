package ops

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"biso/internal/board"
	"biso/internal/match"
	"biso/internal/model"
)

// This file is docs/spec/cmd/ls.md: the filters, the rule of order and the
// limit. It answers the tasks to print and the effective filter that
// produced them; the columns and the JSON envelope are internal/cli's job.

// DefaultLimit is how many rows `biso ls` prints when nothing at all says
// otherwise: not the call, not BISO_LIMIT and not the machine's
// `default_limit` key (docs/spec/cmd/ls.md). It is the last rung of the
// precedence of docs/spec/invocacion.md#variables-de-entorno, and the same
// number board.DefaultLimit fills that key in with.
const DefaultLimit = board.DefaultLimit

// Filters is the filter of one `biso ls` call. The same struct carries what
// the caller typed and, once resolved, what really produced the listing,
// which is what `data.filters` of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls asks for: not an echo of
// the flags, but the effective filter, defaults resolved included.
//
// The JSON tags are that table, one for one. They live on this struct and
// not on a second one in the rendering layer on purpose: a translation
// written by hand between the parameters and the envelope is exactly where
// the two would drift apart.
type Filters struct {
	Status       []string `json:"status"`
	NotStatus    []string `json:"notStatus"`
	AnyStatus    bool     `json:"anyStatus"`
	Archived     bool     `json:"archived"`
	OnlyArchived bool     `json:"onlyArchived"`
	Type         []string `json:"type"`
	Priority     []string `json:"priority"`
	Label        []string `json:"label"`
	LabelOr      []string `json:"labelOr"`
	Assignee     []string `json:"assignee"`
	Unassigned   bool     `json:"unassigned"`
	Parent       *string  `json:"parent"`
	Blocked      *bool    `json:"blocked"`
	Waiting      *bool    `json:"waiting"`
	Active       *bool    `json:"active"`
	Overdue      bool     `json:"overdue"`
	DueBefore    *string  `json:"dueBefore"`
	Search       *string  `json:"search"`
	Unchecked    bool     `json:"unchecked"`
}

// ListParams is one `biso ls` call, already read off the command line: the
// filters plus the flags that shape the answer instead of narrowing it.
//
// The shaping flags carry no JSON tag because they are not filters and
// `data.filters` does not have them.
type ListParams struct {
	Filters

	// HasStatus says the call wrote --status at all, which is what tells an
	// explicit status apart from the default that excludes the terminal
	// one (docs/spec/cmd/ls.md).
	HasStatus bool `json:"-"`
	// Mine is resolved into Assignee, so the effective filter names the
	// identity that was used and not the flag that asked for it.
	Mine bool `json:"-"`

	Sort     string `json:"-"`
	Reverse  bool   `json:"-"`
	Limit    int    `json:"-"`
	HasLimit bool   `json:"-"`
	All      bool   `json:"-"`
	IDs      bool   `json:"-"`
	Count    bool   `json:"-"`
}

// SortFields are the values --sort admits, in the order of the table of
// docs/spec/cmd/ls.md. The empty string is not one of them: it is the
// absence of the flag, which is the default order.
var SortFields = []string{"urgency", "id", "ordinal", "due", "updated", "created", "title"}

// DefaultSort is what `data.sort` says when no --sort was written.
const DefaultSort = "default"

// ListResult is what `biso ls` answers.
type ListResult struct {
	// Tasks are the ones to print, already filtered, ordered and cut to
	// the limit.
	Tasks []TaskView

	Shown     int
	Matched   int
	Hidden    int
	Truncated bool
	Skipped   []string
	Sort      string
	Filters   Filters

	Warnings []Warning
	Notes    []string
}

// List lists the tasks of the board (docs/spec/cmd/ls.md).
func List(env Env, p ListParams) (*ListResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return ListOn(b, env, p)
}

// ListOn is List over a board that is already open.
func ListOn(b *board.Board, env Env, p ListParams) (*ListResult, error) {
	r := newReader(b, env)
	if err := r.load(); err != nil {
		return nil, err
	}

	filters, err := r.resolveFilters(p)
	if err != nil {
		return nil, err
	}

	sortField := p.Sort
	if sortField == "" {
		sortField = DefaultSort
	}

	var matched []TaskView
	for _, t := range r.all {
		if !r.matches(t, filters) {
			continue
		}
		v, viewErr := r.view(t, false)
		if viewErr != nil {
			// A task whose priority the board no longer configures is
			// the undecodable task of docs/spec/garantias.md: on a set
			// read it is skipped and named, never a reason to fail.
			r.undecodable(t, viewErr)
			continue
		}
		matched = append(matched, v)
	}
	r.sortTasks(matched, p)

	result := &ListResult{
		Matched: len(matched),
		Sort:    sortField,
		Filters: filters,
	}
	if !p.Count {
		result.Tasks = matched[:limitOf(p, env, len(matched))]
		result.Shown = len(result.Tasks)
		result.Hidden = result.Matched - result.Shown
		result.Truncated = result.Hidden > 0
	}
	if result.Truncated {
		result.Warnings = append(result.Warnings, truncationWarning(result.Shown, result.Matched))
	}
	// A filter that matched nothing is a fact about the board and not a
	// failure, so it is a note and the code stays 0
	// (docs/spec/cmd/ls.md#comportamiento-caso-a-caso). With --count there
	// is nothing to say: the zero it prints says it.
	if result.Matched == 0 && !p.Count {
		r.note("no tasks match")
	}

	r.warnAboutSkipped()
	result.Skipped = r.skippedIDs()
	result.Warnings = append(r.warnings, result.Warnings...)
	result.Notes = r.notes
	return result, nil
}

// limitOf is how many rows this call prints, reading the precedence of
// docs/spec/invocacion.md#variables-de-entorno from the top: --all asks for
// everything, --limit is the flag of this call, and the environment answers
// for the rest, with the built-in thirty behind it. Never more rows than
// there are.
func limitOf(p ListParams, env Env, matched int) int {
	limit := DefaultLimit
	if env.ListLimit != nil {
		limit = *env.ListLimit
	}
	switch {
	case p.All:
		return matched
	case p.HasLimit:
		limit = p.Limit
	}
	if limit > matched {
		return matched
	}
	return limit
}

// truncationWarning is the one of docs/spec/cmd/ls.md#salida, with the hint
// that names the three ways out.
func truncationWarning(shown, matched int) Warning {
	return Warning{
		Code: "list_truncated",
		Message: fmt.Sprintf("%d more tasks match; showing %d of %d",
			matched-shown, shown, matched),
		Hints: []string{
			"narrow with --status, --type or --label, or ask for everything with --all",
		},
		Fields: map[string]any{"shown": shown, "matched": matched},
	}
}

// resolveFilters turns what the caller typed into the effective filter:
// every closed vocabulary matched against the board's own spelling, --mine
// turned into the identity it used, --parent resolved to an identifier, and the
// default of --status spelled out.
//
// It is the whole of docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué,
// and it calls the one matching algorithm there is, so that the same text
// is worth the same when filtering as when writing.
func (r *reader) resolveFilters(p ListParams) (Filters, error) {
	f := p.Filters
	var err error

	if f.Status, err = r.matchAll(match.Status, f.Status, r.b.Config.Statuses); err != nil {
		return f, err
	}
	if f.NotStatus, err = r.matchAll(match.Status, f.NotStatus, r.b.Config.Statuses); err != nil {
		return f, err
	}
	if f.Type, err = r.matchAll(match.Type, f.Type, r.b.Config.Types); err != nil {
		return f, err
	}
	if f.Priority, err = r.matchAll(match.Priority, f.Priority, r.b.Config.Priorities); err != nil {
		return f, err
	}

	// Without an explicit --status the default is every status but the terminal
	// one, and --any-status is the only other way to bring it back. That
	// is the default value of the flag spelled out, not a rule laid on top
	// of it (docs/spec/cmd/ls.md).
	if !p.HasStatus {
		f.Status = r.everyStatus(f.AnyStatus)
	}

	// The form of a label filter is judged before the board is looked at,
	// and --unchecked does not turn it off: a malformed label is the shape
	// of the token and not a check against this board
	// (docs/spec/cmd/ls.md#comportamiento-caso-a-caso). Judging it here is
	// also what lets the key form travel on in one spelling below.
	for _, field := range []struct {
		name   string
		values *[]string
	}{{"label", &f.Label}, {"labelOr", &f.LabelOr}} {
		normalized, err := labelFilterValues(field.name, *field.values)
		if err != nil {
			return f, err
		}
		*field.values = normalized
	}

	if !f.Unchecked {
		if err := r.checkLabelFilter("label", f.Label); err != nil {
			return f, err
		}
		if err := r.checkLabelFilter("labelOr", f.LabelOr); err != nil {
			return f, err
		}
		if err := r.checkKnown("assignee", f.Assignee, r.assigneesOfTheBoard()); err != nil {
			return f, err
		}
	}

	if p.Mine {
		if r.env.Me == "" {
			return f, &model.Error{
				ExitCode: 6,
				Code:     "mine_requires_identity",
				Message: "--mine needs an identity; set BISO_ME, " +
					"or add \"me\" to ~/.biso/config.json",
				Field: "mine",
			}
		}
		f.Assignee = []string{r.env.Me}
	}

	if f.Parent != nil {
		// A reference that is the value of a flag has no --id and no
		// --match of its own: the grammar decides alone
		// (docs/spec/referencias.md#la-gramática).
		resolved, err := resolveRefWith(r.b, r.all, *f.Parent, RefAuto)
		if err != nil {
			return f, withCandidates(r.b, r.env, r.all, err)
		}
		if resolved.Note != "" {
			r.note(resolved.Note)
		}
		id := resolved.Task.ID
		f.Parent = &id
	}
	// Every list of the effective filter is a list and never the absence
	// of one: `[]` is how the contract writes a filter nobody asked for
	// (docs/spec/contrato-json.md#números-fechas-y-ausencias).
	for _, values := range []*[]string{
		&f.Status, &f.NotStatus, &f.Type, &f.Priority,
		&f.Label, &f.LabelOr, &f.Assignee,
	} {
		if *values == nil {
			*values = []string{}
		}
	}
	return f, nil
}

// everyStatus is the resolved default of --status: every status the board
// configures, with the terminal one left out unless --any-status asked for
// it.
func (r *reader) everyStatus(anyStatus bool) []string {
	out := make([]string, 0, len(r.b.Config.Statuses))
	for _, status := range r.b.Config.Statuses {
		if !anyStatus && status == r.b.Config.TerminalStatus {
			continue
		}
		out = append(out, status)
	}
	return out
}

// matchAll resolves every value of one filter against a configured
// vocabulary, answering the board's own spelling.
func (r *reader) matchAll(field match.Field, values, configured []string) ([]string, error) {
	if len(values) == 0 {
		return values, nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		resolved, err := match.Match(field, v, configured)
		if err != nil {
			return nil, err
		}
		out = append(out, resolved)
	}
	return out, nil
}

// checkKnown is the other half of that table: a label or a person the board
// does not have is exit code 3 with up to five of the closest ones, which
// --unchecked is the one thing that turns off
// (docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué).
//
// The comparison folds the case, because this is a reading filter and
// docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma says a
// reading filter does not distinguish it. A board whose only label is
// "Parser" knows the label "parser", and --label parser is a listing and not
// an unknown value.
func (r *reader) checkKnown(field string, values, known []string) *model.Error {
	noun := "label"
	if field == "assignee" {
		noun = "assignee"
	}
	for _, v := range values {
		if containsFold(known, v) {
			continue
		}
		e := &model.Error{
			ExitCode: 3,
			Code:     "unknown_" + noun,
			Message:  fmt.Sprintf("unknown %s: %q", noun, v),
			Field:    field,
			Given:    v,
		}
		if closest := match.Suggest(v, known, 5); len(closest) > 0 {
			e.Hints = []string{"did you mean: " + strings.Join(closest, ", ") + "?"}
		}
		return e
	}
	return nil
}

// labelsOfTheBoard is the set --label validates against: the labels the
// configuration declares plus the ones any task carries, archived and
// finished tasks included
// (docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué).
func (r *reader) labelsOfTheBoard() []string {
	return r.setOf(r.b.Config.Labels, func(t *model.Task) []string { return t.Labels })
}

// labelKeysOfTheBoard is the other set of that table, the one the key form
// of a filter validates against: the keys of the labels the configuration
// declares and of the ones any task carries, archived and finished included
// (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito).
// A declared key counts as known although no task uses it yet, which is
// exactly what makes declaring it useful.
func (r *reader) labelKeysOfTheBoard() []string {
	return sortedCopy(labelKeys(r.labelsOfTheBoard()))
}

// labelFilterValues judges the form of every value of --label or --label-or
// and answers them in the spelling the effective filter carries. The key
// form travels with a single colon however it was typed, because the two
// spellings are the same filter, and the key keeps the spelling it was
// typed in, because there is no configured one to resolve it to
// (docs/spec/contrato-json.md#los-filtros-de-biso-ls).
func labelFilterValues(field string, values []string) ([]string, *model.Error) {
	if len(values) == 0 {
		return values, nil
	}
	out := make([]string, 0, len(values))
	for _, v := range values {
		l, err := model.ParseLabelKeyOrLabel(v)
		if err != nil {
			err.Field = field
			return nil, err
		}
		if l.IsKey() {
			out = append(out, l.Key+model.LabelSeparatorSeveral)
			continue
		}
		out = append(out, v)
	}
	return out, nil
}

// checkLabelFilter is checkKnown for the two label filters, which validate
// against one of two sets depending on the form of the value: the key form
// against the keys of the board, and everything else against its labels,
// where the separator does not count, so --label milestone:m1 finds a task
// labelled milestone::m1
// (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito).
func (r *reader) checkLabelFilter(field string, values []string) *model.Error {
	if len(values) == 0 {
		return nil
	}
	labels := r.labelsOfTheBoard()
	keys := sortedCopy(labelKeys(labels))
	for _, v := range values {
		l := model.SplitLabel(v)
		if l.IsKey() {
			if containsFold(keys, l.Key) {
				continue
			}
			return unknownLabelKey(field, v, l.Key, keys)
		}
		if anyLabelMatches(labels, v) {
			continue
		}
		e := &model.Error{
			ExitCode: 3,
			Code:     "unknown_label",
			Message:  fmt.Sprintf("unknown label: %q", v),
			Field:    field,
			Given:    v,
		}
		if closest := match.Suggest(v, labels, 5); len(closest) > 0 {
			e.Hints = []string{"did you mean: " + strings.Join(closest, ", ") + "?"}
		}
		return e
	}
	return nil
}

// unknownLabelKey is the error of a key the board does not have, with up to
// five of the closest ones. Its `given` is the filter as it was typed, and
// its `valid` the keys of the board.
func unknownLabelKey(field, given, key string, keys []string) *model.Error {
	e := &model.Error{
		ExitCode: 3,
		Code:     "unknown_label_key",
		Message:  fmt.Sprintf("unknown label key: %q", key),
		Field:    field,
		Given:    given,
		Valid:    keys,
	}
	if closest := match.Suggest(key, keys, 5); len(closest) > 0 {
		e.Hints = []string{"did you mean: " + strings.Join(closest, ", ") + "?"}
	}
	return e
}

// anyLabelMatches answers whether any label of a set is what one filter
// value names, by the comparison of labelMatches.
func anyLabelMatches(known []string, wanted string) bool {
	for _, k := range known {
		if labelMatches(k, wanted) {
			return true
		}
	}
	return false
}

// assigneesOfTheBoard is the set --assignee validates against. The authors
// are not in it: there is no --author filter, so a person who only ever
// wrote a task and never had one assigned does not belong to it.
func (r *reader) assigneesOfTheBoard() []string {
	return r.setOf(r.b.Config.Assignees, func(t *model.Task) []string { return t.Assignees })
}

func (r *reader) setOf(configured []string, of func(*model.Task) []string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(values []string) {
		for _, v := range values {
			if !seen[v] {
				seen[v] = true
				out = append(out, v)
			}
		}
	}
	add(configured)
	for _, t := range r.all {
		add(of(t))
	}
	sort.Strings(out)
	return out
}

// matches answers whether one task passes the effective filter. Filters of
// different fields are ANDed and repeated values of the same field ORed,
// with --label the one exception that ANDs (docs/spec/cmd/ls.md).
func (r *reader) matches(t *model.Task, f Filters) bool {
	switch {
	case f.OnlyArchived && !t.Archived:
		return false
	case !f.OnlyArchived && !f.Archived && t.Archived:
		return false
	}
	if !containsString(f.Status, t.Status) {
		return false
	}
	if containsString(f.NotStatus, t.Status) {
		return false
	}
	if len(f.Type) > 0 && !containsString(f.Type, t.Type) {
		return false
	}
	if len(f.Priority) > 0 && !containsString(f.Priority, t.Priority) {
		return false
	}
	// The three filters over a label or a person fold the case, the same
	// way checkKnown does, and for the same reason: a label is stored
	// letter for letter and read without distinguishing
	// (docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma).
	for _, label := range f.Label {
		if !anyLabelMatches(t.Labels, label) {
			return false
		}
	}
	if len(f.LabelOr) > 0 && !anyLabelOf(t.Labels, f.LabelOr) {
		return false
	}
	if len(f.Assignee) > 0 && !anyOfFold(t.Assignees, f.Assignee) {
		return false
	}
	if f.Unassigned && len(t.Assignees) > 0 {
		return false
	}
	if f.Parent != nil && t.Parent != *f.Parent {
		return false
	}
	if f.Blocked != nil && r.blocked(t) != *f.Blocked {
		return false
	}
	if f.Waiting != nil && t.Waiting() != *f.Waiting {
		return false
	}
	if f.Active != nil && (t.Status == r.b.Config.ActiveStatus) != *f.Active {
		return false
	}
	if f.Overdue && !r.overdue(t) {
		return false
	}
	if f.DueBefore != nil && !dueBefore(t, *f.DueBefore) {
		return false
	}
	if f.Search != nil && !TaskMatchesText(t, *f.Search) {
		return false
	}
	return true
}

// overdue is "days < 0", the same count the proximity term of the urgency
// uses: a task that is due today has zero days left and is not overdue
// (docs/spec/cmd/ls.md).
func (r *reader) overdue(t *model.Task) bool {
	return !t.Due.IsZero() && t.Due.Before(r.today())
}

func dueBefore(t *model.Task, day string) bool {
	if t.Due.IsZero() {
		return false
	}
	limit, err := time.ParseInLocation(model.DateLayout, day, time.UTC)
	if err != nil {
		return false
	}
	return t.Due.Before(limit)
}

// sortTasks is the rule of order of docs/spec/cmd/ls.md#la-regla-de-orden-completa,
// whole: the default tuple, or the field --sort named, with the identifier
// breaking every tie and --reverse flipping the finished list.
func (r *reader) sortTasks(views []TaskView, p ListParams) {
	sort.SliceStable(views, func(i, j int) bool {
		a, b := views[i], views[j]
		if less, decided := lessBySort(a, b, p.Sort); decided {
			return less
		}
		return taskNumber(a.Task.ID) < taskNumber(b.Task.ID)
	})
	if p.Reverse {
		for i, j := 0, len(views)-1; i < j; i, j = i+1, j-1 {
			views[i], views[j] = views[j], views[i]
		}
	}
}

// lessBySort compares two tasks by the order that was asked for, and says
// whether the comparison decided anything: when it did not, the caller
// breaks the tie by identifier, always.
func lessBySort(a, b TaskView, field string) (bool, bool) {
	switch field {
	case "":
		// The default order: the tasks that have an ordinal come first,
		// ascending; the ones that do not come after, by urgency
		// descending.
		if (a.Task.Ordinal != nil) != (b.Task.Ordinal != nil) {
			return a.Task.Ordinal != nil, true
		}
		if a.Task.Ordinal != nil {
			return *a.Task.Ordinal < *b.Task.Ordinal, *a.Task.Ordinal != *b.Task.Ordinal
		}
		return a.Urgency > b.Urgency, a.Urgency != b.Urgency
	case "urgency":
		// Descending, because it is a measure of priority.
		return a.Urgency > b.Urgency, a.Urgency != b.Urgency
	case "id":
		return false, false
	case "ordinal":
		if (a.Task.Ordinal != nil) != (b.Task.Ordinal != nil) {
			return a.Task.Ordinal != nil, true
		}
		if a.Task.Ordinal == nil {
			return false, false
		}
		return *a.Task.Ordinal < *b.Task.Ordinal, *a.Task.Ordinal != *b.Task.Ordinal
	case "due":
		if a.Task.Due.IsZero() != b.Task.Due.IsZero() {
			return !a.Task.Due.IsZero(), true
		}
		return a.Task.Due.Before(b.Task.Due), !a.Task.Due.Equal(b.Task.Due)
	case "updated":
		return a.Task.UpdatedAt.Before(b.Task.UpdatedAt),
			!a.Task.UpdatedAt.Equal(b.Task.UpdatedAt)
	case "created":
		return a.Task.CreatedAt.Before(b.Task.CreatedAt),
			!a.Task.CreatedAt.Equal(b.Task.CreatedAt)
	case "title":
		// By code point and not by the collation rules of any language,
		// so that the same board comes out in the same order on every
		// machine (docs/spec/cmd/ls.md#la-regla-de-orden-completa).
		return a.Task.Title < b.Task.Title, a.Task.Title != b.Task.Title
	}
	return false, false
}

// taskNumber is the number of an identifier, which is what the tie-break
// compares: MYP-9 comes before MYP-10, and comparing the text would not.
func taskNumber(id string) int {
	n := 0
	rest := id
	if at := strings.LastIndexByte(id, '-'); at >= 0 {
		rest = id[at+1:]
	}
	for _, c := range rest {
		if c < '0' || c > '9' {
			return n
		}
		n = n*10 + int(c-'0')
	}
	return n
}

func containsString(values []string, v string) bool {
	for _, x := range values {
		if x == v {
			return true
		}
	}
	return false
}

// containsFold and anyOfFold are the membership tests of a reading filter
// over a label or a person: the same comparison as containsString with the
// case folded on both sides.
func containsFold(values []string, v string) bool {
	folded := match.FoldCase(v)
	for _, x := range values {
		if match.FoldCase(x) == folded {
			return true
		}
	}
	return false
}

func anyOfFold(values, wanted []string) bool {
	for _, w := range wanted {
		if containsFold(values, w) {
			return true
		}
	}
	return false
}

// anyLabelOf is anyOfFold for the labels of a task, with the comparison a
// scoped label brings: the OR of --label-or over the same rule --label uses.
func anyLabelOf(labels, wanted []string) bool {
	for _, w := range wanted {
		if anyLabelMatches(labels, w) {
			return true
		}
	}
	return false
}

// listing is a set of tasks printed the way `biso ls` would print them: the
// default order, the default limit and the same truncation warning. It is
// what docs/spec/referencias.md asks for when a text matches several tasks,
// so that the candidates of an exit code 5 are not a second listing format.
func (r *reader) listing(tasks []*model.Task) *ListResult {
	var views []TaskView
	for _, t := range tasks {
		v, err := r.view(t, false)
		if err != nil {
			r.undecodable(t, err)
			continue
		}
		views = append(views, v)
	}
	r.sortTasks(views, ListParams{})

	result := &ListResult{Matched: len(views), Sort: DefaultSort}
	// Thirty, and not the limit this machine configured: what
	// docs/spec/referencias.md promises about a list of candidates is the
	// order, the limit of thirty and the truncation warning, and a
	// reference resolves the same on every machine.
	result.Tasks = views[:min(DefaultLimit, len(views))]
	result.Shown = len(result.Tasks)
	result.Hidden = result.Matched - result.Shown
	result.Truncated = result.Hidden > 0
	if result.Truncated {
		result.Warnings = append(result.Warnings, truncationWarning(result.Shown, result.Matched))
	}
	return result
}

// withCandidates fills in the listing of an ambiguous reference, so that
// the layer that prints it has the candidates already ordered and cut, and
// answers every other error untouched.
//
// It exists because the resolution of a reference is a function of
// internal/ops that knows nothing about the urgency or the board's
// configuration, and the order of that listing is the order of `biso ls`,
// which needs both.
func withCandidates(b *board.Board, env Env, all []*model.Task, err error) error {
	var ambiguous *AmbiguousRef
	if !errors.As(err, &ambiguous) {
		return err
	}
	r := newReader(b, env)
	r.all = all
	r.byID = make(map[string]*model.Task, len(all))
	for _, t := range all {
		r.byID[t.ID] = t
	}
	ambiguous.Listing = r.listing(ambiguous.Candidates)
	return err
}
