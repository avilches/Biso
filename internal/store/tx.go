package store

import (
	"database/sql"
	"fmt"
)

// WithTx runs fn inside a single write transaction, the transaction that
// the six guarantees of docs/spec/garantias.md depend on. The connection
// was opened with _txlock=immediate (see Open), so the write lock is
// already requested at BEGIN: if another process holds it, this call waits
// at most the time configured in Open before failing, never partway
// through fn.
//
// fn writes through the *sql.Tx it receives and never through the store,
// which that transaction holds: the store refuses those calls with
// ErrTxInProgress, and so does a second WithTx, because waiting for the
// single connection of Open would never end and guarantee 5 promises a
// failure in five seconds, not silence.
//
// If fn returns an error, the transaction is rolled back entirely and that
// error is returned as-is, unwrapped: the caller can recognize its own
// with errors.Is or errors.As. The one thing that does change on the way
// out is a failure of the driver itself, which is classified like any
// other (see Classify): a statement that could not be written because the
// file is read only is the environment failing and has to end up as exit
// code 8, not as the unforeseen failure of exit code 1. If BEGIN or COMMIT
// itself fails because the write lock was not available, the *model.Error
// of guarantee 5 is returned instead: exit code 8 and the literal text of
// the specification.
//
// A panic inside fn rolls the transaction back and travels on, so no
// failure can leave the single connection of Open held by an open
// transaction.
func (s *Store) WithTx(fn func(*sql.Tx) error) error {
	if !s.inTx.CompareAndSwap(false, true) {
		return fmt.Errorf("WithTx: %w", ErrTxInProgress)
	}
	// Registered first, so it runs last: the rollback below still has the
	// transaction it needs, and the store is free again afterwards even if
	// fn panicked.
	defer s.inTx.Store(false)

	tx, err := s.db.Begin()
	if err != nil {
		return s.classify(err, "begin transaction")
	}

	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := fn(tx); err != nil {
		return s.Classify(err)
	}

	if err := tx.Commit(); err != nil {
		return s.classify(err, "commit transaction")
	}
	committed = true
	return nil
}
