package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"backlog.md-migrate/internal/convert"
)

// This file is the integration test for TASK-70's acceptance criterion #7:
// a small, real Backlog.md board, generated with the actual CLI, that
// exercises a Definition of Done, an id collision, a second run over the
// same source and destination, a title mentioning another source task's id,
// and a lowercase branch-like mention, all in the SAME board, compared field
// by field against the resulting NDJSON, not just a successful exit code.

// findBacklogBinary returns the path to a real Backlog.md CLI, skipping the
// test with a clear message when it is not on PATH, mirroring
// findBisoBinary's skip-rather-than-fail behavior in import_run_test.go.
func findBacklogBinary(t *testing.T) string {
	t.Helper()
	path, err := exec.LookPath("backlog")
	if err != nil {
		t.Skip("backlog CLI not found on PATH, cannot generate a real source board for this test")
	}
	return path
}

// runBacklog runs the backlog CLI with args inside dir and fails the test,
// with its combined output, if it exits with anything other than 0.
func runBacklog(t *testing.T, backlogCLI, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(backlogCLI, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("backlog %s: %v\n%s", strings.Join(args, " "), err, out)
	}
	return string(out)
}

// backlogTaskIDPattern matches the id backlog task create --plain prints on
// its first line ("Task TASK-1 - Title").
var backlogTaskIDPattern = regexp.MustCompile(`Task (\S+) -`)

// createBacklogTask runs "backlog task create" with args inside sourceRoot
// and returns the source id Backlog.md assigned to the new task, parsed
// from its --plain output rather than assumed, so this test does not depend
// on ids starting at 1.
func createBacklogTask(t *testing.T, backlogCLI, sourceRoot string, args ...string) string {
	t.Helper()
	fullArgs := append([]string{"task", "create"}, args...)
	fullArgs = append(fullArgs, "--plain")
	out := runBacklog(t, backlogCLI, sourceRoot, fullArgs...)
	m := backlogTaskIDPattern.FindStringSubmatch(out)
	if m == nil {
		t.Fatalf("could not parse the created task id from backlog output: %s", out)
	}
	return m[1]
}

// sourceNumberPattern extracts the trailing number of a source id
// (TASK-12 -> 12).
var sourceNumberPattern = regexp.MustCompile(`-(\d+)$`)

// sourceNumber returns the numeric part of a simple source id.
func sourceNumber(t *testing.T, id string) int {
	t.Helper()
	m := sourceNumberPattern.FindStringSubmatch(id)
	if m == nil {
		t.Fatalf("id %q does not end in -<number>", id)
	}
	n, err := strconv.Atoi(m[1])
	if err != nil {
		t.Fatalf("id %q: %v", id, err)
	}
	return n
}

// decodeNDJSON reads path as NDJSON and decodes each line into a
// convert.Line, the exact struct backlog.md-migrate's own encoder writes, so
// the test compares against the real shape rather than a hand rolled one.
// An empty file decodes to a nil slice.
func decodeNDJSON(t *testing.T, path string) []convert.Line {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%q): %v", path, err)
	}
	trimmed := strings.TrimSpace(string(content))
	if trimmed == "" {
		return nil
	}
	var lines []convert.Line
	for _, raw := range strings.Split(trimmed, "\n") {
		var line convert.Line
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			t.Fatalf("invalid NDJSON line %q: %v", raw, err)
		}
		lines = append(lines, line)
	}
	return lines
}

// findLineByTitle returns the line in lines whose Title matches, failing the
// test with the full list of titles found when there is none.
func findLineByTitle(t *testing.T, lines []convert.Line, title string) convert.Line {
	t.Helper()
	for _, l := range lines {
		if l.Title == title {
			return l
		}
	}
	var titles []string
	for _, l := range lines {
		titles = append(titles, l.Title)
	}
	t.Fatalf("no line with title %q, got: %v", title, titles)
	return convert.Line{}
}

