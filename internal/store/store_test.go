package store

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOpenCreatesFileInWALModeWithForeignKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	var mode string
	if err := s.db.QueryRow("PRAGMA journal_mode").Scan(&mode); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}

	var fk int
	if err := s.db.QueryRow("PRAGMA foreign_keys").Scan(&fk); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}

	var timeout int
	if err := s.db.QueryRow("PRAGMA busy_timeout").Scan(&timeout); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if timeout != busyTimeoutMillis {
		t.Fatalf("busy_timeout = %d, want %d (guarantee 5 of docs/spec/garantias.md)", timeout, busyTimeoutMillis)
	}
}

func TestOpenOnExistingFileSucceeds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s1, err := Open(path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	s1.Close()

	s2, err := Open(path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()
}

func TestOpenOnAPathWithURICharactersUsesThatFile(t *testing.T) {
	dir := t.TempDir()
	odd := filepath.Join(dir, "board #1")
	if err := os.Mkdir(odd, 0o755); err != nil {
		t.Fatalf("create the directory: %v", err)
	}
	path := filepath.Join(odd, "board.sqlite")

	s, err := Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := s.db.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("write to the database: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the file was not created where it was asked for: %v", err)
	}
}
