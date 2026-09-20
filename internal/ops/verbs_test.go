package ops

import (
	"strings"
	"testing"
	"time"

	"biso/internal/model"
)

// These are the behaviour tables of docs/spec/cmd/verbos-del-ciclo.md, one
// test per row, driven over an open board exactly as the tests of `biso new`
// and `biso set` are.

// texts is the positional text of a verb as the command line hands it over:
// what was typed and what it resolved to, which are the same thing for a
// literal.
func texts(values ...string) []Text {
	out := make([]Text, 0, len(values))
	for _, v := range values {
		out = append(out, Text{Typed: v, Value: v})
	}
	return out
}

// lease writes a lease straight onto a task, which is the one state no
// command of this package can produce on its own: an expired one.
func (h *harness) lease(id, holder string, expiresAt time.Time) {
	h.t.Helper()
	task := h.load(id)
	task.LeaseHolder, task.LeaseExpiresAt = holder, expiresAt
	if err := h.b.Tasks.Save(task); err != nil {
		h.t.Fatal(err)
	}
}

// active leaves a task in the state `biso start` leaves it in, without
// going through the verb under test.
func (h *harness) active(title string, changes ...Change) string {
	h.t.Helper()
	id := h.create(title, changes...)
	h.set(id, scalar("status", "In Progress"), add("add-assignees", "@claude"))
	return id
}

func (h *harness) start(refs []string, p StartParams) (*WriteResult, error) {
	h.t.Helper()
	p.Refs = refs
	return StartOn(h.b, h.env, p)
}

// `biso start`

func TestStartMovesTheTaskToTheActiveStatusAndClaimsTheLease(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	result, err := h.start([]string{id}, StartParams{
		Changes: []Change{add("append-plan", "1. Read the parser.")},
	})
	if err != nil {
		t.Fatal(err)
	}

	task := h.load(id)
	if task.Status != "In Progress" {
		t.Errorf("status = %q, want the board's active status", task.Status)
	}
	if strings.Join(task.Assignees, ",") != "@claude" {
		t.Errorf("assignees = %v, want the caller alone", task.Assignees)
	}
	if task.LeaseHolder != "@claude" {
		t.Errorf("leaseHolder = %q, want the caller", task.LeaseHolder)
	}
	if want := writeClock.Add(240 * time.Minute); !task.LeaseExpiresAt.Equal(want) {
		t.Errorf("leaseExpiresAt = %v, want %v", task.LeaseExpiresAt, want)
	}
	if task.Plan != "1. Read the parser." {
		t.Errorf("plan = %q, and --append-plan wrote it", task.Plan)
	}
	if result.Tasks[0].Status != "In Progress" {
		t.Errorf("the status line says %q", result.Tasks[0].Status)
	}
}

