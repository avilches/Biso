package ops

import (
	"strings"
	"testing"

	"biso/internal/board"
	"biso/internal/model"
)

// These are the rules of the scoped labels, one test per rule and per
// `code`: writing them
// (docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito),
// querying them
// (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito),
// the `labels` list of the configuration
// (docs/spec/cmd/config.md#la-lista-labels) and the three checks of
// `biso doctor` (docs/spec/cmd/doctor.md#qué-comprueba).

// declare writes the `labels` list of the harness's board.
func (h *harness) declare(values ...string) {
	h.t.Helper()
	if _, err := ConfigOn(h.b, h.env, ConfigParams{
		Action: ConfigSet, Key: board.KeyLabels,
		Value: strings.Join(values, ","), Values: values,
	}); err != nil {
		h.t.Fatalf("biso config set labels %v: %v", values, err)
	}
}

// declareFails is declare over a list the board refuses.
func (h *harness) declareFails(values ...string) *model.Error {
	h.t.Helper()
	_, err := ConfigOn(h.b, h.env, ConfigParams{
		Action: ConfigSet, Key: board.KeyLabels,
		Value: strings.Join(values, ","), Values: values,
	})
	return specError(h.t, err)
}

// setFails runs a `biso set` that must not succeed and answers its error.
func (h *harness) setFails(ref string, changes ...Change) *model.Error {
	h.t.Helper()
	_, err := SetOn(h.b, h.env, SetParams{Refs: []string{ref}, Changes: changes})
	return specError(h.t, err)
}

// store writes a task straight through the board, which is the only way to
// reach the states `biso doctor` reports: every write of the program refuses
// them, so a board that holds one had its database written from outside.
func (h *harness) store(title string, labels ...string) string {
	h.t.Helper()
	task := &model.Task{
		Title: title, Status: h.b.Config.InitialStatus, Labels: labels,
		CreatedAt: writeClock, UpdatedAt: writeClock,
	}
	if err := h.b.Tasks.Create(task); err != nil {
		h.t.Fatalf("storing %q: %v", title, err)
	}
	return task.ID
}

// ------------------------------------------------------------ writing

// TestWritingAnExclusiveLabelDropsTheOthersOfItsKey is the first row of the
// table of docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito.
func TestWritingAnExclusiveLabelDropsTheOthersOfItsKey(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task",
		add("add-labels", "size::s"), add("add-labels", "urgent"))
	// A second value of the same key, written with :, is fine while no
	// exclusive one is there to contradict it.
	h.set(id, remove("rm-labels", "size::s"),
		add("add-labels", "size:l"), add("add-labels", "size:s"))

	result := h.set(id, add("add-labels", "size::m"))

	assertLabels(t, h.load(id), "urgent", "size::m")
	w := warningOf(result, "exclusive_label_replaced")
	if w == nil {
		t.Fatalf("warnings = %v, want an exclusive_label_replaced", result.Warnings)
	}
	// The labels are named in the order the task held them, because no
	// list of this program is ever sorted on its own.
	want := `--add-labels: "size::m" replaced size:l, size:s on ` + id
	if w.Message != want {
		t.Errorf("the warning says %q, want %q", w.Message, want)
	}
	if w.Fields["task"] != id || w.Fields["value"] != "size::m" || w.Fields["flag"] != "--add-labels" {
		t.Errorf("the warning carries %v", w.Fields)
	}
	replaced, _ := w.Fields["replaced"].([]string)
	if strings.Join(replaced, "|") != "size:l|size:s" {
		t.Errorf("the warning replaced %v", w.Fields["replaced"])
	}
}

// TestWritingAKeyWithOneColonOverAnExclusiveOneIsRefused is the second row,
// which is exit code 6 and writes nothing.
func TestWritingAKeyWithOneColonOverAnExclusiveOneIsRefused(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	e := h.setFails(id, add("add-labels", "size:m"))

	if e.ExitCode != 6 || e.Code != "exclusive_label_conflict" {
		t.Fatalf("error = %d/%s, want 6/exclusive_label_conflict", e.ExitCode, e.Code)
	}
	want := id + ` already has "size::s", and :: allows at most one value of the key "size"`
	if e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "drop it first, as in --rm-labels size::s --add-labels size:m" {
		t.Errorf("the hint is %v", e.Hints)
	}
	// Nothing was written, which is what makes it a check of the
	// validation phase.
	assertLabels(t, h.load(id), "size::s")
}

