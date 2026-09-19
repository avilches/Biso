package vcs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// run builds the runner and executes it, failing the test on an unexpected error.
func run(t *testing.T, cfg Config, mode Mode, dir string) *Result {
	t.Helper()
	runner, err := New(cfg, mode)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	result, err := runner.Run(request(dir))
	if err != nil {
		t.Fatalf("Run: unexpected error %+v", err)
	}
	return result
}

func TestGitCreatesItsOwnRepositoryWhenThereIsNone(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	if result.VCS != "git" {
		t.Fatalf("VCS = %q, want git", result.VCS)
	}
	if !result.OwnRepository {
		t.Fatalf("OwnRepository = false, want true")
	}
	if result.Repository != board {
		t.Fatalf("Repository = %q, want %q", result.Repository, board)
	}
	if len(result.Commit) != 40 {
		t.Fatalf("Commit = %q, want a full identifier of 40 characters", result.Commit)
	}
	if result.Commit != git(t, board, "rev-parse", "HEAD") {
		t.Fatalf("Commit = %q, does not match HEAD", result.Commit)
	}
	if result.Pushed {
		t.Fatalf("Pushed = true without --vcs push")
	}
	if result.StagedOutsideBoard != 0 {
		t.Fatalf("StagedOutsideBoard = %d, want 0", result.StagedOutsideBoard)
	}

	committed := strings.Fields(git(t, board, "show", "--name-only", "--format=", "HEAD"))
	if len(committed) != 3 {
		t.Fatalf("the revision has %d files, want the three of the board: %v", len(committed), committed)
	}
	for _, name := range threeFiles {
		if !strings.Contains(strings.Join(committed, " "), name) {
			t.Fatalf("%s is not in the revision: %v", name, committed)
		}
	}
	if got := git(t, board, "log", "-1", "--format=%s"); got != "biso snapshot: 3 tasks" {
		t.Fatalf("the message is %q, want the one biso composes", got)
	}
	if len(result.Output) == 0 {
		t.Fatalf("Output is empty, the orders that act are forwarded")
	}
}

func TestGitCommitsIntoTheBoardOwnRepositoryWhenItAlreadyIsOne(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	git(t, board, "init")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted || !result.OwnRepository {
		t.Fatalf("Outcome = %q, OwnRepository = %v, want a commit in the board's own repository", result.Outcome, result.OwnRepository)
	}
	if result.Repository != board {
		t.Fatalf("Repository = %q, want %q", result.Repository, board)
	}
}