func TestStartOnATaskThatIsAlreadyActiveAppliesTheRestAndSaysSo(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	result, err := h.start([]string{id}, StartParams{
		Changes: []Change{add("append-plan", "2. Add the CRLF case.")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, id+" was already In Progress") {
		t.Errorf("notes = %v, want the one of an already active task", result.Notes)
	}
	if h.load(id).Plan != "2. Add the CRLF case." {
		t.Errorf("the rest of the call was not applied")
	}
}

func TestStartOnAFinishedTaskRefusesUnlessReopen(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")
	h.set(id, scalar("status", "Done"))

	_, err := h.start([]string{id}, StartParams{})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "already_finished" {
		t.Fatalf("error = %d/%s, want 6/already_finished", e.ExitCode, e.Code)
	}
	if len(e.Hints) == 0 || !strings.Contains(e.Hints[0], "--reopen") {
		t.Errorf("hints = %v, want the pointer to --reopen", e.Hints)
	}

	if _, err := h.start([]string{id}, StartParams{Reopen: true}); err != nil {
		t.Fatalf("--reopen: %v", err)
	}
	if h.load(id).Status != "In Progress" {
		t.Errorf("--reopen did not bring the task back to the active status")
	}
}

// The one refusal only `biso start` has, because it is the one verb that
// claims a lease (docs/spec/lease.md#el-vaciado).
func TestStartOnAnArchivedTaskRefusesWithTheHintOfUnarchive(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")
	task := h.load(id)
	task.Archived = true
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	_, err := h.start([]string{id}, StartParams{})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "precondition_failed" {
		t.Fatalf("error = %d/%s, want 6/precondition_failed", e.ExitCode, e.Code)
	}
	if e.Message != id+" is archived" {
		t.Errorf("message = %q", e.Message)
	}
	if len(e.Hints) == 0 || !strings.Contains(e.Hints[0], "--unarchive") {
		t.Errorf("hints = %v, want the pointer to biso archive --unarchive", e.Hints)
	}
}

func TestStartWarnsAboutUnfinishedDependenciesAndStartsAnyway(t *testing.T) {
	h := newHarness(t)
	blocker := h.create("The parser")
	id := h.create("Normalize CRLF", add("add-deps", blocker))

	result, err := h.start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "unresolved_dependencies") {
		t.Errorf("warnings = %v, want the one of a blocked task", result.Warnings)
	}
	if h.load(id).Status != "In Progress" {
		t.Errorf("the warning stopped the write, and it only warns")
	}
}

func TestStartWarnsAboutAnOpenQuestionAndStartsAnyway(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")
	task := h.load(id)
	task.Question = &model.Question{Author: "@sara", AskedAt: writeClock, Body: "Binary too?"}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	result, err := h.start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	w := warningOf(result, "open_question_on_start")
	if w == nil {
		t.Fatalf("warnings = %v, want open_question_on_start", result.Warnings)
	}
	if w.Message != id+" has an open question, asked by @sara" {
		t.Errorf("message = %q", w.Message)
	}
	if h.load(id).Status != "In Progress" {
		t.Errorf("the warning stopped the write, and it only warns")
	}
}

func TestStartTakesALiveLeaseOfAnotherIdentityWithAWarning(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.lease(id, "@sara", writeClock.Add(time.Hour))

	result, err := h.start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	w := warningOf(result, "lease_held")
	if w == nil {
		t.Fatalf("warnings = %v, want lease_held", result.Warnings)
	}
	if !strings.Contains(w.Message, "@sara") {
		t.Errorf("message = %q, want the holder in it", w.Message)
	}
	if h.load(id).LeaseHolder != "@claude" {
		t.Errorf("the lease was not taken over, and this row takes it")
	}
}

func TestStartReclaimsAnExpiredLease(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.lease(id, "@sara", writeClock.Add(-time.Hour))

	result, err := h.start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	if warned(result, "lease_held") {
		t.Errorf("an expired lease is not held by anybody: %v", result.Warnings)
	}
	task := h.load(id)
	if task.LeaseHolder != "@claude" {
		t.Errorf("leaseHolder = %q, want the caller", task.LeaseHolder)
	}
	if want := writeClock.Add(240 * time.Minute); !task.LeaseExpiresAt.Equal(want) {
		t.Errorf("leaseExpiresAt = %v, want %v", task.LeaseExpiresAt, want)
	}
}

// -s with a status that is not the active one claims nothing, and empties
// what the task had, because the fields only hold a value on an active and
// assigned task (docs/spec/lease.md#el-vaciado).
func TestStartWithAnotherStatusClaimsNoLeaseAndEmptiesTheOneItHad(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.lease(id, "@claude", writeClock.Add(time.Hour))

	result, err := h.start([]string{id}, StartParams{
		Changes: []Change{scalar("status", "To Do")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, id+" was moved to To Do, no lease was claimed") {
		t.Errorf("notes = %v, want the one of -s to another status", result.Notes)
	}
	task := h.load(id)
	if task.LeaseHolder != "" || !task.LeaseExpiresAt.IsZero() {
		t.Errorf("the lease survived at %q until %v", task.LeaseHolder, task.LeaseExpiresAt)
	}
}

func TestStartLeavesATaskSomebodyElseHasAssignedAsItIs(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF", add("add-assignees", "@sara"))

	result, err := h.start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, id+" is assigned to @sara, left as is") {
		t.Errorf("notes = %v, want the one of a task somebody else has", result.Notes)
	}
	if strings.Join(h.load(id).Assignees, ",") != "@sara" {
		t.Errorf("assignees = %v, and `me` was added on top", h.load(id).Assignees)
	}
}

func TestStartAddsWhatExplicitAddAssigneesSays(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF", add("add-assignees", "@sara"))

	if _, err := h.start([]string{id}, StartParams{
		Changes: []Change{add("add-assignees", "@juan")},
	}); err != nil {
		t.Fatal(err)
	}

	if got := strings.Join(h.load(id).Assignees, ","); got != "@sara,@juan" {
		t.Errorf("assignees = %q, want what -a said added to what was there", got)
	}
}

func TestStartWithoutAnIdentityAssignsNobodyAndClaimsNothing(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	result, err := h.as("").start([]string{id}, StartParams{})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, "no identity configured, task left unassigned") {
		t.Errorf("notes = %v, want the one of no identity", result.Notes)
	}
	task := h.load(id)
	if len(task.Assignees) != 0 {
		t.Errorf("assignees = %v, want nobody", task.Assignees)
	}
	if task.LeaseHolder != "" {
		t.Errorf("leaseHolder = %q, and there is no identity to attribute it to", task.LeaseHolder)
	}
}

func TestStartAppendsToAPlanTheTaskAlreadyHad(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF", add("append-plan", "1. Read the parser."))

	if _, err := h.start([]string{id}, StartParams{
		Changes: []Change{add("append-plan", "2. Add the CRLF case.")},
	}); err != nil {
		t.Fatal(err)
	}

	if got := h.load(id).Plan; got != "1. Read the parser.\n\n2. Add the CRLF case." {
		t.Errorf("plan = %q, want the second paragraph behind the first", got)
	}
}

func TestStartOverSeveralTasksIsAllOrNothing(t *testing.T) {
	h := newHarness(t)
	first := h.create("First")
	second := h.create("Second")
	h.set(second, scalar("status", "Done"))

	_, err := h.start([]string{first, second}, StartParams{})

	assertSpec(t, err, 6, "already_finished")
	if h.load(first).Status != "To Do" {
		t.Errorf("the first task was written although the call failed")
	}
}

// `biso note`

func TestNoteAppendsOneParagraphPerText(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	if _, err := NoteOn(h.b, h.env, NoteParams{
		Ref: id, Texts: texts("First finding", "Second finding"),
	}); err != nil {
		t.Fatal(err)
	}

	if got := h.load(id).Notes; got != "First finding\n\nSecond finding" {
		t.Errorf("notes = %q, want one paragraph per text, in order", got)
	}
}

func TestNoteRefusesAPositionalThatLooksLikeATaskID(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	for _, typed := range []string{"MYP-2", "myp-2", "2", "#2"} {
		_, err := NoteOn(h.b, h.env, NoteParams{Ref: id, Texts: texts(typed)})
		e := specError(t, err)
		if e.ExitCode != 2 || e.Code != "id_like_positional" {
			t.Fatalf("%q: error = %d/%s, want 2/id_like_positional", typed, e.ExitCode, e.Code)
		}
		if e.Message != "\""+typed+"\" looks like a task id, and `biso note` takes only one task" {
			t.Errorf("message = %q", e.Message)
		}
	}
}

// The flag never goes through that check: it is the way of writing a note
// that really says MYP-2.
func TestNoteWritesAnIDLikeTextThroughTheFieldFlag(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	if _, err := NoteOn(h.b, h.env, NoteParams{
		Ref: id, Changes: []Change{add("append-note", "MYP-2")},
	}); err != nil {
		t.Fatal(err)
	}

	if h.load(id).Notes != "MYP-2" {
		t.Errorf("notes = %q, want the text the flag carried", h.load(id).Notes)
	}
}

func TestNoteWithAnEmptyTextAddsNothingAndWarns(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF", add("append-note", "What was there"))

	result, err := NoteOn(h.b, h.env, NoteParams{Ref: id, Texts: texts("   ")})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "empty_append") {
		t.Errorf("warnings = %v, want empty_append", result.Warnings)
	}
	if h.load(id).Notes != "What was there" {
		t.Errorf("notes = %q, and nothing should have been added", h.load(id).Notes)
	}
}

