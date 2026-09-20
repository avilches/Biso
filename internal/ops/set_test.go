package ops

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"biso/internal/model"
)

// These are the cases of the behaviour table of docs/spec/cmd/set.md, and
// the order of application of
// docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura, which
// is checked with several combinations and not with one: a single one would
// only prove that the two steps it happens to exercise are in the right
// order relative to each other.

func TestSetWithNoChangeFlagIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	_, err := SetOn(h.b, h.env, SetParams{Refs: []string{id}})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "nothing_to_change" {
		t.Fatalf("error = %d/%s, want 2/nothing_to_change", e.ExitCode, e.Code)
	}
	if len(e.Hints) == 0 || !strings.Contains(e.Hints[0], "biso get") {
		t.Errorf("hints = %v, want the pointer to biso get", e.Hints)
	}
}

func TestSetWithSeveralTasksRefusesASelectorThatIsNotAll(t *testing.T) {
	h := newHarness(t)
	first := h.create("First", add("add-ac", "A"))
	second := h.create("Second", add("add-ac", "B"))

	for _, selector := range []string{"1", "1-2", "1,3", "covers CRLF"} {
		_, err := SetOn(h.b, h.env, SetParams{
			Refs: []string{first, second}, Changes: []Change{check("check-ac", selector)},
		})
		assertSpec(t, err, 2, "key_selector_with_many_tasks")
	}
}

func TestSetWithSeveralTasksAcceptsCheckAcAll(t *testing.T) {
	h := newHarness(t)
	first := h.create("First", add("add-ac", "A"))
	second := h.create("Second", add("add-ac", "B"))

	if _, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{first, second}, Changes: []Change{check("check-ac", "all")},
	}); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{first, second} {
		if h.load(id).AcDone() != 1 {
			t.Errorf("%s did not get its own criterion checked", id)
		}
	}
}

// The all-or-nothing of
// docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables,
// checked against the database and not by reading the code: the first task
// is a real one and comes back from SQLite exactly as it was.
func TestSetWritesNoneOfTheTasksWhenOneReferenceDoesNotExist(t *testing.T) {
	h := newHarness(t)
	first := h.create("First")
	second := h.create("Second")
	before := h.load(first)

	_, err := SetOn(h.b, h.env, SetParams{
		Refs:    []string{first, second, "MYP-90"},
		Changes: []Change{add("add-labels", "urgent")},
	})
	assertSpec(t, err, 4, "never_allocated")

	for _, id := range []string{first, second} {
		task := h.load(id)
		if len(task.Labels) != 0 {
			t.Errorf("%s was written although the call failed: labels %v", id, task.Labels)
		}
		if !task.UpdatedAt.Equal(before.UpdatedAt) {
			t.Errorf("%s had its updatedAt moved by a call that wrote nothing", id)
		}
	}
}

// The same guarantee when what fails is not the reference but a value: the
// second task's write is rejected and the first one's does not survive.
func TestSetWritesNoneOfTheTasksWhenAValueIsRejected(t *testing.T) {
	h := newHarness(t)
	first := h.create("First")
	second := h.create("Second")

	_, err := SetOn(h.b, h.env, SetParams{
		Refs:    []string{first, second},
		Changes: []Change{add("add-labels", "urgent"), scalar("status", "Pending")},
	})
	assertSpec(t, err, 3, "unknown_status")

	if labels := h.load(first).Labels; len(labels) != 0 {
		t.Errorf("the first task kept %v from a call that failed on the second", labels)
	}
}

func TestSetWarnsWhenReplaceOverwritesANonEmptyList(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "cli"), add("add-labels", "parser"))

	result := h.set(id, replace("replace-labels", "urgent"))

	if !warned(result, "overwrite") {
		t.Fatalf("replacing two labels did not warn: %v", result.Warnings)
	}
	if got := result.Warnings[0].Fields["count"]; got != 2 {
		t.Errorf("count = %v, want the 2 labels it replaced", got)
	}
	assertLabels(t, h.load(id), "urgent")
}

func TestSetWarnsAboutArrivingAtTheTerminalStatusAndDoesItAnyway(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-ac", "A"), add("add-ac", "B"))

	result := h.set(id, scalar("status", "Done"))

	if !warned(result, "terminal_ac_unchecked") {
		t.Errorf("finishing with unchecked criteria did not warn: %v", result.Warnings)
	}
	if !warned(result, "terminal_no_summary") {
		t.Errorf("finishing without a summary did not warn: %v", result.Warnings)
	}
	if h.load(id).Status != "Done" {
		t.Errorf("the warning stopped the write, and it only warns")
	}
}

