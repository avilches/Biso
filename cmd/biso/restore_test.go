package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/board"
)

// These are the rows of the case table of docs/spec/cmd/init.md that speak
// of --from: what a half snapshot answers, what an unreadable board.json or
// marker answers, and what a snapshot whose tasks do not fit its own
// vocabulary answers.

// snapshotDir is a directory holding the three files of a snapshot, built
// by `biso snapshot` over a board of two tasks so that it is a real one.
func snapshotDir(t *testing.T) (*machine, string) {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.run(t, "init", "My project", "--prefix", "MYP",
		"--extensions", "trello.card").assertCode(t, 0)
	m.run(t, "new", "One").assertCode(t, 0)
	m.run(t, "new", "Two", "--status", "Done").assertCode(t, 0)
	m.run(t, "snapshot", "--vcs", "none").assertCode(t, 0)
	return m, m.boardDir(t)
}

// restoreInto restores a snapshot into a new directory from a project that
// has no board of its own, which is how the specification's own example
// does it.
func (m *machine) restoreInto(t *testing.T, from, into string, extra ...string) call {
	t.Helper()
	elsewhere := filepath.Join(m.home, "elsewhere")
	m.mkdir(t, elsewhere)
	argv := append([]string{"init", "--at", into, "--from", from}, extra...)
	return m.at(elsewhere).run(t, argv...)
}

func TestRestoreBringsTheVocabularyTheSnapshotCarries(t *testing.T) {
	m, dir := snapshotDir(t)
	into := filepath.Join(m.home, "restored")

	got := m.restoreInto(t, dir, into).assertCode(t, 0)

	if !strings.Contains(got.stdout, `Created board "My project"`) {
		t.Errorf("the output does not name the board of the snapshot:\n%s", got.stdout)
	}
	if !strings.Contains(got.stdout, "prefix      MYP") {
		t.Errorf("the prefix did not come from board.json:\n%s", got.stdout)
	}
	assertEqual(t, m.at(into).run(t, "ls", "--any-status", "--count").assertCode(t, 0).stdout,
		"2\n", "the tasks of the restored board")
}

func TestRestoreNamesTheFileAHalfSnapshotIsMissing(t *testing.T) {
	m, dir := snapshotDir(t)
	for _, name := range []string{board.SnapshotTasksFile, board.SnapshotConfigFile} {
		half := filepath.Join(m.home, "half-"+name)
		copySnapshot(t, m, dir, half)
		if err := os.Remove(filepath.Join(half, name)); err != nil {
			t.Fatal(err)
		}

		got := m.restoreInto(t, half, filepath.Join(m.home, "out-"+name))

		if got.code != 4 {
			t.Errorf("a snapshot without %s exited %d and not 4", name, got.code)
		}
		if !strings.Contains(got.stderr, name) {
			t.Errorf("the error does not name %s:\n%s", name, got.stderr)
		}
	}
}

func TestRestoreRefusesADirectoryWithNoMarker(t *testing.T) {
	m, dir := snapshotDir(t)
	half := filepath.Join(m.home, "no-marker")
	copySnapshot(t, m, dir, half)
	if err := os.Remove(filepath.Join(half, board.MarkerName(board.MarkerID(dir)))); err != nil {
		t.Fatal(err)
	}

	got := m.restoreInto(t, half, filepath.Join(m.home, "out"))

	if got.code != 4 || !strings.Contains(got.stderr, "<id>.id") {
		t.Errorf("a snapshot without a marker exited %d:\n%s", got.code, got.stderr)
	}
}

func TestRestoreRefusesAnUnreadableBoardJSON(t *testing.T) {
	m, dir := snapshotDir(t)
	for i, content := range []string{
		`not json at all`,
		`{"project_name":"My project","unknown_key":1}`,
	} {
		broken := filepath.Join(m.home, "broken", string(rune('a'+i)))
		copySnapshot(t, m, dir, broken)
		m.write(t, filepath.Join(broken, board.SnapshotConfigFile), content)

		got := m.restoreInto(t, broken, filepath.Join(m.home, "out-cfg", string(rune('a'+i))))

		if got.code != 2 || !strings.Contains(got.stderr, "board.json cannot be read") {
			t.Errorf("board.json %q exited %d:\n%s", content, got.code, got.stderr)
		}
	}
}

