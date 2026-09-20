package ops

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the harness the tests of the two writing commands run
// through: a board of its own in a temporary directory, with the default
// vocabulary of docs/spec/cmd/init.md, a clock that says the same thing
// twice, and an identity.
//
// They drive NewOn and SetOn over an open board rather than New and Set,
// which is the whole point of the logic taking the board as its first
// argument: nothing here has to resolve a pointer or read a machine
// configuration to exercise a rule about a field.

// writeClock is the instant every one of these tests happens at.
var writeClock = time.Date(2026, 9, 6, 9, 12, 4, 0, time.UTC)

type harness struct {
	t   *testing.T
	b   *board.Board
	env Env
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "my-project-3f9a2b1c")
	cfg := board.DefaultConfig("My project", "MYP")
	cfg.Extensions = []string{"trello.card"}
	b, err := board.Create(dir, "3f9a2b1c", cfg, board.Machine{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { b.Close() })
	return &harness{
		t: t, b: b,
		env: Env{
			Dir: dir, Me: "@claude",
			Now: func() time.Time { return writeClock },
		}.WithDefaults(),
	}
}

// as answers the same harness seen by another identity, which is what the
// rules of docs/spec/lease.md need to be told apart.
func (h *harness) as(me string) *harness {
	clone := *h
	clone.env.Me = me
	return &clone
}

// create runs `biso new` and answers the identifier it minted.
func (h *harness) create(title string, changes ...Change) string {
	h.t.Helper()
	result, err := NewOn(h.b, h.env, NewParams{Title: title, HasTitle: true, Changes: changes})
	if err != nil {
		h.t.Fatalf("biso new %q: %v", title, err)
	}
	return result.Tasks[0].ID
}

// set runs `biso set` and fails the test when it does not succeed.
func (h *harness) set(ref string, changes ...Change) *WriteResult {
	h.t.Helper()
	result, err := SetOn(h.b, h.env, SetParams{Refs: []string{ref}, Changes: changes})
	if err != nil {
		h.t.Fatalf("biso set %s: %v", ref, err)
	}
	return result
}

// load reads a task back out of the database, which is the only way to check
// that a write really happened and not just that a function answered.
func (h *harness) load(id string) *model.Task {
	h.t.Helper()
	task, err := h.b.Tasks.Load(id)
	if err != nil {
		h.t.Fatalf("loading %s: %v", id, err)
	}
	return task
}

// The four builders of a change, in the four shapes the specification has.
func add(flag, value string) Change {
	return Change{Flag: flag, Step: StepAdd, Value: value}
}

func clear(flag string) Change { return Change{Flag: flag, Step: StepClear} }

func remove(flag, value string) Change {
	return Change{Flag: flag, Step: StepRemove, Value: value}
}

func replace(flag, value string) Change {
	return Change{Flag: flag, Step: StepReplace, Value: value}
}

func scalar(flag, value string) Change {
	return Change{Flag: flag, Step: StepScalar, Value: value}
}

func ext(key, value string) Change {
	return Change{Flag: "ext", Step: StepExt, Key: key, Value: value}
}

func check(flag, selector string) Change {
	return Change{Flag: flag, Step: StepCheckAC, Value: selector}
}

func comment(text string) Change {
	return Change{Flag: "comment", Step: StepComment, Value: text}
}

func commentDate(selector, instant string) Change {
	return Change{Flag: "set-comment-date", Step: StepCommentDate, Key: selector, Value: instant}
}

// warned answers whether the result carries a warning with that code, which
// is what a caller branches on, never the prose.
func warned(r *WriteResult, code string) bool {
	for _, w := range r.Warnings {
		if w.Code == code {
			return true
		}
	}
	return false
}

func noted(r *WriteResult, fragment string) bool {
	for _, note := range r.Notes {
		if strings.Contains(note, fragment) {
			return true
		}
	}
	return false
}

func assertLabels(t *testing.T, task *model.Task, want ...string) {
	t.Helper()
	if strings.Join(task.Labels, "|") != strings.Join(want, "|") {
		t.Errorf("labels = %v, want %v", task.Labels, want)
	}
}
