package ops

import (
	"fmt"
	"os"
	"path/filepath"

	"biso/internal/board"
	"biso/internal/model"
	"biso/internal/vcs"
)

// This file is docs/spec/cmd/snapshot.md: the two files a board writes for
// itself, and the revision that records them. The recipe of the version
// control system is not here, it is internal/vcs; what is here is the order
// of the two writes, which is what the specification promises about them.

// DefaultVCSMode is what `biso snapshot` does when --vcs does not say:
// it writes the two files and records a revision
// (docs/spec/cmd/snapshot.md).
const DefaultVCSMode = VCSMode(vcs.ModeCommit)

// VCSMode is what --vcs asked for, named here so that internal/cli can read
// the flag without importing internal/vcs, which the dependency rule keeps
// out of its reach.
type VCSMode = vcs.Mode

// ParseVCSMode reads the value of --vcs, or answers the usage error of a
// value outside its closed domain.
func ParseVCSMode(value string) (VCSMode, *model.Error) {
	return vcs.ParseMode(value)
}

// SnapshotParams is one `biso snapshot` call.
type SnapshotParams struct {
	// Mode is what --vcs asked for in this call, which never changes the
	// system the machine is configured with.
	Mode VCSMode
}

// SnapshotResult is what `biso snapshot` answers, in the shape of the
// `snapshot` envelope of docs/spec/cmd/snapshot.md#el-esquema-json.
type SnapshotResult struct {
	Tasks int
	Files []string
	VCS   string
	// Committed, Commit and Repository describe the revision. Commit and
	// Repository are empty when there was none, and Commit is empty with
	// custom, which returns no identifier.
	Committed     bool
	Commit        string
	Repository    string
	OwnRepository bool
	Pushed        bool

	VCSOutput          []string
	StagedOutsideBoard int
	Skipped            []string

	Warnings []Warning
	Notes    []string
}

// Snapshot writes snapshot.ndjson and board.json into the board's own
// directory and records them (docs/spec/cmd/snapshot.md).
func Snapshot(env Env, p SnapshotParams) (*SnapshotResult, error) {
	b, err := openBoard(env)
	if err != nil {
		return nil, err
	}
	defer b.Close()
	return SnapshotOn(b, env, p)
}

// SnapshotOn is Snapshot over a board that is already open.
func SnapshotOn(b *board.Board, env Env, p SnapshotParams) (*SnapshotResult, error) {
	// The one failure that has to surface before any file is written is a
	// custom system missing the order this call needs, so the runner is
	// built first (docs/spec/cmd/snapshot.md).
	runner, verr := vcs.New(VCSConfig(env.Machine), p.Mode)
	if verr != nil {
		return nil, verr
	}

	dump, err := ExportOn(b, env, ExportParams{})
	if err != nil {
		return nil, err
	}
	config, err := encodeBoardConfig(b.Config)
	if err != nil {
		return nil, &model.Error{
			ExitCode: 8, Code: "io_error",
			Message: "board.json cannot be built: " + err.Error(),
		}
	}

	dir := b.Location.Dir
	if err := writeSnapshotFiles(dir, dump.NDJSON, config); err != nil {
		return nil, err
	}

	result := &SnapshotResult{
		Tasks:    dump.Tasks,
		Files:    []string{board.SnapshotTasksFile, board.SnapshotConfigFile},
		VCS:      string(vcs.KindNone),
		Skipped:  dump.Skipped,
		Warnings: dump.Warnings,
	}

	outcome, verr := runner.Run(vcs.Request{
		BoardDir: dir,
		// The three files of the revision are the two this command writes
		// and the marker, which is what makes the identity travel with the
		// snapshot (docs/spec/cmd/snapshot.md#qué-entra-en-la-revisión).
		Files: []string{
			board.SnapshotTasksFile,
			board.SnapshotConfigFile,
			board.MarkerName(b.Location.ID),
		},
		Message: fmt.Sprintf("biso snapshot: %s", plural(dump.Tasks, "task")),
	})
	if outcome != nil {
		result.VCS = outcome.VCS
		result.VCSOutput = outcome.Output
		result.StagedOutsideBoard = outcome.StagedOutsideBoard
		result.Commit = outcome.Commit
		result.Repository = outcome.Repository
		result.OwnRepository = outcome.OwnRepository
		result.Committed = outcome.Outcome == vcs.OutcomeCommitted
		result.Pushed = outcome.Pushed
		result.Notes = append(result.Notes, snapshotNotes(outcome)...)
	}
	if verr != nil {
		return result, verr
	}
	return result, nil
}

