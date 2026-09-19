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
	s, err := Open(filepath.Join(dir, "board.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	err = s.WithTx(func(tx *sql.Tx) error {
		for i := 1; i <= 3; i++ {
			if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (?, ?)", i, "ok"); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("WithTx: %v", err)
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
}

func TestWithTxRollsBackAllWritesWhenFnFails(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "board.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	wantErr := errors.New("boom")
	err = s.WithTx(func(tx *sql.Tx) error {
		if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'a')"); err != nil {
			return err
		}
		if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (2, 'b')"); err != nil {
			return err
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("WithTx error = %v, want %v", err, wantErr)
	}

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (a failed transaction must write nothing)", count)
	}
}

func TestWithTxRollsBackWhenTheCallbackPanics(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(filepath.Join(dir, "board.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	func() {
		defer func() {
			if recovered := recover(); recovered == nil {
				t.Errorf("the panic did not travel through WithTx")
			}
		}()
		_ = s.WithTx(func(tx *sql.Tx) error {
			if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'a')"); err != nil {
				return err
			}
			panic("boom")
		})
	}()

	var count int
	if err := s.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 0 {
		t.Fatalf("count = %d, want 0 (a panic must leave nothing written and no connection held)", count)
	}
}

func TestWithTxFailsWithBusyWhenLockNotAvailable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	holder, err := open(path, 200)
	if err != nil {
		t.Fatalf("open holder: %v", err)
	}
	defer holder.Close()

	holderTx, err := holder.db.Begin()
	if err != nil {
		t.Fatalf("begin holder tx: %v", err)
	}
	defer holderTx.Rollback()
	if _, err := holderTx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'held')"); err != nil {
		t.Fatalf("insert into the holder's transaction: %v", err)
	}

	writer, err := open(path, 200)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (2, 'blocked')")
		return err
	})

	var busyErr *model.Error
	if !errors.As(err, &busyErr) {
		t.Fatalf("WithTx error = %v, want a *model.Error", err)
	}
	if busyErr.ExitCode != 8 || busyErr.Code != "busy" {
		t.Fatalf("busy error = %+v, want ExitCode 8 and Code busy", busyErr)
	}
	if busyErr.Message != "the board is busy, another process is writing to it" {
		t.Fatalf("busy error message = %q, not the exact text of the concurrency guarantee", busyErr.Message)
	}
	if busyErr.Hint != "retry in a moment; nothing was written" {
		t.Fatalf("busy error hint = %q, not the exact text of the concurrency guarantee", busyErr.Hint)
	}
}

func TestWithTxWaitsForTheConfiguredTimeBeforeGivingUp(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	holder, err := open(path, 200)
	if err != nil {
		t.Fatalf("open holder: %v", err)
	}
	defer holder.Close()

	holderTx, err := holder.db.Begin()
	if err != nil {
		t.Fatalf("begin holder tx: %v", err)
	}
	defer holderTx.Rollback()
	if _, err := holderTx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'held')"); err != nil {
		t.Fatalf("insert into the holder's transaction: %v", err)
	}

	writer, err := open(path, 200)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	start := time.Now()
	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (2, 'blocked')")
		return err
	})
	elapsed := time.Since(start)

	var busyErr *model.Error
	if !errors.As(err, &busyErr) {
		t.Fatalf("WithTx error = %v, want a *model.Error", err)
	}
	if elapsed < 150*time.Millisecond {
		t.Fatalf("WithTx gave up after %s, without waiting out the configured 200ms", elapsed)
	}
}

func TestWithTxAppliesAfterConcurrentWriterCommits(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	holder, err := open(path, 2000)
	if err != nil {
		t.Fatalf("open holder: %v", err)
	}
	defer holder.Close()

	holderTx, err := holder.db.Begin()
	if err != nil {
		t.Fatalf("begin holder tx: %v", err)
	}
	if _, err := holderTx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'held')"); err != nil {
		t.Fatalf("insert into the holder's transaction: %v", err)
	}
	go func() {
		time.Sleep(50 * time.Millisecond)
		holderTx.Commit()
	}()

	writer, err := open(path, 2000)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	err = writer.WithTx(func(tx *sql.Tx) error {
		_, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (2, 'after')")
		return err
	})
	if err != nil {
		t.Fatalf("WithTx after the concurrent commit: %v", err)
	}

	var count int
	if err := writer.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2 (the holder's row and the writer's, one after the other)", count)
	}
}

func TestReadsDoNotBlockOnConcurrentWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	writer, err := open(path, 2000)
	if err != nil {
		t.Fatalf("open writer: %v", err)
	}
	defer writer.Close()

	tx, err := writer.db.Begin()
	if err != nil {
		t.Fatalf("begin writer tx: %v", err)
	}
	defer tx.Rollback()
	if _, err := tx.Exec("INSERT INTO bench_row (id, payload) VALUES (1, 'in-progress')"); err != nil {
		t.Fatalf("insert into the uncommitted transaction: %v", err)
	}

	reader, err := open(path, 2000)
	if err != nil {
		t.Fatalf("open reader: %v", err)
	}
	defer reader.Close()

	done := make(chan error, 1)
	go func() {
		var count int
		done <- reader.db.QueryRow("SELECT COUNT(*) FROM bench_row").Scan(&count)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("the read failed with a write in progress: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("the read did not return in 100ms with an uncommitted write (docs/spec/garantias.md, guarantee 6)")
	}
}