func TestGitCommitsIntoTheProjectRepositoryThatDoesNotIgnoreTheBoard(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")
	if err := os.WriteFile(filepath.Join(project, "code.txt"), []byte("code\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, project, "add", "code.txt")
	git(t, project, "commit", "-m", "first")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	if result.OwnRepository {
		t.Fatalf("OwnRepository = true, the revision had to go to the project's repository")
	}
	if result.Repository != project {
		t.Fatalf("Repository = %q, want the project root %q", result.Repository, project)
	}
	if _, err := os.Stat(filepath.Join(board, ".git")); err == nil {
		t.Fatalf("a nested repository was created inside the board directory")
	}
	if result.Commit != git(t, project, "rev-parse", "HEAD") {
		t.Fatalf("Commit = %q, does not match the project's HEAD", result.Commit)
	}
	committed := git(t, project, "show", "--name-only", "--format=", "HEAD")
	for _, name := range threeFiles {
		if !strings.Contains(committed, ".biso-board/"+name) {
			t.Fatalf("%s is not in the revision, named by its path: %q", name, committed)
		}
	}
}

func TestGitCreatesItsOwnRepositoryWhenTheProjectIgnoresTheBoard(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte(".biso-board/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, project, "add", ".gitignore")
	git(t, project, "commit", "-m", "first")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted || !result.OwnRepository {
		t.Fatalf("Outcome = %q, OwnRepository = %v, want the board's own repository", result.Outcome, result.OwnRepository)
	}
	if result.Repository != board {
		t.Fatalf("Repository = %q, want %q", result.Repository, board)
	}
	if _, err := os.Stat(filepath.Join(board, ".git")); err != nil {
		t.Fatalf("the board's own repository was not created: %v", err)
	}
	if got := git(t, project, "rev-parse", "HEAD"); got == result.Commit {
		t.Fatalf("the revision went to the project's repository, which ignores the board")
	}
}

func TestGitSecondSnapshotWithoutChangesHasNothingToCommit(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	first := run(t, gitConfig(), ModeCommit, board)

	second := run(t, gitConfig(), ModeCommit, board)

	if second.Outcome != OutcomeNothingToCommit {
		t.Fatalf("Outcome = %q, want %q", second.Outcome, OutcomeNothingToCommit)
	}
	if second.Commit != "" || second.Repository != "" {
		t.Fatalf("Commit = %q and Repository = %q, both empty when nothing was committed", second.Commit, second.Repository)
	}
	if len(second.Output) == 0 {
		t.Fatalf("what git commit wrote on its own is forwarded even when there is nothing to commit")
	}
	if got := git(t, board, "rev-parse", "HEAD"); got != first.Commit {
		t.Fatalf("HEAD moved to %q, no second revision was expected", got)
	}
}

func TestGitCountsWhatTheIndexHadStagedOutsideTheBoardAndLeavesItAlone(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")
	for _, name := range []string{"one.txt", "two.txt"} {
		if err := os.WriteFile(filepath.Join(project, name), []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	git(t, project, "add", "one.txt", "two.txt")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.StagedOutsideBoard != 2 {
		t.Fatalf("StagedOutsideBoard = %d, want 2", result.StagedOutsideBoard)
	}
	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, the revision is saved anyway", result.Outcome)
	}
	committed := git(t, project, "show", "--name-only", "--format=", "HEAD")
	if strings.Contains(committed, "one.txt") || strings.Contains(committed, "two.txt") {
		t.Fatalf("the revision swallowed what was staged outside the board: %q", committed)
	}
	staged := git(t, project, "diff", "--cached", "--name-only")
	if !strings.Contains(staged, "one.txt") || !strings.Contains(staged, "two.txt") {
		t.Fatalf("what was staged outside the board is no longer staged: %q", staged)
	}
}

func TestGitDoesNotCountTheBoardFilesAsStagedOutside(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")
	git(t, project, "add", ".biso-board")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.StagedOutsideBoard != 0 {
		t.Fatalf("StagedOutsideBoard = %d, the board's own three files are never outside", result.StagedOutsideBoard)
	}
}

func TestGitCommitThatFailsForAnEnvironmentReasonIsAnError(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	git(t, board, "init")
	hooks := filepath.Join(board, ".git", "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	script(t, hooks, "pre-commit", "echo 'the hook says no' >&2\nexit 1\n")

	runner, err := New(gitConfig(), ModeCommit)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	result, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a commit rejected by a hook has to be an error, got %+v", result)
	}
	if failure.ExitCode != 8 || failure.Code != "vcs_commit_failed" {
		t.Fatalf("ExitCode = %d, Code = %q, want 8 and vcs_commit_failed", failure.ExitCode, failure.Code)
	}
	if len(failure.VCSOutput) == 0 {
		t.Fatalf("VCSOutput is empty, it carries the lines the order that failed wrote")
	}
	if !strings.Contains(strings.Join(failure.VCSOutput, "\n"), "the hook says no") {
		t.Fatalf("VCSOutput does not carry what the hook wrote: %v", failure.VCSOutput)
	}
	if failure.Message == "" {
		t.Fatalf("the message is empty, it has to say that nothing was lost")
	}
}

func TestGitNotInstalledSkipsTheCommitWithoutFailing(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	t.Setenv("PATH", "")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeUnavailable {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeUnavailable)
	}
	if result.VCS != "git" {
		t.Fatalf("VCS = %q, want git: it is the system this call was asked for", result.VCS)
	}
	if _, err := os.Stat(filepath.Join(board, ".git")); err == nil {
		t.Fatalf("a repository was created without git")
	}
}

func TestGitPushPublishesTheRevision(t *testing.T) {
	root := tempRoot(t)
	remote := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, remote, "init", "--bare")
	board := newBoardDir(t, root, "board")
	git(t, board, "init")
	git(t, board, "remote", "add", "origin", remote)
	git(t, board, "config", "branch.main.remote", "origin")
	git(t, board, "config", "branch.main.merge", "refs/heads/main")

	result := run(t, gitConfig(), ModePush, board)

	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	if !result.Pushed {
		t.Fatalf("Pushed = false after a --vcs push that worked")
	}
	if got := git(t, remote, "rev-parse", "main"); got != result.Commit {
		t.Fatalf("the remote is at %q, want %q", got, result.Commit)
	}
}

func TestGitPushThatFailsIsAnErrorAfterTheCommit(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	runner, err := New(gitConfig(), ModePush)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	_, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a push with no destination configured has to fail")
	}
	if failure.ExitCode != 8 || failure.Code != "vcs_push_failed" {
		t.Fatalf("ExitCode = %d, Code = %q, want 8 and vcs_push_failed", failure.ExitCode, failure.Code)
	}
	if len(failure.VCSOutput) == 0 {
		t.Fatalf("VCSOutput is empty, it carries what git push wrote")
	}
	if got := gitStatus(t, board, "rev-parse", "HEAD"); got != 0 {
		t.Fatalf("the revision was not saved before trying to publish")
	}
}

func TestGitRunsEveryOrderInTheBoardDirectory(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	// The three files are named as paths relative to the board directory, so
	// finding them under .biso-board/ in the project's revision proves that the
	// working directory of the orders was the board directory and not the root.
	committed := git(t, project, "show", "--name-only", "--format=", "HEAD")
	if !strings.Contains(committed, ".biso-board/board.json") {
		t.Fatalf("the revision does not name the files by their path: %q", committed)
	}
}

func TestGitNeverForwardsWhatTheQuestionsWrite(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	result := run(t, gitConfig(), ModeCommit, board)

	joined := strings.Join(result.Output, "\n")
	if strings.Contains(joined, "not a git repository") {
		t.Fatalf("the alarming answer of a question reached the output: %v", result.Output)
	}
}

func TestGitPushPublishesEvenWhenThereWasNothingToCommit(t *testing.T) {
	root := tempRoot(t)
	remote := filepath.Join(root, "remote.git")
	if err := os.MkdirAll(remote, 0o755); err != nil {
		t.Fatal(err)
	}
	git(t, remote, "init", "--bare")
	board := newBoardDir(t, root, "board")
	git(t, board, "init")
	git(t, board, "remote", "add", "origin", remote)
	git(t, board, "config", "branch.main.remote", "origin")
	git(t, board, "config", "branch.main.merge", "refs/heads/main")
	run(t, gitConfig(), ModePush, board)

	result := run(t, gitConfig(), ModePush, board)

	if result.Outcome != OutcomeNothingToCommit {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeNothingToCommit)
	}
	if !result.Pushed {
		t.Fatalf("Pushed = false: what --vcs push promises does not depend on this call having recorded a revision")
	}
}

