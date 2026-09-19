package vcs

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestModeNoneRunsNothingAtAll(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	result := run(t, gitConfig(), ModeNone, board)

	if result.Outcome != OutcomeNotRequested {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeNotRequested)
	}
	if result.VCS != "none" {
		t.Fatalf("VCS = %q, want none: nothing was executed", result.VCS)
	}
	if len(result.Output) != 0 {
		t.Fatalf("Output = %v, nothing ran", result.Output)
	}
	if _, err := os.Stat(filepath.Join(board, ".git")); err == nil {
		t.Fatalf("a repository was created with --vcs none")
	}
}

func TestConfiguredNoneRunsNothingAndSaysSo(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")

	for _, mode := range []Mode{ModeCommit, ModePush} {
		result := run(t, Config{Kind: KindNone}, mode, board)

		if result.Outcome != OutcomeDisabled {
			t.Fatalf("with --vcs %s, Outcome = %q, want %q", mode, result.Outcome, OutcomeDisabled)
		}
		if result.VCS != "none" {
			t.Fatalf("with --vcs %s, VCS = %q, want none", mode, result.VCS)
		}
		if _, err := os.Stat(filepath.Join(board, ".git")); err == nil {
			t.Fatalf("a repository was created with vcs set to none")
		}
	}
}

func TestParseMode(t *testing.T) {
	for _, name := range []string{"none", "commit", "push"} {
		mode, err := ParseMode(name)
		if err != nil {
			t.Fatalf("ParseMode(%q): unexpected error %+v", name, err)
		}
		if string(mode) != name {
			t.Fatalf("ParseMode(%q) = %q", name, mode)
		}
	}

	_, err := ParseMode("Commit")
	if err == nil {
		t.Fatalf("ParseMode accepted a mode that is not in the closed domain")
	}
	if err.ExitCode != 2 || err.Code != "invalid_vcs_mode" {
		t.Fatalf("ExitCode = %d, Code = %q, want 2 and invalid_vcs_mode", err.ExitCode, err.Code)
	}
	if err.Field != "--vcs" || err.Given != "Commit" {
		t.Fatalf("Field = %q, Given = %q, want --vcs and Commit", err.Field, err.Given)
	}
	if !reflect.DeepEqual(err.Valid, []string{"none", "commit", "push"}) {
		t.Fatalf("Valid = %v, want the three modes", err.Valid)
	}
}

func TestIgnoreFile(t *testing.T) {
	if got := IgnoreFile(Config{Kind: KindGit}); got != ".gitignore" {
		t.Fatalf("IgnoreFile(git) = %q, want .gitignore", got)
	}
	if got := IgnoreFile(Config{Kind: KindNone}); got != "" {
		t.Fatalf("IgnoreFile(none) = %q, want no exclusion file", got)
	}
	cfg := Config{Kind: KindCustom, Custom: Custom{IgnoreFile: ".jjignore"}}
	if got := IgnoreFile(cfg); got != ".jjignore" {
		t.Fatalf("IgnoreFile(custom) = %q, want .jjignore", got)
	}
	if got := IgnoreFile(Config{Kind: KindCustom}); got != "" {
		t.Fatalf("IgnoreFile(custom without ignore_file) = %q, want none", got)
	}
}

func TestCustomRunsItsCommitCommandInTheBoardDirectory(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	log := filepath.Join(root, "args.txt")
	commit := script(t, root, "commit.sh", `
pwd > `+filepath.Join(root, "cwd.txt")+`
for arg in "$@"; do echo "$arg" >> `+log+`; done
if read line; then echo "stdin: $line"; else echo "stdin is closed"; fi
echo "custom says hello"
`)
	cfg := Config{Kind: KindCustom, Custom: Custom{Commit: []string{commit, "--say", "{message}", "{files}"}}}

	result := run(t, cfg, ModeCommit, board)

	if result.Outcome != OutcomeCommitted {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeCommitted)
	}
	if result.VCS != "custom" {
		t.Fatalf("VCS = %q, want custom", result.VCS)
	}
	if result.Commit != "" || result.Repository != "" {
		t.Fatalf("custom returns no revision identifier and no repository, got %q and %q", result.Commit, result.Repository)
	}
	if result.StagedOutsideBoard != 0 {
		t.Fatalf("StagedOutsideBoard = %d, custom asks the repository nothing", result.StagedOutsideBoard)
	}

	raw, err := os.ReadFile(log)
	if err != nil {
		t.Fatalf("the custom command did not run: %v", err)
	}
	args := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
	want := []string{"--say", "biso snapshot: 3 tasks", "snapshot.ndjson", "board.json", "my-board-3f9a2b1c.id"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("the arguments are %v, want %v", args, want)
	}

	cwd, err := os.ReadFile(filepath.Join(root, "cwd.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(cwd)) != board {
		t.Fatalf("the working directory was %q, want the board directory %q", strings.TrimSpace(string(cwd)), board)
	}

	joined := strings.Join(result.Output, "\n")
	if !strings.Contains(joined, "stdin is closed") {
		t.Fatalf("the standard input was not closed: %v", result.Output)
	}
	if !strings.Contains(joined, "custom says hello") {
		t.Fatalf("what the order wrote was not forwarded: %v", result.Output)
	}
}

