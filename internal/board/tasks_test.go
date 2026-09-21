package board

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"biso/internal/model"
	"biso/internal/store"
)

// testBoardID stands in for a real board id, the eight characters of
// docs/spec/resolucion-del-tablero.md.
const testBoardID = "3f9a2b1c"

// testPrefix is the task_prefix of the board every example of the
// specification belongs to (docs/spec/modelo-de-datos/identificadores.md).
// TASK-10 receives it already decided: deriving it from the board's name is
// `biso init`'s job.
const testPrefix = "MYP"

// testExtensions is the `extensions` list the example board declares
// (docs/spec/modelo-de-datos/campos-externos.md).
var testExtensions = []string{"trello.card", "github.issue"}

func openTasks(t *testing.T, path string) (*Tasks, func()) {
	t.Helper()

	s, err := store.Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open the board at %s: %v", path, err)
	}
	return NewTasks(s, testPrefix, testExtensions), func() { s.Close() }
}

func TestCreateAllocatesConsecutiveIdentifiers(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	for i := 1; i <= 3; i++ {
		task := &model.Task{Title: fmt.Sprintf("Task %d", i), Status: "To Do"}
		if err := tasks.Create(task); err != nil {
			t.Fatalf("Create: %v", err)
		}
		want := fmt.Sprintf("MYP-%d", i)
		if task.ID != want {
			t.Fatalf("id = %q, want %q", task.ID, want)
		}
	}

	last, err := tasks.LastAllocated()
	if err != nil {
		t.Fatalf("LastAllocated: %v", err)
	}
	if last != 3 {
		t.Fatalf("LastAllocated() = %d, want 3", last)
	}
}

func TestCreateFillsTheDatesItDoesNotReceiveAndKeepsTheOnesItDoes(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	fresh := &model.Task{Title: "Fresh", Status: "To Do"}
	if err := tasks.Create(fresh); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if fresh.CreatedAt.IsZero() || fresh.UpdatedAt.IsZero() {
		t.Fatalf("Create left the dates empty: %+v", fresh)
	}
	if fresh.CreatedAt.Location() != time.UTC {
		t.Fatalf("createdAt is not in UTC: %v", fresh.CreatedAt)
	}
	if fresh.CreatedAt.Nanosecond() != 0 {
		t.Fatalf("createdAt keeps sub-second precision: %v", fresh.CreatedAt)
	}

	// docs/spec/modelo-de-datos/fechas.md: the dates can be given only when
	// importing, and then they are kept exactly as they come.
	imported := &model.Task{
		Title:     "Imported",
		Status:    "To Do",
		CreatedAt: time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC),
		UpdatedAt: time.Date(2026, 9, 6, 11, 40, 18, 0, time.UTC),
	}
	if err := tasks.Create(imported); err != nil {
		t.Fatalf("Create: %v", err)
	}
	back, err := tasks.Load(imported.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !back.CreatedAt.Equal(imported.CreatedAt) || !back.UpdatedAt.Equal(imported.UpdatedAt) {
		t.Fatalf("the imported dates did not survive: %v / %v", back.CreatedAt, back.UpdatedAt)
	}
}