// snapshotNotes are the notes each ending of the recipe earns, in the order
// docs/spec/cmd/snapshot.md prints them: what happened with the revision
// first, and then what was left staged outside the board.
func snapshotNotes(outcome *vcs.Result) []string {
	var notes []string
	switch outcome.Outcome {
	case vcs.OutcomeDisabled:
		notes = append(notes, "vcs is set to none, skipping the commit")
	case vcs.OutcomeUnavailable:
		notes = append(notes, "no version control here, skipping the commit")
	case vcs.OutcomeNothingToCommit:
		notes = append(notes,
			"nothing to commit, snapshot.ndjson and board.json are unchanged since the last snapshot")
	}
	if outcome.StagedOutsideBoard > 0 {
		notes = append(notes, fmt.Sprintf(
			"%d staged change(s) outside the board were left untouched",
			outcome.StagedOutsideBoard))
	}
	return notes
}

// writeSnapshotFiles writes the two files the way
// docs/spec/cmd/snapshot.md#cómo-se-escriben-y-qué-pasa-si-el-segundo-falla
// asks for: each one into a temporary file of the same directory, both
// temporaries complete before either is renamed over its predecessor.
//
// That order is what the specification can promise: if writing fails, the
// two previous files are intact, no revision is attempted, and the exit
// code is 8. Between the first rename and the second there is an instant in
// which the pair is not coherent, and saying so is better than promising an
// atomicity of two files that no file system gives.
func writeSnapshotFiles(dir string, tasks, config []byte) *model.Error {
	type pending struct{ temp, final string }
	var written []pending
	cleanup := func() {
		for _, p := range written {
			os.Remove(p.temp)
		}
	}

	for _, f := range []struct {
		name    string
		content []byte
	}{
		{board.SnapshotTasksFile, tasks},
		{board.SnapshotConfigFile, config},
	} {
		final := filepath.Join(dir, f.name)
		temp, err := os.CreateTemp(dir, f.name+".*")
		if err != nil {
			cleanup()
			return cannotWriteSnapshot(final, err)
		}
		if _, err := temp.Write(f.content); err != nil {
			temp.Close()
			os.Remove(temp.Name())
			cleanup()
			return cannotWriteSnapshot(final, err)
		}
		if err := temp.Close(); err != nil {
			os.Remove(temp.Name())
			cleanup()
			return cannotWriteSnapshot(final, err)
		}
		// A temporary file is born readable only by its owner, and these
		// two are ordinary files of the board meant to be read and
		// versioned like any other.
		if err := os.Chmod(temp.Name(), 0o644); err != nil {
			os.Remove(temp.Name())
			cleanup()
			return cannotWriteSnapshot(final, err)
		}
		written = append(written, pending{temp: temp.Name(), final: final})
	}

	for _, p := range written {
		if err := os.Rename(p.temp, p.final); err != nil {
			cleanup()
			return cannotWriteSnapshot(p.final, err)
		}
	}
	return nil
}

func cannotWriteSnapshot(path string, cause error) *model.Error {
	return &model.Error{
		ExitCode: 8,
		Code:     "io_error",
		Message:  fmt.Sprintf("%s cannot be written: %s", path, cause),
		Field:    path,
	}
}