// TestTheSameValueWithTheOtherSeparatorIsStillARefusal is the half of that
// row that covers the value already being there: the separator asserts the
// cardinality and not only the value, so `k:v` over a stored `k::v` is
// refused like any other.
func TestTheSameValueWithTheOtherSeparatorIsStillARefusal(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	e := h.setFails(id, add("add-labels", "size:s"))

	if e.Code != "exclusive_label_conflict" {
		t.Errorf("error = %d/%s, want an exclusive_label_conflict", e.ExitCode, e.Code)
	}
}

// TestTheTwoSeparatorsOfOneKeyInOneCallAreAUsageError is the third row, and
// the error blames neither value.
func TestTheTwoSeparatorsOfOneKeyInOneCallAreAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	e := h.setFails(id, add("add-labels", "size:a"), add("add-labels", "size::b"))

	if e.ExitCode != 2 || e.Code != "mixed_label_separators" {
		t.Fatalf("error = %d/%s, want 2/mixed_label_separators", e.ExitCode, e.Code)
	}
	want := `"size:a" and "size::b" mix the two separators of the key "size"`
	if e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	// It names a field and quotes no value, because neither of the two is
	// more at fault than the other
	// (docs/spec/contrato-json.md#los-errores-en-json).
	if e.Field != "labels" || !e.NoGiven {
		t.Errorf("the error carries field %q and given %q", e.Field, e.Given)
	}
}

// TestTheOrderTheTwoSeparatorsWereWrittenInDoesNotMatter is the half of
// that same row that says the refusal does not look at which of the two
// came first on the command line.
func TestTheOrderTheTwoSeparatorsWereWrittenInDoesNotMatter(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	e := h.setFails(id, add("add-labels", "size::b"), add("add-labels", "size:a"))

	if e.Code != "mixed_label_separators" {
		t.Fatalf("error = %d/%s, want a mixed_label_separators", e.ExitCode, e.Code)
	}
	if want := `"size::b" and "size:a" mix the two separators of the key "size"`; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
}

// TestTheTwoSeparatorsSplitBetweenTwoFlagsAreTheSameRefusal is the rest of
// that row, which admits the two values spread over --add-labels and
// --replace-labels however the caller likes. The message names them in the
// order the write would apply them, replacing before adding, and not in the
// order they were typed, because it blames neither
// (docs/spec/familias-de-flags.md#escribir-una-etiqueta-con-ámbito).
func TestTheTwoSeparatorsSplitBetweenTwoFlagsAreTheSameRefusal(t *testing.T) {
	want := `"zz::b" and "zz:a" mix the two separators of the key "zz"`
	for _, typed := range [][]Change{
		{add("add-labels", "zz:a"), replace("replace-labels", "zz::b")},
		{replace("replace-labels", "zz::b"), add("add-labels", "zz:a")},
	} {
		h := newHarness(t)
		id := h.create("A task")

		e := h.setFails(id, typed...)

		if e.ExitCode != 2 || e.Code != "mixed_label_separators" {
			t.Fatalf("error = %d/%s, want 2/mixed_label_separators", e.ExitCode, e.Code)
		}
		if e.Message != want {
			t.Errorf("the message is %q, want %q", e.Message, want)
		}
	}
}

// TestBetweenTwoFlagsTheOrderOfTheStepsDecidesWhichExclusiveStays is the
// paragraph that follows: within one flag the last value of the command line
// wins, and between --add-labels and --replace-labels the order of the steps
// does, because replacing comes before adding. What comes out is then the
// warning of the replacement and not the one of the last value, which is
// what really happened.
func TestBetweenTwoFlagsTheOrderOfTheStepsDecidesWhichExclusiveStays(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	result := h.set(id, add("add-labels", "g::AA"), replace("replace-labels", "g::BB"))

	assertLabels(t, h.load(id), "g::AA")
	w := warningOf(result, "exclusive_label_replaced")
	if w == nil {
		t.Fatalf("warnings = %v, want an exclusive_label_replaced", result.Warnings)
	}
	if want := `--add-labels: "g::AA" replaced g::BB on ` + id; w.Message != want {
		t.Errorf("the warning says %q, want %q", w.Message, want)
	}
	if warned(result, "exclusive_label_last_wins") {
		t.Errorf("the two came from two flags and it said one of them lost: %v", result.Warnings)
	}
}

// TestTheValuesOfRmLabelsDoNotCountForTheMixedSeparators is what makes
// `--rm-labels k::1 --add-labels k:2` work in one call: the values that
// remove write nothing, and they are precisely what makes room.
func TestTheValuesOfRmLabelsDoNotCountForTheMixedSeparators(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	h.set(id, remove("rm-labels", "size::s"), add("add-labels", "size:m"))

	assertLabels(t, h.load(id), "size:m")
}