// sampleTask is a task with every field of
// docs/spec/modelo-de-datos/index.md at a non-zero value, so that a round
// trip that drops one is visible.
func sampleTask() *model.Task {
	ordinal := 7
	task := &model.Task{
		Title:          "Normalize CRLF in the diff",
		Status:         "In Progress",
		Type:           "bug",
		Priority:       "high",
		Parent:         "MYP-1",
		Assignees:      []string{"@claude", "@sara"},
		Author:         "@avilches",
		Labels:         []string{"parser", "crlf"},
		Dependencies:   []string{"MYP-4", "MYP-5"},
		References:     []string{"docs/bugs/BUG-02.md"},
		ModifiedFiles:  []string{"internal/diff/diff.go"},
		Due:            time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		Ordinal:        &ordinal,
		Ext:            map[string]string{"trello.card": "5f2a8c1e", "github.issue": "42"},
		Description:    "The diff compares byte by byte...",
		Plan:           "1. Read the parser.\n2. Add the CRLF case.",
		Notes:          "The parser already normalized LF.",
		Summary:        "Done by normalizing on read.",
		CreatedAt:      time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC),
		UpdatedAt:      time.Date(2026, 9, 6, 11, 40, 18, 0, time.UTC),
		Archived:       true,
		LeaseExpiresAt: time.Date(2026, 9, 6, 15, 40, 18, 0, time.UTC),
		LeaseHolder:    "@claude",
		Question: &model.Question{
			Author:  "@sara",
			AskedAt: time.Date(2026, 9, 6, 12, 0, 0, 0, time.UTC),
			Body:    "Which encoding does the fixture use?",
		},
	}
	task.AddCriterion("The diff ignores CRLF")
	task.AddCriterion("There is a test that covers it")
	task.Criterion(1).Checked = true
	task.AddComment("@avilches", time.Date(2026, 9, 6, 10, 2, 11, 0, time.UTC), "A user with a Windows clone...")
	task.AddComment("@claude", time.Date(2026, 9, 6, 10, 30, 0, 0, time.UTC), "Reproduced it")
	return task
}

func TestEveryFieldSurvivesTheRoundTrip(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	want := sampleTask()
	if err := tasks.Create(want); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := tasks.Load(want.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("the round trip changed the task:\n got %+v\nwant %+v", got, want)
	}
}

func TestSaveRewritesTheListsWithoutRenumberingTheKeys(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	task := sampleTask()
	if err := tasks.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	task.AddCriterion("It is documented")
	if !task.RemoveCriterion(2) {
		t.Fatalf("RemoveCriterion(2) found nothing")
	}
	task.Labels = []string{"crlf"}
	if err := tasks.Save(task); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := tasks.Load(task.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	keys := got.CriterionKeys()
	if len(keys) != 2 || keys[0] != 1 || keys[1] != 3 {
		t.Fatalf("keys = %v, want [1 3]: removing #2 must not renumber anything", keys)
	}
	if got.NextCriterionKey != 4 {
		t.Fatalf("NextCriterionKey = %d, want 4: the counter is stored and only grows", got.NextCriterionKey)
	}
	if len(got.Labels) != 1 || got.Labels[0] != "crlf" {
		t.Fatalf("labels = %v, want [crlf]", got.Labels)
	}
}

func TestAllReadsEveryTaskInIdentifierOrder(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	for i := 1; i <= 12; i++ {
		task := &model.Task{Title: fmt.Sprintf("Task %d", i), Status: "To Do"}
		task.AddCriterion("one")
		if err := tasks.Create(task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	all, skipped, err := tasks.All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}
	if len(skipped) != 0 {
		t.Fatalf("All() skipped %d tasks on a board it wrote itself", len(skipped))
	}
	if len(all) != 12 {
		t.Fatalf("All() returned %d tasks, want 12", len(all))
	}
	// Ascending by identifier, which is the tie-break of every listing
	// (docs/spec/cmd/ls.md), and not the lexicographic order of the ids.
	for i, task := range all {
		want := fmt.Sprintf("MYP-%d", i+1)
		if task.ID != want {
			t.Fatalf("task %d is %q, want %q", i, task.ID, want)
		}
		if task.AcTotal() != 1 {
			t.Fatalf("%s came back with %d criteria, want 1", task.ID, task.AcTotal())
		}
	}
}

// TestTwoProcessesNeverAllocateTheSameIdentifier is guarantee 3 of
// docs/spec/garantias.md, the one TASK-4 deliberately left unproven
// because it depends on identifiers existing. Each writer has its own
// *Store over the same file, which is what stands in for another process
// working from another copy of the project: two goroutines over one
// connection would prove nothing, since that connection serializes them
// on its own.
func TestTwoProcessesNeverAllocateTheSameIdentifier(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")

	// The board exists before the writers arrive, which is the situation
	// the guarantee describes: `biso init` created it once, and everything
	// after that opens it. Creating the file from several connections at
	// the same instant is a race of its own, on switching it to WAL, and
	// internal/store answers it by waiting out the configured time (see
	// TestConcurrentCreationOfTheSameFileWaitsInsteadOfFailingAtOnce);
	// mixing it in here would measure that wait and not the counter.
	seed, err := store.Open(testBoardID, path)
	if err != nil {
		t.Fatalf("create the board: %v", err)
	}
	seed.Close()

	const writers = 4
	const perWriter = 25

	ids := make(chan string, writers*perWriter)
	errs := make(chan error, writers)

	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()

			s, err := store.Open(testBoardID, path)
			if err != nil {
				errs <- err
				return
			}
			defer s.Close()

			tasks := NewTasks(s, testPrefix, testExtensions)
			for i := 0; i < perWriter; i++ {
				task := &model.Task{Title: fmt.Sprintf("writer %d task %d", w, i), Status: "To Do"}
				if err := tasks.Create(task); err != nil {
					errs <- err
					return
				}
				ids <- task.ID
			}
		}(w)
	}
	wg.Wait()
	close(ids)
	close(errs)

	for err := range errs {
		t.Fatalf("a writer failed: %v", err)
	}

	seen := map[string]bool{}
	for id := range ids {
		if seen[id] {
			t.Fatalf("two writers allocated %s", id)
		}
		seen[id] = true
	}
	if len(seen) != writers*perWriter {
		t.Fatalf("%d identifiers for %d tasks", len(seen), writers*perWriter)
	}
	// Every number from 1 to the total, with no gap: nothing was skipped
	// either, which is what says the counter moved once per task.
	for i := 1; i <= writers*perWriter; i++ {
		if !seen[fmt.Sprintf("MYP-%d", i)] {
			t.Fatalf("MYP-%d was never allocated", i)
		}
	}
}

