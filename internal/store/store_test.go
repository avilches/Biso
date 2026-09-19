package store

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"biso/internal/model"
)

// testBoardID stands in for a real board id, the eight hexadecimal
// characters of docs/spec/resolucion-del-tablero.md. The store only uses it
// to build the message of the unreadable-database error.
const testBoardID = "3f9a2b1c"

// scratchMigration is a table that exists only for this package's tests.
// They are about the mechanism (opening, migrating, transactions and the
// write lock) and not about the model, so they write to a table of their
// own instead of leaning on the real schema, which would make them fail
// every time a field of docs/spec/modelo-de-datos/ moves.
const scratchMigration = `CREATE TABLE scratch (
	id      INTEGER PRIMARY KEY,
	payload TEXT NOT NULL
);`

// testMigrations is the real list with that table appended, so a store
// opened by openScratch has the whole real schema plus the scratch table.
var testMigrations = append(append([]string{}, migrations...), scratchMigration)

// openScratch opens the board at path with testMigrations, failing the
// test if it cannot.
func openScratch(t *testing.T, path string) *Store {
	t.Helper()

	s, err := openAt(testBoardID, path, busyTimeoutMillis, testMigrations)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	return s
}

func TestOpenCreatesFileInWALModeWithForeignKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	var mode string
	if err := s.scanOne(&mode, "PRAGMA journal_mode"); err != nil {
		t.Fatalf("read journal_mode: %v", err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}

	var fk int
	if err := s.scanOne(&fk, "PRAGMA foreign_keys"); err != nil {
		t.Fatalf("read foreign_keys: %v", err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}

	var timeout int
	if err := s.scanOne(&timeout, "PRAGMA busy_timeout"); err != nil {
		t.Fatalf("read busy_timeout: %v", err)
	}
	if timeout != busyTimeoutMillis {
		t.Fatalf("busy_timeout = %d, want %d (guarantee 5 of docs/spec/garantias.md)", timeout, busyTimeoutMillis)
	}
}

func TestOpenOnExistingFileSucceeds(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	s1, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("first Open: %v", err)
	}
	s1.Close()

	s2, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("second Open: %v", err)
	}
	defer s2.Close()
}

// TestConcurrentCreationOfTheSameFileWaitsInsteadOfFailingAtOnce is
// guarantee 5 of docs/spec/garantias.md on the one path that used to skip
// it: creating the file.
//
// Turning a brand new file into WAL mode needs the file to itself and
// SQLite answers SQLITE_BUSY for it without consulting the busy handler,
// so four connections creating the same file at the same instant used to
// give an instant exit code 8 on several attempts out of twenty. Waiting
// is not optional there: the guarantee says five seconds and then code 8,
// and it does not exempt the command that creates the board.
//
// Twenty attempts, because the collision is a race and one attempt proves
// nothing: reverting initialVersion's retry makes this test fail.
func TestConcurrentCreationOfTheSameFileWaitsInsteadOfFailingAtOnce(t *testing.T) {
	const attempts = 20
	const connections = 4

	for attempt := 1; attempt <= attempts; attempt++ {
		path := filepath.Join(t.TempDir(), "board.sqlite")

		start := make(chan struct{})
		failures := make([]error, connections)
		var wg sync.WaitGroup
		for i := range failures {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				<-start
				s, err := openAt(testBoardID, path, busyTimeoutMillis, testMigrations)
				failures[i] = err
				if s != nil {
					s.Close()
				}
			}(i)
		}
		close(start)
		wg.Wait()

		for i, err := range failures {
			if err != nil {
				t.Fatalf("attempt %d, connection %d: %v", attempt, i, err)
			}
		}
	}
}

func TestOpenOnAPathWithURICharactersUsesThatFile(t *testing.T) {
	dir := t.TempDir()
	odd := filepath.Join(dir, "board #1")
	if err := os.Mkdir(odd, 0o755); err != nil {
		t.Fatalf("create the directory: %v", err)
	}
	path := filepath.Join(odd, "board.sqlite")

	s, err := Open(testBoardID, path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if _, err := s.Exec("CREATE TABLE t (id INTEGER PRIMARY KEY)"); err != nil {
		t.Fatalf("write to the database: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the file was not created where it was asked for: %v", err)
	}
}

// TestURIPathEscapesWhatSQLiteReadsAsURISyntax covers the Windows forms
// directly, without a Windows machine: the driver was chosen because it
// cross-compiles there (docs/decisiones/lenguaje-y-rendimiento.md), so the
// path that goes into the DSN has to be right on that platform too.
func TestURIPathEscapesWhatSQLiteReadsAsURISyntax(t *testing.T) {
	cases := []struct {
		name string
		path string
		want string
	}{
		{
			name: "a plain POSIX path is left alone",
			path: "/home/ada/boards/my-project-3f9a2b1c/board.sqlite",
			want: "/home/ada/boards/my-project-3f9a2b1c/board.sqlite",
		},
		{
			name: "the three characters of URI syntax are escaped",
			path: "/home/ada/board #1/what%?/board.sqlite",
			want: "/home/ada/board %231/what%25%3f/board.sqlite",
		},
		{
			name: "a Windows drive letter gets the leading slash SQLite discards",
			path: `C:\Users\ada\boards\board.sqlite`,
			want: "/C:/Users/ada/boards/board.sqlite",
		},
		{
			name: "a Windows path escapes its URI characters too",
			path: `D:\boards\board #1\board.sqlite`,
			want: "/D:/boards/board %231/board.sqlite",
		},
		{
			name: "a Windows UNC path takes the five leading slashes",
			path: `\\server\share\boards\board.sqlite`,
			want: "/////server/share/boards/board.sqlite",
		},
		{
			name: "a relative Windows path only changes its separators",
			path: `boards\board.sqlite`,
			want: "boards/board.sqlite",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := uriPath(c.path); got != c.want {
				t.Fatalf("uriPath(%q) = %q, want %q", c.path, got, c.want)
			}
		})
	}
}

