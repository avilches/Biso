package store

import (
	"errors"
	"fmt"

	"modernc.org/sqlite"

	"biso/internal/model"
)

// SQLite's primary result codes, stable forever in its C API:
// https://www.sqlite.org/rescode.html
const (
	sqliteBusy    = 5  // SQLITE_BUSY, the write lock is held elsewhere
	sqliteLocked  = 6  // SQLITE_LOCKED, the same thing within one connection
	sqliteCorrupt = 11 // SQLITE_CORRUPT, the file is a damaged database
	sqliteNotADB  = 26 // SQLITE_NOTADB, the file is not a database at all
)

// primaryCodeMask keeps the low byte of a result code. SQLite's extended
// codes carry the primary one in their low eight bits, so SQLITE_BUSY,
// SQLITE_BUSY_RECOVERY and SQLITE_BUSY_SNAPSHOT all answer 5 through it:
// https://www.sqlite.org/rescode.html#primary_result_codes_versus_extended_result_codes
const primaryCodeMask = 0xff

// ErrTxInProgress is returned instead of hanging when something tries to
// use the store's handle while one of its write transactions is live. See
// the guard in WithTx for why waiting would never end.
var ErrTxInProgress = errors.New("a write transaction is already open on this store: inside WithTx use the *sql.Tx it gives you, never the store handle")

// classify turns a driver error into the error the specification fixes for
// that case, when there is one. The store is the deepest layer that can
// tell these two apart, and docs/spec/garantias.md gives both of them a
// literal text and an exit code, so they are born here and travel up
// unchanged (see the comment on model.Error).
func (s *Store) classify(err error, action string) error {
	switch primaryResultCode(err) {
	case sqliteBusy, sqliteLocked:
		return newBusyError()
	case sqliteCorrupt, sqliteNotADB:
		return s.newUnreadableError()
	}
	return fmt.Errorf("%s: %w", action, err)
}

// Opening the file has no classifier of its own: initialVersion does that
// classification inline, because there it is inseparable from deciding
// whether to wait and try again.

// primaryResultCode answers SQLite's primary result code carried by err, or
// zero when err does not come from the driver.
func primaryResultCode(err error) int {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return 0
	}
	return sqliteErr.Code() & primaryCodeMask
}

// newBusyError builds the error of guarantee 5 of docs/spec/garantias.md:
// the write lock could not be taken within the wait, nothing was written,
// and the exit code is 8 (see docs/spec/codigos-de-salida.md).
func newBusyError() *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "busy",
		Message:  "the board is busy, another process is writing to it",
		Hints:    []string{"retry in a moment; nothing was written"},
	}
}

// newUnreadableError builds the error of the second case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar:
// the board is where it should be and its database cannot be read, so every
// command aborts with exit code 21 (DAMAGED) and these three lines, with no
// automatic repair.
func (s *Store) newUnreadableError() *model.Error {
	return &model.Error{
		ExitCode: 21,
		Code:     "database_unreadable",
		Message:  fmt.Sprintf("board %s's database could not be read", s.id),
		Hints: []string{
			"it did not open, or it failed its integrity check, and there is no automatic repair",
			"rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id",
		},
	}
}