func TestAnIdentifierIsNotReusedAfterArchivingOrAfterTheRowDisappears(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	tasks, done := openTasks(t, path)
	defer done()

	first := &model.Task{Title: "First", Status: "To Do"}
	if err := tasks.Create(first); err != nil {
		t.Fatalf("Create: %v", err)
	}
	second := &model.Task{Title: "Second", Status: "To Do"}
	if err := tasks.Create(second); err != nil {
		t.Fatalf("Create: %v", err)
	}

	first.Archived = true
	if err := tasks.Save(first); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// And the harder half: the row is gone altogether, the way something
	// outside biso could leave it. The counter still does not go back.
	if err := tasks.store.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM task WHERE id = ?", second.ID)
		return err
	}); err != nil {
		t.Fatalf("delete the second task: %v", err)
	}

	third := &model.Task{Title: "Third", Status: "To Do"}
	if err := tasks.Create(third); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if third.ID != "MYP-3" {
		t.Fatalf("id = %q, want MYP-3: an identifier is never reused", third.ID)
	}
}

func TestLoadTellsNeverAllocatedApartFromNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	tasks, done := openTasks(t, path)
	defer done()

	for i := 1; i <= 2; i++ {
		if err := tasks.Create(&model.Task{Title: "T", Status: "To Do"}); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	if err := tasks.store.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("DELETE FROM task WHERE id = 'MYP-2'")
		return err
	}); err != nil {
		t.Fatalf("delete MYP-2: %v", err)
	}

	// docs/spec/referencias.md#los-tres-mensajes-de-no-la-encuentro
	_, err := tasks.Load("MYP-999")
	modelErr, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("Load(MYP-999) = %v, want a *model.Error", err)
	}
	if modelErr.ExitCode != 4 || modelErr.Code != "never_allocated" {
		t.Fatalf("error = %d/%s, want 4/never_allocated", modelErr.ExitCode, modelErr.Code)
	}
	if modelErr.Message != "MYP-999 has never existed on this board" {
		t.Fatalf("message = %q", modelErr.Message)
	}
	if len(modelErr.Notes) != 1 || modelErr.Notes[0] != "the highest id ever assigned here is MYP-2" {
		t.Fatalf("notes = %v", modelErr.Notes)
	}

	_, err = tasks.Load("MYP-2")
	modelErr, ok = err.(*model.Error)
	if !ok {
		t.Fatalf("Load(MYP-2) = %v, want a *model.Error", err)
	}
	if modelErr.ExitCode != 4 || modelErr.Code != "not_found" {
		t.Fatalf("error = %d/%s, want 4/not_found", modelErr.ExitCode, modelErr.Code)
	}
	if modelErr.Message != "MYP-2 is not on this board" {
		t.Fatalf("message = %q", modelErr.Message)
	}
	if len(modelErr.Notes) != 1 {
		t.Fatalf("notes = %v, want the one that says it was assigned at some point", modelErr.Notes)
	}
	if len(modelErr.Hints) != 1 {
		t.Fatalf("hints = %v, want the one that names biso doctor", modelErr.Hints)
	}
}