func TestNoteWithNoTextAndNoFieldFlagIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	_, err := NoteOn(h.b, h.env, NoteParams{Ref: id})

	assertSpec(t, err, 2, "missing_text")
	assertMissingText(t, err, "biso note needs a text to append")
}

// assertMissingText is the whole of that refusal: its literal message, and
// the two detail keys it does not carry. `field` and `given` belong to the
// errors that name a flag, a configuration key or a concrete value
// (docs/spec/contrato-json.md#los-errores-en-json), and a positional that
// was never written names none of the three, exactly like the reference
// that does not exist.
func assertMissingText(t *testing.T, err error, message string) {
	t.Helper()
	e := specError(t, err)
	if e.Message != message {
		t.Errorf("message = %q, want %q", e.Message, message)
	}
	if e.Field != "" || e.Given != "" {
		t.Errorf("field = %q and given = %q, and this error names no value", e.Field, e.Given)
	}
	if len(e.Hints) != 1 {
		t.Errorf("hints = %v, want the example of the call that would have worked", e.Hints)
	}
}

func TestNoteOverAnArchivedTaskIsDoneWithANote(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")
	task := h.load(id)
	task.Archived = true
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	result, err := NoteOn(h.b, h.env, NoteParams{Ref: id, Texts: texts("A finding")})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, id+" is archived") {
		t.Errorf("notes = %v, want the archived one", result.Notes)
	}
	if h.load(id).Notes != "A finding" {
		t.Errorf("the note was not written, and only `start` refuses here")
	}
}

// `biso comment`

func TestCommentAppendsACommentAndSaysWhichKeyItTook(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	result, err := CommentOn(h.b, h.env, CommentParams{
		Ref: id, Texts: texts("A user with a Windows clone reported this"),
	})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, "comment #1 by @claude") {
		t.Errorf("notes = %v, want the key the comment took", result.Notes)
	}
	comments := h.load(id).Comments
	if len(comments) != 1 || comments[0].Author != "@claude" {
		t.Fatalf("comments = %v", comments)
	}
	if comments[0].Body != "A user with a Windows clone reported this" {
		t.Errorf("body = %q", comments[0].Body)
	}
}

