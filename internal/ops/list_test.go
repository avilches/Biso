package ops

import (
	"encoding/json"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
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
	h.create("Placed second", scalar("ordinal", "last"))
	h.create("Placed first", scalar("ordinal", "first"))
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

// TestSortOrdinalIsTheDefaultOrderWithoutItsStepOfUrgency is the sentence
// of docs/spec/cmd/ls.md#la-regla-de-orden-completa that tells --sort
// ordinal from the default order: the tasks with a key come first and
// ascending in both, and what changes is that the ones without a key are
// ordered among themselves by identifier and not by urgency.
func TestSortOrdinalIsTheDefaultOrderWithoutItsStepOfUrgency(t *testing.T) {
	h := newHarness(t)
	// The most urgent of the tasks with no key is the last identifier, so
	// an order that still looked at the urgency would put it in front.
	h.create("No key, low", scalar("priority", "low"))
	h.create("Placed second", scalar("ordinal", "last"))
	h.create("No key, high", scalar("priority", "high"))
	h.create("Placed first", scalar("ordinal", "first"))

	got := ids(h.list(ListParams{Sort: "ordinal"}))

	want := []string{"MYP-4", "MYP-2", "MYP-1", "MYP-3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--sort ordinal answered %v and not %v", got, want)
	}

	// --reverse flips the whole thing, so the tasks with no key come first.
	flipped := ids(h.list(ListParams{Sort: "ordinal", Reverse: true}))
	wantFlipped := []string{"MYP-3", "MYP-1", "MYP-2", "MYP-4"}
	if !reflect.DeepEqual(flipped, wantFlipped) {
		t.Errorf("--sort ordinal --reverse answered %v and not %v", flipped, wantFlipped)
	}
}

// TestTheOrderOfTwoKeysIsByCodePoint is the comparison of
// docs/spec/modelo-de-datos/orden-manual.md: code point by code point, so a
// digit comes before a letter and a longer key comes after the prefix it
// extends. Written by hand onto the board, because a key cannot be typed.
func TestTheOrderOfTwoKeysIsByCodePoint(t *testing.T) {
	h := newHarness(t)
	for _, c := range []struct{ title, key string }{
		{"Fourth", "b"},
		{"Second", "a"},
		{"First", "9"},
		{"Third", "ai"},
	} {
		id := h.create(c.title)
		task := h.load(id)
		task.Ordinal = c.key
		if err := h.b.Tasks.Save(task); err != nil {
			t.Fatalf("writing the key %q: %v", c.key, err)
		}
	}

	got := ids(h.list(ListParams{Sort: "ordinal"}))

	want := []string{"MYP-3", "MYP-2", "MYP-4", "MYP-1"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("the keys 9, a, ai and b came out as %v and not %v", got, want)
	}
}

