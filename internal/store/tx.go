package store

import (
	"database/sql"
	"errors"
	"fmt"

	"modernc.org/sqlite"

	"biso/internal/model"
)

// sqliteBusy is SQLite's SQLITE_BUSY primary result code, stable forever in
// its C API: https://www.sqlite.org/rescode.html#busy
const sqliteBusy = 5

// primaryCodeMask keeps the low byte of a result code. SQLite's extended
// codes carry the primary one in their low eight bits, so SQLITE_BUSY,
// SQLITE_BUSY_RECOVERY and SQLITE_BUSY_SNAPSHOT all answer 5 through it:
// https://www.sqlite.org/rescode.html#primary_result_codes_versus_extended_result_codes
const primaryCodeMask = 0xff

// WithTx runs fn inside a single write transaction, the transaction that
// the six guarantees of docs/spec/garantias.md depend on. The connection
// was opened with _txlock=immediate (see Open), so the write lock is
// already requested at BEGIN: if another process holds it, this call waits
// at most the time configured in Open before failing, never partway
// through fn.
//
// If fn returns an error, the transaction is rolled back entirely and that
// error is returned as-is, unwrapped: the caller can recognize its own
// with errors.Is or errors.As. If BEGIN or COMMIT itself fails because the
// write lock was not available, the *model.Error of guarantee 5 is
// returned instead: exit code 8 and the literal text of the specification.
//
// A panic inside fn rolls the transaction back and travels on, so no
// failure can leave the single connection of Open held by an open
// transaction.
func (s *Store) WithTx(fn func(*sql.Tx) error) error {
	tx, err := s.db.Begin()
	if err != nil {
		return busyOrWrap(err, "begin transaction")
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return busyOrWrap(err, "commit transaction")
	}
	committed = true
	return nil
}

func busyOrWrap(err error, action string) error {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) && sqliteErr.Code()&primaryCodeMask == sqliteBusy {
		return newBusyError()
	}
	return fmt.Errorf("%s: %w", action, err)
}

// newBusyError builds the error of guarantee 5 of docs/spec/garantias.md:
// the write lock could not be taken within the wait, nothing was written,
// and the exit code is 8 (see docs/spec/codigos-de-salida.md).
func newBusyError() *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "busy",
		Message:  "the board is busy, another process is writing to it",
		Hint:     "retry in a moment; nothing was written",
	}
}
