package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// findBisoBinary returns the path to the bin/biso this repository's own
// Makefile builds (see tools/backlog.md-migrate/CLAUDE.md, "Es un proyecto
// aparte"), skipping the test with a clear message when it has not been
// built yet, rather than failing. Mirrors
// internal/destination/destination_test.go's helper of the same name.
func findBisoBinary(t *testing.T) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not report this file's path")
	}
	// thisFile is <repo>/tools/backlog.md-migrate/internal/cli/import_run_test.go.
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

// newTestProject creates a fresh biso board in a temporary directory outside
// the repository and returns the project directory (the one --project
// resolves the destination board from).
func newTestProject(t *testing.T, biso string) string {
	t.Helper()

	root := t.TempDir()
	project := filepath.Join(root, "project")
	board := filepath.Join(root, "board")
	if err := os.MkdirAll(project, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", project, err)
	}

	runBiso(t, biso, "-C", project, "init", "Fixture", "--prefix", "BISO", "--at", board)
	return project
}

// writeMinimalBacklogTask writes a single task file directly inside
// backlogDir's tasks/ directory, in the minimal frontmatter-only shape
// docs/especificacion.md, "Casos que no son obvios", describes ("Una tarea
// puede no tener ninguna sección"). It writes the file by hand rather than
// invoking the real backlog CLI, since these tests only need control over a
// handful of fields (id, title, status) to exercise runImport's own
// orchestration and exit codes, not the source reader's parsing (which has
// its own exhaustive tests in internal/source).
func writeMinimalBacklogTask(t *testing.T, backlogDir, id, title, status string) {
	t.Helper()

	tasksDir := filepath.Join(backlogDir, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(%q): %v", tasksDir, err)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("id: " + id + "\n")
	b.WriteString("title: " + title + "\n")
	if status != "" {
		b.WriteString("status: " + status + "\n")
	}
	b.WriteString("---\n")

	path := filepath.Join(tasksDir, strings.ToLower(id)+".md")
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", path, err)
	}
}

func TestRunImportSourceBacklogDirDoesNotExist(t *testing.T) {
	nonexistent := filepath.Join(t.TempDir(), "no-such-backlog-dir")

	_, errOut, code := run("import", nonexistent, "--project", ".")
	if code != 3 {
		t.Fatalf("got exit code %d, want 3", code)
	}
	// <backlog-dir> itself missing is a different condition than <backlog-dir>
	// existing without a tasks/ subdirectory (TestRunImportSourceBacklogDirHasNoTasksDirectory
	// below), and gets its own distinct message rather than being described
	// as if it were merely missing tasks/.
	if !strings.Contains(errOut, "does not exist") {
		t.Fatalf("errOut = %q, want it to say the backlog directory does not exist", errOut)
	}
}

func TestRunImportSourceBacklogDirHasNoTasksDirectory(t *testing.T) {
	backlogDir := t.TempDir()
	// backlogDir exists, but has no tasks/ subdirectory at all.

	_, errOut, code := run("import", backlogDir, "--project", ".")
	if code != 3 {
		t.Fatalf("got exit code %d, want 3", code)
	}
	if !strings.Contains(errOut, "tasks/") {
		t.Fatalf("errOut = %q, want it to mention the missing tasks/ directory", errOut)
	}
	if strings.Contains(errOut, "does not exist") {
		t.Fatalf("errOut = %q, want it to distinguish this from the backlog directory itself not existing", errOut)
	}
}

func TestRunImportSourceBacklogDirIsAFileNotADirectory(t *testing.T) {
	notADir := filepath.Join(t.TempDir(), "backlog-is-a-file")
	if err := os.WriteFile(notADir, []byte("not a directory"), 0o644); err != nil {
		t.Fatalf("WriteFile(%q): %v", notADir, err)
	}

	_, errOut, code := run("import", notADir, "--project", ".")
	if code != 3 {
		t.Fatalf("got exit code %d, want 3", code)
	}
	if !strings.Contains(errOut, "is not a directory") {
		t.Fatalf("errOut = %q, want it to say the backlog directory is not a directory", errOut)
	}
}

func TestRunImportDestinationBisoBinaryNotFound(t *testing.T) {
	backlogDir := t.TempDir()
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "A task", "")

	nonexistentBiso := filepath.Join(t.TempDir(), "no-such-biso-binary")

	_, errOut, code := run("import", backlogDir, "--project", ".", "--biso", nonexistentBiso)
	if code != 4 {
		t.Fatalf("got exit code %d, want 4", code)
	}
	if errOut == "" {
		t.Fatal("expected an error message on stderr")
	}
}