// TestTwoTasksWithTheSameKeyAreBrokenByIdentifier is the last step of the
// order rule applied where two keys tie, which a batch can perfectly well
// produce (docs/spec/cmd/ls.md#la-regla-de-orden-completa).
func TestTwoTasksWithTheSameKeyAreBrokenByIdentifier(t *testing.T) {
	h := newHarness(t)
	for _, title := range []string{"The second", "The first"} {
		id := h.create(title)
		task := h.load(id)
		task.Ordinal = "m"
		if err := h.b.Tasks.Save(task); err != nil {
			t.Fatalf("writing a key by hand: %v", err)
		}
	}

	got := ids(h.list(ListParams{Sort: "ordinal"}))

	want := []string{"MYP-1", "MYP-2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("two tasks sharing a key came out as %v and not %v", got, want)
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

// TestSearchNeverReachesReferencesButRefDoes is the other half of that same
// scope, now that --ref exists: docs/spec/cmd/ls.md#parámetros keeps
// `references` out of --search's scope on purpose, and --ref is the
// dedicated filter that reaches it instead
// (docs/spec/referencias.md#la-búsqueda-por-texto).
func TestSearchNeverReachesReferencesButRefDoes(t *testing.T) {
	h := newHarness(t)
	h.create("A task", add("add-refs", "docs/bugs/BUG-02.md"))

	bySearch := h.list(ListParams{Filters: Filters{Search: pointer("BUG-02")}})
	if got := ids(bySearch); len(got) != 0 {
		t.Errorf("--search BUG-02 answered %v, want none: references is out of its scope", got)
	}

	byRef := h.list(ListParams{Filters: Filters{Ref: []string{"BUG-02"}}})
	if got := ids(byRef); !reflect.DeepEqual(got, []string{"MYP-1"}) {
		t.Errorf("--ref BUG-02 answered %v and not the task carrying it", got)
	}
}

// TestTheFiltersOfTheParametersCarryTheKeysOfTheContract is the round trip
// the plan of TASK-13 asks for: the parameters are serialized and read back,
// and the keys are the ones of the table of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls, with no translation
// written by hand between the two.
func TestTheFiltersOfTheParametersCarryTheKeysOfTheContract(t *testing.T) {
	filters := Filters{
		Status:        []string{"To Do"},
		AnyStatus:     true,
		Type:          []string{"bug"},
		NotType:       []string{"docs"},
		Priority:      []string{"high"},
		NotPriority:   []string{"low"},
		Label:         []string{"parser"},
		NotLabel:      []string{"blocked"},
		Assignee:      []string{"@claude"},
		NotAssignee:   []string{"@sara"},
		Unassigned:    true,
		Author:        []string{"@avilches"},
		Parent:        pointer("MYP-1"),
		Root:          false,
		Blocked:       boolPointer(false),
		DueBefore:     pointer("2026-09-20"),
		CreatedAfter:  pointer("2026-09-01"),
		CreatedBefore: pointer("2026-09-08"),
		UpdatedAfter:  pointer("2026-09-02"),
		UpdatedBefore: pointer("2026-09-09"),
		Ref:           []string{"internal/ops/write.go"},
		NotRef:        []string{"docs/bugs"},
		Search:        pointer("CRLF"),
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
		"active", "anyStatus", "archived", "assignee", "author", "blocked",
		"createdAfter", "createdBefore", "dueBefore", "label", "labelOr",
		"notAssignee", "notLabel", "notPriority", "notRef", "notStatus",
		"notType", "onlyArchived", "overdue", "parent", "priority", "ref",
		"root", "search", "status", "type", "unassigned", "unchecked",
		"updatedAfter", "updatedBefore", "waiting",
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

// TestTheFourNegationsSubtractEachValueOnItsOwn is
// docs/spec/cmd/ls.md#parámetros: a negation does not OR its repeated
// values like the positive filters do, because summing exclusions can never
// bring back a task that one of them already threw out.
func TestTheFourNegationsSubtractEachValueOnItsOwn(t *testing.T) {
	h := newHarness(t)
	h.create("A bug", scalar("type", "bug"))
	h.create("A doc", scalar("type", "docs"))
	h.create("A task", scalar("type", "task"))

	got := ids(h.list(ListParams{
		Filters: Filters{NotType: []string{"bug", "docs"}},
	}))

	if want := []string{"MYP-3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--not-type bug --not-type docs answered %v and not %v", got, want)
	}
}

// TestATypeAndItsOwnNegationIsAnEmptyListNotAnError is the general rule of
// docs/spec/cmd/ls.md#parámetros applied to naming the same value on both
// sides of the same field: two valid filters that never overlap are a fact
// about the board and not a usage error.
func TestATypeAndItsOwnNegationIsAnEmptyListNotAnError(t *testing.T) {
	h := newHarness(t)
	h.create("A bug", scalar("type", "bug"))

	r := h.list(ListParams{
		Filters: Filters{Type: []string{"bug"}, NotType: []string{"bug"}},
	})

	if got := ids(r); len(got) != 0 {
		t.Errorf("--type bug --not-type bug answered %v, want none", got)
	}
}

// TestNotPriorityExcludesExactlyLikeNotType mirrors the type case for
// --not-priority, the other vocabulary flag docs/spec/cmd/ls.md#parámetros
// adds a negation for.
func TestNotPriorityExcludesExactlyLikeNotType(t *testing.T) {
	h := newHarness(t)
	h.create("High", scalar("priority", "high"))
	h.create("Low", scalar("priority", "low"))

	got := ids(h.list(ListParams{Filters: Filters{NotPriority: []string{"high"}}}))

	if want := []string{"MYP-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--not-priority high answered %v and not %v", got, want)
	}
}

// TestNotTypeAndNotPriorityWithAnUnknownValueKeepTheFieldOfTheirPositive is
// the mechanism --not-type and --not-priority replicate from --not-status:
// both reuse match.Type/match.Priority the same way --not-status reuses
// match.Status, so the error's `field` stays the domain name ("type",
// "priority") and not a name of its own ("notType", "notPriority"), unlike
// --not-label and --not-assignee just below, which do get their own
// (docs/spec/contrato-json.md#los-filtros-de-biso-ls).
func TestNotTypeAndNotPriorityWithAnUnknownValueKeepTheFieldOfTheirPositive(t *testing.T) {
	h := newHarness(t)
	h.create("A task")

	_, err := ListOn(h.b, h.env, ListParams{Filters: Filters{NotType: []string{"epic"}}})
	e := specError(t, err)
	if e.ExitCode != 3 || e.Code != "unknown_type" || e.Field != "type" {
		t.Errorf("--not-type epic answered %d/%s field=%q, want 3/unknown_type field=\"type\"",
			e.ExitCode, e.Code, e.Field)
	}

	_, err = ListOn(h.b, h.env, ListParams{Filters: Filters{NotPriority: []string{"urgent"}}})
	e = specError(t, err)
	if e.ExitCode != 3 || e.Code != "unknown_priority" || e.Field != "priority" {
		t.Errorf("--not-priority urgent answered %d/%s field=%q, want 3/unknown_priority field=\"priority\"",
			e.ExitCode, e.Code, e.Field)
	}
}

// TestNotLabelExcludesATaskCarryingAnyOfTheGivenLabels is the OR of
// docs/spec/cmd/ls.md#parámetros applied to the subtractive side: a task is
// excluded as soon as it carries any one of the --not-label values.
func TestNotLabelExcludesATaskCarryingAnyOfTheGivenLabels(t *testing.T) {
	h := newHarness(t)
	h.create("Front", add("add-labels", "frontend"))
	h.create("Back", add("add-labels", "backend"))
	h.create("Neither")

	got := ids(h.list(ListParams{
		Filters: Filters{NotLabel: []string{"frontend", "backend"}},
	}))

	if want := []string{"MYP-3"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--not-label frontend --not-label backend answered %v and not %v", got, want)
	}
}

// TestNotLabelAcceptsTheKeyForm is the same "clave:" reading --label and
// --label-or already have (docs/spec/cmd/ls.md#parámetros).
func TestNotLabelAcceptsTheKeyForm(t *testing.T) {
	h := newHarness(t)
	h.create("Milestone 1", add("add-labels", "milestone:m1"))
	h.create("No milestone")

	got := ids(h.list(ListParams{Filters: Filters{NotLabel: []string{"milestone:"}}}))

	if want := []string{"MYP-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--not-label milestone: answered %v and not %v", got, want)
	}
}

// TestNotLabelWithAnUnknownLabelIsExitCodeThree is the same check --label
// already runs, reused by a different flag
// (docs/spec/vocabularios.md#qué-valida-cada-filtro-y-contra-qué).
func TestNotLabelWithAnUnknownLabelIsExitCodeThree(t *testing.T) {
	h := newHarness(t)
	h.create("A task", add("add-labels", "frontend"))

	_, err := ListOn(h.b, h.env, ListParams{
		Filters: Filters{NotLabel: []string{"fronted"}},
	})
	assertSpec(t, err, 3, "unknown_label")

	// --unchecked turns that check off, the same as it does for --label.
	r := h.list(ListParams{Filters: Filters{NotLabel: []string{"fronted"}, Unchecked: true}})
	if got := ids(r); !reflect.DeepEqual(got, []string{"MYP-1"}) {
		t.Errorf("--not-label fronted --unchecked answered %v", got)
	}
}

// TestNotAssigneeExcludesATaskAssignedToAnyOfThem mirrors --not-label for
// people, folding the case like --assignee does.
func TestNotAssigneeExcludesATaskAssignedToAnyOfThem(t *testing.T) {
	h := newHarness(t)
	h.create("Claude's", add("add-assignees", "@claude"))
	h.create("Sara's", add("add-assignees", "@sara"))
	h.create("Nobody's")

	got := ids(h.list(ListParams{Filters: Filters{NotAssignee: []string{"@CLAUDE"}}}))

	want := []string{"MYP-2", "MYP-3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--not-assignee @CLAUDE answered %v and not %v", got, want)
	}
}