func TestGitPushErrorCarriesOnlyTheLinesOfTheOrderThatFailed(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	runner, err := New(gitConfig(), ModePush)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	result, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a push with no destination configured has to fail")
	}
	// The commit did work and wrote its own summary line, which the result
	// carries and the error must not: the error is about git push alone.
	commitLine := "biso snapshot: 3 tasks"
	if !strings.Contains(strings.Join(result.Output, "\n"), commitLine) {
		t.Fatalf("Output = %v, want every line of every order that ran, the commit included", result.Output)
	}
	for _, line := range failure.VCSOutput {
		if strings.Contains(line, commitLine) {
			t.Fatalf("VCSOutput = %v, it carries the lines of the order that failed and of no other one", failure.VCSOutput)
		}
	}
	if len(failure.VCSOutput) == 0 {
		t.Fatalf("VCSOutput is empty, it carries what git push wrote")
	}
	if len(failure.VCSOutput) >= len(result.Output) {
		t.Fatalf("VCSOutput = %v is not shorter than Output = %v, so it is the accumulated output", failure.VCSOutput, result.Output)
	}
}

func TestGitCommitErrorCarriesOnlyTheLinesOfTheOrderThatFailed(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	hooks := filepath.Join(root, "hooks")
	if err := os.MkdirAll(hooks, 0o755); err != nil {
		t.Fatal(err)
	}
	script(t, hooks, "pre-commit", "echo 'the hook says no' >&2\nexit 1\n")
	// The hooks belong to a repository that does not exist yet: git init is
	// the first order of this run, and the one whose line must not end up in
	// the error of the commit that fails afterwards.
	t.Setenv("GIT_CONFIG_COUNT", "1")
	t.Setenv("GIT_CONFIG_KEY_0", "core.hooksPath")
	t.Setenv("GIT_CONFIG_VALUE_0", hooks)

	runner, err := New(gitConfig(), ModeCommit)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	result, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a commit rejected by a hook has to be an error, got %+v", result)
	}
	joined := strings.Join(failure.VCSOutput, "\n")
	if !strings.Contains(joined, "the hook says no") {
		t.Fatalf("VCSOutput does not carry what the order that failed wrote: %v", failure.VCSOutput)
	}
	if strings.Contains(joined, "Initialized empty") {
		t.Fatalf("VCSOutput = %v carries a line of git init, an earlier order", failure.VCSOutput)
	}
	if !strings.Contains(strings.Join(result.Output, "\n"), "Initialized empty") {
		t.Fatalf("Output = %v, want every line of every order that ran, git init included", result.Output)
	}
}