// The author is free text, validated against nothing, and the leading "@"
// is never a file reference.
func TestCommentTakesItsAuthorExactlyAsItIsWritten(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	result, err := CommentOn(h.b, h.env, CommentParams{
		Ref: id, Texts: texts("Moved to Doing from the phone"),
		Changes: []Change{{Flag: "comment-author", Value: "@trello:juan"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	if h.load(id).Comments[0].Author != "@trello:juan" {
		t.Errorf("author = %q", h.load(id).Comments[0].Author)
	}
	if !noted(result, "comment #1 by @trello:juan") {
		t.Errorf("notes = %v, want the real author in it", result.Notes)
	}
}

func TestCommentWithoutAnAuthorAndWithoutAnIdentityIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	_, err := CommentOn(h.b, h.as("").env, CommentParams{Ref: id, Texts: texts("Something")})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "missing_identity" {
		t.Fatalf("error = %d/%s, want 2/missing_identity", e.ExitCode, e.Code)
	}
	if e.Message != "--comment-author is required, no identity is configured" {
		t.Errorf("message = %q", e.Message)
	}
}

// --comment-author is not a text to append and it is not a field flag
// either: it only says who signs a comment that some other flag writes, so
// a call that carries nothing else has nothing to append and is the same
// refusal as a call with nothing at all
// (docs/spec/cmd/verbos-del-ciclo.md#biso-comment).
func TestCommentWithItsAuthorAloneHasNothingToAppend(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	_, err := CommentOn(h.b, h.env, CommentParams{
		Ref:     id,
		Changes: []Change{{Flag: "comment-author", Value: "@trello:juan"}},
	})

	assertSpec(t, err, 2, "missing_text")
	assertMissingText(t, err, "biso comment needs a text to append")
	if len(h.load(id).Comments) != 0 {
		t.Errorf("a call that could write no comment wrote one")
	}
}

func TestCommentRefusesAPositionalThatLooksLikeATaskID(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	_, err := CommentOn(h.b, h.env, CommentParams{Ref: id, Texts: texts("MYP-2")})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "id_like_positional" {
		t.Fatalf("error = %d/%s, want 2/id_like_positional", e.ExitCode, e.Code)
	}
	if !strings.Contains(e.Message, "`biso comment`") {
		t.Errorf("message = %q, want this command's name in it", e.Message)
	}
}

// `biso finish`

func TestFinishClosesTheTaskAndEmptiesTheLease(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF", add("add-ac", "The parser accepts CRLF"))
	h.lease(id, "@claude", writeClock.Add(time.Hour))

	result, err := FinishOn(h.b, h.env, FinishParams{
		Refs:    []string{id},
		Changes: []Change{check("check-ac", "all"), add("append-summary", "Normalizes CRLF")},
	})
	if err != nil {
		t.Fatal(err)
	}

	task := h.load(id)
	if task.Status != "Done" {
		t.Errorf("status = %q, want the board's terminal status", task.Status)
	}
	if task.AcDone() != 1 || task.Summary != "Normalizes CRLF" {
		t.Errorf("the field flags of the call were not applied")
	}
	if task.LeaseHolder != "" || !task.LeaseExpiresAt.IsZero() {
		t.Errorf("the lease survived a finish")
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and nothing was missing", result.Warnings)
	}
}

func TestFinishWarnsAboutUncheckedCriteriaAndListsThem(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF",
		add("add-ac", "The parser accepts CRLF"),
		add("add-ac", "There is a test that covers it"))
	h.set(id, check("check-ac", "1"), add("append-summary", "Done"))

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}})
	if err != nil {
		t.Fatal(err)
	}

	w := warningOf(result, "terminal_ac_unchecked")
	if w == nil {
		t.Fatalf("warnings = %v, want terminal_ac_unchecked", result.Warnings)
	}
	if w.Message != id+" moved to Done with 1 of 2 acceptance criteria unchecked" {
		t.Errorf("message = %q", w.Message)
	}
	if len(w.Detail) != 1 || w.Detail[0] != "  #2 There is a test that covers it" {
		t.Errorf("detail = %v, want the criterion that is missing", w.Detail)
	}
	if h.load(id).Status != "Done" {
		t.Errorf("the task did not close, and this row closes it")
	}
}

func TestFinishWithStrictRefusesAndWritesNothing(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF", add("add-ac", "The parser accepts CRLF"))

	_, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}, Strict: true})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "precondition_failed" {
		t.Fatalf("error = %d/%s, want 6/precondition_failed", e.ExitCode, e.Code)
	}
	if len(e.Detail) != 2 {
		t.Errorf("detail = %v, want the two things that are missing", e.Detail)
	}
	if h.load(id).Status != "In Progress" {
		t.Errorf("--strict wrote after refusing")
	}
}