// TestNotAssigneeWithAnUnknownPersonIsExitCodeThree is checkKnown reused for
// the new flag, with the same suggestion algorithm.
func TestNotAssigneeWithAnUnknownPersonIsExitCodeThree(t *testing.T) {
	h := newHarness(t)
	h.create("A task", add("add-assignees", "@claude"))

	_, err := ListOn(h.b, h.env, ListParams{
		Filters: Filters{NotAssignee: []string{"@clude"}},
	})
	assertSpec(t, err, 3, "unknown_assignee")
}

// TestAuthorComparesTheSameNormalizeAsAssigneeWithNoVocabulary is
// docs/spec/cmd/ls.md#parámetros: --author folds the case, drops the
// diacritics and the separators like --assignee, but never checks the
// result against a configured set, so a value nobody used is an empty list
// and never exit code 3.
func TestAuthorComparesTheSameNormalizeAsAssigneeWithNoVocabulary(t *testing.T) {
	h := newHarness(t)
	h.as("@Alberto-Vilches").create("Written by Alberto")
	h.as("@sara").create("Written by Sara")

	got := ids(h.list(ListParams{Filters: Filters{Author: []string{"@albertovilches"}}}))
	if want := []string{"MYP-1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--author albertovilches answered %v and not %v", got, want)
	}

	nobody := h.list(ListParams{Filters: Filters{Author: []string{"@nobody"}}})
	if got := ids(nobody); len(got) != 0 {
		t.Errorf("--author @nobody answered %v, want none", got)
	}
	if len(nobody.Warnings) != 0 {
		// The board-wide unreadable-task warning is the only one that could
		// appear here, and this board has no such task; an unknown author
		// must never turn into an error of its own.
		t.Errorf("an unknown --author produced warnings: %v", nobody.Warnings)
	}
}

