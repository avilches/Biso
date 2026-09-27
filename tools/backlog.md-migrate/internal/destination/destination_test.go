package destination

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// findBisoBinary returns the path to the bin/biso this repository's own
// Makefile builds (see tools/backlog.md-migrate/CLAUDE.md, "Es un proyecto
// aparte"), skipping the test with a clear message when it has not been
// built yet, rather than failing.
func findBisoBinary(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not report this file's path")
	}
	// thisFile is <repo>/tools/backlog.md-migrate/internal/destination/destination_test.go.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	biso := filepath.Join(repoRoot, "bin", "biso")

	if _, err := os.Stat(biso); err != nil {
		t.Skip("bin/biso not found; run `make build` from the repository root first")
	}
	return biso
}

// runBiso runs the biso binary with args and fails the test, with its
// stdout and stderr, if it exits with anything other than 0.
func runBiso(t *testing.T, biso string, args ...string) string {
	t.Helper()

	cmd := exec.Command(biso, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("biso %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return strings.TrimSpace(string(output))
}

// newTestProject creates a fresh biso board in a temporary directory
// outside the repository, with the given extra biso init flags on top of
// --prefix XYZ, and returns the project directory that resolves to it (the
// one Read's project parameter takes).
func newTestProject(t *testing.T, biso string, extraInitFlags ...string) string {
	t.Helper()

	root := t.TempDir()
	project := filepath.Join(root, "project")
	board := filepath.Join(root, "board")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", project, err)
	}

	args := append([]string{"-C", project, "init", "Fixture", "--prefix", "XYZ", "--at", board}, extraInitFlags...)
	runBiso(t, biso, args...)
	return project
}

func findTask(t *testing.T, tasks []Task, id string) Task {
	t.Helper()
	for _, task := range tasks {
		if task.ID == id {
			return task
		}
	}
	var ids []string
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	t.Fatalf("no task with id %q, got: %v", id, ids)
	return Task{}
}

func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestReadsTheDestinationConfiguration(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso,
		"--statuses", "To Do,In Progress,Blocked,Done",
		"--initial-status", "To Do",
		"--active-status", "In Progress",
		"--terminal-status", "Done",
		"--types", "task,bug",
		"--priorities", "high,low",
	)

	board, err := Read(biso, project)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if board.Config.TaskPrefix != "XYZ" {
		t.Errorf("TaskPrefix = %q, want XYZ", board.Config.TaskPrefix)
	}
	wantStatuses := []string{"To Do", "In Progress", "Blocked", "Done"}
	if !equalStringSlices(board.Config.Statuses, wantStatuses) {
		t.Errorf("Statuses = %v, want %v", board.Config.Statuses, wantStatuses)
	}
	wantTypes := []string{"task", "bug"}
	if !equalStringSlices(board.Config.Types, wantTypes) {
		t.Errorf("Types = %v, want %v", board.Config.Types, wantTypes)
	}
	wantPriorities := []string{"high", "low"}
	if !equalStringSlices(board.Config.Priorities, wantPriorities) {
		t.Errorf("Priorities = %v, want %v", board.Config.Priorities, wantPriorities)
	}
}

func TestReadsDestinationTasksWithAndWithoutManualOrder(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	before := time.Now().UTC()
	id1 := runBiso(t, biso, "-C", project, "new", "Task without order", "--quiet")
	id2 := runBiso(t, biso, "-C", project, "new", "Task with order", "--quiet")
	runBiso(t, biso, "-C", project, "set", id2, "--ordinal", "last")
	runBiso(t, biso, "-C", project, "archive", id1)
	after := time.Now().UTC()

	board, err := Read(biso, project)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(board.Tasks) != 2 {
		t.Fatalf("len(Tasks) = %d, want 2 (archived tasks must be included)", len(board.Tasks))
	}

	withoutOrder := findTask(t, board.Tasks, id1)
	if withoutOrder.Title != "Task without order" {
		t.Errorf("Title = %q, want %q", withoutOrder.Title, "Task without order")
	}
	if withoutOrder.Ordinal != "" {
		t.Errorf("Ordinal = %q, want empty (no manual order set)", withoutOrder.Ordinal)
	}
	createdAt, err := time.Parse(time.RFC3339, withoutOrder.CreatedAt)
	if err != nil {
		t.Fatalf("CreatedAt %q does not parse as RFC3339: %v", withoutOrder.CreatedAt, err)
	}
	if createdAt.Before(before.Add(-time.Second)) || createdAt.After(after.Add(time.Second)) {
		t.Errorf("CreatedAt = %v, want between %v and %v", createdAt, before, after)
	}

	withOrder := findTask(t, board.Tasks, id2)
	if withOrder.Title != "Task with order" {
		t.Errorf("Title = %q, want %q", withOrder.Title, "Task with order")
	}
	if withOrder.Ordinal == "" {
		t.Errorf("Ordinal is empty, want a manual order key after --ordinal last")
	}
}

func TestEmptyDestinationHasNoTasksAndNoError(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	board, err := Read(biso, project)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if len(board.Tasks) != 0 {
		t.Errorf("len(Tasks) = %d, want 0 on a freshly initialized board", len(board.Tasks))
	}
}

func TestNonexistentBisoBinaryProducesAClearError(t *testing.T) {
	// This one does not need a real biso: it exercises the case where the
	// binary itself cannot be executed, so any project directory will do.
	project := t.TempDir()
	nonexistent := filepath.Join(t.TempDir(), "no-such-biso-binary")

	_, err := Read(nonexistent, project)
	if err == nil {
		t.Fatal("Read: got nil error, want one about biso not being runnable")
	}
	if !strings.Contains(err.Error(), "could not run") {
		t.Errorf("error = %q, want it to say the binary could not be run", err.Error())
	}
}

func TestProjectWithoutABoardProducesAClearError(t *testing.T) {
	biso := findBisoBinary(t)
	// A plain empty directory: biso init was never run against it, so
	// "biso config list --json" exits with code 20, "no board here".
	project := t.TempDir()

	_, err := Read(biso, project)
	if err == nil {
		t.Fatal("Read: got nil error, want one about there being no board")
	}
	if !strings.Contains(err.Error(), "20") {
		t.Errorf("error = %q, want it to mention exit code 20", err.Error())
	}
	if !strings.Contains(err.Error(), "no board") {
		t.Errorf("error = %q, want it to include biso's own \"no board\" message", err.Error())
	}
}
