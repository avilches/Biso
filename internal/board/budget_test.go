package board

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"biso/internal/model"
	"biso/internal/store"
)

// budgetMillis is the startup budget of
// docs/spec/presupuestos.md#el-presupuesto-de-arranque: 25ms of wall clock
// on the reference machine, which is the one that runs the project's
// continuous integration suite.
const budgetMillis = 25

// raceBudgetMillis is the limit that applies when the suite runs under the
// race detector, which is how the implementation plan asks for it to be
// run. It is not a second budget of the specification: the 25ms above are
// wall clock of the real program, and the detector multiplies every read
// of this pure-Go SQLite driver by an order of magnitude, so comparing its
// figure against the specification's number would measure the detector and
// not the program. This one is a regression guard, set at roughly twice
// what the same board measures today, so that a change that makes reading
// several times slower still fails in both modes.
const raceBudgetMillis = 150

// budgetTasks is the 300-task board that budget talks about.
const budgetTasks = 300

// TestReadBudgetOnThreeHundredTasks measures opening a board of 300 real
// tasks and reading every one of them with all their lists, criteria and
// comments. It moved here from internal/store when the synthetic table it
// used to read gave way to the real schema, and it measures strictly more
// than it did: five queries and the whole assembly of a task, instead of
// one scan of one table.
//
// It is still not the measurement of the budget, which is over `biso ls`
// and `biso prime` on the compiled binary and belongs to TASK-15 (step 7
// of TASK-55). It is the floor of it: whatever those two commands end up
// costing, they cannot cost less than this.
//
// The figures on the development machine, so that a change that makes it
// slower is recognizable: around 2.5ms without the race detector, against
// the 25ms of the specification, and around 75ms with it, which is why the
// two builds compare against different limits (see raceBudgetMillis).
func TestReadBudgetOnThreeHundredTasks(t *testing.T) {
	path := filepath.Join(t.TempDir(), "board.sqlite")

	seedTasks, done := openTasks(t, path)
	for i := 1; i <= budgetTasks; i++ {
		task := &model.Task{
			Title:        fmt.Sprintf("Task number %d of the budget board", i),
			Status:       "To Do",
			Type:         "feature",
			Priority:     "medium",
			Author:       "@avilches",
			Assignees:    []string{"@claude"},
			Labels:       []string{"parser", "crlf"},
			Dependencies: []string{"MYP-1"},
			Description:  "A description of the length a real task carries, more or less.",
			Ext:          map[string]string{"trello.card": "5f2a8c1e"},
		}
		task.AddCriterion("The first criterion")
		task.AddCriterion("The second criterion")
		task.AddComment("@avilches", time.Now().UTC(), "A comment of the usual length.")
		if err := seedTasks.Create(task); err != nil {
			t.Fatalf("seed task %d: %v", i, err)
		}
	}
	done()

	start := time.Now()

	s, err := store.Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open for reading: %v", err)
	}
	defer s.Close()

	all, err := NewTasks(s, testPrefix, testExtensions).All()
	if err != nil {
		t.Fatalf("All: %v", err)
	}

	elapsed := time.Since(start)

	if len(all) != budgetTasks {
		t.Fatalf("All() returned %d tasks, want %d", len(all), budgetTasks)
	}
	// Reading them is not enough: a read that dropped the children would be
	// faster and wrong, so the check is on an assembled task.
	if all[budgetTasks-1].AcTotal() != 2 || all[budgetTasks-1].CommentCount() != 1 {
		t.Fatalf("the last task came back incomplete: %+v", all[budgetTasks-1])
	}

	limit := budgetMillis
	if raceDetector {
		limit = raceBudgetMillis
	}
	t.Logf("opening and reading %d tasks took %s (limit %dms)", budgetTasks, elapsed, limit)
	if elapsed > time.Duration(limit)*time.Millisecond {
		t.Fatalf(
			"opening and reading %d tasks took %s, want less than %dms (docs/spec/presupuestos.md#el-presupuesto-de-arranque)",
			budgetTasks, elapsed, limit,
		)
	}
}
