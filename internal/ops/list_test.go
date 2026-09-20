package ops

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
)

// These are the rules of docs/spec/cmd/ls.md that are about the answer and
// not about its shape: the order, the limit and the filters. What the rows
// look like is checked against the specification's own block in
// cmd/biso/read_golden_test.go, over the compiled program.

// ids is the identifiers of a listing, in the order it answered them.
func ids(r *ListResult) []string {
	out := make([]string, 0, len(r.Tasks))
	for _, v := range r.Tasks {
		out = append(out, v.Task.ID)
	}
	return out
}

func (h *harness) list(p ListParams) *ListResult {
	h.t.Helper()
	result, err := ListOn(h.b, h.env, p)
	if err != nil {
		h.t.Fatalf("biso ls: %v", err)
	}
	return result
}

// TestTheDefaultOrderIsTheTupleOfTheSpecification is
// docs/spec/cmd/ls.md#la-regla-de-orden-completa: the tasks with an ordinal
// first and ascending, then the ones without it by urgency descending, and
// the identifier breaking every tie.
func TestTheDefaultOrderIsTheTupleOfTheSpecification(t *testing.T) {
	h := newHarness(t)
	h.create("No ordinal, low", scalar("priority", "low"))
	h.create("No ordinal, high", scalar("priority", "high"))
	h.create("Ordinal 5", scalar("ordinal", "5"))
	h.create("Ordinal 1", scalar("ordinal", "1"))
	// Two tasks that tie at exactly the same urgency, so that the
	// tie-break by identifier is what decides between them.
	h.create("Tie a", scalar("priority", "low"))

	got := ids(h.list(ListParams{}))

	want := []string{"MYP-4", "MYP-3", "MYP-2", "MYP-1", "MYP-5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the default order is %v and not %v", got, want)
	}
}

// TestSortPutsTheTasksWithoutThatFieldAtTheEnd is the last paragraph of that
// section, for the two fields a task can be missing.
func TestSortPutsTheTasksWithoutThatFieldAtTheEnd(t *testing.T) {
	h := newHarness(t)
	h.create("No due")
	h.create("Due later", scalar("due", "2027-01-01"))
	h.create("No due either")
	h.create("Due sooner", scalar("due", "2026-01-01"))

	got := ids(h.list(ListParams{Sort: "due"}))

	want := []string{"MYP-4", "MYP-2", "MYP-1", "MYP-3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--sort due answered %v and not %v", got, want)
	}
}

// TestReverseFlipsTheWholeOrderIncludingTheTieBreak is the other half of
// that rule.
func TestReverseFlipsTheWholeOrderIncludingTheTieBreak(t *testing.T) {
	h := newHarness(t)
	for _, title := range []string{"a", "b", "c"} {
		h.create(title, scalar("priority", "low"))
	}

	straight := ids(h.list(ListParams{}))
	flipped := ids(h.list(ListParams{Reverse: true}))

	sort.Sort(sort.Reverse(sort.StringSlice(straight)))
	if !reflect.DeepEqual(flipped, straight) {
		t.Errorf("--reverse answered %v and not the flipped order %v", flipped, straight)
	}
}

// TestTheLimitCutsTheListingAndSaysSo is the row of the behaviour table
// about the limit, with the three numbers the JSON schema carries.
func TestTheLimitCutsTheListingAndSaysSo(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 5; i++ {
		h.create("A task")
	}

	r := h.list(ListParams{Limit: 2, HasLimit: true})

	if r.Shown != 2 || r.Matched != 5 || r.Hidden != 3 || !r.Truncated {
		t.Errorf("shown %d, matched %d, hidden %d, truncated %v",
			r.Shown, r.Matched, r.Hidden, r.Truncated)
	}
	if len(r.Warnings) != 1 || r.Warnings[0].Code != "list_truncated" {
		t.Fatalf("the warnings are %v", r.Warnings)
	}
	if r.Warnings[0].Message != "3 more tasks match; showing 2 of 5" {
		t.Errorf("the truncation warning says %q", r.Warnings[0].Message)
	}
}

// TestLimitZeroPrintsNothingAndCountsEverything is the way of counting
// without --count.
func TestLimitZeroPrintsNothingAndCountsEverything(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 3; i++ {
		h.create("A task")
	}

	r := h.list(ListParams{Limit: 0, HasLimit: true})

	if len(r.Tasks) != 0 || r.Matched != 3 || !r.Truncated {
		t.Errorf("--limit 0 answered %d rows, %d matched, truncated %v",
			len(r.Tasks), r.Matched, r.Truncated)
	}
}