// TestTwoExclusiveValuesOfOneKeyKeepTheLastOne is the fourth row: the last
// value of the command line wins, with a warning that carries no task,
// because it is a fact about the call and not about any one task.
func TestTwoExclusiveValuesOfOneKeyKeepTheLastOne(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	result := h.set(id, add("add-labels", "size::a"), add("add-labels", "size::b"))

	assertLabels(t, h.load(id), "size::b")
	w := warningOf(result, "exclusive_label_last_wins")
	if w == nil {
		t.Fatalf("warnings = %v, want an exclusive_label_last_wins", result.Warnings)
	}
	if want := `--add-labels: key "size" given twice with ::, kept "size::b"`; w.Message != want {
		t.Errorf("the warning says %q, want %q", w.Message, want)
	}
	if w.Fields["key"] != "size" || w.Fields["kept"] != "size::b" {
		t.Errorf("the warning carries %v", w.Fields)
	}
}

// TestThreeExclusiveValuesCountThemselves is the "given 3 times with ::" of
// that warning, the same way duplicate_flag_value counts.
func TestThreeExclusiveValuesCountThemselves(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	result := h.set(id,
		add("add-labels", "size::a"), add("add-labels", "size::b"), add("add-labels", "size::c"))

	assertLabels(t, h.load(id), "size::c")
	w := warningOf(result, "exclusive_label_last_wins")
	if w == nil || w.Message != `--add-labels: key "size" given 3 times with ::, kept "size::c"` {
		t.Fatalf("the warning is %v", w)
	}
}

// TestRemovingALabelIgnoresTheSeparator is the fifth row, and the sixth
// checks the form the flags that write refuse.
func TestRemovingALabelIgnoresTheSeparator(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "milestone::m1"))

	h.set(id, remove("rm-labels", "milestone:m1"))

	assertLabels(t, h.load(id))
}

func TestRemovingALabelFoldsTheKey(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "milestone::m1"))

	h.set(id, remove("rm-labels", "Milestone:m1"))

	assertLabels(t, h.load(id))
}

// TestReplaceLabelsNeverCollidesWithWhatTheTaskHad is the paragraph about
// --replace-labels: the list it leaves is judged whole, and what was stored
// went in its own step.
func TestReplaceLabelsNeverCollidesWithWhatTheTaskHad(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	h.set(id, replace("replace-labels", "size:m"), replace("replace-labels", "size:l"))

	assertLabels(t, h.load(id), "size:m", "size:l")
}

func TestReplaceLabelsIsJudgedByTheSameTwoRules(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	e := h.setFails(id, replace("replace-labels", "k:a"), replace("replace-labels", "k::b"))
	if e.Code != "mixed_label_separators" {
		t.Errorf("--replace-labels k:a,k::b = %d/%s, want a mixed_label_separators",
			e.ExitCode, e.Code)
	}

	result := h.set(id, replace("replace-labels", "k::a"), replace("replace-labels", "k::b"))
	assertLabels(t, h.load(id), "k::b")
	if !warned(result, "exclusive_label_last_wins") {
		t.Errorf("--replace-labels k::a,k::b earned %v", result.Warnings)
	}
}

// TestTheExclusivityIsJudgedOverTheListTheWriteLeaves is the promise the
// section makes about --clear-labels: what was stored went in its own step,
// so nothing of it can collide with what is added.
func TestTheExclusivityIsJudgedOverTheListTheWriteLeaves(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	h.set(id, clear("clear-labels"), add("add-labels", "size:m"))

	assertLabels(t, h.load(id), "size:m")
}

// TestAPlainLabelIsUntouchedByAllOfThis is the sentence that opens the
// section of docs/spec/valores-de-entrada.md: a label with no colon has no
// key, and none of these rules reaches it.
func TestAPlainLabelIsUntouchedByAllOfThis(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "urgent"), add("add-labels", "parser"))

	result := h.set(id, add("add-labels", "urgent"))

	assertLabels(t, h.load(id), "urgent", "parser")
	if !warned(result, "value_already_present") {
		t.Errorf("adding a label the task had earned %v", result.Warnings)
	}
}

// TestWritingTheSameExclusiveLabelTwiceKeepsItWhereItWas is the corner where
// the two rules meet: the value is already there, so nothing is replaced and
// the ordinary warning of a repeated value is the whole answer.
func TestWritingTheSameExclusiveLabelTwiceKeepsItWhereItWas(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::m"), add("add-labels", "urgent"))

	result := h.set(id, add("add-labels", "size::m"))

	assertLabels(t, h.load(id), "size::m", "urgent")
	if !warned(result, "value_already_present") {
		t.Errorf("warnings = %v, want a value_already_present", result.Warnings)
	}
	if warned(result, "exclusive_label_replaced") {
		t.Errorf("nothing was replaced and the call said it was: %v", result.Warnings)
	}
}