// docs/spec/cmd/set.md: every flag leaving the task as it was is exit code
// 0 with a note, no field is written and updatedAt does not move, but the
// lease renews all the same when the caller is its holder
// (docs/spec/lease.md#la-renovación).
func TestSetThatChangesNothingRenewsTheLeaseAndTouchesNothingElse(t *testing.T) {
	h := newHarness(t)
	result, err := NewOn(h.b, h.env, NewParams{Title: "A task", HasTitle: true, Start: true})
	if err != nil {
		t.Fatal(err)
	}
	id := result.Tasks[0].ID
	before := h.load(id)
	// The lease of the creation is about to expire, so that a renewal is
	// visible as a different instant and not as the same one.
	before.LeaseExpiresAt = writeClock.Add(time.Minute)
	if err := h.b.Tasks.Save(before); err != nil {
		t.Fatal(err)
	}

	again := h.set(id, add("add-assignees", "@claude"))

	if len(again.Tasks[0].Changed) != 0 {
		t.Errorf("changed = %v, and nothing changed", again.Tasks[0].Changed)
	}
	if !noted(again, id+" unchanged") {
		t.Errorf("notes = %v, want the unchanged one", again.Notes)
	}
	after := h.load(id)
	if !after.UpdatedAt.Equal(before.UpdatedAt) {
		t.Errorf("updatedAt moved on a write that changed no field")
	}
	if !after.LeaseExpiresAt.Equal(writeClock.Add(240 * time.Minute)) {
		t.Errorf("leaseExpiresAt = %v, want the renewal of the holder's own write",
			after.LeaseExpiresAt)
	}
}

// docs/spec/lease.md#el-vaciado: the fields only have a value on a task that
// is at once active and assigned, so losing either empties them in that very
// write, whoever made it.
func TestTheLeaseIsEmptiedWhenTheTaskStopsBeingActiveOrAssigned(t *testing.T) {
	for _, c := range []struct {
		name   string
		change Change
	}{
		{"leaving the active status", scalar("status", "To Do")},
		{"losing its last assignee", clear("clear-assignees")},
		{"removing its last assignee", remove("rm-assignees", "@claude")},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			created, err := NewOn(h.b, h.env, NewParams{
				Title: "A task", HasTitle: true, Start: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			id := created.Tasks[0].ID

			h.set(id, c.change)

			task := h.load(id)
			if !task.LeaseExpiresAt.IsZero() || task.LeaseHolder != "" {
				t.Errorf("the lease survived: holder %q, expires %v",
					task.LeaseHolder, task.LeaseExpiresAt)
			}
		})
	}
}

// docs/spec/lease.md#la-renovación: a write by another identity while the
// lease is alive touches neither field and warns.
func TestAWriteByAnotherIdentityWarnsAndLeavesTheLeaseAlone(t *testing.T) {
	h := newHarness(t)
	created, err := NewOn(h.b, h.env, NewParams{Title: "A task", HasTitle: true, Start: true})
	if err != nil {
		t.Fatal(err)
	}
	id := created.Tasks[0].ID

	other := h.as("@sara")
	result, err := SetOn(other.b, other.env, SetParams{
		Refs: []string{id}, Changes: []Change{add("append-note", "looked at it")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "lease_held") {
		t.Errorf("writing over somebody else's lease did not warn: %v", result.Warnings)
	}
	task := h.load(id)
	if task.LeaseHolder != "@claude" {
		t.Errorf("leaseHolder = %q, and only biso start transfers one", task.LeaseHolder)
	}
	if !task.LeaseExpiresAt.Equal(created.Tasks[0].leaseOf(h, id)) {
		t.Errorf("the lease of another identity was renewed")
	}
}

// leaseOf is the lease instant the board holds for a task, read back so the
// test above compares against the database and not against its own memory.
func (TaskWrite) leaseOf(h *harness, id string) time.Time {
	return h.load(id).LeaseExpiresAt
}

func TestSetRefusesACommentWithNoAuthorAndNoIdentity(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	anonymous := h.as("")
	_, err := SetOn(anonymous.b, anonymous.env, SetParams{
		Refs: []string{id}, Changes: []Change{comment("said something")},
	})

	assertSpec(t, err, 2, "missing_identity")
}

func TestSetWritesOnAnArchivedTaskAndSaysSo(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	task := h.load(id)
	task.Archived = true
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	result := h.set(id, add("add-labels", "parser"))

	if !noted(result, id+" is archived") {
		t.Errorf("notes = %v, want the archived one", result.Notes)
	}
	assertLabels(t, h.load(id), "parser")
}

func TestSetDryRunAnswersTheSameLineAndWritesNothing(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-ac", "A"))

	preview, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{id}, Changes: []Change{add("add-labels", "parser")}, DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	real := h.set(id, add("add-labels", "parser"))

	if !reflect.DeepEqual(preview.Tasks[0], real.Tasks[0]) {
		t.Errorf("the preview said %+v and the real write said %+v", preview.Tasks[0], real.Tasks[0])
	}
	if !preview.DryRun {
		t.Errorf("the result does not say it was a preview")
	}
}

func TestSetDryRunLeavesTheDatabaseUntouched(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	if _, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{id}, Changes: []Change{add("add-labels", "parser")}, DryRun: true,
	}); err != nil {
		t.Fatal(err)
	}

	if labels := h.load(id).Labels; len(labels) != 0 {
		t.Errorf("labels = %v, and a preview writes nothing", labels)
	}
}