func TestRestoreRefusesAMarkerThatDoesNotHoldWhatItShould(t *testing.T) {
	m, dir := snapshotDir(t)
	broken := filepath.Join(m.home, "broken-marker")
	copySnapshot(t, m, dir, broken)
	m.write(t, filepath.Join(broken, board.MarkerName(board.MarkerID(dir))), "{}\n")

	got := m.restoreInto(t, broken, filepath.Join(m.home, "out-marker"))

	if got.code != 2 || !strings.Contains(got.stderr, "storeVersion") {
		t.Errorf("a marker with the wrong content exited %d:\n%s", got.code, got.stderr)
	}
}

func TestRestoreRefusesAVocabularyThatCannotBeABoard(t *testing.T) {
	m, dir := snapshotDir(t)
	for _, c := range []struct {
		name, config, expect string
	}{
		{
			"two statuses",
			`{"project_name":"P","statuses":["A","B"],"initial_status":"A","active_status":"A",` +
				`"terminal_status":"B","types":[],"priorities":[],"labels":[],"assignees":[],` +
				`"extensions":[],"task_prefix":"MYP","finish_strict":false,"lease_minutes":240,` +
				`"urgency":{"priority":6.0,"active":4.0,"blocking":8.0,"blocked":-5.0,"due":12.0,"criteria":1.0,"age":0.5}}`,
			"at least three",
		},
		{
			"a prefix that is not letters",
			`{"project_name":"P","statuses":["A","B","C"],"initial_status":"A","active_status":"B",` +
				`"terminal_status":"C","types":[],"priorities":[],"labels":[],"assignees":[],` +
				`"extensions":[],"task_prefix":"9","finish_strict":false,"lease_minutes":240,` +
				`"urgency":{"priority":6.0,"active":4.0,"blocking":8.0,"blocked":-5.0,"due":12.0,"criteria":1.0,"age":0.5}}`,
			"letters",
		},
	} {
		broken := filepath.Join(m.home, "vocab", strings.ReplaceAll(c.name, " ", "-"))
		copySnapshot(t, m, dir, broken)
		m.write(t, filepath.Join(broken, board.SnapshotConfigFile), c.config)
		m.write(t, filepath.Join(broken, board.SnapshotTasksFile), "")

		got := m.restoreInto(t, broken, filepath.Join(m.home, "out-vocab",
			strings.ReplaceAll(c.name, " ", "-")))

		if got.code != 2 || !strings.Contains(got.stderr, c.expect) {
			t.Errorf("%s exited %d:\n%s", c.name, got.code, got.stderr)
		}
	}
}

func TestRestoreAcceptsAnEmptySnapshotOfTasks(t *testing.T) {
	m, dir := snapshotDir(t)
	empty := filepath.Join(m.home, "empty")
	copySnapshot(t, m, dir, empty)
	m.write(t, filepath.Join(empty, board.SnapshotTasksFile), "")
	into := filepath.Join(m.home, "restored-empty")

	m.restoreInto(t, empty, into).assertCode(t, 0)

	assertEqual(t, m.at(into).run(t, "ls", "--any-status", "--count").assertCode(t, 0).stdout,
		"0\n", "the tasks of a board restored from an empty snapshot")
}

func TestRestoreRefusesTasksThatDoNotFitTheirOwnVocabulary(t *testing.T) {
	m, dir := snapshotDir(t)
	broken := filepath.Join(m.home, "bad-tasks")
	copySnapshot(t, m, dir, broken)
	m.write(t, filepath.Join(broken, board.SnapshotTasksFile),
		"{\"title\":\"Fine\"}\n{\"title\":\"Bad\",\"status\":\"Pendiente\"}\n")
	into := filepath.Join(m.home, "not-restored")

	got := m.restoreInto(t, broken, into)

	if got.code != 7 {
		t.Errorf("a snapshot whose tasks do not fit exited %d and not 7", got.code)
	}
	if !strings.Contains(got.stderr, "line 2: unknown status") {
		t.Errorf("the failure does not name the line:\n%s", got.stderr)
	}
	// Nothing was written: not even the board.
	if _, err := os.Stat(into); err == nil {
		t.Errorf("a failed restore left %s behind", into)
	}
}