// ------------------------------------------------------------ the list of the configuration

func TestTheLabelsListRestrictsOnlyTheKeysItNames(t *testing.T) {
	h := newHarness(t)
	h.declare("pepe", "size::s", "size::m", "milestone::")
	id := h.create("A task")

	// A key the list does not name is free, and so is a plain label
	// nobody declares.
	h.set(id, add("add-labels", "trello:card:42"), add("add-labels", "whatever"))

	assertLabels(t, h.load(id), "trello:card:42", "whatever")
}

func TestAValueARestrictedKeyDoesNotDeclareIsRefused(t *testing.T) {
	h := newHarness(t)
	h.declare("size::s", "size::m", "size::l")
	id := h.create("A task")

	e := h.setFails(id, add("add-labels", "size::xl"))

	if e.ExitCode != 3 || e.Code != "unknown_label_value" {
		t.Fatalf("error = %d/%s, want 3/unknown_label_value", e.ExitCode, e.Code)
	}
	if want := `unknown label value: "size::xl"`; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	want := `       valid labels for the key "size" on this board: size::s, size::m, size::l`
	if len(e.Detail) != 1 || e.Detail[0] != want {
		t.Errorf("the second line is %v, want %q", e.Detail, want)
	}
	if strings.Join(e.Valid, ",") != "size::s,size::m,size::l" {
		t.Errorf("valid = %v", e.Valid)
	}
}

func TestAnOpenKeyWrittenWithTheOtherSeparatorIsRefused(t *testing.T) {
	h := newHarness(t)
	h.declare("milestone::")
	id := h.create("A task")

	e := h.setFails(id, add("add-labels", "milestone:m1"))

	if e.ExitCode != 3 || e.Code != "wrong_label_separator" {
		t.Fatalf("error = %d/%s, want 3/wrong_label_separator", e.ExitCode, e.Code)
	}
	want := `wrong separator for the label key "milestone": "milestone:m1"`
	if e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "this board declares milestone::, at most one value per task" {
		t.Errorf("the hint is %v", e.Hints)
	}
	// It is the one `code` of exit status 3 with no `valid`, because an
	// open key has no set of values to offer
	// (docs/spec/contrato-json.md#los-errores-en-json).
	if len(e.Valid) != 0 {
		t.Errorf("valid = %v, and an open key has nothing to offer", e.Valid)
	}
}

func TestAnOpenKeyAcceptsAnyValueWithItsOwnSeparator(t *testing.T) {
	h := newHarness(t)
	h.declare("milestone::")
	id := h.create("A task")

	h.set(id, add("add-labels", "milestone::anything-at-all"))

	assertLabels(t, h.load(id), "milestone::anything-at-all")
}

func TestTheListIsCheckedBeforeTheSeparatorsOfTheCall(t *testing.T) {
	h := newHarness(t)
	h.declare("size::s")
	id := h.create("A task")

	// Both rules are broken at once, and the list is the one that answers.
	e := h.setFails(id, add("add-labels", "size:a"), add("add-labels", "size::b"))

	if e.Code != "unknown_label_value" {
		t.Errorf("error = %d/%s, and the list is checked first", e.ExitCode, e.Code)
	}
}

func TestAKeyDeclaredWithTheTwoSeparatorsIsRefused(t *testing.T) {
	h := newHarness(t)

	e := h.declareFails("size::s", "size:l")

	if e.ExitCode != 3 || e.Code != "bad_config_value" {
		t.Fatalf("error = %d/%s, want 3/bad_config_value", e.ExitCode, e.Code)
	}
	if want := `labels: the key "size" is declared with both : and ::`; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	if len(e.Hints) != 1 ||
		e.Hints[0] != "a key takes one separator, : for several values or :: for at most one" {
		t.Errorf("the hint is %v", e.Hints)
	}
}

func TestAKeyDeclaredOpenAndWithValuesIsRefused(t *testing.T) {
	h := newHarness(t)

	e := h.declareFails("milestone::", "milestone::m1")

	if e.ExitCode != 3 || e.Code != "bad_config_value" {
		t.Fatalf("error = %d/%s, want 3/bad_config_value", e.ExitCode, e.Code)
	}
	want := `labels: the key "milestone" is declared both as an open key and with exact values`
	if e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	if len(e.Hints) != 1 ||
		e.Hints[0] != "declare milestone:: on its own, or its values, but not both" {
		t.Errorf("the hint is %v", e.Hints)
	}
}

