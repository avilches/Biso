package store

import (
	"database/sql"
	"fmt"
)

// migrations is the ordered list of schema scripts. Each one's index plus
// one is the value PRAGMA user_version reaches after applying it.
//
// TASK-4 (step 1 of TASK-55) asks for the opening, WAL and transaction
// mechanism, not a task's final schema: that arrives with TASK-10 (step 2,
// the logical data model). This first script creates a synthetic table,
// bench_row, that serves to test the mechanism itself (opening, migration,
// transaction and concurrency) without pre-empting any decision about the
// real model. A later step adds the migrations that create the real
// tables.
//
// Nothing reads this list directly: openAt receives it, so a test can hand
// over a different one without touching what every other test sees.
var migrations = []string{
	`CREATE TABLE bench_row (
		id      INTEGER PRIMARY KEY,
		payload TEXT NOT NULL
	);`,
}

// migrateTo brings s's schema up to date against scripts, reading the
// version it is at now.
func (s *Store) migrateTo(scripts []string) error {
	var version int
	if err := s.scanOne(&version, "PRAGMA user_version"); err != nil {
		return s.classify(err, "read schema version")
	}
	return s.migrateFrom(version, scripts)
}

// migrateFrom applies the scripts that version does not yet reflect.
//
// The version it receives only decides whether there is anything to do: it
// is read outside any transaction, so it can be stale by the time the work
// starts. The work itself re-reads it inside a single immediate
// transaction, which is what makes checking and acting one atomic step. Two
// processes opening the same stale file therefore do not both apply the
// same script: one takes the write lock, migrates and commits, and the
// other reads the version the first one left and finds nothing to do.
//
// Skipping the transaction entirely when there is nothing to apply matters
// as much: that is the ordinary case of every single command, and taking
// the write lock to open a board would make a read fail while another
// process is writing, against guarantee 6 of docs/spec/garantias.md.
func (s *Store) migrateFrom(version int, scripts []string) error {
	if version >= len(scripts) {
		return nil
	}

	if !s.inTx.CompareAndSwap(false, true) {
		return fmt.Errorf("migrate: %w", ErrTxInProgress)
	}
	defer s.inTx.Store(false)

	// _txlock=immediate (see openAt) asks for the write lock right at the
	// BEGIN, so either this transaction owns the schema from here on or it
	// fails without having written anything.
	tx, err := s.db.Begin()
	if err != nil {
		return s.classify(err, "begin the migration")
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	if err := tx.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return s.classify(err, "read schema version")
	}
	for i := version; i < len(scripts); i++ {
		if err := applyMigration(tx, scripts[i], i+1); err != nil {
			return s.classify(err, fmt.Sprintf("apply migration %d", i+1))
		}
	}
	if err := tx.Commit(); err != nil {
		return s.classify(err, "commit the migration")
	}
	committed = true
	return nil
}

// applyMigration runs one script and records the version it leaves behind.
// Both happen inside the caller's transaction, along with every other
// pending script, so an interrupted migration leaves the file exactly on
// the version it already had.
func applyMigration(tx *sql.Tx, script string, toVersion int) error {
	if _, err := tx.Exec(script); err != nil {
		return err
	}
	// PRAGMA user_version takes no bound parameter, so the number is
	// formatted into the statement. It comes from the migration list's own
	// index and never from anything the caller supplies.
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", toVersion)); err != nil {
		return err
	}
	return nil
}
