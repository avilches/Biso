package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"
)

// budgetMillis is the startup budget of
// docs/spec/presupuestos.md#el-presupuesto-de-arranque: 25ms of wall clock
// on the reference machine, which is the one that runs the project's
// continuous integration suite.
const budgetMillis = 25

// budgetRows is the 300-task board that budget talks about. The
// composition of those tasks is deliberately not fixed by the
// specification, so the synthetic rows of this step are a fair stand-in
// for the row count.
const budgetRows = 300

// The figures this test measures on the development machine, so that a
// change that makes it slower is recognizable: around 0.5 to 0.7ms with the
// process already warm, and around 4 to 5ms for the first run of the suite
// under -race, which is how the plan asks for it to be run. The 25ms limit
// below is the one the specification fixes, and the race detector's figure
// is the honest one to compare it against.
//
// TestReadBudgetOnThreeHundredRows is the first real (non-synthetic) step
// of the startup budget test from
// docs/spec/presupuestos.md#el-presupuesto-de-arranque. It does not yet
// exercise the compiled binary or the ls/prime commands, which arrive with
// TASK-13 and TASK-15 (steps 5 and 7 of TASK-55): it measures the path the
// store already controls, opening the file and reading its 300 rows, which
// is a floor of the full budget and not the final measurement. The step
// that adds ls/prime replaces it with the measurement on the real binary.
func TestReadBudgetOnThreeHundredRows(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	seed, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open for seeding: %v", err)
	}
	if err := seed.WithTx(func(tx *sql.Tx) error {
		for i := 1; i <= budgetRows; i++ {
			if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (?, ?)", i, "payload"); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("seed the rows: %v", err)
	}
	seed.Close()

	start := time.Now()

	s, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("open for reading: %v", err)
	}
	defer s.Close()

	rows, err := s.Query("SELECT id, payload FROM bench_row")
	if err != nil {
		t.Fatalf("query: %v", err)
	}
	count := 0
	for rows.Next() {
		var id int
		var payload string
		if err := rows.Scan(&id, &payload); err != nil {
			rows.Close()
			t.Fatalf("scan row: %v", err)
		}
		count++
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		t.Fatalf("iterate rows: %v", err)
	}
	rows.Close()

	elapsed := time.Since(start)
	if count != budgetRows {
		t.Fatalf("count = %d, want %d", count, budgetRows)
	}
	t.Logf("opening and reading %d rows took %s", budgetRows, elapsed)
	if elapsed > budgetMillis*time.Millisecond {
		t.Fatalf(
			"opening and reading %d rows took %s, want less than %dms (docs/spec/presupuestos.md#el-presupuesto-de-arranque)",
			budgetRows, elapsed, budgetMillis,
		)
	}
}
