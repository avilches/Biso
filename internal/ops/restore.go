package ops

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"biso/internal/board"
	"biso/internal/model"
)

// This file is the `--from` of `biso init` (docs/spec/cmd/init.md): a board
// rebuilt whole from the three files `biso snapshot` left in its directory.
//
// It creates no format of its own. The tasks go through the same batch as
// `biso new --from` and the configuration through the same board.json as
// `biso snapshot` writes, which is what makes the restored board the board
// that was snapshotted and not a board that looks like it.

// snapshot is the three files of a snapshot directory, already read and
// already judged as far as they can be judged on their own.
type snapshot struct {
	// ID is the identity the marker carries, which is what travels with the
	// snapshot so that the pointer committed in a project keeps working
	// after a restore (docs/spec/cmd/snapshot.md#qué-entra-en-la-revisión).
	ID     string
	Config board.Config
	Tasks  string
}

// readSnapshot reads the directory `--from` names.
//
// A directory missing any of the three files is a half snapshot, and the
// error names which one is missing, because that is what tells a snapshot
// that was never taken from one that was copied incompletely.
func readSnapshot(dir string) (*snapshot, *model.Error) {
	id := board.MarkerID(dir)
	marker := "<id>.id"
	if id != "" {
		marker = board.MarkerName(id)
	}
	var missing []string
	for _, name := range []string{board.SnapshotTasksFile, board.SnapshotConfigFile, marker} {
		if id == "" && name == marker {
			missing = append(missing, marker)
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, name)); err != nil || !info.Mode().IsRegular() {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, &model.Error{
			ExitCode: 4,
			Code:     "file_not_found",
			Message: fmt.Sprintf("%s is not a snapshot: %s %s missing",
				dir, strings.Join(missing, ", "), isAre(len(missing))),
			Field: "from",
			Given: dir,
			Hints: []string{"a snapshot directory holds snapshot.ndjson, board.json and the <id>.id marker"},
		}
	}

	content, err := os.ReadFile(filepath.Join(dir, board.MarkerName(id)))
	if err != nil {
		return nil, unreadableSnapshotFile(filepath.Join(dir, board.MarkerName(id)), err)
	}
	// The marker's content is the version of the store's format and nothing
	// else, so anything else in it is a directory this program cannot read
	// as a snapshot (docs/spec/cmd/init.md).
	if string(content) != board.MarkerContent {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "invalid_snapshot_id",
			Message: fmt.Sprintf("%s does not hold %s",
				board.MarkerName(id), strings.TrimRight(board.MarkerContent, "\n")),
			Field: "from",
			Given: dir,
		}
	}

	configBytes, err := os.ReadFile(filepath.Join(dir, board.SnapshotConfigFile))
	if err != nil {
		return nil, unreadableSnapshotFile(filepath.Join(dir, board.SnapshotConfigFile), err)
	}
	cfg, cfgErr := decodeBoardConfig(configBytes)
	if cfgErr != nil {
		return nil, cfgErr
	}
	if cfgErr := checkSnapshotConfig(cfg); cfgErr != nil {
		return nil, cfgErr
	}

	tasks, err := os.ReadFile(filepath.Join(dir, board.SnapshotTasksFile))
	if err != nil {
		return nil, unreadableSnapshotFile(filepath.Join(dir, board.SnapshotTasksFile), err)
	}
	return &snapshot{ID: id, Config: cfg, Tasks: string(tasks)}, nil
}

func unreadableSnapshotFile(path string, cause error) *model.Error {
	if errors.Is(cause, fs.ErrNotExist) {
		return &model.Error{
			ExitCode: 4,
			Code:     "file_not_found",
			Message:  "file not found: " + path,
			Field:    "from",
			Given:    path,
		}
	}
	return &model.Error{
		ExitCode: 8,
		Code:     "file_unreadable",
		Message:  fmt.Sprintf("%s cannot be read: %s", path, cause),
		Field:    "from",
		Given:    path,
	}
}

// restore is `biso init --from`: it creates the board the snapshot
// describes and imports its tasks, in one invocation.
//
// The whole file is judged before anything is created, against the
// vocabulary the snapshot itself brings, which is what lets --dry-run
// answer whether the restore would work without leaving a board behind and
// what makes the exit code 7 mean that nothing was written.
func restore(env Env, p InitParams, facts board.Searched, target string) (*InitResult, error) {
	snap, snapErr := readSnapshot(p.From)
	if snapErr != nil {
		return nil, snapErr
	}
	lines, batchErr := validateBatchAgainst(snap.Config, nil, env, snap.Tasks)
	if batchErr != nil {
		return nil, batchErr
	}

	// The identity is the destination's marker when it has one, which is the
	// board being rebuilt in place, and the snapshot's otherwise: that is
	// what makes a committed pointer keep working after a restore
	// (docs/spec/cmd/init.md).
	id := snap.ID
	if target != "" {
		if destination := board.MarkerID(target); destination != "" {
			id = destination
		}
	}
	if _, err := board.FindID(env.Machine, id); err != nil {
		return nil, err
	}

	dir := target
	if dir == "" {
		dir = filepath.Join(env.Machine.BoardsRoot, FolderName(snap.Config.ProjectName, id))
	}
	// A restore always creates a new board and never rewrites one that is
	// already there, so a destination that is still a whole board is the
	// same board_exists of this command, seen from the other side
	// (docs/spec/cmd/init.md). A destination whose database cannot be read
	// is not a whole board, and that one is rebuilt in place.
	readable, readErr := board.DatabaseReadable(dir)
	if readErr != nil {
		return nil, readErr
	}
	if readable {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "board_exists",
			Message:  fmt.Sprintf("%s already holds board %s", dir, board.MarkerID(dir)),
			Hints:    []string{"--from always creates a new board, and never rewrites one that is already there"},
		}
	}

	stored := ""
	switch {
	case p.HasAt:
		stored = p.At
	case target != "":
		stored = pointerPathFor(env, p, target)
	}

	result := &InitResult{
		Action: Created, ID: id, Board: summaryOf(snap.Config), Path: dir,
		StoredPath: stored, Notes: initNotes(env, p, dir, stored),
		Restored: true, RestoredTasks: len(lines),
	}
	if p.DryRun {
		result.DryRun, result.DryRunPath = true, dryRunPath(p, dir)
		return result, nil
	}

	if err := clearUnreadableDatabase(dir); err != nil {
		return nil, err
	}
	b, err := board.Create(dir, id, snap.Config, env.Machine)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	if err := board.WriteMarker(dir, id); err != nil {
		return nil, err
	}
	if err := writeIgnoreFile(env, dir); err != nil {
		return nil, err
	}
	tasks := make([]*model.Task, 0, len(lines))
	for _, line := range lines {
		tasks = append(tasks, line.task)
	}
	if err := b.Tasks.CreateAll(tasks); err != nil {
		return nil, err
	}

	written, err := writePointer(env, facts, id, stored)
	if err != nil {
		return nil, err
	}
	result.PointerWritten = written
	return result, nil
}

func isAre(n int) string {
	if n == 1 {
		return "is"
	}
	return "are"
}