func TestRestorePreviewCountsTheTasksAndCreatesNothing(t *testing.T) {
	m, dir := snapshotDir(t)
	into := filepath.Join(m.home, "previewed")

	got := m.restoreInto(t, dir, into, "--dry-run").assertCode(t, 0)

	assertEqual(t, got.stderr,
		"2 tasks would be created, nothing was written (--dry-run)\n",
		"the preview of a restore")
	if _, err := os.Stat(into); err == nil {
		t.Errorf("a preview created %s", into)
	}
}

func TestRestoreIsIncompatibleWithEveryVocabularyFlag(t *testing.T) {
	m, dir := snapshotDir(t)
	for _, extra := range [][]string{
		{"Another name"},
		{"--prefix", "OTH"},
		{"--types", "task,bug"},
		{"--statuses", "a,b,c"},
		{"--priorities", "high"},
		{"--extensions", "trello.card"},
		{"--overwrite-config"},
	} {
		argv := append([]string{"init", "--from", dir}, extra...)
		if got := m.run(t, argv...); got.code != 2 {
			t.Errorf("biso %s exited %d and not 2", strings.Join(argv, " "), got.code)
		}
	}
}

// TestRestoreRebuildsInPlaceADirectoryWithNoReadableDatabase is the row of
// the case table about a destination that is not a board any more: --from
// rebuilds it there, adopting the identity of its marker, and the pointer a
// project had committed keeps working.
func TestRestoreRebuildsInPlaceADirectoryWithNoReadableDatabase(t *testing.T) {
	m, dir := snapshotDir(t)
	id := board.MarkerID(dir)
	// A clone brought to another machine arrives like this: the folder was
	// versioned, so the marker and the two files are there and the database
	// is not.
	for _, name := range []string{board.DatabaseFile, board.DatabaseFile + "-wal", board.DatabaseFile + "-shm"} {
		os.Remove(filepath.Join(dir, name))
	}

	got := m.at(dir).run(t, "init", "--from", dir).assertCode(t, 0)

	if !strings.Contains(got.stdout, `Created board "My project"`) {
		t.Errorf("the board was not rebuilt in place:\n%s", got.stdout)
	}
	if board.MarkerID(dir) != id {
		t.Errorf("the identity changed from %s to %s", id, board.MarkerID(dir))
	}
	assertEqual(t, m.at(dir).run(t, "ls", "--any-status", "--count").assertCode(t, 0).stdout,
		"2\n", "the tasks the rebuild restored")
}

// copySnapshot copies the three files of a snapshot into a new directory,
// which is how a test builds a broken one without touching the board the
// others read.
func copySnapshot(t *testing.T, m *machine, from, to string) {
	t.Helper()
	m.mkdir(t, to)
	entries, err := os.ReadDir(from)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() || strings.HasPrefix(e.Name(), board.DatabaseFile) {
			continue
		}
		m.write(t, filepath.Join(to, e.Name()), m.read(t, filepath.Join(from, e.Name())))
	}
}

// TestRestoreRefusesADestinationThatIsStillAWholeBoard is the other side of
// the board_exists of this command: --from always creates a board, so a
// destination whose database opens is never rewritten.
func TestRestoreRefusesADestinationThatIsStillAWholeBoard(t *testing.T) {
	m, dir := snapshotDir(t)

	got := m.restoreInto(t, dir, dir)

	if got.code != 2 || !strings.Contains(got.stderr, "already holds board") {
		t.Errorf("restoring over a live board exited %d:\n%s", got.code, got.stderr)
	}
	// The board it refused to touch is still there, whole.
	assertEqual(t, m.run(t, "ls", "--any-status", "--count").assertCode(t, 0).stdout,
		"2\n", "the tasks of the board a refused restore left alone")
}