func TestCustomCommitThatFailsIsAnError(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	commit := script(t, root, "commit.sh", "echo 'no way' >&2\nexit 3\n")
	cfg := Config{Kind: KindCustom, Custom: Custom{Commit: []string{commit}}}

	runner, err := New(cfg, ModeCommit)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	_, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a commit command that exits non-zero has to be an error")
	}
	if failure.ExitCode != 8 || failure.Code != "vcs_commit_failed" {
		t.Fatalf("ExitCode = %d, Code = %q, want 8 and vcs_commit_failed", failure.ExitCode, failure.Code)
	}
	if !reflect.DeepEqual(failure.VCSOutput, []string{"no way"}) {
		t.Fatalf("VCSOutput = %v, want the line the order wrote", failure.VCSOutput)
	}
}

func TestCustomCommandThatIsNotInstalledSkipsTheCommit(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	missing := filepath.Join(root, "there-is-no-such-program")
	cfg := Config{Kind: KindCustom, Custom: Custom{Commit: []string{missing, "{message}"}}}

	result := run(t, cfg, ModeCommit, board)

	if result.Outcome != OutcomeUnavailable {
		t.Fatalf("Outcome = %q, want %q", result.Outcome, OutcomeUnavailable)
	}
	if result.VCS != "custom" {
		t.Fatalf("VCS = %q, want custom", result.VCS)
	}
}

func TestCustomPushWithoutPublishFailsBeforeAnythingRuns(t *testing.T) {
	cfg := Config{Kind: KindCustom, Custom: Custom{Commit: []string{"/bin/true"}}}

	runner, err := New(cfg, ModePush)

	if err == nil {
		t.Fatalf("New accepted --vcs push with no publish command declared, got %+v", runner)
	}
	if err.ExitCode != 2 || err.Code != "vcs_push_unavailable" {
		t.Fatalf("ExitCode = %d, Code = %q, want 2 and vcs_push_unavailable", err.ExitCode, err.Code)
	}
	if err.Field != "--vcs" || err.Given != "push" {
		t.Fatalf("Field = %q, Given = %q, want --vcs and push", err.Field, err.Given)
	}
}

func TestCustomPushRunsThePublishCommand(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	commit := script(t, root, "commit.sh", "echo committed\n")
	publish := script(t, root, "publish.sh", "echo \"published $1\" > "+filepath.Join(root, "published.txt")+"\n")
	cfg := Config{Kind: KindCustom, Custom: Custom{
		Commit:  []string{commit, "{message}"},
		Publish: []string{publish, "{message}"},
	}}

	result := run(t, cfg, ModePush, board)

	if !result.Pushed {
		t.Fatalf("Pushed = false after a publish command that worked")
	}
	published, err := os.ReadFile(filepath.Join(root, "published.txt"))
	if err != nil {
		t.Fatalf("the publish command did not run: %v", err)
	}
	if strings.TrimSpace(string(published)) != "published biso snapshot: 3 tasks" {
		t.Fatalf("the publish command got %q", strings.TrimSpace(string(published)))
	}
}

func TestCustomPublishThatFailsIsAPushError(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	commit := script(t, root, "commit.sh", "exit 0\n")
	publish := script(t, root, "publish.sh", "echo 'the remote said no' >&2\nexit 1\n")
	cfg := Config{Kind: KindCustom, Custom: Custom{
		Commit:  []string{commit},
		Publish: []string{publish},
	}}

	runner, err := New(cfg, ModePush)
	if err != nil {
		t.Fatalf("New: unexpected error %+v", err)
	}
	_, failure := runner.Run(request(board))

	if failure == nil {
		t.Fatalf("a publish command that exits non-zero has to be an error")
	}
	if failure.ExitCode != 8 || failure.Code != "vcs_push_failed" {
		t.Fatalf("ExitCode = %d, Code = %q, want 8 and vcs_push_failed", failure.ExitCode, failure.Code)
	}
	if !reflect.DeepEqual(failure.VCSOutput, []string{"the remote said no"}) {
		t.Fatalf("VCSOutput = %v", failure.VCSOutput)
	}
}

func TestOutputKeepsBothStreamsDiscardsBlankLinesAndCountsTheLastOne(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	commit := script(t, root, "commit.sh", `
printf 'first\n\n  \nsecond\n' >&2
printf ' indented\nno newline at the end'
`)
	cfg := Config{Kind: KindCustom, Custom: Custom{Commit: []string{commit}}}

	result := run(t, cfg, ModeCommit, board)

	var standard, errors []string
	for _, line := range result.Output {
		switch line {
		case "first", "second":
			errors = append(errors, line)
		default:
			standard = append(standard, line)
		}
	}
	if !reflect.DeepEqual(errors, []string{"first", "second"}) {
		t.Fatalf("the lines of the error stream are %v, blank ones are discarded and the order inside a stream is kept", errors)
	}
	if !reflect.DeepEqual(standard, []string{" indented", "no newline at the end"}) {
		t.Fatalf("the lines of the standard stream are %v: leading blanks are kept and a last line without a newline counts", standard)
	}
}

func TestOutputAccumulatesEveryOrderInTheOrderTheyRan(t *testing.T) {
	root := tempRoot(t)
	board := newBoardDir(t, root, "board")
	commit := script(t, root, "commit.sh", "echo from-commit\n")
	publish := script(t, root, "publish.sh", "echo from-publish\n")
	cfg := Config{Kind: KindCustom, Custom: Custom{
		Commit:  []string{commit},
		Publish: []string{publish},
	}}

	result := run(t, cfg, ModePush, board)

	if !reflect.DeepEqual(result.Output, []string{"from-commit", "from-publish"}) {
		t.Fatalf("Output = %v, want one line per order in the order they ran", result.Output)
	}
}