func TestRunImportNoFindingsExitsZeroAndWritesNdjson(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	// The default initial status of a freshly initialized board always
	// matches a task with no status at all (biso assigns the initial status
	// itself when the status field is omitted from the batch line), so this
	// task converts with no findings whatsoever.
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "A clean task", "")

	stdout, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso)
	if code != 0 {
		t.Fatalf("got exit code %d, want 0, stderr: %s", code, errOut)
	}
	if errOut != "" {
		t.Fatalf("errOut = %q, want empty (no findings)", errOut)
	}

	var line map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(stdout)), &line); err != nil {
		t.Fatalf("stdout is not valid JSON: %v\nstdout: %s", err, stdout)
	}
	if line["id"] != "BISO-1" {
		t.Errorf("id = %v, want BISO-1", line["id"])
	}
	if line["title"] != "A clean task" {
		t.Errorf("title = %v, want %q", line["title"], "A clean task")
	}
}

func TestRunImportWithFindingsExitsFiveAndStillWritesNdjson(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	// "Not A Configured Status" matches none of the destination's
	// configured statuses, so MatchField raises a Finding and the status
	// field is omitted from the line: the task still converts and the
	// NDJSON is still written, per docs/especificacion.md, "Códigos de
	// salida", row 5.
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "Odd status task", "Not A Configured Status")

	stdout, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso)
	if code != 5 {
		t.Fatalf("got exit code %d, want 5, stderr: %s", code, errOut)
	}
	if !strings.HasPrefix(errOut, "warning: ") {
		t.Fatalf("errOut = %q, want it to start with a finding line", errOut)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("stdout is empty, want the NDJSON line to still be written (no --strict)")
	}
}

func TestRunImportStrictWithFindingsWritesNothing(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "Odd status task", "Not A Configured Status")

	stdout, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--strict")
	if code != 5 {
		t.Fatalf("got exit code %d, want 5, stderr: %s", code, errOut)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty (--strict writes nothing when there are findings)", stdout)
	}
	if !strings.HasPrefix(errOut, "warning: ") {
		t.Fatalf("errOut = %q, want it to start with a finding line", errOut)
	}
}

func TestRunImportStrictWithoutFindingsStillWrites(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "A clean task", "")

	stdout, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--strict")
	if code != 0 {
		t.Fatalf("got exit code %d, want 0, stderr: %s", code, errOut)
	}
	if errOut != "" {
		t.Fatalf("errOut = %q, want empty", errOut)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("stdout is empty, want the NDJSON line (--strict only changes behavior when there ARE findings)")
	}
}

func TestRunImportOutFileIsWrittenAtomicallyWithNoLeftoverTempFile(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "A clean task", "")

	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "tasks.ndjson")

	_, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--out", outFile)
	if code != 0 {
		t.Fatalf("got exit code %d, want 0, stderr: %s", code, errOut)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", outDir, err)
	}
	if len(entries) != 1 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("outDir has %d entries, want exactly 1 (tasks.ndjson, no leftover temp file): %v", len(entries), names)
	}
	if entries[0].Name() != "tasks.ndjson" {
		t.Fatalf("outDir's only entry is %q, want tasks.ndjson", entries[0].Name())
	}

	info, err := os.Stat(outFile)
	if err != nil {
		t.Fatalf("Stat(%q): %v", outFile, err)
	}
	// os.CreateTemp creates its file with mode 0600, meant for a private
	// scratch file; the final output file must not inherit that, since it
	// is an ordinary file meant to be read like any other.
	if perm := info.Mode().Perm(); perm != 0o644 {
		t.Errorf("permissions = %o, want 0644 (not the restrictive mode of the temporary file it was written through)", perm)
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", outFile, err)
	}
	var line map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(content), &line); err != nil {
		t.Fatalf("output file is not valid JSON: %v\ncontent: %s", err, content)
	}
	if line["id"] != "BISO-1" {
		t.Errorf("id = %v, want BISO-1", line["id"])
	}
}

func TestRunImportStrictWithFindingsAndOutFileWritesNoFileAtAll(t *testing.T) {
	biso := findBisoBinary(t)
	project := newTestProject(t, biso)

	backlogDir := t.TempDir()
	writeMinimalBacklogTask(t, backlogDir, "TASK-1", "Odd status task", "Not A Configured Status")

	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "tasks.ndjson")

	_, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--out", outFile, "--strict")
	if code != 5 {
		t.Fatalf("got exit code %d, want 5, stderr: %s", code, errOut)
	}
	if !strings.HasPrefix(errOut, "warning: ") {
		t.Fatalf("errOut = %q, want it to start with a finding line", errOut)
	}

	entries, err := os.ReadDir(outDir)
	if err != nil {
		t.Fatalf("ReadDir(%q): %v", outDir, err)
	}
	if len(entries) != 0 {
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Fatalf("outDir has %d entries, want 0 (--strict with findings writes neither the final file nor a temporary one): %v", len(entries), names)
	}
}