// TestAMalformedEntryOfTheListIsExitCodeTwo is the row of the case table:
// the shape of the token, and not the coherence of the list, which is 3.
func TestAMalformedEntryOfTheListIsExitCodeTwo(t *testing.T) {
	h := newHarness(t)

	e := h.declareFails("size:::m")

	if e.ExitCode != 2 || e.Code != "malformed_label" {
		t.Errorf("error = %d/%s, want 2/malformed_label", e.ExitCode, e.Code)
	}
}

// TestSettingLabelsFailsWhenATaskUsesWhatTheNewListForbids is the report of
// docs/spec/cmd/config.md#la-lista-labels, with one line per label and the
// tasks that carry it.
func TestSettingLabelsFailsWhenATaskUsesWhatTheNewListForbids(t *testing.T) {
	h := newHarness(t)
	first := h.create("One", add("add-labels", "milestone:m1"))
	second := h.create("Two", add("add-labels", "size::xl"))
	third := h.create("Three", add("add-labels", "size::xl"))

	e := h.declareFails("milestone::", "size::s", "size::m", "size::l")

	if e.ExitCode != 6 || e.Code != "board_inconsistent" {
		t.Fatalf("error = %d/%s, want 6/board_inconsistent", e.ExitCode, e.Code)
	}
	if want := "labels: 3 tasks use a label the new list would not allow"; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	want := []string{
		`  "milestone:m1" on ` + first + `: the new list declares the key "milestone" with ::`,
		`  "size::xl" on ` + second + ", " + third + `: the key "size" allows size::s, size::m, size::l`,
	}
	if strings.Join(e.Detail, "\n") != strings.Join(want, "\n") {
		t.Errorf("the report is\n%s\nand it should be\n%s",
			strings.Join(e.Detail, "\n"), strings.Join(want, "\n"))
	}
}

func TestOneTaskAgainstTheNewListSaysOneTask(t *testing.T) {
	h := newHarness(t)
	h.create("One", add("add-labels", "size::xl"))

	e := h.declareFails("size::s")

	if want := "labels: 1 task uses a label the new list would not allow"; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
}

// TestTakingAPlainLabelOrAWholeKeyOutOfTheListNeverFails is the last
// sentence of that section: what stops being declared goes back to free.
func TestTakingAPlainLabelOrAWholeKeyOutOfTheListNeverFails(t *testing.T) {
	h := newHarness(t)
	h.declare("pepe", "size::s")
	h.create("One", add("add-labels", "pepe"), add("add-labels", "size::s"))

	h.declare("milestone::")

	if got := h.b.Config.Labels; strings.Join(got, ",") != "milestone::" {
		t.Errorf("labels = %v", got)
	}
}

// TestTheListLooksAtArchivedAndFinishedTasksToo is the set the refusal
// judges against, which is the same one the filters validate against.
func TestTheListLooksAtArchivedAndFinishedTasksToo(t *testing.T) {
	h := newHarness(t)
	id := h.create("One", add("add-labels", "size::xl"))
	if _, err := ArchiveOn(h.b, h.env, ArchiveParams{Refs: []string{id}}); err != nil {
		t.Fatal(err)
	}

	e := h.declareFails("size::s")

	if e.Code != "board_inconsistent" {
		t.Errorf("error = %d/%s, and an archived task counts", e.ExitCode, e.Code)
	}
}

// ------------------------------------------------------------ querying

func TestAKeyFilterMatchesAnyValueOfThatKey(t *testing.T) {
	h := newHarness(t)
	one := h.create("One", add("add-labels", "milestone::m1"))
	two := h.create("Two", add("add-labels", "milestone:m2"))
	h.create("Three", add("add-labels", "urgent"))

	for _, filter := range []string{"milestone:", "milestone::", "Milestone:"} {
		got := h.list(ListParams{Filters: Filters{Label: []string{filter}}})
		if strings.Join(ids(got), ",") != one+","+two {
			t.Errorf("--label %s found %v, want %s", filter, ids(got), one+","+two)
		}
	}
}

// TestAWholeLabelFilterIgnoresTheSeparatorToo is the paragraph that follows:
// --label milestone:m1 finds a task labelled milestone::m1.
func TestAWholeLabelFilterIgnoresTheSeparatorToo(t *testing.T) {
	h := newHarness(t)
	one := h.create("One", add("add-labels", "milestone::m1"))

	got := h.list(ListParams{Filters: Filters{Label: []string{"milestone:m1"}}})

	if strings.Join(ids(got), ",") != one {
		t.Errorf("--label milestone:m1 found %v", ids(got))
	}
}