// TestImportSmallBoardCoversDoDCollisionRerunMentionAndBranchName is the
// integration test for TASK-70's acceptance criterion #7. It builds a real
// Backlog.md board with four tasks that together cover all five required
// cases in the same board:
//
//   - A task with its own acceptance criteria plus a Definition of Done.
//   - That same task's id number colliding with a task already on the
//     destination, so it must be reassigned (docs/especificacion.md,
//     "Identificadores", point 8, the "is taken on the destination,
//     reassigned to" finding).
//   - Running "backlog.md-migrate import" a second time, after applying the
//     first batch for real with "biso new --from", to check that no task is
//     duplicated.
//   - A title mentioning another source task's id, rewritten to that task's
//     final destination id.
//   - A lowercase branch-like mention ("task-10-modelo") that must be left
//     untouched, since the mention pattern only matches an uppercase source
//     prefix.
func TestImportSmallBoardCoversDoDCollisionRerunMentionAndBranchName(t *testing.T) {
	biso := findBisoBinary(t)
	backlogCLI := findBacklogBinary(t)

	// The destination: a fresh biso board with one pre-existing task of its
	// own, so its number (1) is already taken before the conversion starts.
	project := newTestProject(t, biso)
	runBiso(t, biso, "-C", project, "new", "Existing destination task", "--quiet")

	// The source: a real Backlog.md board, generated with the actual CLI
	// (docs/decisiones.md, "El formato de Backlog.md se mide con su CLI, no
	// con un tablero"), never a hand-written fixture.
	sourceRoot := t.TempDir()
	runBacklog(t, backlogCLI, sourceRoot, "init", "Fixture", "--defaults", "--no-git", "--agent-instructions", "none")
	backlogDir := filepath.Join(sourceRoot, "backlog")

	collisionID := createBacklogTask(t, backlogCLI, sourceRoot,
		"Collision source task", "--ac", "First criterion", "--ac", "Second criterion", "--dod", "Done item one")
	contextID := createBacklogTask(t, backlogCLI, sourceRoot, "Context task")
	followUpSourceTitle := fmt.Sprintf("Follow-up of %s", contextID)
	followUpID := createBacklogTask(t, backlogCLI, sourceRoot, followUpSourceTitle)
	createBacklogTask(t, backlogCLI, sourceRoot,
		"Branch mention task", "-d", "See branch task-10-modelo for context.")

	contextNumber := sourceNumber(t, contextID)
	followUpNumber := sourceNumber(t, followUpID)

	// --- First run: converts all four tasks. The collision is itself a
	// Finding, so the run exits 5, but docs/especificacion.md, "Códigos de
	// salida", row 5, says the NDJSON is still written (no --strict). ---
	firstOut := filepath.Join(t.TempDir(), "first.ndjson")
	_, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--out", firstOut)
	if code != 5 {
		t.Fatalf("first run: got exit code %d, want 5, stderr: %s", code, errOut)
	}

	firstLines := decodeNDJSON(t, firstOut)
	if len(firstLines) != 4 {
		t.Fatalf("first run: got %d lines, want 4", len(firstLines))
	}

	// Definition of Done alongside the task's own acceptance criteria, and
	// the id collision, both on the same task.
	collisionLine := findLineByTitle(t, firstLines, "Collision source task")
	if collisionLine.ID == "BISO-1" {
		t.Fatalf("collision task kept id BISO-1, want it reassigned (BISO-1 is the pre-existing destination task)")
	}
	wantAC := []convert.AcceptanceCriterion{
		{Key: 1, Text: "First criterion", Checked: false},
		{Key: 2, Text: "Second criterion", Checked: false},
		{Key: 3, Text: "Done item one #dod", Checked: false},
	}
	if !reflect.DeepEqual(collisionLine.AcceptanceCriteria, wantAC) {
		t.Errorf("collision task AcceptanceCriteria = %+v, want %+v", collisionLine.AcceptanceCriteria, wantAC)
	}
	wantReassignFinding := fmt.Sprintf("%s: id BISO-1 is taken on the destination, reassigned to %s", collisionID, collisionLine.ID)
	if !strings.Contains(errOut, wantReassignFinding) {
		t.Errorf("errOut = %q, want it to contain %q", errOut, wantReassignFinding)
	}

	// The context task's own number was free on the destination, so it
	// keeps it (docs/especificacion.md, "Identificadores", point 2).
	contextLine := findLineByTitle(t, firstLines, "Context task")
	wantContextID := fmt.Sprintf("BISO-%d", contextNumber)
	if contextLine.ID != wantContextID {
		t.Errorf("context task id = %q, want %q (its number was free on the destination)", contextLine.ID, wantContextID)
	}

	// The mention of the context task's source id in another task's title
	// is rewritten to the context task's final destination id (point 5).
	wantFollowUpTitle := fmt.Sprintf("Follow-up of %s", wantContextID)
	followUpLine := findLineByTitle(t, firstLines, wantFollowUpTitle)
	wantFollowUpID := fmt.Sprintf("BISO-%d", followUpNumber)
	if followUpLine.ID != wantFollowUpID {
		t.Errorf("follow-up task id = %q, want %q (its own number was free)", followUpLine.ID, wantFollowUpID)
	}

	// A lowercase branch-like mention is never a real mention: the pattern
	// only matches the source prefix in uppercase (point 5's last
	// paragraph), so it must survive untouched.
	branchLine := findLineByTitle(t, firstLines, "Branch mention task")
	if branchLine.Description != "See branch task-10-modelo for context." {
		t.Errorf("branch mention task Description = %q, want the branch-like text left untouched", branchLine.Description)
	}

	// --- Apply the first batch for real, so the destination actually has
	// these four tasks (not just an unapplied NDJSON) before the second run,
	// as the criterion requires: "con bin/biso new --from real entre
	// medias". ---
	runBiso(t, biso, "-C", project, "new", "--from", firstOut, "--dry-run")
	runBiso(t, biso, "-C", project, "new", "--from", firstOut)

	// --- Second run against the same source and destination: every source
	// task was already imported, so each one must be skipped, not
	// duplicated, and the second NDJSON must carry no line for any of
	// them. ---
	secondOut := filepath.Join(t.TempDir(), "second.ndjson")
	_, errOut2, code2 := run("import", backlogDir, "--project", project, "--biso", biso, "--out", secondOut)
	if code2 != 5 {
		t.Fatalf("second run: got exit code %d, want 5, stderr: %s", code2, errOut2)
	}
	secondLines := decodeNDJSON(t, secondOut)
	if len(secondLines) != 0 {
		t.Fatalf("second run: got %d lines, want 0 (every source task was already imported), lines: %+v", len(secondLines), secondLines)
	}
	if got := strings.Count(errOut2, "already on the destination, skipped"); got != 4 {
		t.Errorf("second run: %d \"already on the destination, skipped\" findings, want 4, stderr: %s", got, errOut2)
	}

	// No task was duplicated on the destination: the pre-existing one plus
	// the four imported ones, still five in total.
	exported := strings.TrimSpace(runBiso(t, biso, "-C", project, "export", "--out", "-"))
	if got := strings.Count(exported, "\n") + 1; got != 5 {
		t.Errorf("destination has %d tasks after the second run, want 5 (no duplicates)", got)
	}
}
