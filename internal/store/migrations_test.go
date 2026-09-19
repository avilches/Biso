package store

import (
	"path/filepath"
	"testing"
)

func TestOpenAppliesMigrationsAndTracksVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != len(migrations) {
		t.Fatalf("user_version = %d, want %d", version, len(migrations))
	}

	if _, err := s.db.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'x')"); err != nil {
		t.Fatalf("insert into bench_row: %v", err)
	}
}

func TestReopenDoesNotReapplyMigrations(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	if _, err := s1.db.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'x')"); err != nil {
		t.Fatalf("insert: %v", err)
	}
	s1.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()

	var count int
	if err := s2.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("count = %d, want 1 (reopening must not recreate or empty the table)", count)
	}
}

func TestMigrationFailureLeavesTheVersionUntouched(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	before := len(migrations)
	migrations = append(migrations, `CREATE TABLE this is not valid SQL;`)
	defer func() { migrations = migrations[:before] }()

	if err := s.migrate(); err == nil {
		t.Fatalf("migrate on a broken script returned no error")
	}

	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		t.Fatalf("read user_version: %v", err)
	}
	if version != before {
		t.Fatalf("user_version = %d, want %d (a failed migration must not advance it)", version, before)
	}
}
