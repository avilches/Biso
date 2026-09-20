package ops

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"biso/internal/board"
	"biso/internal/model"
)

// These tests exercise the half of `biso init` that needs a board with tasks
// in it, which is the exit code 6 of --overwrite-config: a rewrite can never
// take away a value some task still carries.

// project builds a machine with a home of its own and a project inside it,
// and answers the environment a command would receive there.
func project(t *testing.T) (Env, string) {
	t.Helper()
	home := t.TempDir()
	dir := filepath.Join(home, "my-project")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	machine, err := board.LoadMachine(home)
	if err != nil {
		t.Fatal(err)
	}
	ids := []string{"3f9a2b1c", "7a1b2c3d"}
	return Env{
		Dir:     dir,
		Machine: machine,
		NewID: func() (string, error) {
			id := ids[0]
			if len(ids) > 1 {
				ids = ids[1:]
			}
			return id, nil
		},
	}.WithDefaults(), home
}

// addTask writes one task into the board the environment resolves to.
func addTask(t *testing.T, env Env, task *model.Task) {
	t.Helper()
	loc, _, err := board.Resolve(board.Search{Dir: env.Dir, Machine: env.Machine})
	if err != nil {
		t.Fatal(err)
	}
	b, openErr := board.Open(loc, env.Machine)
	if openErr != nil {
		t.Fatal(openErr)
	}
	defer b.Close()
	if err := b.Tasks.Create(task); err != nil {
		t.Fatal(err)
	}
}

func TestOverwriteConfigCannotRemoveAStatusInUse(t *testing.T) {
	env, _ := project(t)
	if _, err := Init(env, InitParams{Name: "My project", HasName: true}); err != nil {
		t.Fatal(err)
	}
	addTask(t, env, &model.Task{
		Title: "A task", Status: "In Progress", Type: "task", Priority: "high",
	})

	_, err := Init(env, InitParams{
		OverwriteConfig:  true,
		HasStatuses:      true,
		Statuses:         []string{"To Do", "Doing", "Done"},
		HasInitialStatus: true, InitialStatus: "To Do",
		HasActiveStatus: true, ActiveStatus: "Doing",
		HasTerminalStatus: true, TerminalStatus: "Done",
	})
	assertSpec(t, err, 6, "board_inconsistent")
}

func TestOverwriteConfigCannotRemoveATypeOrAPriorityInUse(t *testing.T) {
	env, _ := project(t)
	if _, err := Init(env, InitParams{Name: "My project", HasName: true}); err != nil {
		t.Fatal(err)
	}
	addTask(t, env, &model.Task{
		Title: "A task", Status: "To Do", Type: "bug", Priority: "low",
	})

	_, err := Init(env, InitParams{OverwriteConfig: true, HasTypes: true, Types: []string{"task"}})
	assertSpec(t, err, 6, "board_inconsistent")

	_, err = Init(env, InitParams{
		OverwriteConfig: true, HasPriorities: true, Priorities: []string{"high", "medium"},
	})
	assertSpec(t, err, 6, "board_inconsistent")
}

func TestOverwriteConfigCannotChangeThePrefixOfABoardWithTasks(t *testing.T) {
	env, _ := project(t)
	if _, err := Init(env, InitParams{
		Name: "My project", HasName: true, Prefix: "MYP", HasPrefix: true,
	}); err != nil {
		t.Fatal(err)
	}

	// While the board is empty the prefix moves freely.
	if _, err := Init(env, InitParams{
		OverwriteConfig: true, Prefix: "OTH", HasPrefix: true,
	}); err != nil {
		t.Fatalf("the prefix of an empty board is free to change: %v", err)
	}

	addTask(t, env, &model.Task{
		Title: "A task", Status: "To Do", Type: "task", Priority: "high",
	})
	_, err := Init(env, InitParams{OverwriteConfig: true, Prefix: "AGN", HasPrefix: true})
	assertSpec(t, err, 6, "board_inconsistent")

	// And without --prefix it is kept, so the error cannot happen at all.
	if _, err := Init(env, InitParams{OverwriteConfig: true}); err != nil {
		t.Fatalf("a rewrite that does not name the prefix keeps it: %v", err)
	}
}

func TestOverwriteConfigNeverTouchesATask(t *testing.T) {
	env, _ := project(t)
	if _, err := Init(env, InitParams{Name: "My project", HasName: true}); err != nil {
		t.Fatal(err)
	}
	addTask(t, env, &model.Task{
		Title: "A task", Status: "To Do", Type: "task", Priority: "high",
	})

	if _, err := Init(env, InitParams{
		OverwriteConfig: true, HasTypes: true, Types: []string{"task", "chore"},
	}); err != nil {
		t.Fatal(err)
	}

	result, err := Where(env)
	if err != nil {
		t.Fatal(err)
	}
	if result.Counts.NotArchived != 1 || result.Counts.HighestEverAssigned != 1 {
		t.Errorf("counts = %+v, want the one task still there", result.Counts)
	}
}

func TestInitFromNamesTheFilesAHalfSnapshotIsMissing(t *testing.T) {
	env, _ := project(t)
	_, err := Init(env, InitParams{From: t.TempDir(), HasFrom: true})
	assertSpec(t, err, 4, "file_not_found")
}

func TestWhereCountsTheArchivedApart(t *testing.T) {
	env, _ := project(t)
	if _, err := Init(env, InitParams{Name: "My project", HasName: true}); err != nil {
		t.Fatal(err)
	}
	addTask(t, env, &model.Task{Title: "Live", Status: "To Do", Type: "task", Priority: "high"})
	addTask(t, env, &model.Task{
		Title: "Gone", Status: "To Do", Type: "task", Priority: "high", Archived: true,
	})

	result, err := Where(env)
	if err != nil {
		t.Fatal(err)
	}
	if result.Counts.NotArchived != 1 || result.Counts.Archived != 1 || result.Counts.HighestEverAssigned != 2 {
		t.Errorf("counts = %+v", result.Counts)
	}
}

func assertSpec(t *testing.T, err error, code int, identifier string) {
	t.Helper()
	e := specError(t, err)
	if e.ExitCode != code || e.Code != identifier {
		t.Errorf("error = %d/%s (%s), want %d/%s", e.ExitCode, e.Code, e.Message, code, identifier)
	}
}

// specError is the *model.Error inside an error. It goes through errors.As
// and not through a type assertion because an error of the specification can
// travel inside a richer one, as the ambiguous reference of
// docs/spec/referencias.md does with its candidates.
func specError(t *testing.T, err error) *model.Error {
	t.Helper()
	if err == nil {
		t.Fatal("no error at all, and one was expected")
	}
	var e *model.Error
	if !errors.As(err, &e) {
		t.Fatalf("error = %v, want a *model.Error", err)
	}
	return e
}