// TestAllIgnoresTheLimitAltogether.
func TestAllIgnoresTheLimitAltogether(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 40; i++ {
		h.create("A task")
	}

	r := h.list(ListParams{All: true})

	if len(r.Tasks) != 40 || r.Truncated {
		t.Errorf("--all answered %d rows, truncated %v", len(r.Tasks), r.Truncated)
	}
}

// TestCountAnswersTheTotalAndNoRow is the decision this branch took about
// --count, written into docs/spec/cmd/ls.md: it answers how many match and
// never a row, so the limit has nothing to cut and there is no truncation
// to warn about.
func TestCountAnswersTheTotalAndNoRow(t *testing.T) {
	h := newHarness(t)
	for i := 0; i < 40; i++ {
		h.create("A task")
	}

	r := h.list(ListParams{Count: true})

	if r.Matched != 40 || len(r.Tasks) != 0 || r.Truncated || len(r.Warnings) != 0 {
		t.Errorf("--count answered %d matched, %d rows, truncated %v, warnings %v",
			r.Matched, len(r.Tasks), r.Truncated, r.Warnings)
	}
}

// TestTheEffectiveFilterIsWhatTheEnvelopeCarries is the promise of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls: what comes back is the
// filter that ran, with the vocabulary resolved to the board's own
// spelling and the defaults spelled out.
func TestTheEffectiveFilterIsWhatTheEnvelopeCarries(t *testing.T) {
	h := newHarness(t)
	h.create("A task")

	r := h.list(ListParams{Filters: Filters{Status: []string{"to-do"}}, HasStatus: true})
	if !reflect.DeepEqual(r.Filters.Status, []string{"To Do"}) {
		t.Errorf("the effective status filter is %v", r.Filters.Status)
	}

	byDefault := h.list(ListParams{})
	if !reflect.DeepEqual(byDefault.Filters.Status, []string{"To Do", "In Progress"}) {
		t.Errorf("the default status filter is %v", byDefault.Filters.Status)
	}

	everything := h.list(ListParams{Filters: Filters{AnyStatus: true}})
	if !reflect.DeepEqual(everything.Filters.Status,
		[]string{"To Do", "In Progress", "Done"}) {
		t.Errorf("--any-status resolved to %v", everything.Filters.Status)
	}
}

// TestMineResolvesToTheIdentityItUsed is the other resolution that table
// asks for.
func TestMineResolvesToTheIdentityItUsed(t *testing.T) {
	h := newHarness(t)
	h.create("Mine", add("add-assignees", "@claude"))
	h.create("Someone else's", add("add-assignees", "@sara"))

	r := h.list(ListParams{Mine: true})

	if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-1"}) {
		t.Errorf("--mine answered %v", got)
	}
	if !reflect.DeepEqual(r.Filters.Assignee, []string{"@claude"}) {
		t.Errorf("the effective assignee filter is %v", r.Filters.Assignee)
	}
}

func TestMineWithoutAnIdentityIsExitCodeSix(t *testing.T) {
	h := newHarness(t)
	h.create("A task")

	_, err := ListOn(h.b, h.as("").env, ListParams{Mine: true})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "mine_requires_identity" {
		t.Fatalf("error = %d/%s, want 6/mine_requires_identity", e.ExitCode, e.Code)
	}
}

// TestOverdueIsDaysBelowZero is the precision of docs/spec/cmd/ls.md: a
// task due today is not overdue.
func TestOverdueIsDaysBelowZero(t *testing.T) {
	h := newHarness(t)
	today := h.env.Now().UTC().Format("2006-01-02")
	h.create("Due today", scalar("due", today))
	h.create("Due yesterday", scalar("due",
		h.env.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")))

	r := h.list(ListParams{Filters: Filters{Overdue: true}})

	if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-2"}) {
		t.Errorf("--overdue answered %v and not the task due yesterday", got)
	}
}

// TestSearchUsesTheOneScopeThereIs is the promise of
// docs/spec/referencias.md#la-búsqueda-por-texto: --search and the
// resolution of a reference look in the same places.
func TestSearchUsesTheOneScopeThereIs(t *testing.T) {
	h := newHarness(t)
	h.create("A title")
	h.create("Another", add("append-desc", "something about CRLF"))
	h.create("Third", add("add-labels", "crlf"))

	r := h.list(ListParams{Filters: Filters{Search: pointer("crlf")}})

	if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-2", "MYP-3"}) {
		t.Errorf("--search answered %v", got)
	}
}

