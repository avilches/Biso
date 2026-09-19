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

	if _, err := s.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'x')"); err != nil {
		t.Fatalf("insert into bench_row: %v", err)
	}
}

func TestReopenDoesNotReapplyMigrations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s1, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if _, err := s1.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'x')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	s1.Close()

	s2, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()

	var count int
	if err := s2.scanOne(&count, "SELECT COUNT(*) FROM bench_row"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 (reopening must not recreate or empty the table)", count)
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