func TestGitDoesNotCountTheBoardFilesAsStagedOutsideWithANonAsciiPath(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	// A board directory with an accent in its name: git quotes it and escapes
	// it in octal unless the question asks for the raw bytes.
	board := newBoardDir(t, project, ".tablero-cañón")
	git(t, project, "init")
	if err := os.WriteFile(filepath.Join(project, "año.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The three files of the board have to be in the index before the count,
	// which is the question this test is about: with the paths quoted they
	// would not match the board's own three and would be counted as outside.
	git(t, project, "add", "año.txt", ".tablero-cañón")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.StagedOutsideBoard != 1 {
		t.Fatalf("StagedOutsideBoard = %d, want 1: only the file outside the board counts, and its own three never do", result.StagedOutsideBoard)
	}
	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	committed := git(t, project, "show", "--name-only", "--format=", "HEAD")
	if !strings.Contains(committed, "snapshot.ndjson") {
		t.Fatalf("the revision does not carry the board files: %q", committed)
	}
}

func TestGitCountsWhatIsStagedOutsideTheBoardWithDiffRelativeOn(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	board := newBoardDir(t, project, ".biso-board")
	git(t, project, "init")
	// diff.relative makes git list only what hangs below the working
	// directory, and the working directory of every question is the board:
	// without --no-relative the file outside would be invisible and the count
	// would be a silent zero.
	git(t, project, "config", "diff.relative", "true")
	if err := os.WriteFile(filepath.Join(project, "one.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, project, "add", "one.txt")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.StagedOutsideBoard != 1 {
		t.Fatalf("StagedOutsideBoard = %d, want 1 with diff.relative on", result.StagedOutsideBoard)
	}
}

func TestGitAsksAboutABoardDirectoryWhoseNameLooksLikeAPattern(t *testing.T) {
	root := tempRoot(t)
	project := filepath.Join(root, "project")
	// Glob characters in the name of the board directory: the question about
	// whether the project ignores it hands git a path, never a pattern.
	board := newBoardDir(t, project, "bo[a]rd*")
	git(t, project, "init")
	if err := os.WriteFile(filepath.Join(project, ".gitignore"), []byte("bo\\[a\\]rd\\*/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	git(t, project, "add", ".gitignore")
	git(t, project, "commit", "-m", "first")

	result := run(t, gitConfig(), ModeCommit, board)

	if result.Outcome != OutcomeCommitted || !result.OwnRepository {
		t.Fatalf("Outcome = %q, OwnRepository = %v, want the board's own repository: the project ignores it", result.Outcome, result.OwnRepository)
	}
	if _, err := os.Stat(filepath.Join(board, ".git")); err != nil {
		t.Fatalf("the board's own repository was not created: %v", err)
	}
}