func TestAnUndeclaredExtensionKeyIsRejectedOnWrite(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	task := &model.Task{
		Title:  "With an extension",
		Status: "To Do",
		Ext:    map[string]string{"jira.key": "PROJ-1"},
	}
	err := tasks.Create(task)
	modelErr, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("Create with an undeclared key = %v, want a *model.Error", err)
	}
	if modelErr.ExitCode != 3 || modelErr.Code != "unknown_extension_key" {
		t.Fatalf("error = %d/%s, want 3/unknown_extension_key", modelErr.ExitCode, modelErr.Code)
	}

	// Nothing was written, not even the identifier: the check runs before
	// the transaction that allocates it.
	last, err := tasks.LastAllocated()
	if err != nil {
		t.Fatalf("LastAllocated: %v", err)
	}
	if last != 0 {
		t.Fatalf("LastAllocated() = %d, want 0: the rejected write allocated an id", last)
	}
}

func TestAMalformedExtensionKeyIsRejectedBeforeTheDeclaredList(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	task := &model.Task{
		Title:  "With a bad key",
		Status: "To Do",
		Ext:    map[string]string{"trello card": "1"},
	}
	err := tasks.Create(task)
	modelErr, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("Create with a malformed key = %v, want a *model.Error", err)
	}
	// A character outside the alphabet is a problem of form, exit code 2,
	// and not the 3 of a key the board does not declare
	// (docs/spec/valores-de-entrada.md#el-juego-de-caracteres-de-un-token).
	if modelErr.ExitCode != 2 || modelErr.Code != "malformed_extension_key" {
		t.Fatalf("error = %d/%s, want 2/malformed_extension_key", modelErr.ExitCode, modelErr.Code)
	}
}

func TestATaskWithNoOptionalFieldSurvivesTheRoundTrip(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	want := &model.Task{Title: "Bare", Status: "To Do"}
	if err := tasks.Create(want); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := tasks.Load(want.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Ordinal != nil {
		t.Fatalf("ordinal = %v, want nil", *got.Ordinal)
	}
	if !got.Due.IsZero() || !got.LeaseExpiresAt.IsZero() {
		t.Fatalf("an unset date came back set: %v / %v", got.Due, got.LeaseExpiresAt)
	}
	if got.Question != nil {
		t.Fatalf("question = %+v, want nil", got.Question)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("the round trip changed the task:\n got %+v\nwant %+v", got, want)
	}
}

func TestOrdinalZeroIsAValueAndNotAnAbsence(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	zero := 0
	task := &model.Task{Title: "First in the list", Status: "To Do", Ordinal: &zero}
	if err := tasks.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := tasks.Load(task.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if got.Ordinal == nil {
		t.Fatalf("ordinal came back absent, but 0 is a value a caller can give")
	}
	if *got.Ordinal != 0 {
		t.Fatalf("ordinal = %d, want 0", *got.Ordinal)
	}
}

func TestSaveKeepsTheStoredCreatedAtAndRefusesAMissingTask(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	task := &model.Task{
		Title:     "First",
		Status:    "To Do",
		CreatedAt: time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC),
		UpdatedAt: time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC),
	}
	if err := tasks.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// createdAt is not mutable: no flag of the program changes it, so a
	// save that carries another one writes the stored value anyway.
	task.CreatedAt = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	task.Notes = "Changed"
	if err := tasks.Save(task); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := tasks.Load(task.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	want := time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC)
	if !got.CreatedAt.Equal(want) {
		t.Fatalf("createdAt = %v, want %v: the field is not mutable", got.CreatedAt, want)
	}
	if got.Notes != "Changed" {
		t.Fatalf("notes = %q, want the saved one", got.Notes)
	}

	// And saving never brings back an identifier the board does not have.
	ghost := &model.Task{ID: "MYP-9", Title: "Ghost", Status: "To Do"}
	err = tasks.Save(ghost)
	modelErr, ok := err.(*model.Error)
	if !ok {
		t.Fatalf("Save of a missing task = %v, want a *model.Error", err)
	}
	if modelErr.ExitCode != 4 || modelErr.Code != "never_allocated" {
		t.Fatalf("error = %d/%s, want 4/never_allocated", modelErr.ExitCode, modelErr.Code)
	}
}