// --strict defaults to the board's finish_strict, which is what makes the
// hard policy settable once instead of typed on every call.
func TestFinishTakesItsStrictDefaultFromTheBoard(t *testing.T) {
	h := newHarness(t)
	h.b.Config.FinishStrict = true
	id := h.active("Normalize CRLF", add("add-ac", "The parser accepts CRLF"))

	_, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}})

	assertSpec(t, err, 6, "precondition_failed")
}

func TestFinishWarnsWhenTheSummaryEndsUpEmpty(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}})
	if err != nil {
		t.Fatal(err)
	}

	w := warningOf(result, "terminal_no_summary")
	if w == nil {
		t.Fatalf("warnings = %v, want terminal_no_summary", result.Warnings)
	}
	if w.Message != id+" finished without a final summary" {
		t.Errorf("message = %q", w.Message)
	}
}

// The checks read the field as this write leaves it, never whether the flag
// was written now: a task that kept the summary of an earlier close does not
// warn on being closed again.
func TestFinishReadsTheSummaryOfTheResultAndNotOfThisCall(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	if _, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Changes: []Change{add("append-summary", "Normalizes CRLF")},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := h.start([]string{id}, StartParams{Reopen: true}); err != nil {
		t.Fatal(err)
	}

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}})
	if err != nil {
		t.Fatal(err)
	}

	if warned(result, "terminal_no_summary") {
		t.Errorf("warnings = %v, and the summary of the first close is still there",
			result.Warnings)
	}
}

func TestFinishWarnsAboutUnfinishedSubtasksAndMarksTheArchivedOnes(t *testing.T) {
	h := newHarness(t)
	parent := h.active("Normalize CRLF", add("append-summary", "Done"))
	h.create("A live child", scalar("parent", parent))
	archived := h.create("An archived child", scalar("parent", parent))
	task := h.load(archived)
	task.Archived = true
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{parent}})
	if err != nil {
		t.Fatal(err)
	}

	w := warningOf(result, "unfinished_subtasks")
	if w == nil {
		t.Fatalf("warnings = %v, want unfinished_subtasks", result.Warnings)
	}
	if !strings.HasSuffix(w.Message, " (archived)") {
		t.Errorf("message = %q, want the archived child marked", w.Message)
	}
}

// The open question warns and never refuses, not even with --strict:
// refusing would only push the caller into `biso set`.
func TestFinishNeverRefusesOverAnOpenQuestion(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF", add("append-summary", "Done"))
	if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")}); err != nil {
		t.Fatal(err)
	}

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}, Strict: true})
	if err != nil {
		t.Fatalf("--strict refused over an open question: %v", err)
	}

	if !warned(result, "open_question_on_terminal") {
		t.Errorf("warnings = %v, want open_question_on_terminal", result.Warnings)
	}
}

func TestFinishWithNoChecksSkipsEveryWarning(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF", add("add-ac", "The parser accepts CRLF"))
	if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")}); err != nil {
		t.Fatal(err)
	}

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}, NoChecks: true})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and --no-checks emits none of them", result.Warnings)
	}
	if h.load(id).Status != "Done" {
		t.Errorf("--no-checks did not close the task")
	}
}

func TestFinishOverAFinishedTaskAppliesTheRestAndSaysSo(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	if _, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Changes: []Change{add("append-summary", "First close")},
	}); err != nil {
		t.Fatal(err)
	}

	result, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Changes: []Change{add("append-note", "One more thing")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if !noted(result, id+" was already Done") {
		t.Errorf("notes = %v, want the one of an already closed task", result.Notes)
	}
	if h.load(id).Notes != "One more thing" {
		t.Errorf("the rest of the call was not applied")
	}
}

// The lease is emptied for two different reasons, and each one is its own
// row of the table (docs/spec/lease.md#el-vaciado).
func TestFinishEmptiesALeaseOfAnotherIdentityAndWarns(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.lease(id, "@sara", writeClock.Add(time.Hour))

	result, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}})
	if err != nil {
		t.Fatal(err)
	}

	if !warned(result, "lease_held") {
		t.Errorf("warnings = %v, want the one of somebody else's lease", result.Warnings)
	}
	task := h.load(id)
	if task.LeaseHolder != "" || !task.LeaseExpiresAt.IsZero() {
		t.Errorf("the lease of @sara survived the close")
	}
}

func TestFinishToAStatusThatIsNotTheTerminalOneEmptiesTheLeaseAllTheSame(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.lease(id, "@claude", writeClock.Add(time.Hour))

	if _, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Changes: []Change{scalar("status", "To Do")},
	}); err != nil {
		t.Fatal(err)
	}

	task := h.load(id)
	if task.LeaseHolder != "" || !task.LeaseExpiresAt.IsZero() {
		t.Errorf("what holds the lease is being active, not reaching the terminal status")
	}
}