// The order of application, step by step. Each case writes the flags in the
// order that would give the wrong answer if the engine obeyed the command
// line, so passing means the fixed order won.
func TestTheOrderOfApplicationDoesNotDependOnTheCommandLine(t *testing.T) {
	t.Run("step 1 before step 4: clearing then adding leaves only what was added", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", add("add-labels", "cli"))

		h.set(id, add("add-labels", "urgent"), clear("clear-labels"))

		assertLabels(t, h.load(id), "urgent")
	})

	t.Run("step 1 before step 4 and step 7: a criterion created here is checked here", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", add("add-ac", "Old"))

		h.set(id, check("check-ac", "all"), add("add-ac", "A"), clear("clear-acs"))

		task := h.load(id)
		if task.AcTotal() != 1 || task.AcDone() != 1 || task.AcceptanceCriteria[0].Text != "A" {
			t.Errorf("criteria = %+v, want only the new one, checked", task.AcceptanceCriteria)
		}
	})

	t.Run("step 2 before step 3: what --replace-labels puts, --rm-labels can take away", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", add("add-labels", "cli"))

		h.set(id, remove("rm-labels", "parser"),
			replace("replace-labels", "parser"), replace("replace-labels", "ui"))

		// Step 2 emptied the list and put the two new labels in it, and
		// only then did step 3 take one of them out. Obeying the command
		// line would have removed a label that was not there yet and left
		// both.
		assertLabels(t, h.load(id), "ui")
	})

	t.Run("step 2 before step 4: --replace-labels does not wipe out an --add-labels", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", add("add-labels", "cli"))

		h.set(id, add("add-labels", "urgent"), replace("replace-labels", "parser"))

		// The replacement threw away "cli", the only label of before the
		// call, and the addition of the same call survived it. Obeying the
		// command line would have left "parser" alone.
		assertLabels(t, h.load(id), "parser", "urgent")
	})

	t.Run("step 1 before step 2: --clear-labels does not undo a --replace-labels", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", add("add-labels", "cli"))

		h.set(id, replace("replace-labels", "parser"), clear("clear-labels"))

		assertLabels(t, h.load(id), "parser")
	})

	t.Run("step 1 before step 5: --ext survives a --clear-ext of the same call", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", ext("trello.card", "old"))

		h.set(id, ext("trello.card", "5f2a8c1e"), clear("clear-ext"))

		if got := h.load(id).Ext["trello.card"]; got != "5f2a8c1e" {
			t.Errorf("ext[trello.card] = %q, and step 5 comes after step 1", got)
		}
	})

	t.Run("step 3 before step 5: --ext survives an --rm-ext of the same key", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", ext("trello.card", "old"))

		h.set(id, ext("trello.card", "5f2a8c1e"), remove("rm-ext", "trello.card"))

		if got := h.load(id).Ext["trello.card"]; got != "5f2a8c1e" {
			t.Errorf("ext[trello.card] = %q, and step 5 comes after step 3", got)
		}
	})

	t.Run("step 5 before step 6: --ext and a scalar do not fight", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task")

		h.set(id, scalar("title", "Renamed"), ext("trello.card", "abc"))

		task := h.load(id)
		if task.Title != "Renamed" || task.Ext["trello.card"] != "abc" {
			t.Errorf("title = %q, ext = %v, want both written", task.Title, task.Ext)
		}
	})

	t.Run("step 8 corrects the date of a comment that was already there", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", comment("Reported from Windows"))

		h.set(id, commentDate("1", "2026-08-14T10:22:00Z"))

		got := h.load(id).Comments[0].CreatedAt.UTC().Format(time.RFC3339)
		if got != "2026-08-14T10:22:00Z" {
			t.Errorf("createdAt = %s, want the corrected instant", got)
		}
	})

	t.Run("step 8 before step 9: a comment this call adds is not a target", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task")

		_, err := SetOn(h.b, h.env, SetParams{
			Refs: []string{id},
			Changes: []Change{
				comment("brand new"),
				commentDate("1", "2026-08-14T10:22:00Z"),
			},
		})

		// The selector is resolved against the list of before the write,
		// which is empty, so it names a comment that does not exist yet.
		assertSpec(t, err, 4, "comment_not_found")
	})

	t.Run("step 3 before step 9: removing every comment and adding one leaves one", func(t *testing.T) {
		h := newHarness(t)
		id := h.create("A task", comment("first"))

		h.set(id, comment("second"), remove("rm-comment", "all"))

		comments := h.load(id).Comments
		if len(comments) != 1 || comments[0].Body != "second" {
			t.Errorf("comments = %+v, want only the one this call added", comments)
		}
	})
}

