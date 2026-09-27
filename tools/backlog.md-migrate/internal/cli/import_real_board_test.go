package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"

	"backlog.md-migrate/internal/convert"
	"backlog.md-migrate/internal/source"
)

// This file is the integration test for TASK-70's acceptance criterion #8:
// converting and importing THIS repository's own real Backlog.md board (over
// a hundred tasks) must conserve, for a representative sample of tasks,
// their title, status, acceptance criteria with their checkmarks,
// description, plan, notes and final summary.
//
// The real backlog/ directory at the repository root is read only, never
// written, and no "backlog" CLI command ever runs against it
// (tools/backlog.md-migrate/CLAUDE.md, "Es una herramienta general": this
// project's own board is circumstantial, not a fixture to shape the tool
// around).

// realBacklogDir returns this repository's own backlog/ directory.
func realBacklogDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller could not report this file's path")
	}
	// thisFile is <repo>/tools/backlog.md-migrate/internal/cli/import_real_board_test.go.
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..")
	dir := filepath.Join(repoRoot, "backlog")
	if _, err := os.Stat(filepath.Join(dir, "tasks")); err != nil {
		t.Skipf("this repository's own backlog/tasks/ was not found at %s, cannot run the criterion #8 validation", dir)
	}
	return dir
}

// realBoardSampleTaskIDs is a deliberately varied sample of real source ids
// from this repository's own backlog/, chosen to cover a task with its own
// acceptance criteria that stay unchecked, one where they end up fully
// checked, a plan, an implementation notes history, a final summary, an
// archived task, and a title (plus one of its own criteria) that mentions
// another task's id, so the mention rewrite is exercised on real prose too,
// not only on the small synthetic board of criterion #7.
//
// All five are SIMPLE ids (never a subtask, which always gets renumbered
// regardless of the destination). The destination board this test creates
// is empty, so every simple id keeps its own number
// (docs/especificacion.md, "Identificadores", point 2): no collision, no
// reassignment. That single fact is what makes wantMentionsRewritten below
// safe to use as a stand-in for the real per-task equivalence table that
// internal/convert computes: on an empty destination, rewriting a mention
// is nothing more than swapping the source prefix for the destination one,
// the number itself never changes.
var realBoardSampleTaskIDs = []string{
	"TASK-70", // in progress; description (with several mentions) and a plan; unchecked criteria; no final summary
	"TASK-10", // done; plan, notes and a final summary; fully checked criteria
	"TASK-72", // archived; notes only (no plan, no summary); unchecked criteria
	"TASK-66", // done; its own title AND one of its criteria mention TASK-19
	"TASK-7",  // blocked; description and notes; unchecked criteria; no plan, no summary
}

// notesContainProtectedMentions lists the sample tasks whose Notes field
// talks ABOUT id mentions rather than just containing plain ones: TASK-70's
// notes are this very project's own changelog of the mention-rewriting
// algorithm, and mention, as text examples, a lowercase "task-1" (measured
// in decisiones.md as a real branch-name look-alike) and a "TASK-1.0" that
// never existed as a real subtask; TASK-10's notes mention the worktree
// "task-10-modelo" it was implemented in, the same branch-name shape
// docs/especificacion.md itself uses as its example. Both are correctly
// left untouched by the converter (docs/especificacion.md, "Identificadores",
// points 5 and 6), but wantMentionsRewritten's plain regex cannot tell a
// protected mention from an ordinary one, so for these two tasks the Notes
// comparison below trusts the conversion engine's own output (already
// covered field by field by internal/convert's tests and by this package's
// small-board test) instead of trying to reproduce the rewrite by hand.
var notesContainProtectedMentions = map[string]bool{
	"TASK-70": true,
	"TASK-10": true,
}

// mentionPattern matches a simple source id mention (TASK-<digits>) for the
// narrow purpose described on wantMentionsRewritten: it deliberately does
// not reproduce every nuance of docs/especificacion.md, "Identificadores",
// point 5 (word boundaries, subtask mentions, an id that does not exist in
// the batch), because the five sample tasks above were read beforehand to
// confirm none of their compared fields contain that kind of mention. The
// exhaustive mention-rewriting rules already have their own test matrix in
// internal/convert/identifiers_test.go and in this package's own
// TestImportSmallBoardCoversDoDCollisionRerunMentionAndBranchName; this test
// is about round-trip fidelity through a real board, not about
// re-verifying that algorithm.
var mentionPattern = regexp.MustCompile(`TASK-(\d+)`)