// TestCreateRefusesAnInvalidTaskWithoutSpendingAnIdentifier is why
// validation happens before the transaction and not inside it: an
// identifier is never reused (docs/spec/modelo-de-datos/identificadores.md),
// so one spent on a task that is then refused would be gone forever.
func TestCreateRefusesAnInvalidTaskWithoutSpendingAnIdentifier(t *testing.T) {
	cases := []struct {
		name string
		task *model.Task
	}{
		{
			name: "an empty title",
			task: &model.Task{Title: "   ", Status: "To Do"},
		},
		{
			name: "a newline in the title",
			task: &model.Task{Title: "first line\nsecond line", Status: "To Do"},
		},
		{
			name: "a label outside its alphabet",
			task: &model.Task{Title: "A task", Status: "To Do", Labels: []string{"urgent!"}},
		},
		{
			name: "an assignee outside its alphabet",
			task: &model.Task{Title: "A task", Status: "To Do", Assignees: []string{"sara smith"}},
		},
		{
			name: "a newline in the value of an extension field",
			task: &model.Task{
				Title:  "A task",
				Status: "To Do",
				Ext:    map[string]string{"trello.card": "5f2a8c1e\n5f2a8c1f"},
			},
		},
		{
			name: "a negative ordinal",
			task: &model.Task{Title: "A task", Status: "To Do", Ordinal: negative()},
		},
		{
			name: "a criterion with a key the program never assigns",
			task: &model.Task{
				Title:              "A task",
				Status:             "To Do",
				AcceptanceCriteria: []model.Criterion{{Key: 0, Text: "No key"}},
			},
		},
		{
			name: "two criteria sharing a key",
			task: &model.Task{
				Title:  "A task",
				Status: "To Do",
				AcceptanceCriteria: []model.Criterion{
					{Key: 1, Text: "The first one"},
					{Key: 1, Text: "The same key again"},
				},
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
			defer done()

			if err := tasks.Create(c.task); err == nil {
				t.Fatalf("Create stored %+v", c.task)
			}
			last, err := tasks.LastAllocated()
			if err != nil {
				t.Fatalf("LastAllocated: %v", err)
			}
			if last != 0 {
				t.Fatalf("LastAllocated() = %d after a refused Create, want 0", last)
			}
			all, _, err := tasks.All()
			if err != nil {
				t.Fatalf("All: %v", err)
			}
			if len(all) != 0 {
				t.Fatalf("the board has %d tasks after a refused Create, want 0", len(all))
			}
		})
	}
}

// negative is the one ordinal the field table of
// docs/spec/modelo-de-datos/index.md does not admit.
func negative() *int {
	value := -1
	return &value
}

// TestSaveRefusesAnInvalidTaskAndLeavesTheStoredOneAlone is the same rule
// on the other write path: a task already on the board keeps what it had.
func TestSaveRefusesAnInvalidTaskAndLeavesTheStoredOneAlone(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	task := &model.Task{Title: "A title of one line", Status: "To Do"}
	if err := tasks.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	task.Title = "first line\nsecond line"
	if err := tasks.Save(task); err == nil {
		t.Fatalf("Save stored a title with a line break")
	}

	back, err := tasks.Load(task.ID)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if back.Title != "A title of one line" {
		t.Fatalf("title = %q, want the one that was there before the refused Save", back.Title)
	}
}