// The four checks of this verb are the checks of arriving at the terminal
// status, so a -s that names another status has no close to check: neither
// the warnings nor the error 6 of --strict come out
// (docs/spec/cmd/verbos-del-ciclo.md#biso-finish).
func TestFinishToAStatusThatIsNotTheTerminalOneChecksNothing(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF", add("add-ac", "The parser accepts CRLF"))

	result, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Changes: []Change{scalar("status", "To Do")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Warnings) != 0 {
		t.Errorf("warnings = %v, and the task arrived at no terminal status", result.Warnings)
	}

	// And the hard policy says the same thing: there is nothing to be
	// strict about when nothing is being closed.
	if _, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{id}, Strict: true, Changes: []Change{scalar("status", "To Do")},
	}); err != nil {
		t.Errorf("biso finish -s --strict = %v, and it closed nothing to refuse", err)
	}
	if h.load(id).Status != "To Do" {
		t.Errorf("status = %q, want the one -s named", h.load(id).Status)
	}
}

func TestFinishOverSeveralTasksIsAllOrNothing(t *testing.T) {
	h := newHarness(t)
	first := h.active("First")
	second := h.active("Second", add("add-ac", "Something"))

	_, err := FinishOn(h.b, h.env, FinishParams{
		Refs: []string{first, second}, Strict: true,
	})

	assertSpec(t, err, 6, "precondition_failed")
	if h.load(first).Status != "In Progress" {
		t.Errorf("the first task was closed although the call failed")
	}
}

// `biso ask`

func TestAskFillsTheQuestionWithoutChangingTheStatus(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	if _, err := AskOn(h.b, h.env, AskParams{
		Ref: id, Texts: texts("Do we normalize binary files too, or only text?"),
	}); err != nil {
		t.Fatal(err)
	}

	task := h.load(id)
	if task.Question == nil {
		t.Fatal("the question was not written")
	}
	if task.Question.Author != "@claude" || !task.Question.AskedAt.Equal(writeClock) {
		t.Errorf("question = %+v, want the caller and the clock of the call", task.Question)
	}
	if task.Question.Body != "Do we normalize binary files too, or only text?" {
		t.Errorf("body = %q", task.Question.Body)
	}
	if task.Status != "In Progress" {
		t.Errorf("status = %q, and asking does not change it", task.Status)
	}
}

func TestAskOverATaskThatAlreadyHasAQuestionRefuses(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("First")}); err != nil {
		t.Fatal(err)
	}

	_, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Second")})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "open_question_exists" {
		t.Fatalf("error = %d/%s, want 6/open_question_exists", e.ExitCode, e.Code)
	}
	if e.Message != id+" already has an open question" {
		t.Errorf("message = %q", e.Message)
	}
	if h.load(id).Question.Body != "First" {
		t.Errorf("the second question overwrote the first in silence")
	}
}

func TestAskOverAFinishedTaskRefusesWithTheHintOfReopening(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	h.set(id, scalar("status", "Done"))

	_, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "already_finished" {
		t.Fatalf("error = %d/%s, want 6/already_finished", e.ExitCode, e.Code)
	}
	if len(e.Hints) == 0 || !strings.Contains(e.Hints[0], "--reopen") {
		t.Errorf("hints = %v, want the pointer to reopening", e.Hints)
	}
}

func TestAskWithAnEmptyQuestionIsExitCodeThree(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("   ")})

	e := specError(t, err)
	if e.ExitCode != 3 || e.Code != "empty_scalar_value" {
		t.Fatalf("error = %d/%s, want 3/empty_scalar_value", e.ExitCode, e.Code)
	}
	if e.Message != "the question cannot be empty" {
		t.Errorf("message = %q", e.Message)
	}
}

func TestAskWithoutAnIdentityIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, err := AskOn(h.b, h.as("").env, AskParams{Ref: id, Texts: texts("Binary too?")})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "missing_identity" {
		t.Fatalf("error = %d/%s, want 2/missing_identity", e.ExitCode, e.Code)
	}
	want := "biso ask needs an identity; set BISO_ME, or add \"me\" to ~/.biso/config.json"
	if e.Message != want {
		t.Errorf("message = %q, want %q", e.Message, want)
	}
}

// The message is this command's own: `biso ask` has no field flag that
// writes the question, so the only way out is @file or -.
func TestAskRefusesAPositionalThatLooksLikeATaskIDWithItsOwnHint(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("MYP-2")})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "id_like_positional" {
		t.Fatalf("error = %d/%s, want 2/id_like_positional", e.ExitCode, e.Code)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "to write that text literally, use @file or - for stdin" {
		t.Errorf("hints = %v", e.Hints)
	}
}

func TestAskWithNoTextAtAllIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, err := AskOn(h.b, h.env, AskParams{Ref: id})

	assertSpec(t, err, 2, "missing_text")
	assertMissingText(t, err, "biso ask needs a question")
}

// `biso answer`