// TestOpenOnAFileThatIsNotADatabaseFailsAsUnreadable is the second case of
// docs/spec/garantias.md#qué-pasa-con-un-dato-que-no-se-puede-interpretar:
// there is no readable board at all, so the command aborts with exit code
// 21 and the literal text of the specification.
func TestOpenOnAFileThatIsNotADatabaseFailsAsUnreadable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")
	if err := os.WriteFile(path, []byte(strings.Repeat("not a database at all. ", 40)), 0o644); err != nil {
		t.Fatalf("write the broken file: %v", err)
	}

	s, err := Open(testBoardID, path)
	if err == nil {
		s.Close()
		t.Fatalf("Open on a file that is not a database returned no error")
	}
	assertUnreadable(t, err)
}

func TestCheckIntegrityPassesOnAFreshBoard(t *testing.T) {
	dir := t.TempDir()
	s, err := Open(testBoardID, filepath.Join(dir, "board.sqlite"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer s.Close()

	if err := s.CheckIntegrity(); err != nil {
		t.Fatalf("CheckIntegrity on a board just created: %v", err)
	}
}

// TestCheckIntegrityFailsOnACorruptedDatabase is the other half of the same
// case of the specification: the file opens, but its integrity check fails,
// which is the check biso doctor runs (docs/spec/cmd/doctor.md).
func TestCheckIntegrityFailsOnACorruptedDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "board.sqlite")

	seed := openScratch(t, path)
	if _, err := seed.Exec("INSERT INTO scratch (id, payload) VALUES (1, 'a'), (2, 'b'), (3, 'c')"); err != nil {
		t.Fatalf("seed rows: %v", err)
	}
	seed.Close()

	corruptPage(t, path)

	s, err := Open(testBoardID, path)
	if err != nil {
		// Damage this deep can already show up when opening, which is the
		// same case of the specification and the same error.
		t.Logf("the damage showed up when opening: %v", err)
		assertUnreadable(t, err)
		return
	}
	defer s.Close()

	err = s.CheckIntegrity()
	if err == nil {
		t.Fatalf("CheckIntegrity on a corrupted file reported no problem")
	}
	assertUnreadable(t, err)
}

// corruptPage overwrites the interior of the database's second page, which
// holds the rows of the only table, leaving the header intact so the file
// still looks like a database.
func corruptPage(t *testing.T, path string) {
	t.Helper()

	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open the file to damage it: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat the file: %v", err)
	}
	if info.Size() < 8192 {
		t.Fatalf("the file is %d bytes, too small to have a second page to damage", info.Size())
	}
	if _, err := f.WriteAt([]byte(strings.Repeat("\xff", 512)), 4096+8); err != nil {
		t.Fatalf("damage the second page: %v", err)
	}
}

func assertUnreadable(t *testing.T, err error) {
	t.Helper()

	var got *model.Error
	if !errors.As(err, &got) {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	if got.ExitCode != 21 {
		t.Fatalf("exit code = %d, want 21 (DAMAGED, docs/spec/codigos-de-salida.md)", got.ExitCode)
	}
	if got.Code != "database_unreadable" {
		t.Fatalf("code = %q, want database_unreadable", got.Code)
	}
	wantMessage := "board 3f9a2b1c's database could not be read"
	if got.Message != wantMessage {
		t.Fatalf("message = %q, want %q (docs/spec/garantias.md)", got.Message, wantMessage)
	}
	wantHints := []string{
		"it did not open, or it failed its integrity check, and there is no automatic repair",
		"rebuild it in place with `biso init --from <snapshot dir>`, which keeps its id",
	}
	if len(got.Hints) != len(wantHints) {
		t.Fatalf("hints = %q, want the two lines of docs/spec/garantias.md", got.Hints)
	}
	for i, want := range wantHints {
		if got.Hints[i] != want {
			t.Fatalf("hint %d = %q, want %q (docs/spec/garantias.md)", i+1, got.Hints[i], want)
		}
	}
}