// TestAuthorCombinesRepeatedValuesWithOr is the same combination rule as
// --type or --assignee, named for --author in
// docs/spec/cmd/ls.md#parámetros.
func TestAuthorCombinesRepeatedValuesWithOr(t *testing.T) {
	h := newHarness(t)
	h.as("@claude").create("Claude's")
	h.as("@sara").create("Sara's")
	h.as("@avilches").create("Avilches's")

	got := ids(h.list(ListParams{Filters: Filters{Author: []string{"@claude", "@sara"}}}))

	want := []string{"MYP-1", "MYP-2"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--author @claude --author @sara answered %v and not %v", got, want)
	}
}

// TestRefMatchesASubstringOfReferencesFoldingCaseAndAccents is
// docs/spec/cmd/ls.md#parámetros: --ref uses the same folding as a text
// selector, case and diacritics alike, unlike a label or a person.
func TestRefMatchesASubstringOfReferencesFoldingCaseAndAccents(t *testing.T) {
	h := newHarness(t)
	h.create("With a reference", add("add-refs", "docs/bugs/BUG-02.md"))
	h.create("Without one")

	got := ids(h.list(ListParams{Filters: Filters{Ref: []string{"bûg-02"}}}))

	if want := []string{"MYP-1"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--ref bûg-02 answered %v and not %v", got, want)
	}
}

// TestRefNeverLooksAtTheTitleOrOtherFields is the scope
// docs/spec/cmd/ls.md#parámetros carves out: --ref only ever looks at
// `references`, unlike --search.
func TestRefNeverLooksAtTheTitleOrOtherFields(t *testing.T) {
	h := newHarness(t)
	h.create("Mentions CRLF in its title")
	h.create("Has it in a reference", add("add-refs", "docs/CRLF.md"))

	got := ids(h.list(ListParams{Filters: Filters{Ref: []string{"CRLF"}}}))

	if want := []string{"MYP-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--ref CRLF answered %v and not %v", got, want)
	}
}

// TestRefWithAValueNoTaskHasIsAnEmptyListNeverAnError is the same rule as
// --search and --author: --ref has no vocabulary to fail against.
func TestRefWithAValueNoTaskHasIsAnEmptyListNeverAnError(t *testing.T) {
	h := newHarness(t)
	h.create("A task", add("add-refs", "docs/a.md"))

	r := h.list(ListParams{Filters: Filters{Ref: []string{"nothing/like/this"}}})
	if got := ids(r); len(got) != 0 {
		t.Errorf("--ref with an unused value answered %v, want none", got)
	}
}

// TestNotRefExcludesATaskWhoseReferencesContainTheSubstring is the negation
// of --ref, with the same subtraction rule as the other four.
func TestNotRefExcludesATaskWhoseReferencesContainTheSubstring(t *testing.T) {
	h := newHarness(t)
	h.create("Bug doc", add("add-refs", "docs/bugs/BUG-02.md"))
	h.create("Other doc", add("add-refs", "docs/other.md"))

	got := ids(h.list(ListParams{Filters: Filters{NotRef: []string{"bugs"}}}))

	if want := []string{"MYP-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--not-ref bugs answered %v and not %v", got, want)
	}
}