func TestAnswerWritesTheTwoCommentsInOrderAndEmptiesTheQuestion(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	asked := writeClock.Add(-2 * time.Hour)
	task := h.load(id)
	task.Question = &model.Question{
		Author: "@sara", AskedAt: asked, Body: "Do we normalize binary files too?",
	}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	if _, err := AnswerOn(h.b, h.env, AnswerParams{
		Ref: id, Texts: texts("Only text files. Binary ones are skipped entirely."),
	}); err != nil {
		t.Fatal(err)
	}

	written := h.load(id)
	if written.Question != nil {
		t.Errorf("question = %+v, want it emptied", written.Question)
	}
	if len(written.Comments) != 2 {
		t.Fatalf("comments = %v, want the question and the answer", written.Comments)
	}
	question, answer := written.Comments[0], written.Comments[1]
	if question.Author != "@sara" || !question.CreatedAt.Equal(asked) {
		t.Errorf("the question comment did not keep its author and its instant: %+v", question)
	}
	if question.Body != "Do we normalize binary files too?" {
		t.Errorf("question body = %q", question.Body)
	}
	if answer.Author != "@claude" || !answer.CreatedAt.Equal(writeClock) {
		t.Errorf("the answer is not signed by the caller now: %+v", answer)
	}
}

// The two comments of the verb go first whatever the command line said,
// because the order of a write does not depend on it.
func TestAnswerWritesItsTwoCommentsBeforeAnyCommentFlagOfTheSameCall(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	task := h.load(id)
	task.Question = &model.Question{Author: "@sara", AskedAt: writeClock, Body: "Binary too?"}
	if err := h.b.Tasks.Save(task); err != nil {
		t.Fatal(err)
	}

	if _, err := AnswerOn(h.b, h.env, AnswerParams{
		Ref: id, Texts: texts("Only text files."),
		Changes: []Change{comment("And one more thing")},
	}); err != nil {
		t.Fatal(err)
	}

	var bodies []string
	for _, c := range h.load(id).Comments {
		bodies = append(bodies, c.Body)
	}
	if strings.Join(bodies, "|") != "Binary too?|Only text files.|And one more thing" {
		t.Errorf("comments = %v", bodies)
	}
}

func TestAnswerWithoutAnOpenQuestionRefuses(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, err := AnswerOn(h.b, h.env, AnswerParams{Ref: id, Texts: texts("Only text files.")})

	e := specError(t, err)
	if e.ExitCode != 6 || e.Code != "no_open_question" {
		t.Fatalf("error = %d/%s, want 6/no_open_question", e.ExitCode, e.Code)
	}
	if len(e.Hints) == 0 || !strings.Contains(e.Hints[0], "biso comment") {
		t.Errorf("hints = %v, want the pointer to biso comment", e.Hints)
	}
}

// Not symmetric with `biso ask`, and on purpose: answering is the one way of
// recovering a question that stayed open when the task was closed.
func TestAnswerWorksOverAFinishedTask(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")}); err != nil {
		t.Fatal(err)
	}
	if _, err := FinishOn(h.b, h.env, FinishParams{Refs: []string{id}}); err != nil {
		t.Fatal(err)
	}

	if _, err := AnswerOn(h.b, h.env, AnswerParams{
		Ref: id, Texts: texts("Only text files."),
	}); err != nil {
		t.Fatalf("answering a closed task: %v", err)
	}

	if h.load(id).Question != nil {
		t.Errorf("the question was not emptied")
	}
}

func TestAnswerWithNoTextAndWithAnEmptyOneAreTwoDifferentErrors(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")

	_, missing := AnswerOn(h.b, h.env, AnswerParams{Ref: id})
	assertSpec(t, missing, 2, "missing_text")
	assertMissingText(t, missing, "biso answer needs an answer")

	_, empty := AnswerOn(h.b, h.env, AnswerParams{Ref: id, Texts: texts("  ")})
	e := specError(t, empty)
	if e.ExitCode != 3 || e.Message != "the answer cannot be empty" {
		t.Errorf("error = %d %q, want 3 and the answer's own message", e.ExitCode, e.Message)
	}
}

func TestAnswerWithoutAnIdentityIsAUsageError(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")}); err != nil {
		t.Fatal(err)
	}

	_, err := AnswerOn(h.b, h.as("").env, AnswerParams{Ref: id, Texts: texts("Only text.")})

	e := specError(t, err)
	if e.ExitCode != 2 || e.Code != "missing_identity" {
		t.Fatalf("error = %d/%s, want 2/missing_identity", e.ExitCode, e.Code)
	}
	if !strings.HasPrefix(e.Message, "biso answer needs an identity") {
		t.Errorf("message = %q", e.Message)
	}
}

