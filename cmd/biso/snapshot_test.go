package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"biso/internal/board"
)

// These are the cases of docs/spec/cmd/snapshot.md#comportamiento-caso-a-caso
// seen from outside the program. The recipe of git itself is tested inside
// internal/vcs; what is checked here is what `biso snapshot` does with it:
// the two files, where the revision lands and the notes each ending earns.

// git runs a git command in a directory and fails the test if it does not
// work, which is how these tests build a real repository.
func git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=biso", "GIT_AUTHOR_EMAIL=biso@example.com",
		"GIT_COMMITTER_NAME=biso", "GIT_COMMITTER_EMAIL=biso@example.com",
	)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
}

// snapshotBoard is a board of three tasks whose directory is a plain folder,
// in no repository at all.
func snapshotBoard(t *testing.T) (*machine, string) {
	t.Helper()
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	m.env["GIT_AUTHOR_NAME"] = "biso"
	m.env["GIT_AUTHOR_EMAIL"] = "biso@example.com"
	m.env["GIT_COMMITTER_NAME"] = "biso"
	m.env["GIT_COMMITTER_EMAIL"] = "biso@example.com"
	m.run(t, "init", "My project", "--prefix", "MYP").assertCode(t, 0)
	for _, title := range []string{"One", "Two", "Three"} {
		m.run(t, "new", title).assertCode(t, 0)
	}
	return m, m.boardDir(t)
}

func TestSnapshotWritesTheTwoFilesWithoutRunningAnything(t *testing.T) {
	m, dir := snapshotBoard(t)

	got := m.run(t, "snapshot", "--vcs", "none").assertCode(t, 0)

	assertEqual(t, got.stdout,
		"Snapshot written: snapshot.ndjson, board.json (3 tasks)\n",
		"the output of biso snapshot --vcs none")
	if got.stderr != "" {
		t.Errorf("--vcs none wrote to stderr: %s", got.stderr)
	}
	if lines := strings.Count(m.read(t, filepath.Join(dir, board.SnapshotTasksFile)), "\n"); lines != 3 {
		t.Errorf("snapshot.ndjson has %d lines and the board has three tasks", lines)
	}
	if !strings.Contains(m.read(t, filepath.Join(dir, board.SnapshotConfigFile)), `"task_prefix": "MYP"`) {
		t.Errorf("board.json does not carry the prefix:\n%s",
			m.read(t, filepath.Join(dir, board.SnapshotConfigFile)))
	}
	// Nothing was created: --vcs none does not even ask.
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Errorf("--vcs none created a repository")
	}
}