// TestRootFiltersTasksWithNoParent is docs/spec/cmd/ls.md#parámetros:
// --root is the complement of --parent.
func TestRootFiltersTasksWithNoParent(t *testing.T) {
	h := newHarness(t)
	parent := h.create("Parent")
	h.create("Child", scalar("parent", parent))
	h.create("Another root")

	got := ids(h.list(ListParams{Filters: Filters{Root: true}}))

	want := []string{"MYP-1", "MYP-3"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--root answered %v and not %v", got, want)
	}
}

// TestCreatedAndUpdatedFiltersAreAHalfOpenInterval is
// docs/spec/cmd/ls.md#parámetros: --created-after/--updated-after include
// the day named, --created-before/--updated-before exclude it, so the two
// together never overlap on their shared boundary.
func TestCreatedAndUpdatedFiltersAreAHalfOpenInterval(t *testing.T) {
	h := newHarness(t)
	h.env.Now = func() time.Time { return time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC) }
	early := h.create("Created on the 1st")
	h.env.Now = func() time.Time { return time.Date(2026, 9, 8, 10, 0, 0, 0, time.UTC) }
	onTheBoundary := h.create("Created on the 8th")
	h.env.Now = func() time.Time { return time.Date(2026, 9, 15, 10, 0, 0, 0, time.UTC) }
	h.create("Created on the 15th")

	firstWeek := ids(h.list(ListParams{
		Filters: Filters{
			CreatedAfter:  pointer("2026-09-01"),
			CreatedBefore: pointer("2026-09-08"),
		},
	}))
	if want := []string{early}; !reflect.DeepEqual(firstWeek, want) {
		t.Errorf("the first week answered %v and not %v", firstWeek, want)
	}

	secondWeek := ids(h.list(ListParams{
		Filters: Filters{
			CreatedAfter:  pointer("2026-09-08"),
			CreatedBefore: pointer("2026-09-15"),
		},
	}))
	if want := []string{onTheBoundary}; !reflect.DeepEqual(secondWeek, want) {
		t.Errorf("the second week answered %v and not %v", secondWeek, want)
	}

	// --updated-after/--updated-before follow the same rule, over
	// `updatedAt`: setting a field moves it without touching `createdAt`.
	h.env.Now = func() time.Time { return time.Date(2026, 9, 20, 10, 0, 0, 0, time.UTC) }
	h.set(early, scalar("priority", "high"))

	updatedOnThe20th := ids(h.list(ListParams{
		Filters: Filters{UpdatedAfter: pointer("2026-09-20")},
	}))
	if want := []string{early}; !reflect.DeepEqual(updatedOnThe20th, want) {
		t.Errorf("--updated-after 2026-09-20 answered %v and not %v", updatedOnThe20th, want)
	}
}

// TestSortPriorityOrdersByTheConfiguredPositionWithNoneLast is
// docs/spec/cmd/ls.md#la-regla-de-orden-completa: ascending by the position
// in `priorities`, 0-indexed from the most urgent, with a task that has none
// at the end, in a block, by identifier.
func TestSortPriorityOrdersByTheConfiguredPositionWithNoneLast(t *testing.T) {
	h := newHarness(t)
	h.create("Low", scalar("priority", "low"))
	h.create("No priority")
	h.create("High", scalar("priority", "high"))
	h.create("Medium", scalar("priority", "medium"))
	h.create("No priority either")

	got := ids(h.list(ListParams{Sort: "priority"}))

	want := []string{"MYP-3", "MYP-4", "MYP-1", "MYP-2", "MYP-5"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("--sort priority answered %v and not %v", got, want)
	}

	flipped := ids(h.list(ListParams{Sort: "priority", Reverse: true}))
	wantFlipped := []string{"MYP-5", "MYP-2", "MYP-1", "MYP-4", "MYP-3"}
	if !reflect.DeepEqual(flipped, wantFlipped) {
		t.Errorf("--sort priority --reverse answered %v and not %v", flipped, wantFlipped)
	}
}

// TestSortPriorityOnABoardWithNoPriorityIsDeterministic is the same
// determinism --sort due already has for a board with no `due`
// (docs/spec/cmd/ls.md#la-regla-de-orden-completa).
func TestSortPriorityOnABoardWithNoPriorityIsDeterministic(t *testing.T) {
	h := newHarness(t)
	h.create("Z")
	h.create("A")

	got := ids(h.list(ListParams{Sort: "priority"}))

	if want := []string{"MYP-1", "MYP-2"}; !reflect.DeepEqual(got, want) {
		t.Errorf("--sort priority with no priority anywhere answered %v and not %v", got, want)
	}
}
