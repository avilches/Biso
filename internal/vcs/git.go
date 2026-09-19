package vcs

import (
	"path/filepath"
	"strings"

	"biso/internal/model"
)

// runGit is the whole recipe of git, the only system biso knows in version
// 1.0. It asks the repository where the revision goes instead of guessing from
// the disk, because the two situations of a board that lives inside a project
// have the same files in the same places and only an exclusion rule tells them
// apart.
func (r *Runner) runGit(req Request) (*Result, *model.Error) {
	result := &Result{VCS: string(KindGit)}
	var lines collector

	// First question: is there a repository containing this directory, and
	// where is its root? Failing to launch git at all is the system not being
	// installed, and a non-zero exit code means there is no repository.
	toplevel := ask(req.BoardDir, "git", "rev-parse", "--show-toplevel")
	if !toplevel.started {
		result.Outcome = OutcomeUnavailable
		return result, nil
	}

	board := resolve(req.BoardDir)
	repository := ""
	switch {
	case toplevel.exit != 0:
		// No repository anywhere above: the board gets its own, lazily.
		repository = ""
	default:
		root := resolve(strings.TrimSpace(firstLine(toplevel.stdout)))
		if root == board {
			repository = root
		} else if ignoresBoard(req.BoardDir) {
			// The project ignores the board, so it cannot record it: the board
			// gets its own repository.
			repository = ""
		} else {
			repository = root
		}
	}

	if repository == "" {
		created := order(&lines, req.BoardDir, "git", "init")
		result.Output = lines.take()
		if !created.ok() {
			result.Outcome = OutcomeUnavailable
			return result, nil
		}
		repository = board
	}

	// Third question: what did the index have staged outside the three files
	// of the board? Measured over the whole repository the revision goes to,
	// not only over the board directory.
	result.StagedOutsideBoard = stagedOutsideBoard(req, repository, board)

	// git commit with paths does not need them staged beforehand, but it does
	// need them known: an untracked file makes it fail with "pathspec did not
	// match any file(s) known to git", which is exactly the first snapshot of
	// a repository this command has just created. Adding the three files by
	// name changes nothing else, because naming them on git commit keeps its
	// --only mode and leaves the rest of the index alone.
	added := order(&lines, req.BoardDir, "git", append([]string{"add", "--"}, req.Files...)...)
	result.Output = lines.take()
	if !added.ok() {
		return result, commitFailed(result.Output)
	}

	// Fourth question: do the three files differ from what is already
	// recorded? It is what tells "there was nothing to commit" from a commit
	// that really failed, without reading a single message of git.
	pending := ask(req.BoardDir, "git", append([]string{"diff", "--cached", "--quiet", "--"}, req.Files...)...)
	changes := !pending.ok()

	committed := order(&lines, req.BoardDir, "git",
		append([]string{"commit", "-m", req.Message, "--"}, req.Files...)...)
	result.Output = lines.take()
	switch {
	case committed.ok():
		result.Outcome = OutcomeCommitted
		result.Repository = repository
		result.OwnRepository = repository == board
		// Fifth question: the full identifier of the revision that resulted.
		// git commit does not print it whole, and the JSON carries it whole.
		if head := ask(req.BoardDir, "git", "rev-parse", "HEAD"); head.ok() {
			result.Commit = strings.TrimSpace(firstLine(head.stdout))
		}
	case !changes:
		result.Outcome = OutcomeNothingToCommit
	default:
		return result, commitFailed(result.Output)
	}

	if r.mode == ModePush {
		published := order(&lines, req.BoardDir, "git", "push")
		result.Output = lines.take()
		if !published.ok() {
			return result, pushFailed(result.Output)
		}
		result.Pushed = true
	}
	return result, nil
}

// ignoresBoard asks git whether the repository above ignores the board's
// directory. A non-zero exit code means it does not, which is a legitimate
// answer and not a failure.
func ignoresBoard(boardDir string) bool {
	return ask(boardDir, "git", "check-ignore", boardDir).ok()
}

// stagedOutsideBoard counts the paths the index has staged that are not one of
// the three files of the board. --no-relative forces the paths to be relative
// to the root of the repository whatever the working directory and whatever
// diff.relative says in the configuration of the machine.
func stagedOutsideBoard(req Request, repository, board string) int {
	staged := ask(req.BoardDir, "git", "diff", "--cached", "--no-relative", "--name-only")
	if !staged.ok() {
		return 0
	}
	own := make(map[string]bool, len(req.Files))
	prefix := ""
	if rel, err := filepath.Rel(repository, board); err == nil && rel != "." {
		prefix = filepath.ToSlash(rel) + "/"
	}
	for _, file := range req.Files {
		own[prefix+filepath.ToSlash(file)] = true
	}

	count := 0
	for _, line := range strings.Split(staged.stdout, "\n") {
		path := strings.TrimSpace(line)
		if path == "" || own[path] {
			continue
		}
		count++
	}
	return count
}

// resolve returns the path with its symbolic links resolved, so that comparing
// what git reports against the board's directory is not defeated by a link on
// the way, like the one macOS puts in front of every temporary directory.
func resolve(path string) string {
	if path == "" {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return filepath.Clean(path)
	}
	return resolved
}

func firstLine(text string) string {
	if i := strings.IndexByte(text, '\n'); i >= 0 {
		return text[:i]
	}
	return text
}