// TestAKeyFilterIsAndedLikeAnyOtherValueOfItsFlag is the combination rule of
// docs/spec/cmd/ls.md.
func TestAKeyFilterIsAndedLikeAnyOtherValueOfItsFlag(t *testing.T) {
	h := newHarness(t)
	both := h.create("Both", add("add-labels", "milestone::m1"), add("add-labels", "bug"))
	h.create("Only the milestone", add("add-labels", "milestone::m2"))

	got := h.list(ListParams{Filters: Filters{Label: []string{"milestone:", "bug"}}})

	if strings.Join(ids(got), ",") != both {
		t.Errorf("--label milestone: --label bug found %v", ids(got))
	}
}

func TestAKeyFilterIsOredInLabelOr(t *testing.T) {
	h := newHarness(t)
	one := h.create("One", add("add-labels", "milestone::m1"))
	two := h.create("Two", add("add-labels", "bug"))

	got := h.list(ListParams{Filters: Filters{LabelOr: []string{"milestone:", "bug"}}})

	if strings.Join(ids(got), ",") != one+","+two {
		t.Errorf("--label-or milestone: --label-or bug found %v", ids(got))
	}
}

func TestAKeyTheBoardDoesNotHaveIsAnUnknownLabelKey(t *testing.T) {
	h := newHarness(t)
	h.create("One", add("add-labels", "milestone::m1"))

	_, err := ListOn(h.b, h.env, ListParams{Filters: Filters{Label: []string{"milestne:"}}})

	e := specError(t, err)
	if e.ExitCode != 3 || e.Code != "unknown_label_key" {
		t.Fatalf("error = %d/%s, want 3/unknown_label_key", e.ExitCode, e.Code)
	}
	if want := `unknown label key: "milestne"`; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "did you mean: milestone?" {
		t.Errorf("the hint is %v", e.Hints)
	}
	if e.Field != "label" || e.Given != "milestne:" {
		t.Errorf("the error carries field %q and given %q", e.Field, e.Given)
	}
	if strings.Join(e.Valid, ",") != "milestone" {
		t.Errorf("valid = %v, want the keys of the board", e.Valid)
	}
}

// TestTheGivenOfAnUnknownKeyIsTheFilterAlreadyNormalized is the rest of that
// paragraph: `given` is the filter that travels in data.filters.label, with
// one colon however it was typed and with the spelling of the key kept
// (docs/spec/vocabularios.md#consultar-por-la-clave-de-una-etiqueta-con-ámbito).
func TestTheGivenOfAnUnknownKeyIsTheFilterAlreadyNormalized(t *testing.T) {
	h := newHarness(t)
	h.create("One", add("add-labels", "milestone::m1"))

	_, err := ListOn(h.b, h.env, ListParams{Filters: Filters{Label: []string{"Milestne::"}}})

	e := specError(t, err)
	if e.Code != "unknown_label_key" {
		t.Fatalf("error = %d/%s, want an unknown_label_key", e.ExitCode, e.Code)
	}
	if e.Given != "Milestne:" {
		t.Errorf("given = %q, want the filter with one colon and the key as typed", e.Given)
	}
	if want := `unknown label key: "Milestne"`; e.Message != want {
		t.Errorf("the message is %q, want %q", e.Message, want)
	}
}

// TestADeclaredKeyCountsAsKnownWithNoTaskUsingIt is what makes declaring a
// key useful before anything carries it.
func TestADeclaredKeyCountsAsKnownWithNoTaskUsingIt(t *testing.T) {
	h := newHarness(t)
	h.declare("milestone::")
	h.create("One")

	if _, err := ListOn(h.b, h.env,
		ListParams{Filters: Filters{Label: []string{"milestone:"}}}); err != nil {
		t.Errorf("--label milestone: on a board that declares it = %v", err)
	}
}

func TestUncheckedTurnsOffTheKeyCheckAndNotTheFormOne(t *testing.T) {
	h := newHarness(t)
	h.create("One")

	if _, err := ListOn(h.b, h.env, ListParams{
		Filters: Filters{Label: []string{"milestne:"}, Unchecked: true},
	}); err != nil {
		t.Errorf("--label milestne: --unchecked = %v, and --unchecked turns that off", err)
	}

	_, err := ListOn(h.b, h.env, ListParams{
		Filters: Filters{Label: []string{":m"}, Unchecked: true},
	})
	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "malformed_label" {
		t.Errorf("--label :m --unchecked = %d/%s, want 2/malformed_label", e.ExitCode, e.Code)
	}
}