// wantMentionsRewritten rewrites every TASK-<n> mention in s to
// destPrefix-<n>, the number unchanged, matching what the real converter
// does for THIS test's specific destination (empty, so no id ever gets
// reassigned). See realBoardSampleTaskIDs's doc comment for why this
// simplification is safe here.
func wantMentionsRewritten(s, destPrefix string) string {
	return mentionPattern.ReplaceAllString(s, destPrefix+"-$1")
}

// bisoGetTask is the subset of "biso get <ref> --json"'s data.task object
// this test needs.
type bisoGetTask struct {
	Title              string `json:"title"`
	Status             string `json:"status"`
	Archived           bool   `json:"archived"`
	Description        string `json:"description"`
	Plan               string `json:"plan"`
	Notes              string `json:"notes"`
	Summary            string `json:"summary"`
	AcceptanceCriteria []struct {
		Key     int    `json:"key"`
		Text    string `json:"text"`
		Checked bool   `json:"checked"`
	} `json:"acceptanceCriteria"`
}

type bisoGetEnvelope struct {
	Data struct {
		Task bisoGetTask `json:"task"`
	} `json:"data"`
}

// bisoGetJSON runs "biso get <id> --json" against project and decodes its
// task object. It decodes only the leading JSON value rather than the whole
// output, because "biso get" on an archived task appends a trailing plain
// text "note: ... is archived" line after the JSON envelope.
func bisoGetJSON(t *testing.T, biso, project, id string) bisoGetTask {
	t.Helper()
	out := runBiso(t, biso, "-C", project, "--json", "get", id)
	var env bisoGetEnvelope
	if err := json.NewDecoder(strings.NewReader(out)).Decode(&env); err != nil {
		t.Fatalf("biso get %s --json: invalid JSON: %v\noutput: %s", id, err, out)
	}
	return env.Data.Task
}