// docs/spec/familias-de-flags.md#comentarios: the two comment selectors are
// resolved before either is applied, so naming the same comment in both is a
// conflict and not a "does not exist" from the one that ran first.
func TestRemovingAndDatingTheSameCommentIsAConflict(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", comment("Reported from Windows"))

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{id},
		Changes: []Change{
			remove("rm-comment", "1"),
			commentDate("1", "2026-08-14T10:22:00Z"),
		},
	})

	assertSpec(t, err, 2, "comment_selector_overlap")
}

func TestTheTolerantWarningsOfAddingAndRemoving(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task", add("add-labels", "urgent"), ext("trello.card", "abc"))

	for _, c := range []struct {
		change Change
		code   string
	}{
		{add("add-labels", "urgent"), "value_already_present"},
		{remove("rm-labels", "nothing-like-this"), "value_not_present"},
		{remove("rm-ext", "priority_score"), "value_not_present"},
	} {
		result := h.set(id, c.change)
		if !warned(result, c.code) {
			t.Errorf("%s did not produce %s: %v", c.change.Flag, c.code, result.Warnings)
		}
	}
}

func TestSetRefusesToCloseACycle(t *testing.T) {
	h := newHarness(t)
	first := h.create("First")
	second := h.create("Second", add("add-deps", first))

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{first}, Changes: []Change{add("add-deps", second)},
	})
	assertSpec(t, err, 2, "dependency_cycle")

	_, err = SetOn(h.b, h.env, SetParams{
		Refs: []string{first}, Changes: []Change{add("add-deps", first)},
	})
	assertSpec(t, err, 2, "self_dependency")
}

func TestSetRefusesToCloseAParentCycle(t *testing.T) {
	h := newHarness(t)
	parent := h.create("The parent")
	child := h.create("The child", scalar("parent", parent))

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{parent}, Changes: []Change{scalar("parent", child)},
	})

	assertSpec(t, err, 2, "parent_cycle")
}

// docs/spec/referencias.md: the three endings of resolving a reference by
// text, and the note the one that works owes its caller.
func TestTheThreeEndingsOfATextReference(t *testing.T) {
	h := newHarness(t)
	h.create("Normalize CRLF in the diff")
	h.create("Another task about CRLF handling")

	result := h.set("Normalize", add("add-labels", "parser"))
	if !noted(result, `"Normalize" matched MYP-1`) {
		t.Errorf("notes = %v, want the matched one", result.Notes)
	}

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{"CRLF"}, Changes: []Change{add("add-labels", "parser")},
	})
	assertSpec(t, err, 5, "ambiguous_reference")

	_, err = SetOn(h.b, h.env, SetParams{
		Refs: []string{"nothing like this"}, Changes: []Change{add("add-labels", "parser")},
	})
	assertSpec(t, err, 4, "not_found")
}

// The candidates of an ambiguous reference travel with the error, because
// that ending prints them.
func TestAnAmbiguousReferenceCarriesItsCandidates(t *testing.T) {
	h := newHarness(t)
	h.create("Normalize CRLF in the diff")
	h.create("Another task about CRLF handling")

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{"CRLF"}, Changes: []Change{add("add-labels", "parser")},
	})

	var ambiguous *AmbiguousRef
	if !asAmbiguous(err, &ambiguous) {
		t.Fatalf("error = %T, want an *AmbiguousRef", err)
	}
	if len(ambiguous.Candidates) != 2 {
		t.Errorf("candidates = %d, want the two tasks that match", len(ambiguous.Candidates))
	}
}

