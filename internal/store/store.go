// Package store is the only biso layer that knows what SQLite is. It opens
// or creates a board's file and exposes the write transaction that the six
// guarantees of docs/spec/garantias.md#concurrencia-atomicidad-y-garantias-observables
// depend on. It does not know what a Task is, only rows and columns.
package store

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	_ "modernc.org/sqlite"

	"biso/internal/model"
)

// busyTimeoutMillis is the limit from guarantee 5 of
// docs/spec/garantias.md#concurrencia-atomicidad-y-garantias-observables: if
// the program cannot get the write lock it needs, it waits at most this
// long before failing with exit code 8.
const busyTimeoutMillis = 5000

// Store wraps a SQLite database opened over a board's file.
//
// id is the board's identifier, the eight characters of
// docs/spec/resolucion-del-tablero.md. The store does nothing with it
// except name the board in the message of the unreadable-database error,
// which docs/spec/garantias.md fixes word for word; whoever opens the store
// already read that id from the board's pointer.
//
// inTx says whether one of this store's write transactions is live. Open
// leaves a single connection (see Open), so a query through the handle
// while that transaction holds it would wait for a connection that only the
// transaction can give back: forever, with no error and no timeout. Every
// access through the handle checks this first and fails with
// ErrTxInProgress instead.
//
// path is the file the store was opened over, which is what the error of an
// environment that refuses to write names: that failure can happen before
// anything has read the id, so the path is the only thing it can name.
type Store struct {
	db   *sql.DB
	id   string
	path string
	inTx atomic.Bool
}

// Open opens the file at path for the board with the given id, creating it
// if it does not exist, in WAL mode, with foreign keys on and with the write
// lock already requested at the BEGIN of every transaction (see WithTx).
func Open(id, path string) (*Store, error) {
	return openAt(id, path, busyTimeoutMillis, migrations)
}

