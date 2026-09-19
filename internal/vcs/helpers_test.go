package vcs

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain isolates every test from the machine's git configuration and from
// the identity of whoever runs the suite. Without this, a global gitconfig
// with hooks, a signing key, a different default branch name or no identity at
// all would change what the recipe observes, and the tests would be measuring
// the machine instead of the code.
func TestMain(m *testing.M) {
	home, err := os.MkdirTemp("", "biso-vcs-home")
	if err != nil {
		panic(err)
	}
	config := filepath.Join(home, "gitconfig")
	err = os.WriteFile(config, []byte("[init]\n\tdefaultBranch = main\n"), 0o644)
	if err != nil {
		panic(err)
	}

	for key, value := range map[string]string{
		"HOME":                home,
		"XDG_CONFIG_HOME":     filepath.Join(home, "config"),
		"GIT_CONFIG_GLOBAL":   config,
		"GIT_CONFIG_SYSTEM":   os.DevNull,
		"GIT_CONFIG_NOSYSTEM": "1",
		"GIT_TERMINAL_PROMPT": "0",
		"GIT_AUTHOR_NAME":     "Biso Test",
		"GIT_AUTHOR_EMAIL":    "test@example.invalid",
		"GIT_COMMITTER_NAME":  "Biso Test",
		"GIT_COMMITTER_EMAIL": "test@example.invalid",
		"GIT_PAGER":           "cat",
	} {
		if err := os.Setenv(key, value); err != nil {
			panic(err)
		}
	}
	for _, key := range []string{"GIT_DIR", "GIT_WORK_TREE", "GIT_INDEX_FILE", "GIT_COMMON_DIR", "GIT_OBJECT_DIRECTORY"} {
		if err := os.Unsetenv(key); err != nil {
			panic(err)
		}
	}

	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}

// The three files that go into every revision, in the order the specification
// fixes: snapshot.ndjson, board.json and the <id>.id marker.
var threeFiles = []string{"snapshot.ndjson", "board.json", "my-board-3f9a2b1c.id"}

// tempRoot returns a temporary directory with its symbolic links resolved, so
// that comparing it against what git reports never fails for the /var to
// /private/var link that macOS puts in front of every temporary directory.
func tempRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatalf("resolving the temporary directory: %v", err)
	}
	return root
}

// git runs a git command in dir and fails the test if it does not succeed.
func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s in %s: %v\n%s", strings.Join(args, " "), dir, err, out)
	}
	return strings.TrimSpace(string(out))
}

// gitStatus runs a git command and returns its exit code instead of failing.
func gitStatus(t *testing.T, dir string, args ...string) int {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		t.Fatalf("git %s in %s: %v", strings.Join(args, " "), dir, err)
	}
	return 0
}

// writeBoardFiles writes the three files of a board into dir, each one with
// the given content suffix so that a second call produces a real change.
func writeBoardFiles(t *testing.T, dir, content string) {
	t.Helper()
	for _, name := range threeFiles {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, []byte(name+" "+content+"\n"), 0o644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}
}

// newBoardDir creates a board directory with its three files already written.
func newBoardDir(t *testing.T, parent, name string) string {
	t.Helper()
	dir := filepath.Join(parent, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating %s: %v", dir, err)
	}
	writeBoardFiles(t, dir, "first")
	return dir
}

// gitConfig is the configuration of a machine whose vcs key is git.
func gitConfig() Config { return Config{Kind: KindGit} }

// request is the request ops would build for a board directory.
func request(dir string) Request {
	return Request{BoardDir: dir, Files: threeFiles, Message: "biso snapshot: 3 tasks"}
}

// script writes an executable shell script in dir and returns its path. The
// custom system needs a real external program, not a double.
func script(t *testing.T, dir, name, body string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}
