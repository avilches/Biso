package store

import (
	"database/sql"
	"fmt"
)

// migrations is the ordered list of schema scripts. Each one's index plus
// one is the value PRAGMA user_version reaches after applying it.
//
// This first script is the whole logical model of
// docs/spec/modelo-de-datos/index.md, written as tables. It replaced the
// synthetic bench_row table that TASK-4 (step 1 of TASK-55) used to test
// the mechanism itself before there was any model to store: biso has never
// shipped, so no board exists on any earlier version and there is nothing
// to migrate from.
//
// Three shapes of the script are decisions of this layer and not of the
// specification:
//
//   - **Every string column is NOT NULL and holds the empty string when
//     the field has no value**, because
//     docs/spec/valores-de-entrada.md#el-valor-vacío makes the empty
//     string a value no caller can ever store: on a scalar it is an error,
//     and clearing a field is a flag of its own. So "" and "no value"
//     cannot be told apart by anything observable, and carrying the
//     difference into SQL would only mean scanning every column through a
//     nullable type. The one exception is task.ordinal, where 0 is a value
//     a caller can legitimately give (int >= 0), so absence needs NULL.
//   - **The six list<string> fields share one table**, task_list_item,
//     keyed by the field's name, instead of one table each. They have the
//     same shape (ordered values with no structure of their own,
//     addressed by their value and never by position), the CHECK keeps the
//     field name a closed vocabulary just as internal/model does, and one
//     table means one query reads every list of every task.
//   - **Each list keeps an explicit position**, because
//     docs/spec/garantias.md#orden-de-aplicación-dentro-de-una-escritura
//     says lists are never sorted on their own: the order they were
//     written in is data, so it is stored and not recomputed.
//
// Nothing reads this list directly: openAt receives it, so a test can hand
// over a different one without touching what every other test sees.
var migrations = []string{
	`CREATE TABLE board_counter (
		id            INTEGER PRIMARY KEY CHECK (id = 1),
		last_task_num INTEGER NOT NULL
	);

	INSERT INTO board_counter (id, last_task_num) VALUES (1, 0);

	CREATE TABLE task (
		id                 TEXT    PRIMARY KEY,
		num                INTEGER NOT NULL UNIQUE,
		title              TEXT    NOT NULL,
		status             TEXT    NOT NULL,
		type               TEXT    NOT NULL,
		priority           TEXT    NOT NULL,
		parent             TEXT    NOT NULL,
		author             TEXT    NOT NULL,
		due                TEXT    NOT NULL,
		ordinal            INTEGER,
		description        TEXT    NOT NULL,
		plan               TEXT    NOT NULL,
		notes              TEXT    NOT NULL,
		summary            TEXT    NOT NULL,
		created_at         TEXT    NOT NULL,
		updated_at         TEXT    NOT NULL,
		archived           INTEGER NOT NULL,
		lease_expires_at   TEXT    NOT NULL,
		lease_holder       TEXT    NOT NULL,
		question_author    TEXT    NOT NULL,
		question_asked_at  TEXT    NOT NULL,
		question_body      TEXT    NOT NULL,
		next_criterion_key INTEGER NOT NULL,
		next_comment_key   INTEGER NOT NULL
	);

	CREATE TABLE task_list_item (
		task_id  TEXT    NOT NULL REFERENCES task(id) ON DELETE CASCADE,
		field    TEXT    NOT NULL CHECK (field IN (
			'assignees', 'labels', 'dependencies',
			'references', 'documentation', 'modifiedFiles'
		)),
		position INTEGER NOT NULL,
		value    TEXT    NOT NULL,
		PRIMARY KEY (task_id, field, position)
	) WITHOUT ROWID;

	CREATE TABLE task_ext (
		task_id TEXT NOT NULL REFERENCES task(id) ON DELETE CASCADE,
		key     TEXT NOT NULL,
		value   TEXT NOT NULL,
		PRIMARY KEY (task_id, key)
	) WITHOUT ROWID;

	CREATE TABLE task_criterion (
		task_id  TEXT    NOT NULL REFERENCES task(id) ON DELETE CASCADE,
		key      INTEGER NOT NULL,
		position INTEGER NOT NULL,
		text     TEXT    NOT NULL,
		checked  INTEGER NOT NULL,
		PRIMARY KEY (task_id, key)
	) WITHOUT ROWID;

	CREATE TABLE task_comment (
		task_id    TEXT    NOT NULL REFERENCES task(id) ON DELETE CASCADE,
		key        INTEGER NOT NULL,
		position   INTEGER NOT NULL,
		author     TEXT    NOT NULL,
		created_at TEXT    NOT NULL,
		body       TEXT    NOT NULL,
		PRIMARY KEY (task_id, key)
	) WITHOUT ROWID;`,
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