// TestASetReadSkipsTheTaskItCannotDecodeAndNamesIt is the first case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar
// on a set read: the task is skipped, it is counted, and the rest of the
// result is valid. Never aborting is the substance of that guarantee; a
// listing that failed whole because of one bad row would leave the caller
// with nothing, and one that dropped it in silence would read as a fact
// about the board.
func TestASetReadSkipsTheTaskItCannotDecodeAndNamesIt(t *testing.T) {
	cases := []struct {
		name    string
		damage  string
		wantIDs []string
	}{
		{
			name:    "a date no version of this program wrote",
			damage:  "UPDATE task SET created_at = 'the day before yesterday' WHERE id = 'MYP-2'",
			wantIDs: []string{"MYP-2"},
		},
		{
			name:    "a comment whose date cannot be read",
			damage:  "UPDATE task_comment SET created_at = 'yesterday' WHERE task_id = 'MYP-2'",
			wantIDs: []string{"MYP-2"},
		},
		{
			name:    "two tasks at once, named in identifier order",
			damage:  "UPDATE task SET due = 'next tuesday' WHERE id IN ('MYP-3', 'MYP-1')",
			wantIDs: []string{"MYP-1", "MYP-3"},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			// Each case damages a board of its own, built the same way, so
			// that one case cannot explain another's result.
			path := filepath.Join(t.TempDir(), "board.sqlite")
			s, err := store.Open(testBoardID, path)
			if err != nil {
				t.Fatalf("open the board: %v", err)
			}
			defer s.Close()
			tasks := NewTasks(s, testPrefix, testExtensions)
			for i := 1; i <= 3; i++ {
				task := &model.Task{Title: fmt.Sprintf("Task %d", i), Status: "To Do"}
				task.AddComment("@avilches", time.Date(2026, 9, 6, 10, 0, 0, 0, time.UTC), "A comment.")
				if err := tasks.Create(task); err != nil {
					t.Fatalf("Create: %v", err)
				}
			}
			// Written straight into the table, because nothing biso does can
			// put this there: that is exactly the situation the guarantee
			// describes, data touched from outside.
			if _, err := s.Exec(c.damage); err != nil {
				t.Fatalf("damage the board: %v", err)
			}

			all, skipped, err := tasks.All()
			if err != nil {
				t.Fatalf("All aborted because of a task it could not decode: %v", err)
			}
			if len(all) != 3-len(c.wantIDs) {
				t.Fatalf("All() returned %d tasks, want %d", len(all), 3-len(c.wantIDs))
			}
			var gotIDs []string
			for _, s := range skipped {
				gotIDs = append(gotIDs, s.ID)
				if s.Reason == nil || s.Reason.ExitCode != 3 || s.Reason.Code != "undecodable_task" {
					t.Fatalf("skipped %s with reason %+v, want exit 3 and code undecodable_task", s.ID, s.Reason)
				}
			}
			if !reflect.DeepEqual(gotIDs, c.wantIDs) {
				t.Fatalf("skipped = %v, want %v", gotIDs, c.wantIDs)
			}
			for _, task := range all {
				for _, bad := range c.wantIDs {
					if task.ID == bad {
						t.Fatalf("%s came back in the listing although it could not be decoded", bad)
					}
				}
			}
		})
	}
}

// TestATargetedReadOfAnUndecodableTaskIsAnError is the other half of the
// same table: asking for that one task has nothing valid to answer, so it
// is exit code 3 with the reason.
func TestATargetedReadOfAnUndecodableTaskIsAnError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	s, err := store.Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open the board: %v", err)
	}
	defer s.Close()
	tasks := NewTasks(s, testPrefix, testExtensions)

	task := &model.Task{Title: "A task", Status: "To Do"}
	if err := tasks.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if _, err := s.Exec("UPDATE task SET updated_at = 'a while ago' WHERE id = ?", task.ID); err != nil {
		t.Fatalf("damage the board: %v", err)
	}

	got, err := tasks.Load(task.ID)
	if err == nil {
		t.Fatalf("Load returned %+v for a task that cannot be decoded", got)
	}
	var specErr *model.Error
	if !errors.As(err, &specErr) {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	if specErr.ExitCode != 3 || specErr.Code != "undecodable_task" {
		t.Fatalf("error = exit %d, code %q; want exit 3, code undecodable_task", specErr.ExitCode, specErr.Code)
	}
	if !strings.Contains(specErr.Message, task.ID) {
		t.Fatalf("message = %q, and it does not name the task", specErr.Message)
	}
}