func TestSnapshotCreatesTheBoardItsOwnRepository(t *testing.T) {
	m, dir := snapshotBoard(t)

	got := m.run(t, "snapshot").assertCode(t, 0)

	if !strings.Contains(got.stdout, "to the board's own repository") {
		t.Errorf("the output does not name the board's own repository:\n%s", got.stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err != nil {
		t.Errorf("the repository was not created: %v", err)
	}
	// The three files of the revision, and only those three: the database
	// never goes in, and neither does anything else the folder holds.
	files := committedFiles(t, dir)
	for _, want := range []string{
		board.SnapshotTasksFile, board.SnapshotConfigFile, board.MarkerName(board.MarkerID(dir)),
	} {
		if !contains(files, want) {
			t.Errorf("the revision does not carry %s: %v", want, files)
		}
	}
	if len(files) != 3 {
		t.Errorf("the revision carries %v and only three files go in", files)
	}
}

func TestSnapshotSaysThereWasNothingToCommit(t *testing.T) {
	m, _ := snapshotBoard(t)
	m.run(t, "snapshot").assertCode(t, 0)

	got := m.run(t, "snapshot").assertCode(t, 0)

	if !strings.Contains(got.stderr,
		"note: nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot") {
		t.Errorf("the note is not there:\n%s", got.stderr)
	}
	assertEqual(t, got.stdout,
		"Snapshot written: snapshot.ndjson, board.json (3 tasks)\n",
		"the output of a snapshot with nothing to commit")
}

// TestSnapshotCommitsIntoTheProjectRepositoryThatDoesNotIgnoreIt is the
// third situation of docs/spec/cmd/snapshot.md#dónde-va-la-revisión-y-cómo-se-decide:
// the board lives inside the repository of the project and that repository
// versions it, so no repository of its own is created and the revision goes
// beside the code.
func TestSnapshotCommitsIntoTheProjectRepositoryThatDoesNotIgnoreIt(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	git(t, m.dir, "init", "-q")
	m.write(t, filepath.Join(m.dir, "README.md"), "a project\n")
	git(t, m.dir, "add", "README.md")
	git(t, m.dir, "commit", "-q", "-m", "the project")

	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "tablero").assertCode(t, 0)
	m.run(t, "new", "One").assertCode(t, 0)
	dir := filepath.Join(m.dir, "tablero")

	got := m.run(t, "snapshot").assertCode(t, 0)

	if !strings.Contains(got.stdout, ", the repository this board lives in") {
		t.Errorf("the output does not name the project's repository:\n%s", got.stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Errorf("a repository of its own was created inside the project's")
	}
	// The three files went in by their path, and the database did not: the
	// exclusion file `biso init` wrote inside the board keeps it out.
	files := committedFiles(t, m.dir)
	if !contains(files, "tablero/"+board.SnapshotTasksFile) ||
		!contains(files, "tablero/"+board.SnapshotConfigFile) {
		t.Errorf("the revision carries %v", files)
	}
	if contains(files, "tablero/"+board.DatabaseFile) {
		t.Errorf("the database went into the revision: %v", files)
	}
}

func TestSnapshotLeavesWhatWasStagedOutsideTheBoardAlone(t *testing.T) {
	m := newMachine(t)
	m.env["BISO_ME"] = "@claude"
	git(t, m.dir, "init", "-q")
	m.write(t, filepath.Join(m.dir, "README.md"), "a project\n")
	git(t, m.dir, "add", "README.md")
	git(t, m.dir, "commit", "-q", "-m", "the project")
	m.run(t, "init", "My project", "--prefix", "MYP", "--at", "tablero").assertCode(t, 0)
	m.run(t, "new", "One").assertCode(t, 0)
	m.write(t, filepath.Join(m.dir, "code.go"), "package main\n")
	m.write(t, filepath.Join(m.dir, "other.go"), "package other\n")
	git(t, m.dir, "add", "code.go", "other.go")

	got := m.run(t, "snapshot").assertCode(t, 0)

	if !strings.Contains(got.stderr,
		"note: 2 staged change(s) outside the board were left untouched") {
		t.Errorf("the note about what was staged is not there:\n%s", got.stderr)
	}
	// The revision carries the board's files and never the staged ones.
	if files := committedFiles(t, m.dir); contains(files, "code.go") {
		t.Errorf("a staged file outside the board went into the revision: %v", files)
	}
}

func TestSnapshotForwardsWhatTheOrdersPrinted(t *testing.T) {
	m, _ := snapshotBoard(t)

	got := m.run(t, "snapshot").assertCode(t, 0)

	if !strings.Contains(got.stderr, "git: ") {
		t.Errorf("nothing was forwarded from git:\n%s", got.stderr)
	}
}

func TestSnapshotRefusesAPublicationItCannotRunBeforeWritingAnything(t *testing.T) {
	m, dir := snapshotBoard(t)
	// A custom system with a commit order and no publish order: --vcs push
	// is a usage error, and it comes before the two files are written
	// (docs/spec/cmd/snapshot.md).
	m.write(t, filepath.Join(m.home, ".biso", "config.json"),
		`{ "vcs": "custom", "vcs_custom": { "commit": ["true"] } }`+"\n")

	m.run(t, "snapshot", "--vcs", "push").assertCode(t, 2)

	if _, err := os.Stat(filepath.Join(dir, board.SnapshotTasksFile)); err == nil {
		t.Errorf("snapshot.ndjson was written before the refusal")
	}
}

func TestSnapshotRefusesAnUnknownVCSMode(t *testing.T) {
	m, _ := snapshotBoard(t)
	m.run(t, "snapshot", "--vcs", "fossil").assertCode(t, 2)
}

func TestSnapshotRefusesTheFlagsOfAWritingCommand(t *testing.T) {
	m, _ := snapshotBoard(t)
	m.run(t, "snapshot", "--dry-run").assertCode(t, 2)
	m.run(t, "snapshot", "--print").assertCode(t, 2)
}

func TestSnapshotWithoutABoard(t *testing.T) {
	m := newMachine(t)
	m.run(t, "snapshot").assertCode(t, 20)
}

func TestSnapshotEnvelopeMatchesTheSchemaOfTheSpecification(t *testing.T) {
	m, _ := snapshotBoard(t)

	got := m.run(t, "snapshot", "--vcs", "none", "--json").assertCode(t, 0)

	// The schema of the specification with the values this board really
	// has: three tasks, no revision, and nothing forwarded because nothing
	// ran.
	assertSameJSON(t, got.stdout, `{
		"schemaVersion": 1,
		"kind": "snapshot",
		"generatedAt": "2026-09-06T09:12:04Z",
		"data": {
			"tasks": 3,
			"files": ["snapshot.ndjson", "board.json"],
			"vcs": "none",
			"committed": false,
			"commit": null,
			"repository": null,
			"pushed": false,
			"vcsOutput": [],
			"stagedOutsideBoard": 0,
			"skipped": []
		}
	}`)
	if got.stderr != "" {
		t.Errorf("--json wrote text on stderr: %s", got.stderr)
	}
}

// committedFiles answers the paths the last revision of a repository holds,
// sorted as git lists them.
func committedFiles(t *testing.T, dir string) []string {
	t.Helper()
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", "HEAD")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git ls-tree in %s: %v", dir, err)
	}
	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if line != "" {
			files = append(files, line)
		}
	}
	return files
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
