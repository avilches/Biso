// Package vcs runs the version control system this machine is configured for,
// which is what biso snapshot uses to record the two files it writes. It is
// the only package in biso that launches another program, and only
// internal/ops calls it, from the snapshot command and from nowhere else.
//
// The catalog of systems, the questions each one answers and the orders each
// one takes are specified in docs/spec/cmd/snapshot.md; the configuration keys
// that choose between them are in docs/spec/invocacion.md, section
// "Configuracion de maquina".
package vcs

import (
	"fmt"

	"biso/internal/model"
)

// Kind is the version control system the machine is configured for: the vcs
// key of docs/spec/invocacion.md.
type Kind string

const (
	// KindGit is the default and the only system biso knows in version 1.0.
	KindGit Kind = "git"
	// KindNone runs nothing at all.
	KindNone Kind = "none"
	// KindCustom runs the commands declared in the vcs_custom object, and asks
	// the repository nothing.
	KindCustom Kind = "custom"
)

// Mode is what the --vcs flag of biso snapshot asks for in this one call. It
// never changes the configured Kind.
type Mode string

const (
	// ModeNone writes the files and runs nothing, not even a question.
	ModeNone Mode = "none"
	// ModeCommit writes the files and records a revision. It is the default.
	ModeCommit Mode = "commit"
	// ModePush records a revision and publishes it.
	ModePush Mode = "push"
)

// modes is the closed domain of --vcs, in the order the help text lists it.
var modes = []string{"none", "commit", "push"}

// ParseMode turns the value of --vcs into a Mode, or returns the usage error
// of a value outside its closed domain.
func ParseMode(value string) (Mode, *model.Error) {
	for _, mode := range modes {
		if value == mode {
			return Mode(mode), nil
		}
	}
	return "", &model.Error{
		ExitCode: 2,
		Code:     "invalid_vcs_mode",
		Message:  fmt.Sprintf("--vcs must be none, commit or push, not %q", value),
		Field:    "--vcs",
		Given:    value,
		Valid:    append([]string(nil), modes...),
	}
}

// Custom is the vcs_custom object: the orders of a system biso does not know.
type Custom struct {
	// Commit is the order that records a revision. It is required.
	Commit []string
	// Publish is the order that publishes it. Without it, --vcs push is a
	// usage error.
	Publish []string
	// IgnoreFile is the name of the exclusion file biso init writes inside the
	// board directory. Empty means the system has none.
	IgnoreFile string
}

// Config is what this package needs from the machine configuration.
type Config struct {
	Kind   Kind
	Custom Custom
}

// IgnoreFile returns the name of the exclusion file that the configured system
// uses inside a board directory, or the empty string when it has none. It is
// what biso init writes and what biso doctor compares against.
func IgnoreFile(cfg Config) string {
	switch cfg.Kind {
	case KindGit:
		return ".gitignore"
	case KindCustom:
		return cfg.Custom.IgnoreFile
	default:
		return ""
	}
}

// Request is one snapshot's worth of work. The two files are already written
// when it arrives here: this package only records them.
type Request struct {
	// BoardDir is the board's own directory, absolute, and the working
	// directory of every question and every order.
	BoardDir string
	// Files are the three paths that go into the revision, relative to
	// BoardDir and in the order the specification fixes: snapshot.ndjson,
	// board.json and the <id>.id marker.
	Files []string
	// Message is the revision message biso composes, "biso snapshot: N tasks".
	Message string
}

// Outcome says which of the five endings of docs/spec/cmd/snapshot.md this
// call reached. Each one carries a different note, or none.
type Outcome string

const (
	// OutcomeNotRequested is --vcs none: nothing ran and there is no note.
	OutcomeNotRequested Outcome = "not_requested"
	// OutcomeDisabled is the vcs key set to none while --vcs asked for commit
	// or push: nothing ran, and the note says so.
	OutcomeDisabled Outcome = "disabled"
	// OutcomeUnavailable is the system not being installed, or creating the
	// repository having failed. Nothing was recorded and it is not an error.
	OutcomeUnavailable Outcome = "unavailable"
	// OutcomeNothingToCommit is the system reporting that the files have not
	// changed since the last snapshot.
	OutcomeNothingToCommit Outcome = "nothing_to_commit"
	// OutcomeCommitted is a revision recorded.
	OutcomeCommitted Outcome = "committed"
)

// Result is what happened, in the shape the JSON of biso snapshot needs.
type Result struct {
	// VCS is the system this call used: "git", "custom", or "none" when no
	// command was even attempted.
	VCS string
	// Outcome is the ending this call reached.
	Outcome Outcome
	// Commit is the full identifier of the revision, empty when none was
	// recorded and always empty with custom, which returns no identifier.
	Commit string
	// Repository is the root of the repository the revision went to, empty
	// when no revision was recorded and always empty with custom.
	Repository string
	// OwnRepository says whether Repository is the board's own directory,
	// which is what tells the board's own history from the project's.
	OwnRepository bool
	// Pushed is true only after a publication that worked.
	Pushed bool
	// Output are the lines the orders that act wrote, both streams, in the
	// order they ran and without the prefix they carry on a terminal.
	Output []string
	// StagedOutsideBoard counts the paths the index had staged outside the
	// three files of the board. It is always 0 with none and with custom.
	StagedOutsideBoard int
}

// Runner records snapshots with one configured system in one mode. Build it
// with New, which is where the only failure that has to happen before any file
// is written is detected.
type Runner struct {
	config Config
	mode   Mode
}

// New validates the pair of configuration and mode, and returns the runner
// that executes it. Its only error is asking to publish with a custom system
// that declares no publish order: it has to surface before biso snapshot
// writes anything, which is why it lives here and not in Run.
func New(cfg Config, mode Mode) (*Runner, *model.Error) {
	if mode == ModePush && cfg.Kind == KindCustom && len(cfg.Custom.Publish) == 0 {
		return nil, &model.Error{
			ExitCode: 2,
			Code:     "vcs_push_unavailable",
			Message:  "--vcs push needs a publish command, and the custom vcs of this machine declares none",
			Field:    "--vcs",
			Given:    "push",
		}
	}
	return &Runner{config: cfg, mode: mode}, nil
}

// Run executes the recipe of the configured system over an already written
// snapshot. The result is never nil: on an error it carries what had been
// achieved before the failure, and the error carries the lines of the order
// that failed.
func (r *Runner) Run(req Request) (*Result, *model.Error) {
	if r.mode == ModeNone {
		return &Result{VCS: string(KindNone), Outcome: OutcomeNotRequested}, nil
	}
	if r.config.Kind == KindNone {
		return &Result{VCS: string(KindNone), Outcome: OutcomeDisabled}, nil
	}
	if r.config.Kind == KindCustom {
		return r.runCustom(req)
	}
	return r.runGit(req)
}

// commitFailed is the error of a revision that was really attempted and did
// not work. The message says that nothing was lost because the two files of
// the snapshot are on disk before the first order runs.
func commitFailed(output []string) *model.Error {
	return &model.Error{
		ExitCode:  8,
		Code:      "vcs_commit_failed",
		Message:   "the commit failed, nothing was lost: snapshot.ndjson and board.json are written, run biso snapshot again once the reason is fixed",
		VCSOutput: output,
	}
}

// pushFailed is the error of a publication that did not work. The revision is
// already recorded when it happens.
func pushFailed(output []string) *model.Error {
	return &model.Error{
		ExitCode:  8,
		Code:      "vcs_push_failed",
		Message:   "the push failed, the revision is recorded: run biso snapshot --vcs push again once the reason is fixed",
		VCSOutput: output,
	}
}