// Every verb prints the status line of docs/spec/cmd/set.md#salida, and
// --dry-run computes it without writing anything.
func TestEveryVerbPreviewsWithDryRunAndWritesNothing(t *testing.T) {
	h := newHarness(t)
	id := h.active("Normalize CRLF")
	before := h.load(id)

	result, err := NoteOn(h.b, h.env, NoteParams{
		Ref: id, Texts: texts("A finding"), DryRun: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.DryRun || len(result.Tasks) != 1 {
		t.Fatalf("result = %+v, want one hypothetical task", result)
	}
	if h.load(id).Notes != before.Notes {
		t.Errorf("the preview wrote the note after all")
	}
}

func TestCommentAppendsOneCommentPerTextInOrder(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")

	result, err := CommentOn(h.b, h.env, CommentParams{
		Ref: id, Texts: texts("First thing", "Second thing"),
	})
	if err != nil {
		t.Fatal(err)
	}

	var bodies []string
	for _, c := range h.load(id).Comments {
		bodies = append(bodies, c.Body)
	}
	if strings.Join(bodies, "|") != "First thing|Second thing" {
		t.Errorf("comments = %v, want one per text, in order", bodies)
	}
	for _, key := range []string{"comment #1 by @claude", "comment #2 by @claude"} {
		if !noted(result, key) {
			t.Errorf("notes = %v, want %q among them", result.Notes, key)
		}
	}
}

// Which comments are new is asked of the keys, so a call that removes two
// and adds one names the one it added instead of walking off the list.
func TestCommentNamesOnlyTheCommentsItAddedWhenTheCallAlsoRemovesSome(t *testing.T) {
	h := newHarness(t)
	id := h.create("Normalize CRLF")
	h.set(id, comment("One"), comment("Two"), comment("Three"))

	result, err := CommentOn(h.b, h.env, CommentParams{
		Ref: id, Texts: texts("The new one"),
		Changes: []Change{remove("rm-comment", "1"), remove("rm-comment", "2")},
	})
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Notes) != 1 || !noted(result, "comment #4 by @claude") {
		t.Errorf("notes = %v, want only the comment this call added", result.Notes)
	}
}

// The verbs of the cycle need no field flag to have something to do, so
// their message about a missing reference does not point at one.
func TestAVerbWithNoReferenceNamesItselfAndNotBisoSet(t *testing.T) {
	h := newHarness(t)

	_, err := h.start(nil, StartParams{})

	e := specError(t, err)
	if e.Message != "biso start needs at least one task reference" {
		t.Errorf("message = %q", e.Message)
	}
	if len(e.Hints) != 1 || e.Hints[0] != "biso start MYP-11" {
		t.Errorf("hints = %v", e.Hints)
	}
}

// And `biso set` keeps the one docs/spec/cmd/set.md prints.
func TestSetWithNoReferenceKeepsTheHintOfItsOwnPage(t *testing.T) {
	h := newHarness(t)

	_, err := SetOn(h.b, h.env, SetParams{Changes: []Change{add("add-labels", "parser")}})

	e := specError(t, err)
	if len(e.Hints) != 1 || e.Hints[0] != "biso set MYP-11 --priority high" {
		t.Errorf("hints = %v", e.Hints)
	}
}

// Only `biso start` refuses over an archived task, because only `biso start`
// claims a lease (docs/spec/lease.md#el-vaciado). The other five write over
// it like over any other, saying so on stderr exactly as `biso get` does.
func TestTheFiveVerbsThatClaimNoLeaseWriteOverAnArchivedTask(t *testing.T) {
	for _, c := range []struct {
		name string
		run  func(h *harness, id string) (*WriteResult, error)
	}{
		{"note", func(h *harness, id string) (*WriteResult, error) {
			return NoteOn(h.b, h.env, NoteParams{Ref: id, Texts: texts("A finding")})
		}},
		{"comment", func(h *harness, id string) (*WriteResult, error) {
			return CommentOn(h.b, h.env, CommentParams{Ref: id, Texts: texts("Something")})
		}},
		{"finish", func(h *harness, id string) (*WriteResult, error) {
			return FinishOn(h.b, h.env, FinishParams{Refs: []string{id}, NoChecks: true})
		}},
		{"ask", func(h *harness, id string) (*WriteResult, error) {
			return AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")})
		}},
		{"answer", func(h *harness, id string) (*WriteResult, error) {
			if _, err := AskOn(h.b, h.env, AskParams{Ref: id, Texts: texts("Binary too?")}); err != nil {
				return nil, err
			}
			return AnswerOn(h.b, h.env, AnswerParams{Ref: id, Texts: texts("Only text.")})
		}},
	} {
		t.Run(c.name, func(t *testing.T) {
			h := newHarness(t)
			id := h.active("Normalize CRLF")
			task := h.load(id)
			task.Archived = true
			if err := h.b.Tasks.Save(task); err != nil {
				t.Fatal(err)
			}

			result, err := c.run(h, id)
			if err != nil {
				t.Fatalf("biso %s over an archived task: %v", c.name, err)
			}
			if !noted(result, id+" is archived") {
				t.Errorf("notes = %v, want the archived one", result.Notes)
			}
		})
	}
}