// TestASetReadSkipsATaskWithAListFieldTheModelDoesNotKnow covers the other
// way a row stops being a task: a name in task_list_item that is not one
// of the five list fields. The schema's CHECK keeps biso itself from
// writing one, so it can only arrive from outside, like the dates above.
func TestASetReadSkipsATaskWithAListFieldTheModelDoesNotKnow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	s, err := store.Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open the board: %v", err)
	}
	defer s.Close()
	tasks := NewTasks(s, testPrefix, testExtensions)

	for i := 1; i <= 2; i++ {
		task := &model.Task{
			Title:  fmt.Sprintf("Task %d", i),
			Status: "To Do",
			Labels: []string{"parser"},
		}
		if err := tasks.Create(task); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}
	// The schema's CHECK refuses this name, which is the point of having
	// it, so the test turns the constraints off for the one statement that
	// puts the damage there.
	if _, err := s.Exec("PRAGMA ignore_check_constraints = 1"); err != nil {
		t.Fatalf("turn the check constraints off: %v", err)
	}
	if _, err := s.Exec(
		"UPDATE task_list_item SET field = 'tags' WHERE task_id = 'MYP-2'",
	); err != nil {
		t.Fatalf("damage the board: %v", err)
	}
	if _, err := s.Exec("PRAGMA ignore_check_constraints = 0"); err != nil {
		t.Fatalf("turn the check constraints back on: %v", err)
	}

	all, skipped, err := tasks.All()
	if err != nil {
		t.Fatalf("All aborted because of a list field it does not know: %v", err)
	}
	if len(all) != 1 || all[0].ID != "MYP-1" {
		t.Fatalf("All() = %v, want only MYP-1", all)
	}
	if len(skipped) != 1 || skipped[0].ID != "MYP-2" {
		t.Fatalf("skipped = %+v, want MYP-2", skipped)
	}
}

// TestSaveAllIsOneTransaction is guarantee 2 of
// docs/spec/garantias.md#concurrencia-atomicidad-y-garantías-observables
// checked against SQLite and not against the code that calls it: a batch
// where the last task cannot be written leaves the first ones exactly as
// they were, and not "as they were plus the first change".
func TestSaveAllIsOneTransaction(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	first := &model.Task{Title: "First", Status: "To Do"}
	second := &model.Task{Title: "Second", Status: "To Do"}
	for _, task := range []*model.Task{first, second} {
		if err := tasks.Create(task); err != nil {
			t.Fatal(err)
		}
	}

	// The third one was never allocated, so writing it fails inside the
	// transaction, after the first two rows have already been rewritten.
	first.Title = "First, renamed"
	second.Title = "Second, renamed"
	missing := &model.Task{ID: testPrefix + "-90", Title: "Never existed", Status: "To Do"}

	err := tasks.SaveAll([]*model.Task{first, second, missing})
	if err == nil {
		t.Fatal("SaveAll succeeded over a task the board never allocated")
	}
	e, ok := err.(*model.Error)
	if !ok || e.ExitCode != 4 || e.Code != "never_allocated" {
		t.Fatalf("error = %v, want exit code 4 with never_allocated", err)
	}

	for _, c := range []struct{ id, title string }{
		{testPrefix + "-1", "First"},
		{testPrefix + "-2", "Second"},
	} {
		stored, err := tasks.Load(c.id)
		if err != nil {
			t.Fatal(err)
		}
		if stored.Title != c.title {
			t.Errorf("%s was left titled %q, and the batch wrote nothing", c.id, stored.Title)
		}
	}
}