// TestTheFiltersOfTheParametersCarryTheKeysOfTheContract is the round trip
// the plan of TASK-13 asks for: the parameters are serialized and read back,
// and the keys are the ones of the table of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls, with no translation
// written by hand between the two.
func TestTheFiltersOfTheParametersCarryTheKeysOfTheContract(t *testing.T) {
	filters := Filters{
		Status:     []string{"To Do"},
		AnyStatus:  true,
		Type:       []string{"bug"},
		Label:      []string{"parser"},
		Assignee:   []string{"@claude"},
		Unassigned: true,
		Parent:     pointer("MYP-1"),
		Blocked:    boolPointer(false),
		DueBefore:  pointer("2026-09-20"),
		Search:     pointer("CRLF"),
	}

	encoded, err := json.Marshal(filters)
	if err != nil {
		t.Fatal(err)
	}
	var object map[string]any
	if err := json.Unmarshal(encoded, &object); err != nil {
		t.Fatal(err)
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	want := []string{
		"active", "anyStatus", "archived", "assignee", "blocked", "dueBefore",
		"label", "labelOr", "notStatus", "onlyArchived", "overdue", "parent",
		"priority", "search", "status", "type", "unassigned", "unchecked", "waiting",
	}
	if !reflect.DeepEqual(keys, want) {
		t.Errorf("the filters serialize as %v and the contract has %v", keys, want)
	}

	var back Filters
	if err := json.Unmarshal(encoded, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, filters) {
		t.Errorf("the round trip changed the filter:\n%#v\n%#v", back, filters)
	}
}

// TestAnUnreadableTaskIsSkippedAndNamed is the set read of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// at the layer that skips it.
func TestAnUnreadableTaskIsSkippedAndNamed(t *testing.T) {
	h := newHarness(t)
	h.create("Fine")
	h.create("Broken")
	if _, err := h.b.Store.Exec(
		`UPDATE task SET priority = 'urgent' WHERE id = 'MYP-2'`); err != nil {
		t.Fatal(err)
	}

	r := h.list(ListParams{})

	if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-1"}) {
		t.Errorf("the listing is %v and the readable task is MYP-1", got)
	}
	if !reflect.DeepEqual(r.Skipped, []string{"MYP-2"}) {
		t.Errorf("the skipped tasks are %v", r.Skipped)
	}
	if len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0].Message, "MYP-2") {
		t.Errorf("the warning does not name the task: %v", r.Warnings)
	}
}

func pointer(s string) *string { return &s }
func boolPointer(b bool) *bool { return &b }

// TestALabelOrAssigneeFilterDoesNotDistinguishCase is the second half of the
// rule of docs/spec/familias-de-flags.md#campos-de-lista-que-admiten-coma: a
// label and a person are stored letter for letter, so "Parser" and "parser"
// are two labels, and a reading filter folds the case, so either spelling
// finds both. The board here has one label and one assignee, each written
// with a capital, and every filter is typed in lowercase.
func TestALabelOrAssigneeFilterDoesNotDistinguishCase(t *testing.T) {
	h := newHarness(t)
	h.create("A task",
		add("add-labels", "Parser"),
		add("add-assignees", "@Sara"))

	cases := []struct {
		name string
		p    ListParams
	}{
		{"--label", ListParams{Filters: Filters{Label: []string{"parser"}}}},
		{"--label in another case", ListParams{Filters: Filters{Label: []string{"PARSER"}}}},
		{"--label-or", ListParams{Filters: Filters{LabelOr: []string{"parser"}}}},
		{"--assignee", ListParams{Filters: Filters{Assignee: []string{"@sara"}}}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r, err := ListOn(h.b, h.env, c.p)
			if err != nil {
				t.Fatalf("biso ls: %v", err)
			}
			if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-1"}) {
				t.Errorf("the listing is %v and the filter should have found MYP-1", got)
			}
		})
	}
}

// TestAFilterThatFoldsTheCaseStillRefusesAValueTheBoardDoesNotHave is the
// other side of that rule: folding the case is not accepting anything, and a
// label nobody wrote is still the exit code 3 of
// docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué. It also
// fixes that folding is not dropping the accents, which that same rule
// reserves for the text selectors and denies to a label.
func TestAFilterThatFoldsTheCaseStillRefusesAValueTheBoardDoesNotHave(t *testing.T) {
	h := newHarness(t)
	h.create("A task", add("add-labels", "Parsé"))

	for _, value := range []string{"parsers", "parse"} {
		_, err := ListOn(h.b, h.env, ListParams{
			Filters: Filters{Label: []string{value}},
		})
		assertSpec(t, err, 3, "unknown_label")
	}
}
