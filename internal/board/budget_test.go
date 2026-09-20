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
// continuous integration suite. It is the only limit this test knows.
const budgetMillis = 25

// budgetTasks is the 300-task board that budget talks about.
const budgetTasks = 300

// TestReadBudgetOnThreeHundredTasks measures opening a board of 300 real
// tasks and reading every one of them with all their lists, criteria and
// comments. It moved here from internal/store when the synthetic table it
// used to read gave way to the real schema, and it measures strictly more
// than it did: five queries and the whole assembly of a task, instead of
// one scan of one table.
//
// It is not the measurement of the budget: that one runs `biso ls` and
// `biso prime` as processes and lives in cmd/biso/budget_test.go. This is
// its floor, and it is worth keeping separate, because a regression here
// is a regression of the store and says so, while the same regression seen
// from the process could have come from anywhere along the way.
//
// The figures on the development machine, so that a change that makes it
// slower is recognizable: around 2.7ms without the race detector and
// around 80ms with it. The detector multiplies every read of this pure-Go
// SQLite driver by more than an order of magnitude, so a run under it
// cannot say anything about the 25ms of the specification: it would
// measure the detector. It does not get a looser limit either, because a
// limit nobody checks is not a limit and the implementation plan runs the
// suite under the detector, which would mean the number of the
// specification was never checked at all. The build under the detector
// skips this test and says so, and the build without it asserts the
// figure.
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

	all, _, err := NewTasks(s, testPrefix, testExtensions).All()
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

	t.Logf("opening and reading %d tasks took %s (budget %dms)", budgetTasks, elapsed, budgetMillis)
	if raceDetector {
		t.Skipf(
			"took %s under the race detector, which measures the detector and not the program; "+
				"the %dms of docs/spec/presupuestos.md#el-presupuesto-de-arranque are checked by the build without it",
			elapsed, budgetMillis,
		)
	}
	if elapsed > budgetMillis*time.Millisecond {
		t.Fatalf(
			"opening and reading %d tasks took %s, want less than %s (docs/spec/presupuestos.md#el-presupuesto-de-arranque)",
			budgetTasks, elapsed, budgetMillis*time.Millisecond,
		)
	}
}
