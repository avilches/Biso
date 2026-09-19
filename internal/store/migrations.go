package store

import "fmt"

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
var migrations = []string{
	`CREATE TABLE bench_row (
		id      INTEGER PRIMARY KEY,
		payload TEXT NOT NULL
	);`,
}

// migrate brings s's schema up to date, applying in order the scripts from
// migrations that PRAGMA user_version does not yet reflect. Each script and
// the version bump that records it share one transaction, so an interrupted
// migration leaves the file on the version it already had.
func (s *Store) migrate() error {
	var version int
	if err := s.db.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return fmt.Errorf("read schema version: %w", err)
	}

	for i := version; i < len(migrations); i++ {
		if err := s.applyMigration(migrations[i], i+1); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) applyMigration(script string, toVersion int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %d: %w", toVersion, err)
	}
	if _, err := tx.Exec(script); err != nil {
		tx.Rollback()
		return fmt.Errorf("apply migration %d: %w", toVersion, err)
	}
	// PRAGMA user_version takes no bound parameter, so the number is
	// formatted into the statement. It comes from this file's own list
	// index and never from anything the caller supplies.
	if _, err := tx.Exec(fmt.Sprintf("PRAGMA user_version = %d", toVersion)); err != nil {
		tx.Rollback()
		return fmt.Errorf("set schema version to %d: %w", toVersion, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %d: %w", toVersion, err)
	}
	return nil
}