// TestImportRealProjectBoardPreservesTitleStatusAcceptanceCriteriaDescriptionPlanNotesAndSummary
// is the integration test for TASK-70's acceptance criterion #8. It converts
// and imports this repository's own real backlog/ into a fresh, empty biso
// board with a vocabulary wide enough to match the real board's own
// statuses, types and priorities, then, for the sample above, checks that
// the destination (read back for real with "biso get --json") carries the
// same title, status, description, plan, notes, summary and acceptance
// criteria (with their checkmarks) that the source task had, mentions
// rewritten to the destination's own prefix.
// This test is not gated behind testing.Short(): converting and importing
// the whole real board (over a hundred tasks) takes well under a second
// (measured and logged below), so it runs as an ordinary part of
// "go test ./...", not as a separate slow suite.
func TestImportRealProjectBoardPreservesTitleStatusAcceptanceCriteriaDescriptionPlanNotesAndSummary(t *testing.T) {
	biso := findBisoBinary(t)
	backlogDir := realBacklogDir(t)

	sourceBoard, err := source.Read(backlogDir)
	if err != nil {
		t.Fatalf("source.Read(%q): %v", backlogDir, err)
	}
	sourceByID := make(map[string]source.Task, len(sourceBoard.Tasks))
	for _, task := range sourceBoard.Tasks {
		sourceByID[task.ID] = task
	}
	for _, id := range realBoardSampleTaskIDs {
		if _, ok := sourceByID[id]; !ok {
			t.Fatalf("sample task %s was not found in this repository's own backlog/; the sample needs updating", id)
		}
	}

	// A vocabulary wide enough to match nearly every real value on this
	// board, mirroring backlog/config.yml's own statuses, types and
	// priorities (statuses: Ideas, To Do, In Progress, Blocked, Done; types:
	// memory, task, bug, docs; priorities: High, Medium, Low), so a value
	// failing to match is the rare exception, not the rule.
	const destPrefix = "REAL"
	project := t.TempDir()
	board := filepath.Join(t.TempDir(), "board")
	runBiso(t, biso, "-C", project, "init", "RealBoard", "--prefix", destPrefix, "--at", board,
		"--statuses", "Ideas,To Do,In Progress,Blocked,Done",
		"--initial-status", "To Do",
		"--active-status", "In Progress",
		"--terminal-status", "Done",
		"--types", "memory,task,bug,docs",
		"--priorities", "High,Medium,Low",
	)

	outFile := filepath.Join(t.TempDir(), "real-board.ndjson")
	start := time.Now()
	_, errOut, code := run("import", backlogDir, "--project", project, "--biso", biso, "--out", outFile)
	elapsed := time.Since(start)
	t.Logf("import of the real board (%d source tasks) took %s", len(sourceBoard.Tasks), elapsed)
	if code != 0 && code != 5 {
		t.Fatalf("import of the real board: got exit code %d, want 0 or 5, stderr: %s", code, errOut)
	}

	lines := decodeNDJSON(t, outFile)
	linesByID := make(map[string]convert.Line, len(lines))
	for _, l := range lines {
		linesByID[l.ID] = l
	}

	runBiso(t, biso, "-C", project, "new", "--from", outFile, "--dry-run")
	runBiso(t, biso, "-C", project, "new", "--from", outFile)

	for _, sourceID := range realBoardSampleTaskIDs {
		sourceTask := sourceByID[sourceID]
		number := sourceNumber(t, sourceID)
		destID := fmt.Sprintf("%s-%d", destPrefix, number)

		line, ok := linesByID[destID]
		if !ok {
			t.Errorf("%s: no NDJSON line with the expected id %s (it was skipped or reassigned unexpectedly)", sourceID, destID)
			continue
		}

		got := bisoGetJSON(t, biso, project, destID)

		wantTitle := wantMentionsRewritten(sourceTask.Title, destPrefix)
		if got.Title != wantTitle {
			t.Errorf("%s (%s): destination title = %q, want %q", sourceID, destID, got.Title, wantTitle)
		}
		if got.Archived != sourceTask.Archived {
			t.Errorf("%s (%s): destination archived = %v, want %v", sourceID, destID, got.Archived, sourceTask.Archived)
		}

		// docs/especificacion.md, "Vocabularios del destino": a status that
		// matches none of the destination's declared values is omitted from
		// the line and reported as a Finding, by design, not a bug. This
		// vocabulary was chosen to cover every real status this board uses,
		// so this branch is not expected to trigger for the sample above,
		// but per this task's own instructions a mismatch here excludes only
		// the status comparison for that one task, not the whole task.
		if line.Status == "" && sourceTask.Status != "" {
			t.Logf("%s: status %q did not match the destination vocabulary and was omitted by design; skipping the status comparison for this task only", sourceID, sourceTask.Status)
		} else if got.Status != sourceTask.Status {
			t.Errorf("%s (%s): destination status = %q, want %q", sourceID, destID, got.Status, sourceTask.Status)
		}

		wantDescription := wantMentionsRewritten(sourceTask.Description, destPrefix)
		if got.Description != wantDescription {
			t.Errorf("%s (%s): destination description = %q, want %q", sourceID, destID, got.Description, wantDescription)
		}
		wantPlan := wantMentionsRewritten(sourceTask.Plan, destPrefix)
		if got.Plan != wantPlan {
			t.Errorf("%s (%s): destination plan = %q, want %q", sourceID, destID, got.Plan, wantPlan)
		}
		wantNotes := wantMentionsRewritten(sourceTask.Notes, destPrefix)
		if notesContainProtectedMentions[sourceID] {
			t.Logf("%s: notes mentions a lowercase or dotted id example that must survive untouched; comparing against what import itself computed instead of a raw source cross-check", sourceID)
			wantNotes = line.Notes
		}
		if got.Notes != wantNotes {
			t.Errorf("%s (%s): destination notes = %q, want %q", sourceID, destID, got.Notes, wantNotes)
		}
		wantSummary := wantMentionsRewritten(sourceTask.Summary, destPrefix)
		if got.Summary != wantSummary {
			t.Errorf("%s (%s): destination summary = %q, want %q", sourceID, destID, got.Summary, wantSummary)
		}

		if len(got.AcceptanceCriteria) != len(sourceTask.AcceptanceCriteria) {
			t.Errorf("%s (%s): destination has %d acceptance criteria, want %d", sourceID, destID, len(got.AcceptanceCriteria), len(sourceTask.AcceptanceCriteria))
			continue
		}
		for i, wantCriterion := range sourceTask.AcceptanceCriteria {
			gotCriterion := got.AcceptanceCriteria[i]
			wantText := wantMentionsRewritten(wantCriterion.Text, destPrefix)
			if gotCriterion.Key != wantCriterion.Number || gotCriterion.Text != wantText || gotCriterion.Checked != wantCriterion.Checked {
				t.Errorf("%s (%s): acceptance criterion %d = %+v, want {Key:%d Text:%q Checked:%v}",
					sourceID, destID, i, gotCriterion, wantCriterion.Number, wantText, wantCriterion.Checked)
			}
		}

		// Cross-check against what the conversion engine itself produced,
		// in case a difference from the source is a real destination-side
		// data loss (a biso storage or JSON round-trip bug) rather than a
		// conversion decision already covered by internal/convert's own
		// tests.
		if got.Title != line.Title {
			t.Errorf("%s (%s): destination title %q does not even match what import wrote (%q); this points at a destination-side loss, not a conversion decision", sourceID, destID, got.Title, line.Title)
		}
	}
}