// TestTheKeyFilterTravelsWithOneColon is the rule of
// docs/spec/contrato-json.md#los-filtros-de-biso-ls: however it was typed,
// and keeping the spelling of the key.
func TestTheKeyFilterTravelsWithOneColon(t *testing.T) {
	h := newHarness(t)
	h.create("One", add("add-labels", "milestone::m1"))

	result, err := ListOn(h.b, h.env, ListParams{
		Filters: Filters{Label: []string{"Milestone::"}, LabelOr: []string{"milestone::"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(result.Filters.Label, ",") != "Milestone:" {
		t.Errorf("label = %v", result.Filters.Label)
	}
	if strings.Join(result.Filters.LabelOr, ",") != "milestone:" {
		t.Errorf("labelOr = %v", result.Filters.LabelOr)
	}
}

// TestNoTaskCarriesADerivedFieldPerKey is the promise of
// docs/spec/vocabularios.md: whoever wants the value of a key splits
// `labels` by the first colon.
func TestNoTaskCarriesADerivedFieldPerKey(t *testing.T) {
	h := newHarness(t)
	id := h.create("One", add("add-labels", "milestone::m1"))

	encoded, err := encodeTask(h.load(id))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "milestone\":") {
		t.Errorf("the exported task carries a field of its own per key:\n%s", encoded)
	}
}

// ------------------------------------------------------------ biso doctor

// soundBoard writes the <id>.id marker that board.Create leaves to whoever
// creates the board, so that the reports below carry the findings about the
// labels and nothing else.
func (h *harness) soundBoard() {
	h.t.Helper()
	if err := board.WriteMarker(h.b.Location.Dir, "3f9a2b1c"); err != nil {
		h.t.Fatalf("writing the marker: %v", err)
	}
}

func TestDoctorReportsALabelTheListDoesNotAllow(t *testing.T) {
	h := newHarness(t)
	h.soundBoard()
	value := h.store("A value", "size::xl")
	separator := h.store("A separator", "milestone:m1")
	h.b.Config.Labels = []string{"size::s", "size::m", "size::l", "milestone::"}

	result := h.doctor()

	want := []string{
		`label "size::xl" is not one of the values the key "size" declares: size::s, size::m, size::l`,
		`label "milestone:m1" uses :, and the key "milestone" is declared with ::`,
	}
	assertFindings(t, result.Problems, "label_not_declared",
		[]string{value, separator}, want)
}

func TestDoctorReportsExclusivityBrokenInsideOneTask(t *testing.T) {
	h := newHarness(t)
	h.soundBoard()
	id := h.store("Two of one key", "size::a", "size::b")

	result := h.doctor()

	assertFindings(t, result.Problems, "label_exclusive_violated", []string{id},
		[]string{`the label key "size" has 2 values on this task, and :: allows at most one: size::a, size::b`})
}

// TestDoctorReportsAKeyUsedWithBothSeparators is the third check, which is a
// warning and not an error: the writes that produced it were each
// legitimate, because the exclusivity is judged inside one task.
func TestDoctorReportsAKeyUsedWithBothSeparators(t *testing.T) {
	h := newHarness(t)
	h.soundBoard()
	h.store("One", "milestone:m1")
	h.store("Two", "milestone:m2")
	h.store("Three", "milestone::m3")

	result := h.doctor()

	if len(result.Problems) != 0 {
		t.Fatalf("it reported errors and this one is a warning: %v", result.Problems)
	}
	want := `label key "milestone" is used with both separators, 2 tasks with : and 1 with ::`
	assertFindings(t, result.Warnings, "label_key_mixed_separators", []string{""},
		[]string{want})
}

// TestTheThreeLabelChecksAreNotRepairable is the paragraph of
// docs/spec/cmd/doctor.md: choosing which value was meant is exactly the
// decision that destroys information when it is taken alone.
func TestTheThreeLabelChecksAreNotRepairable(t *testing.T) {
	h := newHarness(t)
	h.soundBoard()
	h.store("Two of one key", "size::a", "size::b")

	result, err := DoctorOn(h.b, h.env, DoctorParams{Fix: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Fixed) != 0 {
		t.Errorf("--fix repaired %v", result.Fixed)
	}
	if len(result.Problems) != 1 {
		t.Errorf("the problem did not stay: %v", result.Problems)
	}
}

// TestTheLabelChecksComeInTheOrderOfTheTable is the order of
// docs/spec/cmd/doctor.md#el-orden-en-que-sale-el-informe, with the three new
// rows between the criterion keys and the lease.
func TestTheLabelChecksComeInTheOrderOfTheTable(t *testing.T) {
	h := newHarness(t)
	h.soundBoard()
	h.b.Config.Labels = []string{"size::s"}
	h.store("Broken twice", "size::a", "size::b")

	result := h.doctor()

	var codes []string
	for _, p := range result.Problems {
		codes = append(codes, p.Code)
	}
	want := "label_not_declared,label_not_declared,label_exclusive_violated"
	if strings.Join(codes, ",") != want {
		t.Errorf("the report came in the order %v, want %s", codes, want)
	}
}

func (h *harness) doctor() *DoctorResult {
	h.t.Helper()
	result, err := DoctorOn(h.b, h.env, DoctorParams{})
	if err != nil {
		h.t.Fatalf("biso doctor: %v", err)
	}
	return result
}

// assertFindings checks that the report carries exactly those findings of
// one code, each on its task and with its message.
func assertFindings(t *testing.T, findings []Finding, code string, tasks, messages []string) {
	t.Helper()
	var got []Finding
	for _, f := range findings {
		if f.Code == code {
			got = append(got, f)
		}
	}
	if len(got) != len(messages) {
		t.Fatalf("%s reported %d findings and %d were expected: %v",
			code, len(got), len(messages), got)
	}
	for i, f := range got {
		if f.Task != tasks[i] || f.Message != messages[i] {
			t.Errorf("finding %d = %s / %q, want %s / %q",
				i, f.Task, f.Message, tasks[i], messages[i])
		}
	}
}

// ------------------------------------------------------------ the batch

// TestABatchNamesEveryLabelFailureTheSpecificationPrints is the block of
// docs/spec/cmd/new.md#el-modo-lote, message for message.
func TestABatchNamesEveryLabelFailureTheSpecificationPrints(t *testing.T) {
	h := newHarness(t)
	h.declare("size::s", "size::m", "size::l", "milestone::")

	for _, c := range []struct {
		labels   string
		code     string
		exitCode int
		message  string
	}{
		{`["size:"]`, "malformed_label", 2, `malformed label: "size:"`},
		{`["team:a","team::b"]`, "mixed_label_separators", 2,
			`labels mix the two separators of the key "team": "team:a" and "team::b"`},
		{`["team::a","team::b"]`, "exclusive_label_conflict", 6,
			`labels give the key "team" more than one value, and :: allows at most one: "team::a" and "team::b"`},
		{`["size::xl"]`, "unknown_label_value", 3, `unknown label value: "size::xl"`},
		{`["milestone:m1"]`, "wrong_label_separator", 3,
			`wrong separator for the label key "milestone": "milestone:m1"`},
	} {
		_, err := h.batch(`{"title":"A task","labels":` + c.labels + `}`)
		e := specError(t, err)
		if len(e.Details) != 1 {
			t.Fatalf("%s: %d failures", c.labels, len(e.Details))
		}
		d := e.Details[0]
		if d.Code != c.code || d.ExitCode != c.exitCode {
			t.Errorf("%s: code = %d/%s, want %d/%s", c.labels, d.ExitCode, d.Code, c.exitCode, c.code)
		}
		if d.Message != "line 1: "+c.message {
			t.Errorf("%s: message = %q, want %q", c.labels, d.Message, "line 1: "+c.message)
		}
		// A `code` carries the same keys inside a batch as outside it, or a
		// caller could not branch on it
		// (docs/spec/contrato-json.md#los-errores-en-json). The one of exit
		// code 6 carries neither, because the table promises `field` and
		// `given` to the errors of 3 and to the ones of 2 that name a flag.
		if c.exitCode == 6 && (d.Field != "" || d.Given != "") {
			t.Errorf("%s: the 6 carries field %q and given %q, and it carries neither outside a batch",
				c.labels, d.Field, d.Given)
		}
	}
}

// TestTheExclusivityOfALineCarriesTheSameKeysAsOutsideALine is the other
// half of that check, over the error the write itself answers.
func TestTheExclusivityOfALineCarriesTheSameKeysAsOutsideALine(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "size::s"))

	e := h.setFails(id, add("add-labels", "size:m"))

	if e.Field != "" || e.Given != "" {
		t.Errorf("exclusive_label_conflict carries field %q and given %q", e.Field, e.Given)
	}
}

// TestTwoExclusiveValuesOnALineDoNotKeepTheLastOne is the one difference the
// specification declares between a line and a command line: a line describes
// a state that was saved, and dropping one of its values in silence would
// lose a fact the file asserted.
func TestTwoExclusiveValuesOnALineDoNotKeepTheLastOne(t *testing.T) {
	h := newHarness(t)

	_, err := h.batch(`{"title":"A task","labels":["size::a","size::b"]}`)

	e := specError(t, err)
	if len(e.Details) != 1 || e.Details[0].Code != "exclusive_label_conflict" {
		t.Fatalf("the batch answered %v", e.Detail)
	}
}

func TestABatchAcceptsAWellFormedScopedLabel(t *testing.T) {
	h := newHarness(t)

	if _, err := h.batch(
		`{"title":"A task","labels":["size::m","trello:card:42","urgent"]}`); err != nil {
		t.Fatal(err)
	}

	assertLabels(t, h.load("MYP-1"), "size::m", "trello:card:42", "urgent")
}