// openAt is Open with the two things the tests need to vary: how long to
// wait for the write lock, and which migrations to apply. Passing a nil
// list creates or opens the file without touching its schema.
func openAt(id, path string, busyTimeoutMs int, scripts []string) (*Store, error) {
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
	// and never the Store's database handle, which that transaction holds;
	// the inTx guard of Store makes that a clear error instead of a wait
	// with no end.
	db.SetMaxOpenConns(1)

	s := &Store{db: db, id: id, path: path}

	version, err := s.initialVersion(busyTimeoutMs)
	if err != nil {
		db.Close()
		return nil, err
	}
	if err := s.migrateFrom(version, scripts); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// retryFloor and retryCeiling bound the wait between two attempts of
// initialVersion. It starts at the floor and doubles up to the ceiling, so
// a collision that clears at once costs about a millisecond and a long one
// does not spin.
const (
	retryFloor   = 1 * time.Millisecond
	retryCeiling = 50 * time.Millisecond
)

// initialVersion reads the schema version of the file, which is also the
// statement that first connects: sql.Open connects to nothing, so this is
// what proves the file is a database at all, and it is the version the
// migration starts from.
//
// It is the one place that waits for the write lock by hand instead of
// letting SQLite's busy timeout do it, because there is one lock SQLite
// refuses without ever consulting the busy handler: turning a brand new
// file into WAL mode (PRAGMA journal_mode = WAL, which the driver runs when
// it opens the connection) needs the file to itself, and a second
// connection creating the same file at the same instant gets SQLITE_BUSY
// back immediately. Measured: four connections creating the same file at
// once fail on several attempts out of twenty, and zero out of twenty over
// a file that already exists.
//
// Guarantee 5 of docs/spec/garantias.md promises a wait of the configured
// time before exit code 8, and it does not exempt the command that creates
// the board, so this retries until that time is up and only then answers
// the busy error. Anything that is not the write lock is not retried, and
// it splits in two: an environment that refuses the file answers the
// cannot-write error of exit code 8
// (docs/spec/garantias.md#qué-pasa-cuando-el-almacén-no-se-puede-escribir),
// and anything else means the process could not read the database's header,
// which is the second case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar.
// Opening WAL mode needs to write, so a read-only directory fails here and
// not later, which is why this one place cannot settle for "it could not be
// read".
func (s *Store) initialVersion(busyTimeoutMs int) (int, error) {
	deadline := time.Now().Add(time.Duration(busyTimeoutMs) * time.Millisecond)
	wait := retryFloor
	for {
		var version int
		err := s.scanOne(&version, "PRAGMA user_version")
		if err == nil {
			return version, nil
		}
		// scanOne already classified what the driver said. The busy
		// error is the one this loop waits for; any other one it
		// recognized is final, and anything it did not recognize means
		// the header could not be read.
		classified, ok := err.(*model.Error)
		if !ok {
			return 0, s.newUnreadableError()
		}
		if classified.Code != "busy" {
			return 0, classified
		}
		if !time.Now().Add(wait).Before(deadline) {
			return 0, newBusyError()
		}
		time.Sleep(wait)
		if wait < retryCeiling {
			wait *= 2
		}
	}
}

// uriPath turns a filesystem path into the path of a file: URI, which is how
// the driver is called here (it passes SQLITE_OPEN_URI).
//
// It escapes the three characters SQLite reads as URI syntax: without this,
// a board kept under a directory whose name contains one of them would open
// the wrong file or fail. And it writes Windows paths the way SQLite's URI
// parser expects them, because the driver was picked precisely for
// cross-compiling there (docs/decisiones/lenguaje-y-rendimiento.md):
// separators become forward slashes, a drive letter takes a leading slash
// that SQLite's Windows layer discards again, and a UNC path takes the five
// leading slashes that SQLite documents as the way to write one, so that
// the server name is never read as the URI's authority.
func uriPath(path string) string {
	unc := strings.HasPrefix(path, `\\`)

	escaped := strings.NewReplacer("%", "%25", "?", "%3f", "#", "%23").Replace(path)
	escaped = strings.ReplaceAll(escaped, `\`, "/")

	switch {
	case unc:
		return "///" + escaped
	case hasDriveLetter(escaped):
		return "/" + escaped
	default:
		return escaped
	}
}

// hasDriveLetter answers whether path starts with a Windows drive letter,
// as in C:\boards or C:/boards.
func hasDriveLetter(path string) bool {
	if len(path) < 2 || path[1] != ':' {
		return false
	}
	c := path[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// Query runs a read query through the store's handle, outside any
// transaction. Inside a WithTx callback, use the transaction it gives you:
// this call refuses with ErrTxInProgress rather than wait forever.
func (s *Store) Query(query string, args ...any) (*sql.Rows, error) {
	if s.inTx.Load() {
		return nil, fmt.Errorf("query: %w", ErrTxInProgress)
	}
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, s.Classify(err)
	}
	return rows, nil
}

// Exec runs a statement through the store's handle, outside any
// transaction, with the same guard as Query. Anything that has to be atomic
// with something else goes through WithTx instead.
func (s *Store) Exec(query string, args ...any) (sql.Result, error) {
	if s.inTx.Load() {
		return nil, fmt.Errorf("exec: %w", ErrTxInProgress)
	}
	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, s.Classify(err)
	}
	return result, nil
}

// scanOne reads the single value of a single row into dest, with the same
// guard as Query. It is how this package reads its pragmas.
func (s *Store) scanOne(dest any, query string, args ...any) error {
	if s.inTx.Load() {
		return fmt.Errorf("query: %w", ErrTxInProgress)
	}
	return s.Classify(s.db.QueryRow(query, args...).Scan(dest))
}

// CheckIntegrity runs SQLite's own integrity check over the whole file. A
// database that fails it is the second case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar,
// the same one as a file that does not open, with the same error and the
// same exit code 21.
//
// Opening a board does not run it: the check reads every page, which the
// startup budget of docs/spec/presupuestos.md#el-presupuesto-de-arranque
// does not pay for on every command. It is `biso doctor` that runs it
// (docs/spec/cmd/doctor.md), and a board damaged in a way that the check
// would catch fails on its own as soon as a command reads the damaged part.
func (s *Store) CheckIntegrity() error {
	rows, err := s.Query("PRAGMA integrity_check")
	if err != nil {
		return s.classify(err, "run the integrity check")
	}
	defer rows.Close()

	ok := false
	lines := 0
	for rows.Next() {
		var line string
		if err := rows.Scan(&line); err != nil {
			return s.classify(err, "read the integrity check")
		}
		lines++
		ok = lines == 1 && line == "ok"
	}
	if err := rows.Err(); err != nil {
		return s.classify(err, "read the integrity check")
	}
	if !ok {
		return s.newUnreadableError()
	}
	return nil
}

// Close closes the connection to the store.
func (s *Store) Close() error {
	return s.db.Close()
}
