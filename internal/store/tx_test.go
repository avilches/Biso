package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"biso/internal/model"
)

func TestWithTxCommitsAllWritesOnSuccess(t *testing.T) {
	dir := t.TempDir()
	s := openScratch(t, filepath.Join(dir, "board.sqlite"))
	defer s.Close()

	err := s.WithTx(func(tx *sql.Tx) error {
		for i := 1; i <= 3; i++ {
			if _, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (?, ?)", i, "ok"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	var count int
	if err := s.scanOne(&count, "SELECT COUNT(*) FROM scratch"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestWithTxRollsBackAllWritesWhenFnFails(t *testing.T) {
	dir := t.TempDir()
	s := openScratch(t, filepath.Join(dir, "board.sqlite"))
	defer s.Close()

	wantErr := errors.New("boom")
	err := s.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'a')"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (2, 'b')"); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx error = %v, want %v", err, wantErr)
	}

	var count int
	if err := s.scanOne(&count, "SELECT COUNT(*) FROM scratch"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (a failed transaction must write nothing)", count)
	}
}

func TestWithTxRollsBackWhenTheCallbackPanics(t *testing.T) {
	dir := t.TempDir()
	s := openScratch(t, filepath.Join(dir, "board.sqlite"))
	defer s.Close()

	func() {
		defer func() {
			if recovered := recover(); recovered == nil {
				t.Errorf("the panic did not travel through WithTx")
			}
		}()
		_ = s.WithTx(func(tx *sql.Tx) error {
			if _, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'a')"); err != nil {
				return err
			}
			panic("boom")
		})
	}()

	var count int
	if err := s.scanOne(&count, "SELECT COUNT(*) FROM scratch"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (a panic must leave nothing written and no connection held)", count)
	}

	// The panic must also leave the store usable: if the guard against
	// using the handle inside a transaction stayed armed, everything after
	// this would fail.
	if err := s.WithTx(func(tx *sql.Tx) error { return nil }); err != nil {
		t.Fatalf("WithTx after a panic: %v", err)
	}
}

// TestUsingTheStoreHandleInsideWithTxFailsInsteadOfHanging covers the
// consequence of the single connection of Open: the handle is held by the
// live transaction, so a query through it would wait for a connection that
// only the transaction can release, forever and with no error. Guarantee 5
// of docs/spec/garantias.md promises a failure with exit code 8 within five
// seconds, so hanging is not an option: the store refuses the call.
func TestUsingTheStoreHandleInsideWithTxFailsInsteadOfHanging(t *testing.T) {
	dir := t.TempDir()
	s := openScratch(t, filepath.Join(dir, "board.sqlite"))
	defer s.Close()

	done := make(chan error, 1)
	go func() {
		done <- s.WithTx(func(tx *sql.Tx) error {
			var count int
			return s.scanOne(&count, "SELECT COUNT(*) FROM scratch")
		})
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ErrTxInProgress) {
			t.Fatalf("error = %v, want ErrTxInProgress", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("using the store handle inside WithTx hung instead of failing")
	}

	if _, err := s.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'a')"); err != nil {
		t.Fatalf("the store stayed unusable after the refused call: %v", err)
	}
}

func TestNestedWithTxFailsInsteadOfHanging(t *testing.T) {
	dir := t.TempDir()
	s := openScratch(t, filepath.Join(dir, "board.sqlite"))
	defer s.Close()

	done := make(chan error, 1)
	go func() {
		done <- s.WithTx(func(outer *sql.Tx) error {
			return s.WithTx(func(inner *sql.Tx) error { return nil })
		})
	}()

	select {
	case err := <-done:
		if !errors.Is(err, ErrTxInProgress) {
			t.Fatalf("error = %v, want ErrTxInProgress", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("a nested WithTx hung instead of failing")
	}
}

func TestWithTxFailsWithBusyWhenLockNotAvailable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	holder, release := holdTheWriteLock(t, path)
	defer release()
	_ = holder

	writer, err := openAt(testBoardID, path, 200, testMigrations)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (2, 'blocked')")
		return err
	})

	assertBusy(t, err)
}

func TestWithTxWaitsForTheConfiguredTimeBeforeGivingUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	_, release := holdTheWriteLock(t, path)
	defer release()

	writer, err := openAt(testBoardID, path, 200, testMigrations)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	start := time.Now()
	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (2, 'blocked')")
		return err
	})
	elapsed := time.Since(start)

	assertBusy(t, err)
	if elapsed < 150*time.Millisecond {
		t.Fatalf("WithTx gave up after %s, without waiting out the configured 200ms", elapsed)
	}
}

