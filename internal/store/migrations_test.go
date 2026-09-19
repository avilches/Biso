package store

import (
	"path/filepath"
	"sync"
	"testing"
)

func TestOpenAppliesMigrationsAndTracksVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	var version int
	if err := s.scanOne(&version, "PRAGMA user_version"); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != len(migrations) {
		t.Fatalf("user_version = %d, want %d", version, len(migrations))
	}

	// The real schema is there: board_counter is the table the migration
	// creates and seeds, and it is what allocates task ids.
	var last int
	if err := s.scanOne(&last, "SELECT last_task_num FROM board_counter"); err != nil {
		t.Fatalf("read the task counter: %v", err)
	}
	if last != 0 {
		t.Fatalf("last_task_num on a fresh board = %d, want 0", last)
	}
}

// TestSchemaHasEveryTableOfTheModel checks that the whole first migration
// ran and not only its first statement: the script creates six tables in
// one Exec, so this is also what proves the driver applies a multi
// statement script whole.
func TestSchemaHasEveryTableOfTheModel(t *testing.T) {
	s, err := Open(testBoardID, filepath.Join(t.TempDir(), "board.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	for _, table := range []string{
		"board_counter", "task", "task_list_item",
		"task_ext", "task_criterion", "task_comment",
	} {
		var n int
		if err := s.scanOne(&n, "SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = ?", table); err != nil {
			t.Fatalf("look for %s: %v", table, err)
		}
		if n != 1 {
			t.Fatalf("table %s is missing from the schema", table)
		}
	}
}

func TestReopenDoesNotReapplyMigrations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s1, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if _, err := s1.Exec("UPDATE board_counter SET last_task_num = 7"); err != nil {
		t.Fatalf("write the task counter: %v", err)
	}
	s1.Close()

	s2, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()

	var last int
	if err := s2.scanOne(&last, "SELECT last_task_num FROM board_counter"); err != nil {
		t.Fatalf("read the task counter: %v", err)
	}
	if last != 7 {
		t.Fatalf("last_task_num = %d, want 7 (reopening must not recreate or reseed the table)", last)
	}
}

// TestMigrationFailureLeavesTheVersionUntouched injects its broken script
// instead of appending it to the package's migrations list: a test that
// mutates that global would break any other test in the package the day one
// of them runs in parallel.
func TestMigrationFailureLeavesTheVersionUntouched(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	broken := append(append([]string{}, migrations...), `CREATE TABLE this is not valid SQL;`)
	if err := s.migrateTo(broken); err == nil {
		t.Fatalf("migrateTo on a broken script returned no error")
	}

	var version int
	if err := s.scanOne(&version, "PRAGMA user_version"); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != len(migrations) {
		t.Fatalf("user_version = %d, want %d (a failed migration must not advance it)", version, len(migrations))
	}
}

// TestMigrateIsAtomicAcrossTwoStores opens the same stale file from two
// stores at once. Reading PRAGMA user_version outside a transaction would
// let both see the old version and both try to create the same tables, so
// one of them would fail with "table already exists"; inside a single
// immediate transaction, the loser sees the version the winner left.
func TestMigrateIsAtomicAcrossTwoStores(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	scripts := []string{
		`CREATE TABLE first (id INTEGER PRIMARY KEY);`,
		`CREATE TABLE second (id INTEGER PRIMARY KEY);`,
	}

	// A file at version 0, with no schema at all: both stores below have
	// the whole list to apply.
	empty, err := openAt(testBoardID, path, busyTimeoutMillis, nil)
	if err != nil {
		t.Fatalf("create the empty file: %v", err)
	}
	empty.Close()

	one, err := openAt(testBoardID, path, busyTimeoutMillis, nil)
	if err != nil {
		t.Fatalf("open the first store: %v", err)
	}
	defer one.Close()
	two, err := openAt(testBoardID, path, busyTimeoutMillis, nil)
	if err != nil {
		t.Fatalf("open the second store: %v", err)
	}
	defer two.Close()

	errs := make([]error, 2)
	var wg sync.WaitGroup
	for i, s := range []*Store{one, two} {
		wg.Add(1)
		go func(i int, s *Store) {
			defer wg.Done()
			errs[i] = s.migrateTo(scripts)
		}(i, s)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("store %d failed to migrate concurrently: %v", i+1, err)
		}
	}

	var version int
	if err := one.scanOne(&version, "PRAGMA user_version"); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != len(scripts) {
		t.Fatalf("user_version = %d, want %d", version, len(scripts))
	}
	for _, table := range []string{"first", "second"} {
		var name string
		if err := one.scanOne(&name, "SELECT name FROM sqlite_master WHERE name = ?", table); err != nil {
			t.Fatalf("look for table %s: %v", table, err)
		}
	}
}