// docs/spec/referencias.md#la-gramática: --id forces an identifier, and a
// value the grammar does not admit is a usage error there instead of
// becoming a text query.
func TestTheForcedInterpretationsOfAReference(t *testing.T) {
	h := newHarness(t)
	id := h.create("42")

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{"MYP-1.1"}, Mode: RefID, Changes: []Change{add("add-labels", "x")},
	})
	assertSpec(t, err, 2, "malformed_id")

	// Without --match, "42" reads as the identifier MYP-42; with it, as the
	// title of the task that is called that.
	result, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{"42"}, Mode: RefText, Changes: []Change{add("add-labels", "x")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Tasks[0].ID != id {
		t.Errorf("--match resolved to %s, want the task titled 42", result.Tasks[0].ID)
	}
}

func TestTheThreeShapesOfAnIdentifierResolveToTheSameTask(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	for _, ref := range []string{"MYP-1", "myp-1", "1", "#1"} {
		result := h.set(ref, add("add-labels", "x"))
		if result.Tasks[0].ID != id {
			t.Errorf("%q resolved to %s, want %s", ref, result.Tasks[0].ID, id)
		}
	}
}

func TestTheSameTaskNamedTwiceIsWrittenOnce(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")

	result, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{id, "1"}, Changes: []Change{comment("said something")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Tasks) != 1 {
		t.Errorf("tasks = %d, want the one task it named twice", len(result.Tasks))
	}
	if got := len(h.load(id).Comments); got != 1 {
		t.Errorf("comments = %d, want the one the call added", got)
	}
}

// asAmbiguous is errors.As for the one error type of this package that wraps
// a *model.Error.
func asAmbiguous(err error, target **AmbiguousRef) bool {
	if a, ok := err.(*AmbiguousRef); ok {
		*target = a
		return true
	}
	return false
}

// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar: a
// targeted read of a task the board cannot decode is exit code 3, and
// nothing is written.
func TestSetOverAnUnreadableTaskIsExitCodeThreeAndWritesNothing(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	other := h.create("Another task")
	// A priority the configuration does not list is one of the ways a
	// stored task stops being decodable.
	if _, err := h.b.Store.Exec(
		"UPDATE task SET priority = 'urgent' WHERE id = ?", id); err != nil {
		t.Fatal(err)
	}

	_, err := SetOn(h.b, h.env, SetParams{
		Refs: []string{id, other}, Changes: []Change{add("add-labels", "parser")},
	})

	assertSpec(t, err, 3, "undecodable_task")
	if labels := h.load(other).Labels; len(labels) != 0 {
		t.Errorf("the readable task was written anyway: %v", labels)
	}
}

// docs/spec/cmd/set.md: arriving at the terminal status with an open
// question warns and does it anyway, exactly as `biso finish` does.
func TestArrivingAtTheTerminalStatusWithAnOpenQuestionOnlyWarns(t *testing.T) {
	h := newHarness(t)
	id := h.create("A task")
	task := h.load(id)
	task.Question = &model.Question{
		Author: "@sara", AskedAt: writeClock, Body: "Is it a CRLF, or also a lone CR?",
	}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	result := h.set(id, scalar("status", "Done"))

	if !warned(result, "open_question_on_terminal") {
		t.Errorf("finishing with an open question did not warn: %v", result.Warnings)
	}
	if h.load(id).Status != "Done" {
		t.Errorf("the warning stopped the write, and it only warns")
	}
}

// A preview runs every check the real write runs, the model's validation
// included: a --dry-run that answered 0 where the write answers 3 would be a
// preview that lies about what the call would do
// (docs/spec/cmd/flags-globales.md).
func TestSetDryRunRefusesWhatTheRealWriteRefuses(t *testing.T) {
	for _, c := range []struct {
		name    string
		change  Change
		code    int
		problem string
	}{
		{"an extension key the board does not declare", ext("trello.board", "42"), 3, "unknown_extension_key"},
		{"an ordinal that is not a positive number", scalar("ordinal", "-5"), 2, "invalid_number"},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			id := h.create("A task")

			_, preview := SetOn(h.b, h.env, SetParams{
				Refs: []string{id}, Changes: []Change{c.change}, DryRun: true,
			})
			_, real := SetOn(h.b, h.env, SetParams{
				Refs: []string{id}, Changes: []Change{c.change},
			})

			assertSpec(t, preview, c.code, c.problem)
			assertSpec(t, real, c.code, c.problem)
		})
	}
}