func TestWithTxAppliesAfterConcurrentWriterCommits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	holder, err := openAt(testBoardID, path, 2000, testMigrations)
	if err != nil {
		t.Fatalf("open holder: %v", err)
	}
	defer holder.Close()

	// The holder stands in for another process, so it takes the write lock
	// through the handle directly and keeps it open across the call below.
	holderTx, err := holder.db.Begin()
	if err != nil {
		t.Fatalf("begin holder tx: %v", err)
	}
	if _, err := holderTx.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'held')"); err != nil {
		t.Fatalf("insert into the holder's transaction: %v", err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		holderTx.Commit()
	}()

	writer, err := openAt(testBoardID, path, 2000, testMigrations)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (2, 'after')")
		return err
	})
	if err != nil {
		t.Fatalf("WithTx after the concurrent commit: %v", err)
	}

	var count int
	if err := writer.scanOne(&count, "SELECT COUNT(*) FROM scratch"); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2 (the holder's row and the writer's, one after the other)", count)
	}
}

// TestReadsDoNotBlockOnConcurrentWrite is guarantee 6 and, with the count
// it checks, also guarantee 1: the reader neither waits for the write in
// progress nor sees the row it has not committed yet.
func TestReadsDoNotBlockOnConcurrentWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	_, release := holdTheWriteLock(t, path)
	defer release()

	reader, err := openAt(testBoardID, path, 2000, testMigrations)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	defer reader.Close()

	type read struct {
		count int
		err   error
	}
	done := make(chan read, 1)
	go func() {
		var r read
		r.err = reader.scanOne(&r.count, "SELECT COUNT(*) FROM scratch")
		done <- r
	}()

	select {
	case r := <-done:
		if r.err != nil {
			t.Fatalf("the read failed with a write in progress: %v", r.err)
		}
		if r.count != 0 {
			t.Fatalf("count = %d, want 0 (docs/spec/garantias.md, guarantee 1: no write is observed halfway)", r.count)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("the read did not return in 100ms with an uncommitted write (docs/spec/garantias.md, guarantee 6)")
	}
}

// holdTheWriteLock opens the board at path, writes one row inside a
// transaction and leaves it open, the way another process holding the write
// lock would. It returns the store and the call that lets the lock go.
func holdTheWriteLock(t *testing.T, path string) (*Store, func()) {
	t.Helper()

	holder, err := openAt(testBoardID, path, 200, testMigrations)
	if err != nil {
		t.Fatalf("open the holder: %v", err)
	}
	// Directly on the handle, because WithTx cannot keep a transaction open
	// after it returns: this is the one place that needs that.
	tx, err := holder.db.Begin()
	if err != nil {
		holder.Close()
		t.Fatalf("begin the holder's transaction: %v", err)
	}
	if _, err := tx.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'held')"); err != nil {
		tx.Rollback()
		holder.Close()
		t.Fatalf("insert into the holder's transaction: %v", err)
	}
	return holder, func() {
		tx.Rollback()
		holder.Close()
	}
}

func assertBusy(t *testing.T, err error) {
	t.Helper()

	var busyErr *model.Error
	if !errors.As(err, &busyErr) {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	if busyErr.ExitCode != 8 || busyErr.Code != "busy" {
		t.Fatalf("busy error = %+v, want ExitCode 8 and Code busy", busyErr)
	}
	if busyErr.Message != "the board is busy, another process is writing to it" {
		t.Fatalf("busy error message = %q, not the exact text of the concurrency guarantee", busyErr.Message)
	}
	wantHints := []string{"retry in a moment; nothing was written"}
	if len(busyErr.Hints) != 1 || busyErr.Hints[0] != wantHints[0] {
		t.Fatalf("busy error hints = %q, not the exact text of the concurrency guarantee", busyErr.Hints)
	}
}

// TestOpenReportsBusyWhenTheSchemaMustBeMigrated covers the other path that
// can meet the write lock: a file whose schema is behind is migrated when it
// is opened, and that migration needs the lock like any other write.
func TestOpenReportsBusyWhenTheSchemaMustBeMigrated(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	// A file at version 0: opening it with the real list has migrations to
	// apply, so it needs the write lock.
	empty, err := openAt(testBoardID, path, 200, nil)
	if err != nil {
		t.Fatalf("create the empty file: %v", err)
	}
	if _, err := empty.Exec("CREATE TABLE lock_me (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("create the table the holder writes to: %v", err)
	}

	tx, err := empty.db.Begin()
	if err != nil {
		t.Fatalf("begin the holder's transaction: %v", err)
	}
	if _, err := tx.Exec("INSERT INTO lock_me (id) VALUES (1)"); err != nil {
		t.Fatalf("take the write lock: %v", err)
	}
	defer func() {
		tx.Rollback()
		empty.Close()
	}()

	s, err := openAt(testBoardID, path, 200, testMigrations)
	if err == nil {
		s.Close()
		t.Fatalf("opening a stale file against a held write lock returned no error")
	}
	assertBusy(t, err)
}
