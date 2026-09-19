// Package store is the only biso layer that knows what SQLite is. It opens
// or creates a board's file and exposes the write transaction that the six
// guarantees of docs/spec/garantias.md#concurrencia-atomicidad-y-garantias-observables
// depend on. It does not know what a Task is, only rows and columns.
package store

import (
	"database/sql"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

// busyTimeoutMillis is the limit from guarantee 5 of
// docs/spec/garantias.md#concurrencia-atomicidad-y-garantias-observables: if
// the program cannot get the write lock it needs, it waits at most this
// long before failing with exit code 8.
const busyTimeoutMillis = 5000

// Store wraps a SQLite database opened over a board's file.
type Store struct {
	db *sql.DB
}

// Open opens the file at path, creating it if it does not exist, in WAL
// mode, with foreign keys on and with the write lock already requested at
// the BEGIN of every transaction (see WithTx).
func Open(path string) (*Store, error) {
	return open(path, busyTimeoutMillis)
}

func open(path string, busyTimeoutMs int) (*Store, error) {
	dsn := fmt.Sprintf(
		"file:%s?_journal_mode=WAL&_busy_timeout=%d&_foreign_keys=1&_txlock=immediate",
		uriPath(path), busyTimeoutMs,
	)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	// One connection per process. SQLite in WAL mode allows a single writer
	// at a time, so a second connection of this same process could only
	// wait on the first one and time out against itself. It also means the
	// code inside a WithTx callback must use the transaction it receives
	// and never the Store's database handle, which that transaction holds.
	db.SetMaxOpenConns(1)

	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// uriPath escapes the three characters SQLite reads as URI syntax when the
// filename is given as a file: URI, which is how the driver is called here
// (it passes SQLITE_OPEN_URI). Without this, a board kept under a directory
// whose name contains one of them would open the wrong file or fail.
func uriPath(path string) string {
	return strings.NewReplacer("%", "%25", "?", "%3f", "#", "%23").Replace(path)
}

// Close closes the connection to the store.
func (s *Store) Close() error {
	return s.db.Close()
}