// TestSaveAllValidatesEverythingBeforeItOpensTheTransaction is the other
// half of the same guarantee: a batch that is going to fail on its second
// task does not write the first one and then undo it, it never starts.
func TestSaveAllValidatesEverythingBeforeItOpensTheTransaction(t *testing.T) {
	tasks, done := openTasks(t, filepath.Join(t.TempDir(), "board.sqlite"))
	defer done()

	first := &model.Task{Title: "First", Status: "To Do"}
	second := &model.Task{Title: "Second", Status: "To Do"}
	for _, task := range []*model.Task{first, second} {
		if err := tasks.Create(task); err != nil {
			t.Fatal(err)
		}
	}
	first.Title = "First, renamed"
	second.Title = "" // a title is required, and the model says so

	if err := tasks.SaveAll([]*model.Task{first, second}); err == nil {
		t.Fatal("SaveAll accepted a task with no title")
	}

	stored, err := tasks.Load(testPrefix + "-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.Title != "First" {
		t.Errorf("the first task was written although the batch failed: %q", stored.Title)
	}
}

// TestAnExpiredLeaseIsClaimedByOnlyOneOfTwoSimultaneousClaims is the
// conditional half of `biso start`
// (docs/spec/cmd/verbos-del-ciclo.md#biso-start). Two processes read the
// same expired lease, one of them writes first, and the second one is told
// it lost instead of overwriting the winner in silence.
//
// The two go through two independent stores over the same file, which is
// what a second process really is.
func TestAnExpiredLeaseIsClaimedByOnlyOneOfTwoSimultaneousClaims(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	now := time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC)

	first, closeFirst := openTasks(t, path)
	defer closeFirst()
	task := &model.Task{
		Title: "Normalize CRLF", Status: "In Progress",
		Assignees:      []string{"@sara"},
		LeaseHolder:    "@sara",
		LeaseExpiresAt: now.Add(-time.Hour),
	}
	if err := first.Create(task); err != nil {
		t.Fatalf("Create: %v", err)
	}

	// Both read the task while the lease is expired, and each one prepares
	// its own claim of it.
	mine, err := first.Load(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	second, closeSecond := openTasks(t, path)
	defer closeSecond()
	theirs, err := second.Load(task.ID)
	if err != nil {
		t.Fatal(err)
	}

	theirs.Assignees = []string{"@juan"}
	theirs.LeaseHolder, theirs.LeaseExpiresAt = "@juan", now.Add(time.Hour)
	if err := second.SaveAll([]*model.Task{theirs},
		LeaseClaim{TaskID: task.ID, Holder: "@juan", Now: now}); err != nil {
		t.Fatalf("the first claim did not win: %v", err)
	}

	mine.Assignees = []string{"@claude"}
	mine.LeaseHolder, mine.LeaseExpiresAt = "@claude", now.Add(time.Hour)
	err = first.SaveAll([]*model.Task{mine},
		LeaseClaim{TaskID: task.ID, Holder: "@claude", Now: now})

	var e *model.Error
	if !errors.As(err, &e) || e.ExitCode != 8 || e.Code != "lease_lost" {
		t.Fatalf("error = %v, want exit code 8 with lease_lost", err)
	}
	// And the loser wrote nothing at all, not even the fields that had
	// nothing to do with the lease.
	after, err := first.Load(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.LeaseHolder != "@juan" {
		t.Errorf("leaseHolder = %q, want the winner's", after.LeaseHolder)
	}
	if strings.Join(after.Assignees, ",") != "@juan" {
		t.Errorf("assignees = %v, and the loser's write was rolled back whole", after.Assignees)
	}
}

// A claim over a lease that is still expired at write time goes through:
// nobody else got there first.
func TestAClaimOverALeaseThatIsStillExpiredGoesThrough(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")
	now := time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC)

	tasks, done := openTasks(t, path)
	defer done()
	task := &model.Task{
		Title: "Normalize CRLF", Status: "In Progress",
		Assignees:      []string{"@sara"},
		LeaseHolder:    "@sara",
		LeaseExpiresAt: now.Add(-time.Hour),
	}
	if err := tasks.Create(task); err != nil {
		t.Fatal(err)
	}

	task.LeaseHolder, task.LeaseExpiresAt = "@claude", now.Add(time.Hour)
	if err := tasks.SaveAll([]*model.Task{task},
		LeaseClaim{TaskID: task.ID, Holder: "@claude", Now: now}); err != nil {
		t.Fatalf("SaveAll: %v", err)
	}

	after, err := tasks.Load(task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if after.LeaseHolder != "@claude" {
		t.Errorf("leaseHolder = %q, want the caller", after.LeaseHolder)
	}
}
